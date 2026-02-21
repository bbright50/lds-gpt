# Specification: Study Help Scraping

## 1. Goal

Scrape supplementary study materials (Topical Guide, Bible Dictionary, Triple Combination Index, Joseph Smith Translation) into structured JSON for knowledge graph ingestion.

## 2. User Stories

- **As a data engineer**, I want topical guide entries scraped so users can discover scriptures by topic.
- **As a data engineer**, I want bible dictionary entries scraped so definitions enrich search context.
- **As a data engineer**, I want index entries and JST passages scraped for comprehensive coverage.

## 3. Technical Requirements

- **Entry Points**:
  - `cmd/scrapers/tg/main.go` — Topical Guide (~3,514 entries)
  - `cmd/scrapers/bd/main.go` — Bible Dictionary (~1,276 entries)
  - `cmd/scrapers/tc/main.go` — Triple Combination Index (~3,059 entries)
  - `cmd/scrapers/jst/main.go` — Joseph Smith Translation (~100 entries)
- **Core Libraries**:
  - `pkg/scraper/topical.go` — TG parser
  - `pkg/scraper/bible_dict.go` — BD parser
  - `pkg/scraper/triple_index.go` — IDX parser
  - `pkg/scraper/jst.go` — JST parser
- **Output**: JSON files per entry type written to `data/` subdirectories

## 4. Acceptance Criteria

- All entries for each study help type are scraped completely.
- Each entry contains name, text/definition, and verse references where applicable.
- Cross-references between study helps (see_also links) are captured.

## 5. Edge Cases

- Entries with no verse references (definition-only entries).
- Self-referential see_also links within the same study help type.
- Cross-type references (e.g., TG entry referencing a BD entry).
