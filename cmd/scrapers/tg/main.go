package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"time"

	"lds-gpt/pkg/scraper"
)

const (
	defaultOutput   = "pkg/data/topical-guide.json"
	defaultCacheDir = "pkg/data/raw"
	indexURL        = "https://www.churchofjesuschrist.org/study/scriptures/tg?lang=eng"
	requestDelay    = 50 * time.Millisecond
)

func main() {
	outPath := flag.String("o", defaultOutput, "output JSON file path")
	cacheDir := flag.String("c", defaultCacheDir, "cache directory for raw HTML")
	limit := flag.Int("n", 0, "limit number of entries to scrape (0 = all)")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if err := run(ctx, *outPath, *cacheDir, *limit); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, outPath, cacheDir string, limit int) error {
	fmt.Fprintln(os.Stderr, "fetching topical guide index...")

	urls, err := scraper.ScrapeTopicalIndex(ctx, indexURL, cacheDir)
	if err != nil {
		return fmt.Errorf("scrape index: %w", err)
	}

	if limit > 0 && limit < len(urls) {
		urls = urls[:limit]
	}

	fmt.Fprintf(os.Stderr, "found %d entries\n", len(urls))

	result := make(map[string][]scraper.TopicalEntry)

	for i, u := range urls {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		title, entries, cached, err := scraper.ScrapeTopicalEntry(ctx, u, cacheDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  warning: %s: %v\n", u, err)
			continue
		}

		if title == "" {
			fmt.Fprintf(os.Stderr, "  warning: empty title for %s\n", u)
			continue
		}

		if existing, ok := result[title]; ok {
			fmt.Fprintf(os.Stderr, "  warning: duplicate title %q, merging entries\n", title)
			result[title] = append(existing, entries...)
		} else {
			result[title] = entries
		}

		if (i+1)%100 == 0 || i+1 == len(urls) {
			fmt.Fprintf(os.Stderr, "  progress: %d/%d\n", i+1, len(urls))
		}

		if !cached {
			time.Sleep(requestDelay)
		}
	}

	if err := scraper.WriteJSON(result, outPath); err != nil {
		return fmt.Errorf("write output: %w", err)
	}

	fmt.Fprintf(os.Stderr, "wrote %s (%d entries)\n", outPath, len(result))
	return nil
}
