# Specification: Graph Relationships

## 1. Goal

Model the interconnections between scripture entities and study helps as typed edges in the knowledge graph, enabling 1-hop graph traversal for context expansion during search.

## 2. User Stories

- **As a search system**, I need verse cross-references so I can expand results to related passages.
- **As a search system**, I need study help to verse links so topical searches return actual scripture text.
- **As a search system**, I need study help to study help links so related topics are discoverable.

## 3. Technical Requirements

### Junction Tables (Edge Types)

| Edge Type | Schema File | From | To | Count |
|-----------|------------|------|-----|-------|
| VerseCrossRef | `verse_cross_ref.go` | Verse | Verse | ~23,616 |
| TGVerseRef | `tg_verse_ref.go` | TopicalGuideEntry | Verse | ~18,000+ |
| BDVerseRef | `bd_verse_ref.go` | BibleDictEntry | Verse | ~33+ |
| IDXVerseRef | `idx_verse_ref.go` | IndexEntry | Verse | ~3,000+ |
| JSTPassage.compare_verses | M2M | JSTPassage | Verse | varies |
| VerseGroup.verses | M2M | VerseGroup | Verse | ~30,000+ |

### Self-Edges (within study help types)

- TG see_also: TopicalGuideEntry to TopicalGuideEntry
- BD see_also: BibleDictEntry to BibleDictEntry
- IDX see_also: IndexEntry to IndexEntry

### Cross-Edges (between study help types)

- IDX.tg_refs: IndexEntry to TopicalGuideEntry
- IDX.bd_refs: IndexEntry to BibleDictEntry
- TG.bd_refs: TopicalGuideEntry to BibleDictEntry

### Traversal Diagram

```
                    see_also
              +----------------+
              |                |
              v                |
+-------------------+    +-------------------+
| TopicalGuideEntry | -- | BibleDictEntry    |
+-------------------+    +-------------------+
       |  tg_refs              |  bd_refs
       v                       v
+-------------------+    +-------------------+
|   IndexEntry      | -- |   Verse           |
+-------------------+    +-------------------+
       |  idx_refs         ^       |
       +-------------------+       |
                            cross_ref
                                   |
                                   v
                          +-------------------+
                          |   Verse           |
                          +-------------------+
```

## 4. Acceptance Criteria

- All junction tables created with correct foreign keys.
- Edges are directional and typed (category field where applicable).
- Graph traversal can follow any edge type from any seed entity.
- No orphaned edges (foreign key constraints enforced).

## 5. Edge Cases

- Bidirectional verse cross-references (stored as two directed edges or one?).
- Study help entries with no outgoing edges (valid; isolated nodes).
- Circular see_also references (handled by traversal depth limit).
