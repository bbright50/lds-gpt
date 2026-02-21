# Specification: Graph Traversal

## 1. Goal

Expand vector search results by traversing 1 hop through the knowledge graph, discovering related entities (especially verses) that add context to the initial search hits.

## 2. User Stories

- **As a user**, I want search results enriched with related scripture verses so I get a fuller picture of a topic.
- **As a search system**, I need to convert abstract entities (topical guide entries, chapters) into concrete verse text for LLM context.

## 3. Technical Requirements

- **Entry Point**: `internal/libsql/graph_traversal.go` -> `traverseEdges()`
- **Traversal Routes** (by seed entity type):

| Seed Entity | Traversal Target | Edge Type |
|-------------|-----------------|-----------|
| VerseGroup | Verses | verse_group_verses M2M |
| Chapter | Verses | chapter_id FK (limited to graphLimit) |
| TopicalGuideEntry | Verses + see_also TG + BD refs | tg_verse_refs + self M2M + bd M2M |
| BibleDictEntry | Verses + see_also BD | bd_verse_refs + self M2M |
| IndexEntry | Verses + see_also IDX + TG refs + BD refs | idx_verse_refs + self M2M + tg/bd M2M |
| JSTPassage | Compare verses | compare_verses M2M |

- **Distance Assignment**: `assignSyntheticDistances()` applies `parentDistance + hopPenalty` to traversed entities
- **Deduplication**: `deduplicateResults()` keeps the entry with the lowest distance when duplicates are found
- **Constants**:
  - `defaultHopPenalty = 0.05` — distance penalty per hop
  - `defaultGraphLimit = 5` — max neighbors per seed entity

## 4. Acceptance Criteria

- Every vector search result is expanded via its entity-specific traversal route.
- Traversed entities receive synthetic distances (parent distance + penalty).
- Duplicate entities (found via multiple paths) keep the lowest distance.
- Graph traversal adds meaningful context without overwhelming the result set.

## 5. Edge Cases

- Seed entities with no outgoing edges (no expansion, original result preserved).
- Circular references via see_also (depth limited to 1 hop).
- Verse groups overlapping the same verses (deduplicated by verse ID).
