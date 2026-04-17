package falkor

import (
	"context"
	"fmt"

	ormql "github.com/tab58/go-ormql/pkg/client"
)

// Stage 2 — 1-hop graph expansion from each Stage 1 seed.
//
// All six traversals flow through the go-ormql generated client via
// `Client.GraphQL().Execute(ctx, query, vars)`. Each query returns a nested
// result that decodes into a strongly typed struct; the per-entity
// functions then project those structs onto `SearchResult`.
//
// Previously TG/BD/IDX lived on raw Cypher because of the translator's
// naive `+s` pluralization (`TopicalGuideEntry` → `topicalGuideEntrys`);
// that bug is fixed in our local fork (see the replace directive in go.mod),
// so every label uses the typed path again.

func (c *Client) graphExpandAndDedup(ctx context.Context, stage1 []SearchResult) ([]SearchResult, error) {
	var all []SearchResult
	for _, seed := range stage1 {
		neighbors, err := c.traverseEdges(ctx, seed, defaultGraphLimit)
		if err != nil {
			return nil, fmt.Errorf("graph traversal (seed %s/%s): %w", seed.EntityType, seed.ID, err)
		}
		all = append(all, assignSyntheticDistances(neighbors, seed.Distance)...)
	}
	return deduplicateResults(stage1, all), nil
}

func (c *Client) traverseEdges(ctx context.Context, seed SearchResult, limit int) ([]SearchResult, error) {
	switch seed.EntityType {
	case EntityVerseGroup:
		return traverseVerseGroup(ctx, c.GraphQL(), seed.ID, limit)
	case EntityChapter:
		return traverseChapter(ctx, c.GraphQL(), seed.ID, limit)
	case EntityTopicalGuide:
		return traverseTopicalGuide(ctx, c.GraphQL(), seed.ID, limit)
	case EntityBibleDict:
		return traverseBibleDict(ctx, c.GraphQL(), seed.ID, limit)
	case EntityIndex:
		return traverseIndex(ctx, c.GraphQL(), seed.ID, limit)
	case EntityJSTPassage:
		return traverseJSTPassage(ctx, c.GraphQL(), seed.ID, limit)
	case EntityVerse:
		return nil, nil // leaf — no outbound edges
	}
	return nil, nil
}

// --- Decode-friendly shapes. Each type-local struct mirrors the shape of
//     one GraphQL query's response so Result.Decode can populate it. ---

type verseNode struct {
	Id string `json:"id"`
	Reference  string `json:"reference"`
	Text       string `json:"text"`
	Number     int    `json:"number"`
}

type namedNode struct {
	Id string `json:"id"`
	Name       string `json:"name"`
}

type namedNodeWithText struct {
	Id string `json:"id"`
	Name       string `json:"name"`
	Text       string `json:"text"`
}

type verseEdge struct {
	Node verseNode `json:"node"`
}
type namedEdge struct {
	Node namedNode `json:"node"`
}
type namedWithTextEdge struct {
	Node namedNodeWithText `json:"node"`
}

// --- Per-entity traversal queries ---

func traverseVerseGroup(ctx context.Context, gc *ormql.Client, id string, limit int) ([]SearchResult, error) {
	var out struct {
		VerseGroups []struct {
			VersesConnection struct {
				Edges []verseEdge `json:"edges"`
			} `json:"versesConnection"`
		} `json:"verseGroups"`
	}
	if err := execQuery(ctx, gc, `
		query ($id: ID, $first: Int) {
		  verseGroups(where: { id: $id }) {
		    versesConnection(first: $first) {
		      edges { node { id, reference, text, number } }
		    }
		  }
		}`, map[string]any{"id": id, "first": limit}, &out); err != nil {
		return nil, fmt.Errorf("vg→verses: %w", err)
	}
	if len(out.VerseGroups) == 0 {
		return nil, nil
	}
	return versesFromEdges(out.VerseGroups[0].VersesConnection.Edges), nil
}

