package parser

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"mime"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"book-reader/backend/internal/model"
)

type Result struct {
	Title       string
	ParseStatus string
	Chapters    []model.BookChapter
}

type Resource struct {
	Name        string
	ContentType string
	Body        io.ReadCloser
}

type Cover struct {
	Name        string
	ContentType string
	Data        []byte
}

var chapterPatterns = []*regexp.Regexp{
	regexp.MustCompile(`^[ 　\t]{0,4}(序章|楔子|正文[ 　\t]+.{1,30}|终章|后记|尾声|番外|第[ \t　]*[0-9〇零一二两三四五六七八九十百千万壹贰叁肆伍陆柒捌玖拾佰仟]+[ \t　]*(章|节|卷|集|部|篇)).{0,40}$`),
	regexp.MustCompile(`^[ 　\t]{0,4}[0-9]{1,5}[:：,.， 、_\-].{1,40}$`),
	regexp.MustCompile(`^[ 　\t]{0,4}(Chapter|chapter|Section|section|Part|part)[ \t]*[0-9]{1,4}.{0,40}$`),
}

func Parse(filePath, format, fallbackTitle string, bookID int64) (*Result, error) {
	title := strings.TrimSuffix(filepath.Base(fallbackTitle), filepath.Ext(fallbackTitle))
	if strings.TrimSpace(title) == "" {
		title = "未命名图书"
	}
	switch format {
	case model.BookFormatTXT:
		return parseTXT(filePath, title, bookID)
	case model.BookFormatPDF:
		return parsePDF(filePath, title, bookID)
	case model.BookFormatEPUB:
		return parseEPUB(filePath, title, bookID)
	default:
		return nil, fmt.Errorf("unsupported format")
	}
}

func ExtractEPUBCover(filePath string) (*Cover, error) {
	zr, err := zip.OpenReader(filePath)
	if err != nil {
		return nil, err
	}
	defer zr.Close()
	opfPath, err := epubOPFPath(&zr.Reader)
	if err != nil {
		return nil, err
	}
	opfData, err := readZipFile(&zr.Reader, opfPath)
	if err != nil {
		return nil, err
	}
	var pkg struct {
		Manifest []struct {
			ID        string `xml:"id,attr"`
			Href      string `xml:"href,attr"`
			MediaType string `xml:"media-type,attr"`
		} `xml:"manifest>item"`
		Metadata struct {
			Metas []struct {
				Name    string `xml:"name,attr"`
				Content string `xml:"content,attr"`
			} `xml:"meta"`
		} `xml:"metadata"`
	}
	if err := xml.Unmarshal(opfData, &pkg); err != nil {
		return nil, err
	}
	var coverHref string
	for _, m := range pkg.Metadata.Metas {
		if strings.EqualFold(strings.TrimSpace(m.Name), "cover") && strings.TrimSpace(m.Content) != "" {
			for _, item := range pkg.Manifest {
				if item.ID == strings.TrimSpace(m.Content) {
					coverHref = item.Href
					break
				}
			}
		}
		if coverHref != "" {
			break
		}
	}
	if coverHref == "" {
		for _, item := range pkg.Manifest {
			if strings.HasPrefix(strings.ToLower(item.MediaType), "image/") && strings.Contains(strings.ToLower(item.ID), "cover") {
				coverHref = item.Href
				break
			}
		}
	}
	if coverHref == "" {
		return nil, fmt.Errorf("epub cover not found")
	}
	base := path.Dir(opfPath)
	resourcePath := cleanZipPath(path.Join(base, coverHref))
	data, err := readZipFile(&zr.Reader, resourcePath)
	if err != nil {
		return nil, err
	}
	ct := mime.TypeByExtension(strings.ToLower(path.Ext(resourcePath)))
	if ct == "" {
		ct = "application/octet-stream"
	}
	return &Cover{Name: path.Base(resourcePath), ContentType: ct, Data: data}, nil
}

