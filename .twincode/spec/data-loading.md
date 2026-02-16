# Data Loading Domain

## Overview

The dataloader is a 6-phase ETL pipeline (`internal/dataloader/`) that reads scraped JSON files from `pkg/data/scriptures/` and populates the LibSQL knowledge graph. It handles structural data, study helps, cross-reference edges, and vector embeddings.

## CLI

- `cmd/dataloader/app.go` -- Entry point
- `task load` -- Run phases 1-5 (drops and recreates DB)
- `task embed` -- Run phase 6 only (against existing DB)
- `task load-and-embed` -- Run all 6 phases

### Flags

- `--embed` -- Include phase 6 (embedding generation) in full load
- `--embed-only` -- Run phase 6 only

## Phases

### Phase 1: Structural Data
- Walks `{dataDir}/{volume}/{book}/{chapter}.json` files
- Creates Volume, Book, Chapter, Verse records
- Populates `VerseIndex` (O(1) lookup: `{volume}/{slug}/{chapter}/{verse}` -> DB ID)
- Populates verse-level footnote JSON fields (translation_notes, alternate_readings, explanatory_notes)

### Phase 2: Study Help Entities
- Loads Topical Guide, Bible Dictionary, Index, and JST entries
- Creates TopicalGuideEntry, BibleDictEntry, IndexEntry, JSTPassage records
- Returns lookup maps (TG/BD/IDX by name, JST by book/chapter)

### Phase 3: Footnote Edges
- Parses footnote markers from each verse's JSON
- Creates VerseCrossRef, VerseTGRef, VerseBDRef, VerseJSTRef edges
- Uses `RefParser` to resolve scripture references to verse IDs

### Phase 4: Study Help Edges
- Parses references within TG, BD, IDX entries
- Creates TGVerseRef, BDVerseRef, IDXVerseRef edges (study help -> verse)
- Creates see-also edges (TG->TG, BD->BD, IDX->IDX)
- Creates cross-type edges (TG->BD, IDX->TG, IDX->BD)
- Links JST passages to their compare verses

### Phase 5: Verse Groups
- Creates sliding-window VerseGroup records for each chapter
- Groups of consecutive verses for embedding (the primary RAG retrieval unit)

### Phase 6: Embeddings
- Generates 1024-dim embeddings via AWS Bedrock Titan Embed v2
- Embeds: verse groups, chapter summaries, TG entries (name + phrases), BD entries (text), IDX entries (name + phrases), JST passages (summary + text)
- Concurrency: 8 parallel embedding requests with rate limiter
- Text truncated to 25,000 characters for Titan token limit
- Idempotent: only embeds rows with NULL embedding columns

## Key Data Structures

- **VerseIndex**: `map[string]int` for O(1) verse lookup by path
- **JSTIndex**: `map[string][]jstEntry` for JST lookup by book/chapter
- **RefParser**: Resolves scripture abbreviations and references to volume/book/chapter/verse paths
- **LoadStats**: Thread-safe counters and warning accumulator

## Abbreviation System

`internal/dataloader/abbreviations.go` contains:
- `buildAbbreviationMap()` -- Maps abbreviations (e.g., "Gen.", "1 Ne.") to BookInfo (volume, slug, display name)
- `buildSlugToAbbrevMap()` -- Maps URL slugs to standard abbreviations
- `buildBookDisplayNameMap()` -- Maps display names to BookInfo
