# Specification: ETL Pipeline

## 1. Goal

Transform scraped JSON data into a fully populated knowledge graph with vector embeddings, executing 6 sequential phases that build structural data, study helps, relationship edges, verse groups, and embeddings.

## 2. User Stories

- **As a data engineer**, I want to load all scraped data into the database in a single command.
- **As a data engineer**, I want to re-run only the embedding phase when the embedding model changes.
- **As a data engineer**, I want warnings for unparseable references so I can fix scraping issues.

## 3. Technical Requirements

- **Entry Point**: `cmd/dataloader/app.go`
- **Core Library**: `internal/dataloader/loader.go`
- **Phases**:
  1. `loadStructuralData()` — Volumes, books, chapters, verses from scripture JSON
  2. `loadStudyHelps()` — TG, BD, IDX, JST entities from study help JSON
  3. `loadFootnoteEdges()` — Parse footnotes, create verse cross-refs and study help refs
  4. `loadStudyRefEdges()` — Link study helps to verses and to each other
  5. `loadVerseGroups()` — Create sliding-window verse groups (2-3 verses per group)
  6. `loadEmbeddings()` — Generate and store vector embeddings for all entity types
- **CLI Flags**:
  - `--embed` — Run all 6 phases
  - `--embed-only` — Run only phase 6 (requires existing DB)
- **Reference Parser**: `internal/dataloader/refparser.go` parses scripture references from footnote text using abbreviation maps
- **Data Model Dependencies**:
  - `internal/dataloader/types.go` — JSON deserialization structs
  - `internal/dataloader/abbreviations.go` — Book name to {volume, slug} mappings

## 4. Acceptance Criteria

- All 6 phases complete without fatal errors.
- ~45K verses, ~7,850 study help entries, ~15K verse groups loaded.
- ~41K+ footnote relations and ~13K+ study help relations created.
- Embeddings generated for all entity types (verse groups, chapters, TG, BD, IDX, JST).
- Unparseable references logged as warnings, not errors.

## 5. Edge Cases

- References with book name inheritance across semicolons (e.g., "Gen. 1:1; 2:3" means Genesis 2:3).
- Parenthetical references that should be ignored.
- Special book slugs (e.g., "js-m", "dc").
- Duplicate cross-references (deduplicated during loading).
- Missing data files (phase skipped with warning).
