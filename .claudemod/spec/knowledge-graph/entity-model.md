# Specification: Entity Model

## 1. Goal

Define the knowledge graph's entity types and their attributes, representing the full structure of LDS scriptures and study materials as an interconnected graph with vector embeddings for semantic search.

## 2. User Stories

- **As a search system**, I need structured entities so I can perform typed vector search across different content categories.
- **As a data loader**, I need clear entity schemas so I can map scraped data to database records.

## 3. Technical Requirements

- **Schema Location**: `internal/libsql/schema/`
- **ORM**: Ent (entgo.io/ent)

### Primary Entities

| Entity | Schema File | Key Fields | Has Embedding |
|--------|------------|-----------|--------------|
| Volume | `volume.go` | name, abbreviation | No |
| Book | `book.go` | name, slug, url_path | No |
| Chapter | `chapter.go` | number, summary | Yes (summary_embedding) |
| Verse | `verse.go` | number, text, reference | No |
| VerseGroup | `verse_group.go` | start_verse, end_verse, text | Yes (embedding) |
| TopicalGuideEntry | `tg_entry.go` | name | Yes (embedding) |
| BibleDictEntry | `bd_entry.go` | name, text | Yes (embedding) |
| IndexEntry | `index_entry.go` | name | Yes (embedding) |
| JSTPassage | `jst_passage.go` | book, chapter, comprises, text | Yes (embedding) |

### Hierarchy

```
Volume (5)
  └── Book (~90)
       └── Chapter (~1,584)
            └── Verse (~45,000)
                 └── VerseGroup (~15,000) [sliding window, 2-3 verses]
```

### Embedding Format

- Dimension: 1024 (AWS Bedrock Titan Embed Text v2)
- Storage: `F32_BLOB(1024)` (little-endian float32 bytes)
- Conversion: `internal/utils/vec/encoding.go`

## 4. Acceptance Criteria

- All entity types defined with correct fields and relationships in Ent schema.
- Ent code generation produces working CRUD operations.
- Embedding columns accept 1024-dimensional F32_BLOB data.
- Hierarchical relationships (volume > book > chapter > verse) enforced via foreign keys.

## 5. Edge Cases

- Chapters with no summary (summary_embedding is nullable).
- Verse groups spanning chapter boundaries (not currently supported; groups are per-chapter).
- Study help entries with no verse references (valid; some are definition-only).
