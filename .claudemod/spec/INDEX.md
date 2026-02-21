# Project: LDS-GPT

## 1. Overview

A Go-based Retrieval-Augmented Generation (RAG) backend for an LDS scripture study application. The system scrapes scripture text, study helps, and cross-references from web sources, loads them into a knowledge graph backed by LibSQL with vector extensions, and provides multi-stage semantic search for context retrieval. Designed to power a chat/study frontend where users ask questions and receive scripture-grounded answers via an LLM.

## 2. System Components

- **Scrapers:** Go binaries (`cmd/scrapers/`) that fetch and parse scripture HTML into structured JSON. Supports standard works, topical guide, bible dictionary, triple combination index, and Joseph Smith Translation.
- **Data Loader (ETL):** 6-phase pipeline (`internal/dataloader/`) that ingests scraped JSON into the knowledge graph: structural data, study helps, footnote edges, study help edges, verse groups, and embeddings.
- **Knowledge Graph:** Ent ORM schema (`internal/libsql/schema/`) modeling volumes, books, chapters, verses, verse groups, study help entries, and their interconnections via junction tables.
- **Search Pipeline:** 3-stage retrieval (`internal/libsql/`) with parallel vector search across 6 entity types, 1-hop graph traversal for context expansion, and heuristic ranking.
- **Embedding Layer:** AWS Bedrock Titan Embed Text v2 client (`internal/bedrockembedding/`) producing 1024-dimensional vectors stored as F32_BLOB.
- **Database:** LibSQL (SQLite fork with DiskANN vector indices) accessed via Ent ORM and sqlx for raw vector queries.
- **Application:** Core search orchestrator (`internal/app/`) exposing `ContextSearch()` for RAG context retrieval. LLM generation layer planned but not yet implemented.

## 3. Data Flow

1. Scrapers fetch HTML from scripture web pages and write structured JSON to `data/`.
2. Data Loader reads JSON and populates the knowledge graph across 6 phases.
3. Embeddings are generated via AWS Bedrock and stored alongside entities.
4. User query is embedded and searched across 6 vector indices in parallel.
5. Top results are expanded via 1-hop graph traversal to gather related context.
6. Results are ranked, deduplicated, and returned as search context.
7. (Planned) Context is fed to an LLM for natural language answer generation.

## 4. Technology Stack

- **Language:** Go 1.24.5
- **ORM:** Ent (entgo.io/ent v0.14.5)
- **Database:** LibSQL (tursodatabase/go-libsql) with vector extensions
- **Raw SQL:** sqlx (jmoiron/sqlx)
- **Embeddings:** AWS Bedrock Runtime SDK (Titan Embed Text v2, 1024-dim)
- **Scraping:** goquery (PuerkitoBio/goquery)
- **Config:** Viper (spf13/viper)
- **Concurrency:** pond/v2 (alitto/pond), x/sync errgroup
- **Testing:** go.uber.org/mock (mockgen)
- **Task Runner:** Taskfile (go-task/task)

## 5. Security & Scaling

- **Secrets:** AWS credentials via environment/IAM; no hardcoded keys.
- **Rate Limiting:** Embedding requests throttled via pond worker pool (20 concurrent).
- **Database:** WAL journal mode, busy_timeout=5000ms for concurrency.
- **Scaling:** Planned LLM integration will require rate limiting and cost management.

## 6. Feature Specifications

- [./scraping/scripture-scraping.md](Scripture Scraping)
- [./scraping/study-help-scraping.md](Study Help Scraping)
- [./data-loading/etl-pipeline.md](ETL Pipeline)
- [./knowledge-graph/entity-model.md](Entity Model)
- [./knowledge-graph/relationships.md](Graph Relationships)
- [./search/vector-search.md](Vector Search)
- [./search/graph-traversal.md](Graph Traversal)
- [./search/ranking.md](Ranking & Deduplication)
- [./embedding/bedrock-client.md](Bedrock Embedding Client)
- [./database/schema-migrations.md](Schema & Migrations)
