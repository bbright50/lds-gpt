# Specification: Ranking & Deduplication

## 1. Goal

Re-rank combined search results (direct vector hits + graph traversal) using heuristic scoring, then trim to the final result set for LLM context.

## 2. User Stories

- **As a user**, I want the most relevant scripture verses prioritized in my results.
- **As a search system**, I need consistent ordering that favors direct verse matches over abstract study help entries.

## 3. Technical Requirements

- **Entry Point**: `internal/libsql/ranking.go` -> `rankResults()`
- **Ranking Algorithm**:
  1. Sort by distance (ascending — lower distance = more relevant)
  2. Apply verse bonus: `EntityVerse` results get `-0.05` distance adjustment (ranked higher)
  3. Stable sort to preserve original order for equal distances
- **Trim**: Final result set limited to `kNN` results (default 20)
- **Constants**:
  - `defaultVerseBonus = 0.05` — bonus subtracted from verse distances
  - Default kNN = 20

## 4. Acceptance Criteria

- Results are sorted by adjusted distance (lowest first).
- Verse entities are ranked higher than non-verse entities at equivalent distances.
- Result count does not exceed kNN limit.
- Ranking is deterministic (stable sort).

## 5. Edge Cases

- All results are verses (bonus applied uniformly; order by raw distance).
- No results at all (empty slice returned).
- Tied distances (stable sort preserves insertion order).
