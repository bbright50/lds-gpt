package scraper

import (
	"context"
	"fmt"
	"strings"
	"unicode"

	"github.com/PuerkitoBio/goquery"
)

// BDEntry represents a single Bible Dictionary entry with its full text
// and extracted scripture references.
type BDEntry struct {
	Text       string   `json:"text"`
	References []string `json:"references"`
}

// ScrapeBDIndex fetches the Bible Dictionary index page and returns all unique entry URLs.
func ScrapeBDIndex(ctx context.Context, indexURL, cacheDir string) ([]string, error) {
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
		if !strings.Contains(href, "/study/scriptures/bd/") {
			return
		}
		if strings.Contains(href, "/bd/introduction") {
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

// ScrapeBDEntry fetches a single BD entry page and returns the title and entry data.
func ScrapeBDEntry(ctx context.Context, entryURL, cacheDir string) (string, BDEntry, bool, error) {
	doc, cached, err := fetchDocument(ctx, entryURL, cacheDir)
	if err != nil {
		return "", BDEntry{}, false, fmt.Errorf("fetch entry: %w", err)
	}

	title := extractBDTitle(doc)
	entry := extractBDEntry(doc)

	return title, entry, cached, nil
}

func extractBDTitle(doc *goquery.Document) string {
	titleEl := doc.Find("article h1").First()
	if titleEl.Length() == 0 {
		return ""
	}
	return normalizeWhitespace(titleEl.Text())
}

func extractBDEntry(doc *goquery.Document) BDEntry {
	text := extractBDText(doc)
	refs := extractBDReferences(doc)
	refs = splitCompositeRefs(refs)
	refs = addRefBookPrefixes(refs)

	return BDEntry{
		Text:       text,
		References: refs,
	}
}

// extractBDText collects all paragraph text from the body-block, joining
// multiple paragraphs with double newlines.
func extractBDText(doc *goquery.Document) string {
	var paragraphs []string

	doc.Find("article div.body-block p").Each(func(_ int, p *goquery.Selection) {
		text := normalizeWhitespace(p.Text())
		if text != "" {
			paragraphs = append(paragraphs, text)
		}
	})

	return strings.Join(paragraphs, "\n\n")
}

// extractBDReferences collects all scripture reference texts from the entry,
// excluding cross-references to other BD or TG entries.
func extractBDReferences(doc *goquery.Document) []string {
	var refs []string

	doc.Find("article div.body-block a.scripture-ref").Each(func(_ int, a *goquery.Selection) {
		href, exists := a.Attr("href")
		if !exists {
			return
		}
		if isCrossReference(href) {
			return
		}
		text := normalizeWhitespace(a.Text())
		if text != "" {
			refs = append(refs, text)
		}
	})

	return refs
}

// isCrossReference returns true if the href points to a BD or TG entry
// rather than an actual scripture passage.
func isCrossReference(href string) bool {
	return strings.Contains(href, "/scriptures/bd/") ||
		strings.Contains(href, "/scriptures/tg/")
}

// splitCompositeRefs expands references that contain comma-separated verse
// ranges into individual references. For example, "Ex. 4:10–16, 27–31"
// becomes ["Ex. 4:10–16", "Ex. 4:27–31"].
func splitCompositeRefs(refs []string) []string {
	var result []string
	for _, ref := range refs {
		colonIdx := strings.LastIndex(ref, ":")
		if colonIdx < 0 {
			result = append(result, ref)
			continue
		}

		versePart := ref[colonIdx+1:]
		if !strings.Contains(versePart, ", ") {
			result = append(result, ref)
			continue
		}

		prefix := ref[:colonIdx+1]
		for _, part := range strings.Split(versePart, ", ") {
			result = append(result, prefix+part)
		}
	}
	return result
}

// addRefBookPrefixes ensures each reference has a book name prefix.
// When a reference starts with a digit (e.g., "24:1, 9"), the book name
// is carried forward from the most recent reference that had one.
func addRefBookPrefixes(refs []string) []string {
	if len(refs) == 0 {
		return refs
	}

	result := make([]string, len(refs))
	var lastBook string

	for i, ref := range refs {
		if len(ref) > 0 && ref[0] >= '0' && ref[0] <= '9' && startsChapterVerse(ref) {
			if lastBook != "" {
				result[i] = lastBook + " " + ref
			} else {
				result[i] = ref
			}
		} else {
			lastBook = extractRefBookPrefix(ref)
			result[i] = ref
		}
	}

	return result
}

// extractRefBookPrefix extracts the book name prefix from a scripture reference.
// Examples: "Ex. 6:16" -> "Ex.", "1 Chr. 6:3" -> "1 Chr.", "D&C 84:18" -> "D&C"
func extractRefBookPrefix(ref string) string {
	for i, c := range ref {
		if unicode.IsDigit(c) && startsChapterVerse(ref[i:]) {
			return strings.TrimRight(ref[:i], " ")
		}
	}
	return ref
}

// startsChapterVerse checks if a string begins with a chapter:verse pattern
// like "6:16", "84:18", or just a bare chapter number like "29".
func startsChapterVerse(s string) bool {
	if len(s) == 0 {
		return false
	}

	for _, c := range s {
		switch {
		case unicode.IsDigit(c):
			continue
		case c == ':' || c == '\u2013' || c == '-':
			return true
		default:
			return false
		}
	}

	// All digits means a bare chapter number
	return true
}
