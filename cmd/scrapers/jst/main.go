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
	defaultOutput   = "pkg/data/jst.json"
	defaultCacheDir = "pkg/data/raw"
	indexURL        = "https://www.churchofjesuschrist.org/study/scriptures/jst?lang=eng"
	requestDelay    = 50 * time.Millisecond
)

func main() {
	outPath := flag.String("o", defaultOutput, "output JSON file path")
	cacheDir := flag.String("c", defaultCacheDir, "cache directory for raw HTML")
	limit := flag.Int("n", 0, "limit number of pages to scrape (0 = all)")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if err := run(ctx, *outPath, *cacheDir, *limit); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, outPath, cacheDir string, limit int) error {
	fmt.Fprintln(os.Stderr, "fetching JST index...")

	urls, err := scraper.ScrapeJSTIndex(ctx, indexURL, cacheDir)
	if err != nil {
		return fmt.Errorf("scrape index: %w", err)
	}

	if limit > 0 && limit < len(urls) {
		urls = urls[:limit]
	}

	fmt.Fprintf(os.Stderr, "found %d pages\n", len(urls))

	var chapters []scraper.JSTChapter

	for i, u := range urls {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		ch, cached, err := scraper.ScrapeJSTPage(ctx, u, cacheDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  warning: %s: %v\n", u, err)
			continue
		}

		if len(ch.Entries) == 0 {
			fmt.Fprintf(os.Stderr, "  warning: no entries for %s\n", u)
			continue
		}

		chapters = append(chapters, ch)

		if (i+1)%25 == 0 || i+1 == len(urls) {
			fmt.Fprintf(os.Stderr, "  progress: %d/%d\n", i+1, len(urls))
		}

		if !cached {
			time.Sleep(requestDelay)
		}
	}

	if err := scraper.WriteJSON(chapters, outPath); err != nil {
		return fmt.Errorf("write output: %w", err)
	}

	fmt.Fprintf(os.Stderr, "wrote %s (%d chapters)\n", outPath, len(chapters))
	return nil
}