func parsePDF(filePath, title string, bookID int64) (*Result, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	pageCount := estimatePDFPages(data)
	if pageCount <= 0 {
		pageCount = 1
	}
	var chapters []model.BookChapter
	for startPage, idx := 0, 0; startPage < pageCount; startPage, idx = startPage+10, idx+1 {
		endPage := startPage + 10
		if endPage > pageCount {
			endPage = pageCount
		}
		start, end := int64(startPage), int64(endPage)
		chapters = append(chapters, model.BookChapter{
			BookID: bookID, ChapterIndex: idx, Title: fmt.Sprintf("分段_%d", idx+1),
			Locator: fmt.Sprintf("pdf_%d", idx), StartOffset: &start, EndOffset: &end,
		})
	}
	return &Result{Title: title, ParseStatus: "partial", Chapters: chapters}, nil
}

func estimatePDFPages(data []byte) int {
	re := regexp.MustCompile(`/Type\s*/Page\b`)
	return len(re.FindAll(data, -1))
}

type containerXML struct {
	Rootfiles []struct {
		FullPath string `xml:"full-path,attr"`
	} `xml:"rootfiles>rootfile"`
}

type opfPackage struct {
	Metadata struct {
		Titles []string `xml:"title"`
	} `xml:"metadata"`
	Manifest []struct {
		ID        string `xml:"id,attr"`
		Href      string `xml:"href,attr"`
		MediaType string `xml:"media-type,attr"`
	} `xml:"manifest>item"`
	Spine []struct {
		IDRef string `xml:"idref,attr"`
	} `xml:"spine>itemref"`
}

func parseEPUB(filePath, fallbackTitle string, bookID int64) (*Result, error) {
	zr, err := zip.OpenReader(filePath)
	if err != nil {
		return nil, err
	}
	defer zr.Close()
	opfPath, err := epubOPFPath(&zr.Reader)
	if err != nil {
		return nil, err
	}
	opfData, err := readZipFile(&zr.Reader, opfPath)
	if err != nil {
		return nil, err
	}
	var pkg opfPackage
	if err := xml.Unmarshal(opfData, &pkg); err != nil {
		return nil, err
	}
	title := fallbackTitle
	for _, t := range pkg.Metadata.Titles {
		if strings.TrimSpace(t) != "" {
			title = strings.TrimSpace(t)
			break
		}
	}
	base := path.Dir(opfPath)
	manifest := map[string]struct {
		Href      string
		MediaType string
	}{}
	for _, item := range pkg.Manifest {
		manifest[item.ID] = struct {
			Href      string
			MediaType string
		}{Href: item.Href, MediaType: item.MediaType}
	}
	var chapters []model.BookChapter
	for _, itemref := range pkg.Spine {
		item, ok := manifest[itemref.IDRef]
		if !ok || !isEPUBDocument(item.MediaType, item.Href) {
			continue
		}
		href := cleanZipPath(path.Join(base, item.Href))
		doc, err := readZipFile(&zr.Reader, href)
		if err != nil {
			continue
		}
		chapterTitle := htmlTitle(doc)
		if chapterTitle == "" {
			chapterTitle = fmt.Sprintf("章节 %d", len(chapters)+1)
		}
		wordCount := int64(len([]rune(stripHTML(string(doc)))))
		hrefCopy := href
		chapters = append(chapters, model.BookChapter{
			BookID: bookID, ChapterIndex: len(chapters), Title: chapterTitle,
			Locator: href, Href: &hrefCopy, WordCount: &wordCount,
		})
	}
	if len(chapters) == 0 {
		href := opfPath
		chapters = append(chapters, model.BookChapter{
			BookID: bookID, ChapterIndex: 0, Title: "正文", Locator: "epub_0", Href: &href,
		})
		return &Result{Title: title, ParseStatus: "partial", Chapters: chapters}, nil
	}
	return &Result{Title: title, ParseStatus: "parsed", Chapters: chapters}, nil
}

func epubOPFPath(zr *zip.Reader) (string, error) {
	data, err := readZipFile(zr, "META-INF/container.xml")
	if err != nil {
		return "", err
	}
	var c containerXML
	if err := xml.Unmarshal(data, &c); err != nil {
		return "", err
	}
	for _, rf := range c.Rootfiles {
		if rf.FullPath != "" {
			return cleanZipPath(rf.FullPath), nil
		}
	}
	return "", fmt.Errorf("epub opf not found")
}

