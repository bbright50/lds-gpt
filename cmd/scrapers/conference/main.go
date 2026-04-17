package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"lds-gpt/pkg/scraper"
)

const (
	defaultOutputDir = "pkg/data/scriptures/conference"
	defaultCacheDir  = "pkg/data/raw"
	defaultFrom      = "1971-04"
	requestDelay     = 200 * time.Millisecond
)

func main() {
	from := flag.String("from", defaultFrom, "earliest conference YYYY-MM to scrape (inclusive)")
	to := flag.String("to", currentConferenceString(), "latest conference YYYY-MM to scrape (inclusive)")
	outDir := flag.String("o", defaultOutputDir, "output directory (one JSON per talk)")
	cacheDir := flag.String("c", defaultCacheDir, "cache directory for raw HTML")
	limit := flag.Int("n", 0, "limit: stop after N talks total (0 = all)")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if err := run(ctx, *from, *to, *outDir, *cacheDir, *limit); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, from, to, outDir, cacheDir string, limit int) error {
	fromY, fromM, err := scraper.ParseYearMonth(from)
	if err != nil {
		return fmt.Errorf("-from: %w", err)
	}
	toY, toM, err := scraper.ParseYearMonth(to)
	if err != nil {
		return fmt.Errorf("-to: %w", err)
	}

	pairs := scraper.GenerateConferences(fromY, fromM, toY, toM)
	if len(pairs) == 0 {
		return fmt.Errorf("no conferences in range %s..%s", from, to)
	}
	fmt.Fprintf(os.Stderr, "scraping %d conferences (%s..%s)\n", len(pairs), from, to)

	var totalTalks int
	for _, pair := range pairs {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		year, month := pair[0], pair[1]
		fmt.Fprintf(os.Stderr, "— %d-%02d: listing talks…\n", year, month)

		refs, err := scraper.ScrapeConferenceSession(ctx, year, month, cacheDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  warning: skipping %d-%02d: %v\n", year, month, err)
			continue
		}
		fmt.Fprintf(os.Stderr, "  %d-%02d: %d talks\n", year, month, len(refs))

		for i, ref := range refs {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			if limit > 0 && totalTalks >= limit {
				fmt.Fprintf(os.Stderr, "hit limit (%d talks), stopping\n", limit)
				return nil
			}

			talk, cached, err := scraper.ScrapeConferenceTalk(ctx, ref, year, month, cacheDir)
			if err != nil {
				fmt.Fprintf(os.Stderr, "  warning: %s: %v\n", ref.URL, err)
				continue
			}

			path := filepath.Join(outDir, fmt.Sprintf("%d/%02d/%s.json", year, month, ref.Slug))
			if err := scraper.WriteJSON(talk, path); err != nil {
				return fmt.Errorf("write %s: %w", path, err)
			}

			totalTalks++
			if (i+1)%10 == 0 || i+1 == len(refs) {
				fmt.Fprintf(os.Stderr, "    progress: %d/%d talks in %d-%02d (%d total)\n",
					i+1, len(refs), year, month, totalTalks)
			}
			if !cached {
				time.Sleep(requestDelay)
			}
		}
	}

	fmt.Fprintf(os.Stderr, "done: %d talks across %d conferences → %s\n", totalTalks, len(pairs), outDir)
	return nil
}

// currentConferenceString returns the most recently concluded conference
// as a YYYY-MM string. Used as the default `-to` so a plain
// `scrape:conference` picks up everything through now without the caller
// editing the CLI each new conference.
func currentConferenceString() string {
	now := time.Now()
	y, m := now.Year(), int(now.Month())
	switch {
	case m >= 10:
		return fmt.Sprintf("%d-10", y)
	case m >= 4:
		return fmt.Sprintf("%d-04", y)
	default:
		return fmt.Sprintf("%d-10", y-1)
	}
}
