# Knowledge Graph Domain

## Overview

The database is a knowledge graph stored in LibSQL (SQLite) using Ent ORM for schema management and code generation. It models the LDS standard works as a hierarchical structure of volumes, books, chapters, and verses, interconnected with study helps (Topical Guide, Bible Dictionary, Triple Combination Index, JST) via edge tables.

## Entity Hierarchy

```
Volume (ot, nt, bofm, dc-testament, pgp)
  └── Book (genesis, exodus, 1-nephi, ...)
       └── Chapter (number, summary, summary_embedding, url)
            ├── Verse (number, text, reference, footnote JSON fields)
            └── VerseGroup (sliding window, text, embedding, start/end verse numbers)
```

## Entities

### Structural Entities

| Entity | Key Fields | Notes |
|--------|-----------|-------|
| Volume | name, abbreviation | 5 canonical collections |
| Book | name, slug, url_path | Within a volume |
| Chapter | number, summary, summary_embedding, url | Holds verses |
| Verse | number, text, reference, translation_notes, alternate_readings, explanatory_notes | Atomic scripture unit |
| VerseGroup | text, embedding, start_verse_number, end_verse_number | Sliding window for RAG |

### Study Help Entities

| Entity | Key Fields | Notes |
|--------|-----------|-------|
| TopicalGuideEntry | name, embedding | ~3514 topics |
| BibleDictEntry | name, text, embedding | ~1276 articles |
| IndexEntry | name, embedding | ~3059 Triple Combination Index entries |
| JSTPassage | book, chapter, comprises, compare_ref, summary, text, embedding | ~100 JST passages |

### Edge/Join Entities

| Entity | From | To | Metadata |
|--------|------|-----|----------|
| VerseCrossRef | Verse | Verse | marker, category |
| VerseTGRef | Verse | TopicalGuideEntry | marker |
| VerseBDRef | Verse | BibleDictEntry | marker |
| VerseJSTRef | Verse | JSTPassage | marker |
| TGVerseRef | TopicalGuideEntry | Verse | phrase |
| BDVerseRef | BibleDictEntry | Verse | -- |
| IDXVerseRef | IndexEntry | Verse | phrase |

### Self-Referencing Edges

- TopicalGuideEntry -> TopicalGuideEntry (`see_also`)
- BibleDictEntry -> BibleDictEntry (`see_also`)
- IndexEntry -> IndexEntry (`see_also`)

### Cross-Type Edges

- TopicalGuideEntry -> BibleDictEntry (`bd_refs`)
- IndexEntry -> TopicalGuideEntry (`tg_refs`)
- IndexEntry -> BibleDictEntry (`bd_refs`)
- JSTPassage -> Verse (`compare_verses`)

## Vector Embeddings

| Entity | Column | Dimensions | Model |
|--------|--------|-----------|-------|
| VerseGroup | embedding | 1024 (F32_BLOB) | Titan Embed v2 |
| Chapter | summary_embedding | 1024 (F32_BLOB) | Titan Embed v2 |
| TopicalGuideEntry | embedding | 1024 (F32_BLOB) | Titan Embed v2 |
| BibleDictEntry | embedding | 1024 (F32_BLOB) | Titan Embed v2 |
| IndexEntry | embedding | 1024 (F32_BLOB) | Titan Embed v2 |
| JSTPassage | embedding | 1024 (F32_BLOB) | Titan Embed v2 |

All embedding columns have DiskANN vector indices via `CREATE INDEX ... ON table(libsql_vector_idx(column))`.

## Schema Location

Ent schema definitions: `internal/libsql/schema/*.go`
Generated code: `internal/libsql/generated/` (run `go generate ./internal/libsql/generated`)
Migrations: `internal/libsql/migrate.go` (auto-migration + vector index DDL)
