# Specification: Scripture Scraping

## 1. Goal

Fetch and parse the full LDS standard works (Old Testament, New Testament, Book of Mormon, Doctrine & Covenants, Pearl of Great Price) from web sources into structured JSON for downstream ETL ingestion.

## 2. User Stories

- **As a data engineer**, I want to scrape all scripture chapters so the knowledge graph has complete text coverage.
- **As a data engineer**, I want footnotes and cross-references extracted per verse so the graph can model relationships.
- **As a data engineer**, I want cached HTML so re-scrapes don't require network access.

## 3. Technical Requirements

- **Entry Point**: `cmd/scrapers/scriptures/main.go`
- **Core Library**: `pkg/scraper/scraper.go`
- **Parser**: goquery-based HTML extraction
- **Data Model**:
  - `Chapter` (book_name, chapter_number, summary, verses, footnotes)
  - `Verse` (number, text, reference)
  - `Footnote` (category: "scripture" | "study_help" | "other", reference_text, text)
- **Output**: JSON files per chapter written to `data/{volume_slug}/{book_slug}/`
- **Caching**: Raw HTML cached to disk in `cacheDir` to avoid repeat fetches

## 4. Acceptance Criteria

- All 1,584 chapters across 5 volumes are scraped.
- Each chapter JSON contains book name, chapter number, summary, verses with text, and footnotes with category classification.
- Footnotes are correctly categorized into scripture cross-refs, study help refs, and other.
- Re-running with existing cache does not make network requests.

## 5. Edge Cases

- Chapters with no summary (some D&C sections).
- Verses with multiple footnote markers.
- Footnotes referencing study helps vs. scripture cross-references (distinguished by category).
- Special book names with punctuation (e.g., "JS--M", "D&C").
