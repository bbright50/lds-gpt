package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strings"

	"lds-gpt/pkg/scraper"
)

const (
	scripturesDir = "pkg/data/scriptures"
	rawDir        = "pkg/data/raw"
)

func main() {
	outDir := flag.String("d", scripturesDir, "output directory for scraped files")
	cacheDir := flag.String("c", rawDir, "cache directory for raw HTML")
	flag.Parse()

	urls := flag.Args()
	if len(urls) == 0 {
		fmt.Fprintln(os.Stderr, "usage: scraper [-d output_dir] <url> [url...]")
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	for _, rawURL := range urls {
		if err := scrapeAndWrite(ctx, rawURL, *outDir, *cacheDir); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	}
}

func scrapeAndWrite(ctx context.Context, rawURL, outDir, cacheDir string) error {
	fmt.Fprintf(os.Stderr, "scraping %s\n", rawURL)

	ch, err := scraper.ScrapeChapter(ctx, rawURL, cacheDir)
	if err != nil {
		return fmt.Errorf("scrape %s: %w", rawURL, err)
	}

	outPath, err := outputPath(rawURL, outDir)
	if err != nil {
		return fmt.Errorf("derive output path: %w", err)
	}

	if err := writeJSON(ch, outPath); err != nil {
		return fmt.Errorf("write %s: %w", outPath, err)
	}

	fmt.Fprintf(os.Stderr, "wrote %s\n", outPath)
	return nil
}

// outputPath derives a file path from the scripture URL.
// e.g. /study/scriptures/ot/gen/1?lang=eng -> {outDir}/ot/gen/1.json
func outputPath(rawURL, outDir string) (string, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("parse url: %w", err)
	}

	const prefix = "/study/scriptures/"
	path := parsed.Path
	if !strings.HasPrefix(path, prefix) {
		return "", fmt.Errorf("unexpected url path: %s", path)
	}

	rel := strings.TrimPrefix(path, prefix)
	return filepath.Join(outDir, rel+".json"), nil
}

func writeJSON(ch scraper.Chapter, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}

	data, err := json.MarshalIndent(ch, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal json: %w", err)
	}

	return os.WriteFile(path, append(data, '\n'), 0644)
}