func isEPUBDocument(mediaType, href string) bool {
	mt := strings.ToLower(mediaType)
	ext := strings.ToLower(path.Ext(href))
	return mt == "application/xhtml+xml" || mt == "text/html" || ext == ".xhtml" || ext == ".html" || ext == ".htm"
}

func ReadEPUBContent(filePath string, chapter model.BookChapter, bookID int64) (string, error) {
	if chapter.Href == nil || strings.TrimSpace(*chapter.Href) == "" {
		return "", nil
	}
	zr, err := zip.OpenReader(filePath)
	if err != nil {
		return "", err
	}
	defer zr.Close()
	data, err := readZipFile(&zr.Reader, cleanZipPath(*chapter.Href))
	if err != nil {
		return "", err
	}
	htmlText := string(data)
	htmlText = removeTagContent(htmlText, "script")
	htmlText = removeTagContent(htmlText, "style")
	htmlText = rewriteEPUBImageSources(htmlText, cleanZipPath(*chapter.Href), bookID)
	return htmlText, nil
}

func ReadEPUBResource(filePath, href string) (*Resource, error) {
	cleanHref := cleanZipPath(href)
	if cleanHref == "" {
		return nil, fmt.Errorf("empty resource href")
	}
	zr, err := zip.OpenReader(filePath)
	if err != nil {
		return nil, err
	}
	data, err := readZipFile(&zr.Reader, cleanHref)
	_ = zr.Close()
	if err != nil {
		return nil, err
	}
	ct := mime.TypeByExtension(strings.ToLower(path.Ext(cleanHref)))
	if ct == "" {
		ct = "application/octet-stream"
	}
	return &Resource{Name: path.Base(cleanHref), ContentType: ct, Body: io.NopCloser(bytes.NewReader(data))}, nil
}

func rewriteEPUBImageSources(htmlText, docHref string, bookID int64) string {
	re := regexp.MustCompile(`(?i)(<img\b[^>]*\bsrc\s*=\s*["'])([^"']+)(["'])`)
	base := path.Dir(docHref)
	return re.ReplaceAllStringFunc(htmlText, func(match string) string {
		parts := re.FindStringSubmatch(match)
		if len(parts) != 4 {
			return match
		}
		src := parts[2]
		if strings.HasPrefix(src, "http://") || strings.HasPrefix(src, "https://") || strings.HasPrefix(src, "data:") {
			return match
		}
		resource := cleanZipPath(path.Join(base, src))
		encoded := url.QueryEscape(resource)
		return parts[1] + fmt.Sprintf("/api/v1/reader/books/%d/resources?href=%s", bookID, encoded) + parts[3]
	})
}

func cleanZipPath(v string) string {
	v = strings.ReplaceAll(v, "\\", "/")
	cleaned := path.Clean("/" + v)
	cleaned = strings.TrimPrefix(cleaned, "/")
	if cleaned == "." || strings.HasPrefix(cleaned, "../") {
		return ""
	}
	return cleaned
}

func readZipFile(zr *zip.Reader, name string) ([]byte, error) {
	name = cleanZipPath(name)
	for _, f := range zr.File {
		if cleanZipPath(f.Name) != name {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return nil, err
		}
		defer rc.Close()
		return io.ReadAll(rc)
	}
	return nil, fmt.Errorf("zip resource not found: %s", name)
}

func htmlTitle(data []byte) string {
	re := regexp.MustCompile(`(?is)<title[^>]*>(.*?)</title>`)
	m := re.FindSubmatch(data)
	if len(m) < 2 {
		return ""
	}
	return strings.TrimSpace(html.UnescapeString(stripHTML(string(m[1]))))
}

func stripHTML(v string) string {
	re := regexp.MustCompile(`(?s)<[^>]+>`)
	return strings.TrimSpace(html.UnescapeString(re.ReplaceAllString(v, "")))
}

