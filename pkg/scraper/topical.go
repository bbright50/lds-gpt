package scraper

import (
	"context"
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// TopicalEntry represents a single item in a Topical Guide entry.
// For scripture references: Phrase + Reference are set (Key is empty).
// For cross-references (TG, BD, etc.): Reference + Key are set (Phrase is empty).
type TopicalEntry struct {
	Phrase    string `json:"phrase,omitempty"`
	Reference string `json:"reference"`
	Key       string `json:"key,omitempty"`
}

// ScrapeTopicalIndex fetches the TG index page and returns all unique entry URLs.
func ScrapeTopicalIndex(ctx context.Context, indexURL, cacheDir string) ([]string, error) {
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
		if !strings.Contains(href, "/study/scriptures/tg/") {
			return
		}
		if strings.Contains(href, "/tg/introduction") {
			return
		}
		// Strip just the ?lang=eng portion to get slug, but keep it for the full URL
		if seen[href] {
			return
		}
		seen[href] = true
		urls = append(urls, "https://www.churchofjesuschrist.org"+href)
	})

	return urls, nil
}

// ScrapeTopicalEntry fetches a single TG entry page and returns the title, entries,
// and whether the result was served from cache.
func ScrapeTopicalEntry(ctx context.Context, entryURL, cacheDir string) (string, []TopicalEntry, bool, error) {
	doc, cached, err := fetchDocument(ctx, entryURL, cacheDir)
	if err != nil {
		return "", nil, false, fmt.Errorf("fetch entry: %w", err)
	}

	title := extractTopicalTitle(doc)
	entries := extractTopicalEntries(doc)

	return title, entries, cached, nil
}

func extractTopicalTitle(doc *goquery.Document) string {
	titleEl := doc.Find("article h1").First()
	if titleEl.Length() == 0 {
		return ""
	}
	return normalizeWhitespace(titleEl.Text())
}

func extractTopicalEntries(doc *goquery.Document) []TopicalEntry {
	var entries []TopicalEntry

	doc.Find("article div.body-block p.title").Each(func(_ int, sel *goquery.Selection) {
		refs := parseCrossReferences(sel, "TG", "/scriptures/tg/")
		entries = append(entries, refs...)
	})

	doc.Find("article div.body-block p.entry").Each(func(_ int, sel *goquery.Selection) {
		entry := parseScriptureEntry(sel)
		entries = append(entries, entry)
	})

	return addTopicalRefBookPrefixes(entries)
}

// parseScriptureEntry extracts a phrase and reference from a p.entry element.
// The phrase is all text before the first scripture-ref link.
// The reference is the combined text of all scripture-ref links with surrounding punctuation.
func parseScriptureEntry(sel *goquery.Selection) TopicalEntry {
	phrase := buildPhrase(sel)
	reference := buildReference(sel)

	return TopicalEntry{
		Phrase:    phrase,
		Reference: reference,
	}
}

// buildPhrase extracts the phrase text from a p.entry, which is everything
// before the first <a class="scripture-ref"> link.
func buildPhrase(sel *goquery.Selection) string {
	clone := sel.Clone()

	// Remove all scripture-ref links and their surrounding punctuation
	clone.Find("a.scripture-ref").Remove()

	text := normalizeWhitespace(clone.Text())

	// Clean up trailing/leading punctuation artifacts
	text = strings.TrimRight(text, " ,;.()")
	text = strings.TrimLeft(text, " ")

	return text
}

