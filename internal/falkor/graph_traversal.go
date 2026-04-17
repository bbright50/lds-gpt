package falkor

import (
	"context"
	"fmt"

	"github.com/FalkorDB/falkordb-go/v2"
)

// graphExpandAndDedup performs 1-hop graph traversal from each Stage 1
// seed, assigns synthetic distances (seedDist + defaultHopPenalty), and
// removes graph hits that duplicate a Stage 1 result.
func (c *Client) graphExpandAndDedup(ctx context.Context, stage1 []SearchResult) ([]SearchResult, error) {
	var all []SearchResult
	for _, seed := range stage1 {
		neighbors, err := traverseEdges(ctx, c.Raw(), seed, defaultGraphLimit)
		if err != nil {
			return nil, fmt.Errorf("graph traversal (seed %s/%s): %w", seed.EntityType, seed.ID, err)
		}
		all = append(all, assignSyntheticDistances(neighbors, seed.Distance)...)
	}
	return deduplicateResults(stage1, all), nil
}

// traverseEdges dispatches to an entity-type-specific 1-hop expansion.
func traverseEdges(ctx context.Context, g *falkordb.Graph, seed SearchResult, limit int) ([]SearchResult, error) {
	switch seed.EntityType {
	case EntityVerseGroup:
		return traverseVerseGroup(ctx, g, seed.ID, limit)
	case EntityChapter:
		return traverseChapter(ctx, g, seed.ID, limit)
	case EntityTopicalGuide:
		return traverseTopicalGuide(ctx, g, seed.ID, limit)
	case EntityBibleDict:
		return traverseBibleDict(ctx, g, seed.ID, limit)
	case EntityIndex:
		return traverseIndex(ctx, g, seed.ID, limit)
	case EntityJSTPassage:
		return traverseJSTPassage(ctx, g, seed.ID, limit)
	case EntityVerse:
		return nil, nil // leaf — no outbound edges
	}
	return nil, nil
}

// --- Per-entity 1-hop queries ---

func traverseVerseGroup(ctx context.Context, g *falkordb.Graph, id string, limit int) ([]SearchResult, error) {
	_ = ctx
	res, err := g.Query(
		`MATCH (:VerseGroup {id: $id})-[:INCLUDES]->(v:Verse)
		 RETURN v.id AS id, v.reference AS ref, v.text AS text, v.number AS number
		 LIMIT $limit`,
		map[string]interface{}{"id": id, "limit": limit}, nil,
	)
	if err != nil {
		return nil, err
	}
	return collectVerses(res), nil
}

func traverseChapter(ctx context.Context, g *falkordb.Graph, id string, limit int) ([]SearchResult, error) {
	_ = ctx
	res, err := g.Query(
		`MATCH (:Chapter {id: $id})-[:HAS_VERSE]->(v:Verse)
		 RETURN v.id AS id, v.reference AS ref, v.text AS text, v.number AS number
		 LIMIT $limit`,
		map[string]interface{}{"id": id, "limit": limit}, nil,
	)
	if err != nil {
		return nil, err
	}
	return collectVerses(res), nil
}

func traverseTopicalGuide(ctx context.Context, g *falkordb.Graph, id string, limit int) ([]SearchResult, error) {
	_ = ctx
	var out []SearchResult

	// Verse-producing
	if rs, err := g.Query(
		`MATCH (:TopicalGuideEntry {id: $id})-[:TG_VERSE_REF]->(v:Verse)
		 RETURN v.id AS id, v.reference AS ref, v.text AS text, v.number AS number
		 LIMIT $limit`,
		map[string]interface{}{"id": id, "limit": limit}, nil,
	); err != nil {
		return nil, fmt.Errorf("tg→verses: %w", err)
	} else {
		out = append(out, collectVerses(rs)...)
	}

	// See-also (TG → TG)
	if rs, err := g.Query(
		`MATCH (:TopicalGuideEntry {id: $id})-[:TG_SEE_ALSO]->(t:TopicalGuideEntry)
		 RETURN t.id AS id, t.name AS name
		 LIMIT $limit`,
		map[string]interface{}{"id": id, "limit": limit}, nil,
	); err != nil {
		return nil, fmt.Errorf("tg→see_also: %w", err)
	} else {
		out = append(out, collectNamedNodes(rs, EntityTopicalGuide)...)
	}

	// TG → BD
	if rs, err := g.Query(
		`MATCH (:TopicalGuideEntry {id: $id})-[:TG_BD_REF]->(b:BibleDictEntry)
		 RETURN b.id AS id, b.name AS name, b.text AS text
		 LIMIT $limit`,
		map[string]interface{}{"id": id, "limit": limit}, nil,
	); err != nil {
		return nil, fmt.Errorf("tg→bd_refs: %w", err)
	} else {
		out = append(out, collectNamedNodesWithText(rs, EntityBibleDict)...)
	}

	return out, nil
}

