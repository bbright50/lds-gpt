package scraper

import (
	"context"
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// ScrapeTripleIndex fetches the Triple Combination Index page and returns all unique entry URLs.
func ScrapeTripleIndex(ctx context.Context, indexURL, cacheDir string) ([]string, error) {
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
		if !strings.Contains(href, "/study/scriptures/triple-index/") {
			return
		}
		if strings.Contains(href, "/triple-index/introduction") {
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

// ScrapeTripleIndexEntry fetches a single Triple Index entry page and returns the title, entries,
// and whether the result was served from cache.
func ScrapeTripleIndexEntry(ctx context.Context, entryURL, cacheDir string) (string, []TopicalEntry, bool, error) {
	doc, cached, err := fetchDocument(ctx, entryURL, cacheDir)
	if err != nil {
		return "", nil, false, fmt.Errorf("fetch entry: %w", err)
	}

	title := extractTripleIndexTitle(doc)
	entries := extractTripleIndexEntries(doc)

	return title, entries, cached, nil
}

func extractTripleIndexTitle(doc *goquery.Document) string {
	titleEl := doc.Find("article h1").First()
	if titleEl.Length() == 0 {
		return ""
	}
	return normalizeWhitespace(titleEl.Text())
}

func extractTripleIndexEntries(doc *goquery.Document) []TopicalEntry {
	var entries []TopicalEntry

	doc.Find("article div.body-block p.title").Each(func(_ int, sel *goquery.Selection) {
		refs := parseCrossReferences(sel, "IDX", "/scriptures/triple-index/")
		entries = append(entries, refs...)
	})

	doc.Find("article div.body-block p.entry").Each(func(_ int, sel *goquery.Selection) {
		entry := parseScriptureEntry(sel)
		entries = append(entries, entry)
	})

	return addTopicalRefBookPrefixes(entries)
}
