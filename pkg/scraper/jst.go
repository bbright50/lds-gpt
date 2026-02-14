package scraper

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// JSTChapter represents a single JST page containing one or more retranslation entries.
type JSTChapter struct {
	Reference string     `json:"reference"`
	Book      string     `json:"book"`
	Chapter   string     `json:"chapter"`
	Entries   []JSTEntry `json:"entries"`
}

// JSTEntry represents a single retranslation section within a JST chapter.
type JSTEntry struct {
	Comprises string     `json:"comprises"`
	Compare   string     `json:"compare"`
	Summary   string     `json:"summary"`
	Verses    []JSTVerse `json:"verses"`
}

// JSTVerse represents a single verse in a JST entry.
type JSTVerse struct {
	Number int    `json:"number"`
	Text   string `json:"text"`
}

// ScrapeJSTIndex fetches the JST index page and returns all unique entry URLs,
// excluding _contents and introduction pages.
func ScrapeJSTIndex(ctx context.Context, indexURL, cacheDir string) ([]string, error) {
	doc, _, err := fetchDocument(ctx, indexURL, cacheDir)
	if err != nil {
		return nil, fmt.Errorf("fetch index: %w", err)
	}

	seen := make(map[string]bool)
	var urls []string

	doc.Find("a[href]").Each(func(_ int, a *goquery.Selection) {
		href, exists := a.Attr("href")
		if !exists {
			return
		}
		// Only include actual JST entry pages (e.g., /jst/jst-gen/15).
		// Exclude book-level pages (/jst/jst-gen), _contents, and non-JST links.
		if !strings.Contains(href, "/study/scriptures/jst/jst-") {
			return
		}
		if strings.Contains(href, "/_contents") {
			return
		}
		if !isJSTChapterURL(href) {
			return
		}
		if seen[href] {
			return
		}
		seen[href] = true
		urls = append(urls, "https://www.churchofjesuschrist.org"+href)
	})

	return urls, nil
}

// ScrapeJSTPage fetches a single JST page and returns the chapter data
// and whether the result was served from cache.
func ScrapeJSTPage(ctx context.Context, pageURL, cacheDir string) (JSTChapter, bool, error) {
	doc, cached, err := fetchDocument(ctx, pageURL, cacheDir)
	if err != nil {
		return JSTChapter{}, false, fmt.Errorf("fetch page: %w", err)
	}

	chapter := jstChapterFromURL(pageURL)
	book := jstBookFromDoc(doc)
	entries := extractJSTEntries(doc, chapter)

	return JSTChapter{
		Reference: "JST",
		Book:      book,
		Chapter:   chapter,
		Entries:   entries,
	}, cached, nil
}

// isJSTChapterURL checks if a URL path points to an actual JST chapter page
// (e.g., /jst/jst-gen/15) rather than a book-level page (e.g., /jst/jst-gen).
func isJSTChapterURL(href string) bool {
	path := href
	if idx := strings.Index(path, "?"); idx >= 0 {
		path = path[:idx]
	}
	// A chapter URL has at least 6 segments: /study/scriptures/jst/jst-gen/15
	// A book URL has only 5: /study/scriptures/jst/jst-gen
	parts := strings.Split(strings.Trim(path, "/"), "/")
	return len(parts) >= 5
}

// jstChapterFromURL extracts the chapter portion from a JST page URL.
// For ".../jst-gen/15?lang=eng" it returns "15".
// For ".../jst-gen/1-8?lang=eng" it returns "1-8".
func jstChapterFromURL(rawURL string) string {
	path := rawURL
	if idx := strings.Index(path, "?"); idx >= 0 {
		path = path[:idx]
	}
	if idx := strings.LastIndex(path, "/"); idx >= 0 {
		return path[idx+1:]
	}
	return ""
}

// jstBookFromDoc extracts the book name from the first section header.
func jstBookFromDoc(doc *goquery.Document) string {
	strong := doc.Find("article section header h2 strong").First()
	if strong.Length() == 0 {
		return ""
	}
	book, _ := parseJSTHeaderRef(normalizeWhitespace(strong.Text()))
	return book
}