// buildReference walks the DOM children to collect all text starting from the
// first <a class="scripture-ref"> element, producing references like:
// "Ex. 20:5 (34:7; Deut. 5:9)" or just "Rom. 5:2"
func buildReference(sel *goquery.Selection) string {
	refs := sel.Find("a.scripture-ref")
	if refs.Length() == 0 {
		return ""
	}

	// Get the first ref node to use as a boundary marker
	firstRef := refs.First()
	firstRefNode := firstRef.Get(0)

	// Walk all descendant text/element nodes and collect text from the first
	// scripture-ref onward, using the raw DOM tree to preserve punctuation.
	var sb strings.Builder
	foundFirst := false

	sel.Contents().Each(func(_ int, child *goquery.Selection) {
		if !foundFirst {
			// Check if this child is or contains the first scripture-ref
			if child.Get(0) == firstRefNode || child.Find("a.scripture-ref").Length() > 0 || child.Is("a.scripture-ref") {
				foundFirst = true
			} else {
				return
			}
		}
		sb.WriteString(child.Text())
	})

	result := normalizeWhitespace(sb.String())
	return strings.TrimRight(result, " .")
}

// parseCrossReferences extracts cross-reference entries from a p.title element.
// These are "See" or "See also" references to TG, BD, or direct scripture refs.
// selfPrefix and selfPath configure how self-references are classified (e.g., "TG" + "/scriptures/tg/").
func parseCrossReferences(sel *goquery.Selection, selfPrefix, selfPath string) []TopicalEntry {
	var entries []TopicalEntry

	sel.Find("a.scripture-ref").Each(func(_ int, a *goquery.Selection) {
		href, exists := a.Attr("href")
		if !exists {
			return
		}

		entry := classifyAndBuildRef(href, a, selfPrefix, selfPath)
		entries = append(entries, entry)
	})

	return entries
}

// classifyAndBuildRef determines the entry type from the href and link content.
// selfPrefix/selfPath configure how self-references are detected (e.g., "TG" + "/scriptures/tg/"
// for the Topical Guide, or "IDX" + "/scriptures/triple-index/" for the Triple Combination Index).
func classifyAndBuildRef(href string, a *goquery.Selection, selfPrefix, selfPath string) TopicalEntry {
	// Check for explicit prefix in <small> tag (e.g., "BD", "GS")
	smallEl := a.Find("small")
	if smallEl.Length() > 0 {
		return TopicalEntry{
			Reference: normalizeWhitespace(smallEl.Text()),
			Key:       extractRefKey(a),
		}
	}

	// Self-reference (e.g., TG→TG or IDX→IDX)
	if strings.Contains(href, selfPath) {
		return TopicalEntry{
			Reference: selfPrefix,
			Key:       extractRefKey(a),
		}
	}

	// TG cross-reference
	if strings.Contains(href, "/scriptures/tg/") {
		return TopicalEntry{
			Reference: "TG",
			Key:       extractRefKey(a),
		}
	}

	// BD cross-reference
	if strings.Contains(href, "/scriptures/bd/") {
		return TopicalEntry{
			Reference: "BD",
			Key:       extractRefKey(a),
		}
	}

	// Direct scripture reference (no phrase, just reference text)
	return TopicalEntry{
		Reference: normalizeWhitespace(a.Text()),
	}
}

// extractRefKey gets the display name of a cross-reference, stripping any <small> prefix.
func extractRefKey(a *goquery.Selection) string {
	clone := a.Clone()
	clone.Find("small").Remove()
	return normalizeWhitespace(clone.Text())
}

// addTopicalRefBookPrefixes carries forward book name prefixes across
// consecutive scripture reference entries. Cross-references (Key != "") are
// skipped. For example, "D&C 29:21" followed by "45:28–33" becomes
// "D&C 29:21" followed by "D&C 45:28–33".
func addTopicalRefBookPrefixes(entries []TopicalEntry) []TopicalEntry {
	if len(entries) == 0 {
		return entries
	}

	result := make([]TopicalEntry, len(entries))
	var lastBook string

	for i, entry := range entries {
		result[i] = entry

		// Skip cross-references (TG, BD, etc.)
		if entry.Key != "" {
			continue
		}

		ref := entry.Reference
		if len(ref) > 0 && ref[0] >= '0' && ref[0] <= '9' && startsChapterVerse(ref) {
			if lastBook != "" {
				result[i].Reference = lastBook + " " + ref
			}
		} else {
			lastBook = extractRefBookPrefix(ref)
		}
	}

	return result
}
