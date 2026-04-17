# Specification: Scraper

## 1. Purpose

Fetches HTML from churchofjesuschrist.org and normalizes it into structured JSON that the Dataloader consumes. Covers five sources: scripture chapters (OT, NT, BofM, D&C, PoGP), the Topical Guide, the Bible Dictionary, the Triple Combination Index, and the Joseph Smith Translation. Scraping is the only path by which new content enters the system — all downstream domains consume its JSON output.

## 2. Key Components

- `pkg/scraper/scraper.go` — Chapter scraping: URL → `Chapter{verses, footnotes, summary}`. Includes the shared `fetchDocument` with on-disk HTML caching and the shared `WriteJSON` helper.
- `pkg/scraper/topical.go` — Topical Guide index discovery + per-entry scraping.
- `pkg/scraper/bible_dict.go` — Bible Dictionary index + entry scraping.
- `pkg/scraper/triple_index.go` — Triple Combination Index scraping.
- `pkg/scraper/jst.go` — Joseph Smith Translation appendix scraping.
- `pkg/scraper/types.go` — Shared JSON output types: `Chapter`, `Verse`, `Footnote`.
- `cmd/scrapers/scriptures/main.go` — CLI: scrape one or more chapter URLs to `pkg/data/scriptures/<volume>/<book>/<chapter>.json`.
- `cmd/scrapers/tg/main.go`, `cmd/scrapers/bd/main.go`, `cmd/scrapers/tc/main.go`, `cmd/scrapers/jst/main.go` — Per-source CLIs; each writes a single consolidated JSON (`tg.json`, `bd.json`, etc.).
- `cmd/scrapers/inspect/main.go` — Debug helper for inspecting raw HTML structure against new selectors.

## 3. Data Models

- **`Chapter`** — `{url, book, chapter, summary, verses[], footnotes{}}` where footnotes are keyed by full marker (e.g. `"1a"`).
- **`Verse`** — `{number, text, footnote_markers[]}`. `text` has verse numbers and footnote markers stripped.
- **`Footnote`** — `{category, reference_text, text}`. Category is a comma-joined set of `data-note-category` values (e.g. `"tg,bd"`, `"ie"`, `"trn"`, `"or"`).
- **`TopicalEntry`** (TG/IDX) — `{phrase?, reference, key?}`. Phrase+reference for scripture refs; reference+key for cross-references to other TG/BD entries.
- **JST types** — `JSTChapterJSON` grouping passages by book+chapter; each `JSTEntryJSON` has `comprises`, optional `compare` verse range, optional `summary`, and a list of `JSTVerseJSON{number, text}`.

## 4. Interfaces

- **`ScrapeChapter(ctx, rawURL, cacheDir) (Chapter, error)`** — Idempotent; returns cached HTML if present under `cacheDir`.
- **`ScrapeTopicalIndex(ctx, indexURL, cacheDir) ([]string, error)`** — Returns fully-qualified entry URLs discovered from the TG index.
- **`ScrapeTopicalEntry(ctx, entryURL, cacheDir) (title, []TopicalEntry, cached, error)`** — Per-entry scrape with cache-hit flag.
- **`WriteJSON(data, path) error`** — Indented JSON write with `SetEscapeHTML(false)` and mkdir-p on the parent.
- **Cache layout** — Raw HTML is stored under `cacheDir` mirroring the source URL's path below `/study/scriptures/` (e.g. `pkg/data/raw/ot/gen/1.html`).

## 5. Dependencies

- **Depends on:** PuerkitoBio/goquery (HTML parsing), Go stdlib `net/http` (30 s client timeout).
- **Depended on by:** Dataloader (reads the JSON tree via `DATA_DIR`); none of the scrapers link back into other project packages.

## 6. Acceptance Criteria

- Running a scripture scraper produces `{url, book, chapter, summary, verses, footnotes}` where each verse has a positive number and non-empty text, and every referenced footnote marker has a corresponding footnote entry.
- Re-running with a populated cache directory produces identical output without any network calls.
- TG/BD/IDX scrapers produce a single JSON file whose top-level keys are topic/entry names, each mapping to an array or record matching the Dataloader's expected shape (`TGEntryJSON`, `BDEntryJSON`, `IDXEntryJSON`).
- JST scraper produces `JSTChapterJSON` entries that Dataloader's `JSTIndex.Put` can key by `{book-slug, chapter}`.
- `WriteJSON` does not HTML-escape `&`, `<`, `>` in scripture text.

## 7. Edge Cases

- **Missing chapter number or book name** — Verses parse to `Number = 0`; Dataloader will reject `Positive()` verse numbers, so upstream malformed pages surface as load-phase errors.
- **Non-200 HTTP response** — Returns `"unexpected status: N"` error; no partial write to cache.
- **Cache path outside the `/study/scriptures/` prefix** — Returns an error from `cachePath` rather than writing a surprising file path.
- **Footnote markers with no matching `<li data-full-marker=...>`** — The marker is retained on the verse but omitted from the `footnotes` map; Dataloader tolerates dangling markers.
- **Reference/alternate/explanatory footnote categories (`trn`, `or`, `ie`)** — Stored under `category` and consumed by Dataloader into Verse-level JSON fields (`translation_notes`, `alternate_readings`, `explanatory_notes`) rather than graph edges.
