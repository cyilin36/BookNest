package parser

import (
	"archive/zip"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"booknest/backend/internal/model"
)

func TestParseTXTChapters(t *testing.T) {
	file := filepath.Join(t.TempDir(), "book.txt")
	body := "\xEF\xBB\xBF第一章 开始\n    正文一\n\n第二段正文\n第二章 继续\n正文二\n"
	if err := os.WriteFile(file, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	result, err := Parse(file, model.BookFormatTXT, "book.txt", 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Chapters) != 2 {
		t.Fatalf("expected 2 chapters, got %d", len(result.Chapters))
	}
	wantStart := int64(len([]byte("\xEF\xBB\xBF第一章 开始\n")))
	if result.Chapters[0].StartOffset == nil || *result.Chapters[0].StartOffset != wantStart {
		t.Fatalf("unexpected first chapter start offset: %#v, want %d", result.Chapters[0].StartOffset, wantStart)
	}
	content, err := ReadTXTContent(file, result.Chapters[0])
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(content, "正文一") {
		t.Fatalf("unexpected content: %q", content)
	}
	if !strings.HasPrefix(content, "    正文一") {
		t.Fatalf("txt content did not preserve existing leading indent: %q", content)
	}
	if strings.Contains(content, "\uFFFD") {
		t.Fatalf("content contains replacement character: %q", content)
	}
}

func TestReadTXTContentLegacyBOMOffsets(t *testing.T) {
	file := filepath.Join(t.TempDir(), "legacy-bom.txt")
	body := "\xEF\xBB\xBF第一章 开始\n正文一\n第二段正文"
	if err := os.WriteFile(file, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	start, end := int64(0), int64(len([]byte(body))-3)
	chapter := model.BookChapter{StartOffset: &start, EndOffset: &end}
	content, err := ReadTXTContent(file, chapter)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(content, "\uFEFF") || strings.Contains(content, "\uFFFD") {
		t.Fatalf("legacy BOM offsets were not corrected: %q", content)
	}
	if !strings.Contains(content, "第二段正文") {
		t.Fatalf("legacy BOM offsets truncated content: %q", content)
	}

	titleBytes := len([]byte("第一章 开始\n"))
	legacyStart := int64(titleBytes)
	legacyEnd := int64(len([]byte(body)) - 3)
	chapter = model.BookChapter{StartOffset: &legacyStart, EndOffset: &legacyEnd}
	content, err = ReadTXTContent(file, chapter)
	if err != nil {
		t.Fatal(err)
	}
	if strings.HasPrefix(content, "始") || strings.Contains(content, "\uFFFD") {
		t.Fatalf("legacy chapter boundary was not corrected: %q", content)
	}
	if !strings.HasPrefix(content, "正文一") || !strings.Contains(content, "第二段正文") {
		t.Fatalf("legacy chapter content was not preserved: %q", content)
	}
}

func TestNormalizeTXTContentLeadingWhitespace(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "preserve existing half width indent", in: "\n    正文", want: "    正文"},
		{name: "preserve existing full width indent", in: "\r\n　　正文", want: "　　正文"},
		{name: "add indent after blank lines when missing", in: "\n\n正文", want: "　　正文"},
		{name: "leave already started text unchanged", in: "正文", want: "正文"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeTXTContent(tt.in); got != tt.want {
				t.Fatalf("normalizeTXTContent(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestParseEPUBAndResource(t *testing.T) {
	file := filepath.Join(t.TempDir(), "book.epub")
	createTestEPUB(t, file)
	result, err := Parse(file, model.BookFormatEPUB, "fallback.epub", 9)
	if err != nil {
		t.Fatal(err)
	}
	if result.Title != "EPUB Title" {
		t.Fatalf("unexpected title: %s", result.Title)
	}
	if len(result.Chapters) != 2 {
		t.Fatalf("expected 2 chapters, got %d", len(result.Chapters))
	}
	content, err := ReadEPUBContent(file, result.Chapters[0], 9)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(content, "/api/v1/reader/books/9/resources?href=OEBPS%2Fimages%2Fcover.png") {
		t.Fatalf("image src was not rewritten: %s", content)
	}
	svgContent, err := ReadEPUBContent(file, result.Chapters[1], 9)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(svgContent, "<image") || strings.Contains(svgContent, "xlink:href") {
		t.Fatalf("svg image was not normalized: %s", svgContent)
	}
	if !strings.Contains(svgContent, `<img src="/api/v1/reader/books/9/resources?href=OEBPS%2Fimages%2Fcover.png"`) {
		t.Fatalf("svg image src was not rewritten: %s", svgContent)
	}
	resource, err := ReadEPUBResource(file, "OEBPS/images/cover.png")
	if err != nil {
		t.Fatal(err)
	}
	defer resource.Body.Close()
	if resource.ContentType != "image/png" {
		t.Fatalf("unexpected content type: %s", resource.ContentType)
	}
	cover, err := ExtractEPUBCover(file)
	if err != nil {
		t.Fatal(err)
	}
	if cover.Name != "cover.png" || cover.ContentType != "image/png" || len(cover.Data) == 0 {
		t.Fatalf("unexpected cover: %#v", cover)
	}
}

func TestParsePDFEstimatePages(t *testing.T) {
	file := filepath.Join(t.TempDir(), "book.pdf")
	data := []byte("%PDF-1.4\n1 0 obj << /Type /Page >> endobj\n2 0 obj << /Type /Page >> endobj\n")
	if err := os.WriteFile(file, data, 0o644); err != nil {
		t.Fatal(err)
	}
	result, err := Parse(file, model.BookFormatPDF, "book.pdf", 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Chapters) != 1 || result.Chapters[0].EndOffset == nil || *result.Chapters[0].EndOffset != 2 {
		t.Fatalf("unexpected pdf chapters: %#v", result.Chapters)
	}
}

func createTestEPUB(t *testing.T, file string) {
	t.Helper()
	out, err := os.Create(file)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(out)
	addZipFile(t, zw, "META-INF/container.xml", `<?xml version="1.0"?>
<container version="1.0"><rootfiles><rootfile full-path="OEBPS/content.opf" media-type="application/oebps-package+xml"/></rootfiles></container>`)
	addZipFile(t, zw, "OEBPS/content.opf", `<?xml version="1.0"?>
<package xmlns:dc="http://purl.org/dc/elements/1.1/">
  <metadata><dc:title>EPUB Title</dc:title></metadata>
  <manifest>
    <item id="c1" href="chap1.xhtml" media-type="application/xhtml+xml"/>
    <item id="c2" href="chap2.xhtml" media-type="application/xhtml+xml"/>
    <item id="cover-image" href="images/cover.png" media-type="image/png"/>
  </manifest>
  <spine><itemref idref="c1"/><itemref idref="c2"/></spine>
</package>`)
	addZipFile(t, zw, "OEBPS/chap1.xhtml", `<html><head><title>One</title></head><body><p>Hello</p><img src="images/cover.png"/></body></html>`)
	addZipFile(t, zw, "OEBPS/chap2.xhtml", `<html><head><title>Two</title></head><body><svg xmlns="http://www.w3.org/2000/svg"><image width="600" height="800" xlink:href="images/cover.png"></image></svg></body></html>`)
	addZipFile(t, zw, "OEBPS/images/cover.png", "\x89PNG\r\n\x1a\n")
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := out.Close(); err != nil {
		t.Fatal(err)
	}
}

func addZipFile(t *testing.T, zw *zip.Writer, name, body string) {
	t.Helper()
	w, err := zw.Create(name)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write([]byte(body)); err != nil {
		t.Fatal(err)
	}
}
