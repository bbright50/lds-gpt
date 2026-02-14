package scraper

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var (
	chapterNumberRe = regexp.MustCompile(`\d+`)

	httpClient = &http.Client{
		Timeout: 30 * time.Second,
	}
)

// ScrapeChapter fetches a scripture chapter URL and extracts verses and footnotes.
// If cacheDir is non-empty, raw HTML is read from / written to that directory.
func ScrapeChapter(ctx context.Context, rawURL, cacheDir string) (Chapter, error) {
	if strings.TrimSpace(rawURL) == "" {
		return Chapter{}, fmt.Errorf("url cannot be empty")
	}

	doc, err := fetchDocument(ctx, rawURL, cacheDir)
	if err != nil {
		return Chapter{}, fmt.Errorf("fetch document: %w", err)
	}

	book := extractBookName(doc)
	chapter := extractChapterNumber(doc)
	summary := extractSummary(doc)
	verses := extractVerses(doc)
	refTexts := extractReferenceTexts(doc)
	footnotes := extractFootnotes(doc, refTexts)

	return Chapter{
		URL:       rawURL,
		Book:      book,
		Chapter:   chapter,
		Summary:   summary,
		Verses:    verses,
		Footnotes: footnotes,
	}, nil
}

func fetchDocument(ctx context.Context, rawURL, cacheDir string) (*goquery.Document, error) {
	cachePath, _ := cachePath(rawURL, cacheDir)

	if cachePath != "" {
		if f, err := os.Open(cachePath); err == nil {
			defer f.Close()
			doc, err := goquery.NewDocumentFromReader(f)
			if err != nil {
				return nil, fmt.Errorf("parse cached html: %w", err)
			}
			return doc, nil
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http get: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	if cachePath != "" {
		if err := writeCache(cachePath, body); err != nil {
			return nil, fmt.Errorf("write cache: %w", err)
		}
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("parse html: %w", err)
	}

	return doc, nil
}

func cachePath(rawURL, cacheDir string) (string, error) {
	if cacheDir == "" {
		return "", nil
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("parse url: %w", err)
	}

	const prefix = "/study/scriptures/"
	if !strings.HasPrefix(parsed.Path, prefix) {
		return "", fmt.Errorf("unexpected url path: %s", parsed.Path)
	}

	rel := strings.TrimPrefix(parsed.Path, prefix)
	return filepath.Join(cacheDir, rel+".html"), nil
}

func writeCache(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}

func extractBookName(doc *goquery.Document) string {
	dominant := doc.Find("h1 span.dominant").First()
	if dominant.Length() > 0 {
		return strings.TrimSpace(dominant.Text())
	}

	title := doc.Find("h1#title1").First()
	if title.Length() > 0 {
		return strings.TrimSpace(title.Text())
	}

	return ""
}

func extractChapterNumber(doc *goquery.Document) int {
	text := strings.TrimSpace(doc.Find("p.title-number").First().Text())
	match := chapterNumberRe.FindString(text)
	if match == "" {
		return 0
	}

	num, err := strconv.Atoi(match)
	if err != nil {
		return 0
	}

	return num
}

func extractSummary(doc *goquery.Document) string {
	return strings.TrimSpace(doc.Find("p.study-summary").First().Text())
}

func extractVerses(doc *goquery.Document) []Verse {
	var verses []Verse

	doc.Find("p.verse").Each(func(_ int, sel *goquery.Selection) {
		verse := parseVerse(sel)
		verses = append(verses, verse)
	})

	return verses
}

func parseVerse(sel *goquery.Selection) Verse {
	numText := strings.TrimSpace(sel.Find("span.verse-number").First().Text())
	num, err := strconv.Atoi(numText)
	if err != nil {
		num = 0
	}

	var markers []string
	sel.Find("a.study-note-ref sup.marker").Each(func(_ int, sup *goquery.Selection) {
		if val, exists := sup.Attr("data-value"); exists {
			markers = append(markers, val)
		}
	})

	text := buildVerseText(sel)

	return Verse{
		Number:          num,
		Text:            text,
		FootnoteMarkers: markers,
	}
}

func buildVerseText(sel *goquery.Selection) string {
	clone := sel.Clone()

	// Remove verse number span
	clone.Find("span.verse-number").Remove()

	// Remove footnote marker superscripts
	clone.Find("sup.marker").Remove()

	// Remove icon/button elements
	clone.Find("span.iconPointer-OKie_").Remove()

	return normalizeWhitespace(clone.Text())
}

func extractReferenceTexts(doc *goquery.Document) map[string]string {
	refs := make(map[string]string)

	doc.Find("p.verse a.study-note-ref").Each(func(_ int, a *goquery.Selection) {
		href, exists := a.Attr("href")
		if !exists {
			return
		}

		_, noteID, found := strings.Cut(href, "#")
		if !found {
			return
		}

		clone := a.Clone()
		clone.Find("sup.marker").Remove()
		text := normalizeWhitespace(clone.Text())

		if text != "" {
			refs[noteID] = text
		}
	})

	return refs
}

func extractFootnotes(doc *goquery.Document, refTexts map[string]string) map[string]Footnote {
	footnotes := make(map[string]Footnote)

	doc.Find("footer.study-notes li[data-full-marker]").Each(func(_ int, li *goquery.Selection) {
		fullMarker, exists := li.Attr("data-full-marker")
		if !exists {
			return
		}

		footnote := parseFootnote(li)

		if id, exists := li.Attr("id"); exists {
			footnote.ReferenceText = refTexts[id]
		}

		footnotes[fullMarker] = footnote
	})

	return footnotes
}

func parseFootnote(li *goquery.Selection) Footnote {
	categories := collectCategories(li)
	text := normalizeWhitespace(li.Find("p").First().Text())

	return Footnote{
		Category: strings.Join(categories, ","),
		Text:     text,
	}
}

func collectCategories(li *goquery.Selection) []string {
	seen := make(map[string]bool)
	var categories []string

	li.Find("span[data-note-category]").Each(func(_ int, span *goquery.Selection) {
		cat, exists := span.Attr("data-note-category")
		if !exists || seen[cat] {
			return
		}
		seen[cat] = true
		categories = append(categories, cat)
	})

	return categories
}

func normalizeWhitespace(s string) string {
	fields := strings.Fields(s)
	return strings.Join(fields, " ")
}
