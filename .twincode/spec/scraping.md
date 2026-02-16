# Scraping Domain

## Overview

Web scrapers fetch LDS scripture content from churchofjesuschrist.org, cache raw HTML locally, and extract structured data into JSON files. These JSON files serve as the input to the dataloader ETL pipeline.

## Content Types Scraped

| Content | Scraper | CLI | Output Format |
|---------|---------|-----|---------------|
| Scriptures (OT, NT, BoM, D&C, PGP) | `pkg/scraper/scraper.go` | `cmd/scrapers/scriptures/` | One JSON per chapter |
| Topical Guide | `pkg/scraper/topical.go` | `cmd/scrapers/tg/` | Topic-keyed JSON |
| Bible Dictionary | `pkg/scraper/bible_dict.go` | `cmd/scrapers/bd/` | Entry-keyed JSON |
| Triple Combination Index | `pkg/scraper/triple_index.go` | `cmd/scrapers/tc/` | Entry-keyed JSON |
| Joseph Smith Translation | `pkg/scraper/jst.go` | `cmd/scrapers/jst/` | Chapter-grouped JSON |

## Scripture Scraper (`pkg/scraper/`)

### Data Types

- **Chapter**: URL, book name, chapter number, summary, verses, footnotes
- **Verse**: number, text (cleaned of markers), footnote marker references
- **Footnote**: category (cross-ref, TG, BD, JST, trn, or, ie), reference text, content text

### Caching Strategy

- Raw HTML cached at `pkg/data/raw/{volume}/{book}/{chapter}.html`
- On subsequent runs, cached HTML is read instead of re-fetching
- Cache path derived from URL structure: `/study/scriptures/{volume}/{book}/{chapter}`

### Text Extraction

- Book name: `h1 span.dominant` or `h1#title1`
- Chapter number: `p.title-number` (regex for digits)
- Summary: `p.study-summary`
- Verses: `p.verse` elements, with verse number from `span.verse-number`
- Footnote markers: `a.study-note-ref sup.marker[data-value]`
- Footnotes: `footer.study-notes li[data-full-marker]`
- Verse text is cleaned: verse numbers, footnote markers, and icon elements are removed; whitespace is normalized

### Output

- JSON written via `WriteJSON()` with `SetEscapeHTML(false)` to preserve `&`, `<`, `>` in scripture text
- Output directory: `pkg/data/scriptures/{volume}/{book}/{chapter}.json`

## Task Runner Commands

All scraping is orchestrated via `Taskfile.yml`:

- `task scrape` -- Scrape specific URLs (pass as CLI args)
- `task scrape:ot` / `scrape:nt` / `scrape:bofm` / `scrape:dc` / `scrape:pgp` -- Scrape entire volumes
- `task scrape:tg` / `scrape:bd` / `scrape:tc` / `scrape:jst` -- Scrape study helps
- `task scrape:all` -- Scrape all standard works (1584 chapters)
