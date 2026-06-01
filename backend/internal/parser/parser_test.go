package parser

import (
	"archive/zip"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"book-reader/backend/internal/model"
)

func TestParseTXTChapters(t *testing.T) {
	file := filepath.Join(t.TempDir(), "book.txt")
	if err := os.WriteFile(file, []byte("第一章 开始\n正文一\n\n第二章 继续\n正文二\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	result, err := Parse(file, model.BookFormatTXT, "book.txt", 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Chapters) != 2 {
		t.Fatalf("expected 2 chapters, got %d", len(result.Chapters))
	}
	content, err := ReadTXTContent(file, result.Chapters[0])
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(content, "正文一") {
		t.Fatalf("unexpected content: %q", content)
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
	addZipFile(t, zw, "OEBPS/chap2.xhtml", `<html><head><title>Two</title></head><body><p>World</p></body></html>`)
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
