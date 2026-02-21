# Specification: Schema & Migrations

## 1. Goal

Manage database schema creation, vector index setup, and connection configuration for the LibSQL knowledge graph.

## 2. User Stories

- **As a data loader**, I want the database schema auto-created so I don't manage SQL migrations manually.
- **As a search system**, I need DiskANN vector indices for fast similarity search.

## 3. Technical Requirements

- **Client**: `internal/libsql/client.go`
  - Dual access: Ent ORM for structured queries + sqlx for raw SQL (vector search)
  - Connection pragmas: `journal_mode=WAL`, `busy_timeout=5000`, `synchronous=NORMAL`
- **Migration**: `internal/libsql/migrate.go`
  - Ent auto-migration via `client.Schema.Create()`
  - 6 DiskANN vector indices created post-migration:
    - `idx_verse_groups_embedding`
    - `idx_chapters_summary_embedding`
    - `idx_topical_guide_entries_embedding`
    - `idx_bible_dict_entries_embedding`
    - `idx_index_entries_embedding`
    - `idx_jst_passages_embedding`
- **Testing**: `internal/libsql/testing_helpers.go`
  - `TestClient(t)` — in-memory SQLite with auto-migration for tests

## 4. Acceptance Criteria

- Schema created successfully from Ent definitions.
- All 6 vector indices created with `IF NOT EXISTS` (idempotent).
- WAL mode enabled for concurrent read/write.
- Test helper provides isolated in-memory database per test.

## 5. Edge Cases

- Running migrations on an existing database (idempotent; no data loss).
- Vector index creation on empty tables (valid; index populated on insert).
- Concurrent database access (WAL mode + busy_timeout handles this).