func traverseChapter(ctx context.Context, gc *ormql.Client, id string, limit int) ([]SearchResult, error) {
	var out struct {
		Chapters []struct {
			VersesConnection struct {
				Edges []verseEdge `json:"edges"`
			} `json:"versesConnection"`
		} `json:"chapters"`
	}
	if err := execQuery(ctx, gc, `
		query ($id: ID, $first: Int) {
		  chapters(where: { id: $id }) {
		    versesConnection(first: $first) {
		      edges { node { id, reference, text, number } }
		    }
		  }
		}`, map[string]any{"id": id, "first": limit}, &out); err != nil {
		return nil, fmt.Errorf("chapter→verses: %w", err)
	}
	if len(out.Chapters) == 0 {
		return nil, nil
	}
	return versesFromEdges(out.Chapters[0].VersesConnection.Edges), nil
}

func traverseTopicalGuide(ctx context.Context, gc *ormql.Client, id string, limit int) ([]SearchResult, error) {
	var out struct {
		TopicalGuideEntries []struct {
			VerseRefsConnection struct {
				Edges []verseEdge `json:"edges"`
			} `json:"verseRefsConnection"`
			SeeAlsoConnection struct {
				Edges []namedEdge `json:"edges"`
			} `json:"seeAlsoConnection"`
			BdRefsConnection struct {
				Edges []namedWithTextEdge `json:"edges"`
			} `json:"bdRefsConnection"`
		} `json:"topicalGuideEntries"`
	}
	if err := execQuery(ctx, gc, `
		query ($id: ID, $first: Int) {
		  topicalGuideEntries(where: { id: $id }) {
		    verseRefsConnection(first: $first) {
		      edges { node { id, reference, text, number } }
		    }
		    seeAlsoConnection(first: $first) {
		      edges { node { id, name } }
		    }
		    bdRefsConnection(first: $first) {
		      edges { node { id, name, text } }
		    }
		  }
		}`, map[string]any{"id": id, "first": limit}, &out); err != nil {
		return nil, fmt.Errorf("tg expansion: %w", err)
	}
	if len(out.TopicalGuideEntries) == 0 {
		return nil, nil
	}
	tg := out.TopicalGuideEntries[0]
	var results []SearchResult
	results = append(results, versesFromEdges(tg.VerseRefsConnection.Edges)...)
	results = append(results, namedFromEdges(tg.SeeAlsoConnection.Edges, EntityTopicalGuide)...)
	results = append(results, namedWithTextFromEdges(tg.BdRefsConnection.Edges, EntityBibleDict)...)
	return results, nil
}

func traverseBibleDict(ctx context.Context, gc *ormql.Client, id string, limit int) ([]SearchResult, error) {
	var out struct {
		BibleDictEntries []struct {
			VerseRefsConnection struct {
				Edges []verseEdge `json:"edges"`
			} `json:"verseRefsConnection"`
			SeeAlsoConnection struct {
				Edges []namedWithTextEdge `json:"edges"`
			} `json:"seeAlsoConnection"`
		} `json:"bibleDictEntries"`
	}
	if err := execQuery(ctx, gc, `
		query ($id: ID, $first: Int) {
		  bibleDictEntries(where: { id: $id }) {
		    verseRefsConnection(first: $first) {
		      edges { node { id, reference, text, number } }
		    }
		    seeAlsoConnection(first: $first) {
		      edges { node { id, name, text } }
		    }
		  }
		}`, map[string]any{"id": id, "first": limit}, &out); err != nil {
		return nil, fmt.Errorf("bd expansion: %w", err)
	}
	if len(out.BibleDictEntries) == 0 {
		return nil, nil
	}
	bd := out.BibleDictEntries[0]
	var results []SearchResult
	results = append(results, versesFromEdges(bd.VerseRefsConnection.Edges)...)
	results = append(results, namedWithTextFromEdges(bd.SeeAlsoConnection.Edges, EntityBibleDict)...)
	return results, nil
}

