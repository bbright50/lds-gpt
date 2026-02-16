# Specification Changelog

## 2026-02-16 -- Bootstrap

**Action**: Initial specification generated from existing codebase via `/bootstrap`.

**Specs created**:
- `INDEX.md` -- System architecture, project layout, tech stack
- `scraping.md` -- Web scraping pipeline (5 content types, caching, text extraction)
- `knowledge-graph.md` -- Database schema (9 entities, 7 edge types, 6 vector indices)
- `data-loading.md` -- 6-phase ETL pipeline (structural data through embeddings)
- `search.md` -- RAG vector search (6-table parallel cosine similarity)

**Codebase state at bootstrap**:
- Scraping pipeline: complete (all standard works + study helps)
- Data loading: complete (6 phases functional)
- Vector search: functional demo CLI
- HTTP API: not yet implemented
- Chat interface: not yet implemented
