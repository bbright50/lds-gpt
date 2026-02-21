# Specification: Vector Search

## 1. Goal

Perform parallel semantic similarity search across 6 entity types using DiskANN vector indices, returning the most relevant entities for a given query embedding.

## 2. User Stories

- **As a user**, I want to search by meaning (not just keywords) so I can find relevant scriptures even when I don't know exact phrases.
- **As a search system**, I need to search multiple entity types simultaneously for comprehensive coverage.

## 3. Technical Requirements

- **Entry Point**: `internal/libsql/rag_search.go` -> `ContextSearch()`
- **Search Functions** (executed in parallel via errgroup):
  - `searchVerseGroups(ctx, embedding, limit)` — verse_groups.embedding
  - `searchChapters(ctx, embedding, limit)` — chapters.summary_embedding
  - `searchTopicalGuide(ctx, embedding, limit)` — topical_guide_entries.embedding
  - `searchBibleDict(ctx, embedding, limit)` — bible_dict_entries.embedding
  - `searchIndex(ctx, embedding, limit)` — index_entries.embedding
  - `searchJSTPassages(ctx, embedding, limit)` — jst_passages.embedding
- **SQL Pattern**: `SELECT ... FROM {table} WHERE ... ORDER BY vector_distance_cos(embedding, ?) LIMIT ?`
- **Default Limit**: 10 results per entity type (60 total before dedup)
- **Distance Metric**: Cosine distance via `vector_distance_cos()`

### Vector Index Configuration

6 DiskANN indices created in `internal/libsql/migrate.go`:
```sql
CREATE INDEX IF NOT EXISTS idx_{table}_{column}
  ON {table}(libsql_vector_idx({column}))
```

## 4. Acceptance Criteria

- All 6 search functions execute in parallel.
- Results include entity type, ID, text content, and cosine distance.
- Search completes within acceptable latency (single-digit seconds for local SQLite).
- Empty embedding columns are handled gracefully (filtered out).

## 5. Edge Cases

- Entities with NULL embeddings (not yet embedded; excluded from results).
- All 6 searches returning 0 results (valid; empty result set).
- Very similar distances across entity types (ranking stage handles prioritization).