// parseJSTHeaderRef parses a strong header text like "JST, Genesis 15:9–12."
// returning the book name ("Genesis") and the chapter:verse reference ("15:9–12").
func parseJSTHeaderRef(text string) (book string, ref string) {
	text = strings.TrimSpace(text)

	// Strip "JST, " prefix
	if after, found := strings.CutPrefix(text, "JST, "); found {
		text = after
	}

	// Strip trailing period
	text = strings.TrimRight(text, ". ")

	// The last word containing a colon is the chapter:verse reference.
	// Everything before it is the book name.
	words := strings.Fields(text)
	refIdx := -1
	for i := len(words) - 1; i >= 0; i-- {
		if strings.Contains(words[i], ":") {
			refIdx = i
			break
		}
	}

	if refIdx < 0 {
		return text, ""
	}

	book = strings.Join(words[:refIdx], " ")
	ref = strings.Join(words[refIdx:], " ")
	return book, ref
}

// jstComprises converts a chapter:verse reference into a comprises string.
// For a single-chapter ref like "15:9–12" with pageChapter="15", returns "9-12".
// For a multi-chapter ref like "1:1–8:18" with pageChapter="1-8", returns "1:1-8:18".
func jstComprises(ref, pageChapter string) string {
	if ref == "" {
		return ""
	}

	chapterPart, versePart, hasColon := strings.Cut(ref, ":")
	if !hasColon {
		return jstNormalizeHyphens(ref)
	}

	if chapterPart == pageChapter {
		return jstNormalizeHyphens(versePart)
	}

	return jstNormalizeHyphens(ref)
}

// jstNormalizeHyphens replaces en-dashes and em-dashes with regular hyphens.
func jstNormalizeHyphens(s string) string {
	s = strings.ReplaceAll(s, "\u2013", "-")
	s = strings.ReplaceAll(s, "\u2014", "-")
	return s
}

// extractJSTEntries collects all JST retranslation entries from the document.
func extractJSTEntries(doc *goquery.Document, pageChapter string) []JSTEntry {
	var entries []JSTEntry

	doc.Find("article section").Each(func(_ int, section *goquery.Selection) {
		entry := parseJSTSection(section, pageChapter)
		if len(entry.Verses) > 0 || entry.Summary != "" {
			entries = append(entries, entry)
		}
	})

	return entries
}

// parseJSTSection extracts a single JST entry from an article section element.
func parseJSTSection(section *goquery.Selection, pageChapter string) JSTEntry {
	strong := section.Find("header h2 strong").First()
	_, ref := parseJSTHeaderRef(normalizeWhitespace(strong.Text()))

	compareLink := section.Find("header h2 a.scripture-ref").First()
	compare := normalizeWhitespace(compareLink.Text())

	summary := normalizeWhitespace(section.Find("header p.study-intro").First().Text())

	verses := make([]JSTVerse, 0)
	section.Find("p.verse").Each(func(_ int, p *goquery.Selection) {
		verses = append(verses, parseJSTVerse(p))
	})

	return JSTEntry{
		Comprises: jstComprises(ref, pageChapter),
		Compare:   compare,
		Summary:   summary,
		Verses:    verses,
	}
}

// parseJSTVerse extracts a verse number and cleaned text from a p.verse element.
func parseJSTVerse(sel *goquery.Selection) JSTVerse {
	numText := strings.TrimSpace(sel.Find("span.verse-number").First().Text())
	num, err := strconv.Atoi(numText)
	if err != nil {
		num = 0
	}

	clone := sel.Clone()
	clone.Find("span.verse-number").Remove()
	clone.Find("sup.marker").Remove()
	clone.Find("span.iconPointer-OKie_").Remove()

	return JSTVerse{
		Number: num,
		Text:   normalizeWhitespace(clone.Text()),
	}
}