func traverseBibleDict(ctx context.Context, g *falkordb.Graph, id string, limit int) ([]SearchResult, error) {
	_ = ctx
	var out []SearchResult

	if rs, err := g.Query(
		`MATCH (:BibleDictEntry {id: $id})-[:BD_VERSE_REF]->(v:Verse)
		 RETURN v.id AS id, v.reference AS ref, v.text AS text, v.number AS number
		 LIMIT $limit`,
		map[string]interface{}{"id": id, "limit": limit}, nil,
	); err != nil {
		return nil, fmt.Errorf("bd→verses: %w", err)
	} else {
		out = append(out, collectVerses(rs)...)
	}

	if rs, err := g.Query(
		`MATCH (:BibleDictEntry {id: $id})-[:BD_SEE_ALSO]->(b:BibleDictEntry)
		 RETURN b.id AS id, b.name AS name, b.text AS text
		 LIMIT $limit`,
		map[string]interface{}{"id": id, "limit": limit}, nil,
	); err != nil {
		return nil, fmt.Errorf("bd→see_also: %w", err)
	} else {
		out = append(out, collectNamedNodesWithText(rs, EntityBibleDict)...)
	}

	return out, nil
}

func traverseIndex(ctx context.Context, g *falkordb.Graph, id string, limit int) ([]SearchResult, error) {
	_ = ctx
	var out []SearchResult

	if rs, err := g.Query(
		`MATCH (:IndexEntry {id: $id})-[:IDX_VERSE_REF]->(v:Verse)
		 RETURN v.id AS id, v.reference AS ref, v.text AS text, v.number AS number
		 LIMIT $limit`,
		map[string]interface{}{"id": id, "limit": limit}, nil,
	); err != nil {
		return nil, fmt.Errorf("idx→verses: %w", err)
	} else {
		out = append(out, collectVerses(rs)...)
	}

	if rs, err := g.Query(
		`MATCH (:IndexEntry {id: $id})-[:IDX_SEE_ALSO]->(x:IndexEntry)
		 RETURN x.id AS id, x.name AS name
		 LIMIT $limit`,
		map[string]interface{}{"id": id, "limit": limit}, nil,
	); err != nil {
		return nil, fmt.Errorf("idx→see_also: %w", err)
	} else {
		out = append(out, collectNamedNodes(rs, EntityIndex)...)
	}

	if rs, err := g.Query(
		`MATCH (:IndexEntry {id: $id})-[:IDX_TG_REF]->(t:TopicalGuideEntry)
		 RETURN t.id AS id, t.name AS name
		 LIMIT $limit`,
		map[string]interface{}{"id": id, "limit": limit}, nil,
	); err != nil {
		return nil, fmt.Errorf("idx→tg_refs: %w", err)
	} else {
		out = append(out, collectNamedNodes(rs, EntityTopicalGuide)...)
	}

	if rs, err := g.Query(
		`MATCH (:IndexEntry {id: $id})-[:IDX_BD_REF]->(b:BibleDictEntry)
		 RETURN b.id AS id, b.name AS name, b.text AS text
		 LIMIT $limit`,
		map[string]interface{}{"id": id, "limit": limit}, nil,
	); err != nil {
		return nil, fmt.Errorf("idx→bd_refs: %w", err)
	} else {
		out = append(out, collectNamedNodesWithText(rs, EntityBibleDict)...)
	}

	return out, nil
}

func traverseJSTPassage(ctx context.Context, g *falkordb.Graph, id string, limit int) ([]SearchResult, error) {
	_ = ctx
	res, err := g.Query(
		`MATCH (:JSTPassage {id: $id})-[:COMPARES]->(v:Verse)
		 RETURN v.id AS id, v.reference AS ref, v.text AS text, v.number AS number
		 LIMIT $limit`,
		map[string]interface{}{"id": id, "limit": limit}, nil,
	)
	if err != nil {
		return nil, fmt.Errorf("jst→compares: %w", err)
	}
	return collectVerses(res), nil
}

// --- Row extractors ---

func collectVerses(res *falkordb.QueryResult) []SearchResult {
	var out []SearchResult
	for res.Next() {
		rec := res.Record()
		id, _ := rec.Get("id")
		ref, _ := rec.Get("ref")
		text, _ := rec.Get("text")
		number, _ := rec.Get("number")
		out = append(out, SearchResult{
			EntityType: EntityVerse,
			ID:         asString(id),
			Name:       asString(ref),
			Text:       asString(text),
			Metadata: ResultMeta{
				VerseNumber: asInt(number),
				Reference:   asString(ref),
			},
		})
	}
	return out
}

func collectNamedNodes(res *falkordb.QueryResult, t EntityType) []SearchResult {
	var out []SearchResult
	for res.Next() {
		rec := res.Record()
		id, _ := rec.Get("id")
		name, _ := rec.Get("name")
		out = append(out, SearchResult{
			EntityType: t,
			ID:         asString(id),
			Name:       asString(name),
		})
	}
	return out
}

func collectNamedNodesWithText(res *falkordb.QueryResult, t EntityType) []SearchResult {
	var out []SearchResult
	for res.Next() {
		rec := res.Record()
		id, _ := rec.Get("id")
		name, _ := rec.Get("name")
		text, _ := rec.Get("text")
		out = append(out, SearchResult{
			EntityType: t,
			ID:         asString(id),
			Name:       asString(name),
			Text:       asString(text),
		})
	}
	return out
}
