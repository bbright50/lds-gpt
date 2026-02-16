# Search Domain

## Overview

The search layer provides vector similarity search across the knowledge graph for RAG (Retrieval-Augmented Generation). It embeds a query using Bedrock Titan, then runs parallel cosine distance searches across 6 entity types, merging and ranking results.

## Search Flow

```
Query string
  │
  ▼
EmbedText() via Bedrock Titan
  │
  ▼
Parallel vector_distance_cos() across 6 tables
  │
  ▼
Merge + sort by distance
  │
  ▼
Top-K results (default: 20)
```

## Entity Types Searched

| Entity | Table | Vector Column | Text in Result |
|--------|-------|--------------|----------------|
| verse_group | verse_groups | embedding | concatenated verse texts |
| chapter | chapters | summary_embedding | chapter summary |
| topical_guide | topical_guide_entries | embedding | topic name |
| bible_dict | bible_dict_entries | embedding | name + article text |
| index | index_entries | embedding | entry name |
| jst_passage | jst_passages | embedding | passage text + metadata |

## API

### `App.DoContextualSearch(ctx, query, ...options) ([]SearchResult, error)`

Top-level entry point in `internal/app/app.go`. Embeds the query and delegates to `Client.DoContextualSearch`.

### `Client.DoContextualSearch(ctx, embedding, ...options) ([]SearchResult, error)`

In `internal/libsql/rag_search.go`. Runs 6 search functions in parallel via `errgroup`, collects results, sorts by distance, and trims to kNN.

### Options

- `WithKNN(n)` -- Limit total results returned (default: 20)

### SearchResult

```go
type SearchResult struct {
    EntityType EntityType  // which table
    ID         int         // database row ID
    Name       string      // entry name (TG, BD, IDX)
    Text       string      // full text content
    Distance   float64     // cosine distance (lower = more similar)
    Metadata   ResultMeta  // entity-specific fields
}
```

## Current State

The `cmd/app/app.go` is a demo CLI that runs a hardcoded query ("What is faith?") and prints results. There is no HTTP API server yet.

## Planned Improvements

- HTTP API server with search endpoints
- Chat interface using retrieved context
- Improved ranking (hybrid search, re-ranking, graph traversal)