func traverseIndex(ctx context.Context, gc *ormql.Client, id string, limit int) ([]SearchResult, error) {
	var out struct {
		IndexEntries []struct {
			VerseRefsConnection struct {
				Edges []verseEdge `json:"edges"`
			} `json:"verseRefsConnection"`
			SeeAlsoConnection struct {
				Edges []namedEdge `json:"edges"`
			} `json:"seeAlsoConnection"`
			TgRefsConnection struct {
				Edges []namedEdge `json:"edges"`
			} `json:"tgRefsConnection"`
			BdRefsConnection struct {
				Edges []namedWithTextEdge `json:"edges"`
			} `json:"bdRefsConnection"`
		} `json:"indexEntries"`
	}
	if err := execQuery(ctx, gc, `
		query ($id: ID, $first: Int) {
		  indexEntries(where: { id: $id }) {
		    verseRefsConnection(first: $first) {
		      edges { node { id, reference, text, number } }
		    }
		    seeAlsoConnection(first: $first) { edges { node { id, name } } }
		    tgRefsConnection(first: $first)   { edges { node { id, name } } }
		    bdRefsConnection(first: $first)   { edges { node { id, name, text } } }
		  }
		}`, map[string]any{"id": id, "first": limit}, &out); err != nil {
		return nil, fmt.Errorf("idx expansion: %w", err)
	}
	if len(out.IndexEntries) == 0 {
		return nil, nil
	}
	idx := out.IndexEntries[0]
	var results []SearchResult
	results = append(results, versesFromEdges(idx.VerseRefsConnection.Edges)...)
	results = append(results, namedFromEdges(idx.SeeAlsoConnection.Edges, EntityIndex)...)
	results = append(results, namedFromEdges(idx.TgRefsConnection.Edges, EntityTopicalGuide)...)
	results = append(results, namedWithTextFromEdges(idx.BdRefsConnection.Edges, EntityBibleDict)...)
	return results, nil
}

func traverseJSTPassage(ctx context.Context, gc *ormql.Client, id string, limit int) ([]SearchResult, error) {
	var out struct {
		JSTPassages []struct {
			CompareVersesConnection struct {
				Edges []verseEdge `json:"edges"`
			} `json:"compareVersesConnection"`
		} `json:"jSTPassages"` // note: generator camel-cases "JST" → "jST"
	}
	if err := execQuery(ctx, gc, `
		query ($id: ID, $first: Int) {
		  jSTPassages(where: { id: $id }) {
		    compareVersesConnection(first: $first) {
		      edges { node { id, reference, text, number } }
		    }
		  }
		}`, map[string]any{"id": id, "first": limit}, &out); err != nil {
		return nil, fmt.Errorf("jst→compares: %w", err)
	}
	if len(out.JSTPassages) == 0 {
		return nil, nil
	}
	return versesFromEdges(out.JSTPassages[0].CompareVersesConnection.Edges), nil
}

// --- Helpers ---

// execQuery runs a GraphQL query through the go-ormql client and decodes
// the response into `out`.
func execQuery(ctx context.Context, gc *ormql.Client, query string, vars map[string]any, out any) error {
	result, err := gc.Execute(ctx, query, vars)
	if err != nil {
		return err
	}
	return result.Decode(out)
}

func versesFromEdges(edges []verseEdge) []SearchResult {
	out := make([]SearchResult, 0, len(edges))
	for _, e := range edges {
		out = append(out, SearchResult{
			EntityType: EntityVerse,
			ID:         e.Node.Id,
			Name:       e.Node.Reference,
			Text:       e.Node.Text,
			Metadata: ResultMeta{
				VerseNumber: e.Node.Number,
				Reference:   e.Node.Reference,
			},
		})
	}
	return out
}

func namedFromEdges(edges []namedEdge, t EntityType) []SearchResult {
	out := make([]SearchResult, 0, len(edges))
	for _, e := range edges {
		out = append(out, SearchResult{
			EntityType: t,
			ID:         e.Node.Id,
			Name:       e.Node.Name,
		})
	}
	return out
}

func namedWithTextFromEdges(edges []namedWithTextEdge, t EntityType) []SearchResult {
	out := make([]SearchResult, 0, len(edges))
	for _, e := range edges {
		out = append(out, SearchResult{
			EntityType: t,
			ID:         e.Node.Id,
			Name:       e.Node.Name,
			Text:       e.Node.Text,
		})
	}
	return out
}