func removeTagContent(v, tag string) string {
	re := regexp.MustCompile(`(?is)<` + regexp.QuoteMeta(tag) + `\b[^>]*>.*?</` + regexp.QuoteMeta(tag) + `>`)
	return re.ReplaceAllString(v, "")
}

func parseTXT(filePath, title string, bookID int64) (*Result, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	if bytes.HasPrefix(data, []byte{0xEF, 0xBB, 0xBF}) {
		data = data[3:]
	}
	lines := bytes.SplitAfter(data, []byte{'\n'})
	type hit struct {
		title     string
		lineStart int64
		bodyStart int64
	}
	var hits []hit
	var offset int64
	for _, line := range lines {
		text := strings.TrimSpace(string(bytes.TrimRight(line, "\r\n")))
		if isChapterTitle(text) {
			hits = append(hits, hit{title: text, lineStart: offset, bodyStart: offset + int64(len(line))})
		}
		offset += int64(len(line))
	}
	var chapters []model.BookChapter
	if len(hits) == 0 {
		start, end := int64(0), int64(len(data))
		wc := int64(len([]rune(string(data))))
		chapters = append(chapters, model.BookChapter{
			BookID: bookID, ChapterIndex: 0, Title: "正文", Locator: "txt_0",
			StartOffset: &start, EndOffset: &end, WordCount: &wc,
		})
	} else {
		if hits[0].lineStart > 0 {
			start, end := int64(0), hits[0].lineStart
			wc := wordCount(data[start:end])
			chapters = append(chapters, model.BookChapter{
				BookID: bookID, ChapterIndex: 0, Title: "前言", Locator: "txt_0",
				StartOffset: &start, EndOffset: &end, WordCount: wc,
			})
		}
		for i, h := range hits {
			start := h.bodyStart
			end := int64(len(data))
			if i+1 < len(hits) {
				end = hits[i+1].lineStart
			}
			wc := wordCount(data[start:end])
			chapters = append(chapters, model.BookChapter{
				BookID: bookID, ChapterIndex: len(chapters), Title: h.title,
				Locator: fmt.Sprintf("txt_%d", len(chapters)), StartOffset: &start,
				EndOffset: &end, WordCount: wc, IsVolume: strings.TrimSpace(string(data[start:end])) == "",
			})
		}
	}
	return &Result{Title: title, ParseStatus: "parsed", Chapters: chapters}, nil
}

func isChapterTitle(s string) bool {
	if s == "" || len([]rune(s)) > 60 {
		return false
	}
	for _, re := range chapterPatterns {
		if re.MatchString(s) {
			return true
		}
	}
	return false
}

func wordCount(b []byte) *int64 {
	v := int64(len([]rune(strings.TrimSpace(string(b)))))
	return &v
}

func ReadTXTContent(filePath string, chapter model.BookChapter) (string, error) {
	if chapter.StartOffset == nil || chapter.EndOffset == nil || *chapter.EndOffset < *chapter.StartOffset {
		return "", nil
	}
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()
	size := *chapter.EndOffset - *chapter.StartOffset
	buf := make([]byte, size)
	if _, err := f.ReadAt(buf, *chapter.StartOffset); err != nil && size > 0 {
		return "", err
	}
	return strings.TrimLeft(string(buf), " \r\n\t　"), nil
}

func PDFChapterHTML(chapter model.BookChapter) string {
	if chapter.StartOffset == nil || chapter.EndOffset == nil {
		return `<p>PDF 请通过文件流接口交给前端 pdf.js 渲染。</p>`
	}
	var pages []int
	for i := *chapter.StartOffset; i < *chapter.EndOffset; i++ {
		pages = append(pages, int(i))
	}
	sort.Ints(pages)
	var b strings.Builder
	for _, page := range pages {
		b.WriteString(fmt.Sprintf(`<div data-pdf-page="%d"></div>`, page+1))
	}
	return b.String()
}

func EscapeTextAsHTML(s string) string {
	parts := strings.Split(html.EscapeString(s), "\n")
	for i := range parts {
		parts[i] = strings.TrimRight(parts[i], "\r")
	}
	return strings.Join(parts, "<br>")
}
