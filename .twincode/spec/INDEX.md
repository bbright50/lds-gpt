# System Architecture

## Overview

**Application**: lds-gpt
**Type**: Scripture knowledge graph with semantic search API
**Tech Stack**: Go 1.24, LibSQL (SQLite), Ent ORM, sqlx, AWS Bedrock Titan Embed v2, goquery
**Audience**: Small team / family
**Purpose**: API backend for a web app that enables conversational scripture study via RAG-powered semantic search over LDS standard works

## High-Level Architecture

```
┌──────────────┐    ┌──────────────┐    ┌──────────────┐    ┌──────────────┐
│   Scrapers   │───>│  JSON Files  │───>│  Dataloader  │───>│  LibSQL DB   │
│  (goquery)   │    │ (pkg/data/)  │    │  (6 phases)  │    │ (knowledge   │
│              │    │              │    │              │    │   graph)     │
└──────────────┘    └──────────────┘    └──────────────┘    └──────┬───────┘
                                                                   │
                                                          ┌────────▼───────┐
                                                          │  App / Search  │
                                                          │ (vector search │
                                                          │  across 6      │
                                                          │  entity types) │
                                                          └────────────────┘
```

### Pipeline Stages

1. **Scrape** - Fetch HTML from churchofjesuschrist.org, cache locally, extract structured data to JSON
2. **Load** - 6-phase ETL into LibSQL: structural data, study helps, footnote edges, study ref edges, verse groups, embeddings
3. **Search** - Parallel cosine similarity search across 6 vector-indexed entity types, merged and ranked

## Project Layout

```
lds-gpt/
├── cmd/
│   ├── app/              # Main application entry point (search demo)
│   ├── dataloader/       # ETL pipeline CLI
│   └── scrapers/         # Scraper CLIs (scriptures, tg, bd, tc, jst, inspect)
├── internal/
│   ├── app/              # Application layer (contextual search orchestration)
│   ├── bedrockembedding/ # AWS Bedrock Titan embedding client
│   ├── dataloader/       # ETL pipeline (6 loading phases + embedding)
│   ├── libsql/           # Database client, migrations, RAG search, schema
│   │   ├── generated/    # Ent-generated ORM code (do not edit)
│   │   └── schema/       # Ent schema definitions (source of truth)
│   └── utils/            # Shared utilities (vec encoding, rate limiter, etc.)
├── pkg/
│   ├── scraper/          # Scripture scraping library
│   └── data/             # Scraped data (raw HTML + parsed JSON)
├── data/                 # SQLite database files
├── Taskfile.yml          # Task runner commands
└── .env                  # Environment config
```

## Database

**Engine**: LibSQL (SQLite-compatible with vector extensions)
**ORM**: Ent (schema-first, code generation)
**Raw SQL**: sqlx (for vector search queries)
**Connection**: Shared `*sql.DB` pool with WAL mode enabled
**Vector Index**: DiskANN via `libsql_vector_idx()` on F32_BLOB(1024) columns

## External Dependencies

| Dependency | Purpose |
|-----------|---------|
| AWS Bedrock (Titan Embed v2) | 1024-dim text embeddings |
| churchofjesuschrist.org | Source data for scraping |
| LibSQL (go-libsql driver) | SQLite with vector search |
| Ent ORM | Schema management and ORM |
| goquery | HTML parsing for scrapers |
| Viper | Configuration management |
| Task (taskfile.dev) | Build/run automation |

## Spec Directory

| File | Description |
|------|-------------|
| `INDEX.md` | This file -- architecture overview and spec directory |
| `scraping.md` | Web scraping pipeline: scrapers, types, caching |
| `knowledge-graph.md` | Database schema: entities, edges, vector indices |
| `data-loading.md` | ETL pipeline: 6 phases, abbreviation maps, ref parsing |
| `search.md` | RAG search: vector queries, result types, ranking |
| `CHANGELOG.md` | Rolling log of spec changes and completed tasks |
| `SESSION_STATE.md` | Current session status and handoff notes |
