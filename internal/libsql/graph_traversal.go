package libsql

import (
	"context"
	"fmt"

	"lds-gpt/internal/libsql/generated"
)

// traverseEdges performs a 1-hop graph traversal from the given seed result,
// following all edge types available for the seed's entity type.
// Returns related entities as SearchResult values (without distance assigned).
// Limit controls the maximum results per edge type.
func traverseEdges(ctx context.Context, ec *generated.Client, seed SearchResult, limit int) ([]SearchResult, error) {
	switch seed.EntityType {
	case EntityVerseGroup:
		return traverseVerseGroup(ctx, ec, seed.ID, limit)
	case EntityChapter:
		return traverseChapter(ctx, ec, seed.ID, limit)
	case EntityTopicalGuide:
		return traverseTopicalGuide(ctx, ec, seed.ID, limit)
	case EntityBibleDict:
		return traverseBibleDict(ctx, ec, seed.ID, limit)
	case EntityIndex:
		return traverseIndex(ctx, ec, seed.ID, limit)
	case EntityJSTPassage:
		return traverseJSTPassage(ctx, ec, seed.ID, limit)
	case EntityVerse:
		// Verses are leaf nodes — no outbound edges to traverse.
		return nil, nil
	default:
		return nil, nil
	}
}

// traverseVerseGroup follows verse_group → verses (M2M).
func traverseVerseGroup(ctx context.Context, ec *generated.Client, id, limit int) ([]SearchResult, error) {
	vg, err := ec.VerseGroup.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("graph: get verse_group %d: %w", id, err)
	}

	verses, err := ec.VerseGroup.QueryVerses(vg).Limit(limit).All(ctx)
	if err != nil {
		return nil, fmt.Errorf("graph: verse_group→verses: %w", err)
	}

	return versesToResults(verses), nil
}

// traverseChapter follows chapter → verses (O2M).
func traverseChapter(ctx context.Context, ec *generated.Client, id, limit int) ([]SearchResult, error) {
	ch, err := ec.Chapter.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("graph: get chapter %d: %w", id, err)
	}

	verses, err := ec.Chapter.QueryVerses(ch).Limit(limit).All(ctx)
	if err != nil {
		return nil, fmt.Errorf("graph: chapter→verses: %w", err)
	}

	return versesToResults(verses), nil
}

// traverseTopicalGuide follows:
//   - tg → verse_refs (TGVerseRef) — verse-producing
//   - tg → see_also (M2M) — study-help-producing
//   - tg → bd_refs (M2M) — study-help-producing
func traverseTopicalGuide(ctx context.Context, ec *generated.Client, id, limit int) ([]SearchResult, error) {
	tg, err := ec.TopicalGuideEntry.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("graph: get tg %d: %w", id, err)
	}

	var results []SearchResult

	// Verse-producing: tg → verse_refs.
	verses, err := ec.TopicalGuideEntry.QueryVerseRefs(tg).Limit(limit).All(ctx)
	if err != nil {
		return nil, fmt.Errorf("graph: tg→verse_refs: %w", err)
	}
	results = append(results, versesToResults(verses)...)

	// Study-help: tg → see_also.
	seeAlso, err := ec.TopicalGuideEntry.QuerySeeAlso(tg).Limit(limit).All(ctx)
	if err != nil {
		return nil, fmt.Errorf("graph: tg→see_also: %w", err)
	}
	for _, sa := range seeAlso {
		results = append(results, SearchResult{
			EntityType: EntityTopicalGuide,
			ID:         sa.ID,
			Name:       sa.Name,
		})
	}

	// Study-help: tg → bd_refs.
	bdRefs, err := ec.TopicalGuideEntry.QueryBdRefs(tg).Limit(limit).All(ctx)
	if err != nil {
		return nil, fmt.Errorf("graph: tg→bd_refs: %w", err)
	}
	for _, bd := range bdRefs {
		results = append(results, SearchResult{
			EntityType: EntityBibleDict,
			ID:         bd.ID,
			Name:       bd.Name,
			Text:       bd.Text,
		})
	}

	return results, nil
}

// traverseBibleDict follows:
//   - bd → verse_refs (BDVerseRef) — verse-producing
//   - bd → see_also (M2M) — study-help-producing
func traverseBibleDict(ctx context.Context, ec *generated.Client, id, limit int) ([]SearchResult, error) {
	bd, err := ec.BibleDictEntry.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("graph: get bd %d: %w", id, err)
	}

	var results []SearchResult

	// Verse-producing: bd → verse_refs.
	verses, err := ec.BibleDictEntry.QueryVerseRefs(bd).Limit(limit).All(ctx)
	if err != nil {
		return nil, fmt.Errorf("graph: bd→verse_refs: %w", err)
	}
	results = append(results, versesToResults(verses)...)

	// Study-help: bd → see_also.
	seeAlso, err := ec.BibleDictEntry.QuerySeeAlso(bd).Limit(limit).All(ctx)
	if err != nil {
		return nil, fmt.Errorf("graph: bd→see_also: %w", err)
	}
	for _, sa := range seeAlso {
		results = append(results, SearchResult{
			EntityType: EntityBibleDict,
			ID:         sa.ID,
			Name:       sa.Name,
			Text:       sa.Text,
		})
	}

	return results, nil
}

// traverseIndex follows:
//   - idx → verse_refs (IDXVerseRef) — verse-producing
//   - idx → see_also, tg_refs, bd_refs (M2M) — study-help-producing
func traverseIndex(ctx context.Context, ec *generated.Client, id, limit int) ([]SearchResult, error) {
	idx, err := ec.IndexEntry.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("graph: get idx %d: %w", id, err)
	}

	var results []SearchResult

	// Verse-producing: idx → verse_refs.
	verses, err := ec.IndexEntry.QueryVerseRefs(idx).Limit(limit).All(ctx)
	if err != nil {
		return nil, fmt.Errorf("graph: idx→verse_refs: %w", err)
	}
	results = append(results, versesToResults(verses)...)

	// Study-help edges.
	studyHelps, err := traverseIndexStudyHelps(ctx, ec, idx, limit)
	if err != nil {
		return nil, err
	}
	results = append(results, studyHelps...)

	return results, nil
}

// traverseIndexStudyHelps follows study-help-producing edges from an index entry:
// see_also, tg_refs, and bd_refs.
func traverseIndexStudyHelps(ctx context.Context, ec *generated.Client, idx *generated.IndexEntry, limit int) ([]SearchResult, error) {
	var results []SearchResult

	seeAlso, err := ec.IndexEntry.QuerySeeAlso(idx).Limit(limit).All(ctx)
	if err != nil {
		return nil, fmt.Errorf("graph: idx→see_also: %w", err)
	}
	for _, sa := range seeAlso {
		results = append(results, SearchResult{
			EntityType: EntityIndex,
			ID:         sa.ID,
			Name:       sa.Name,
		})
	}

	tgRefs, err := ec.IndexEntry.QueryTgRefs(idx).Limit(limit).All(ctx)
	if err != nil {
		return nil, fmt.Errorf("graph: idx→tg_refs: %w", err)
	}
	for _, tg := range tgRefs {
		results = append(results, SearchResult{
			EntityType: EntityTopicalGuide,
			ID:         tg.ID,
			Name:       tg.Name,
		})
	}

	bdRefs, err := ec.IndexEntry.QueryBdRefs(idx).Limit(limit).All(ctx)
	if err != nil {
		return nil, fmt.Errorf("graph: idx→bd_refs: %w", err)
	}
	for _, bd := range bdRefs {
		results = append(results, SearchResult{
			EntityType: EntityBibleDict,
			ID:         bd.ID,
			Name:       bd.Name,
			Text:       bd.Text,
		})
	}

	return results, nil
}

// traverseJSTPassage follows jst_passage → compare_verses (M2M).
func traverseJSTPassage(ctx context.Context, ec *generated.Client, id, limit int) ([]SearchResult, error) {
	jst, err := ec.JSTPassage.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("graph: get jst %d: %w", id, err)
	}

	verses, err := ec.JSTPassage.QueryCompareVerses(jst).Limit(limit).All(ctx)
	if err != nil {
		return nil, fmt.Errorf("graph: jst→compare_verses: %w", err)
	}

	return versesToResults(verses), nil
}

// versesToResults converts Ent Verse entities to SearchResult values.
func versesToResults(verses []*generated.Verse) []SearchResult {
	results := make([]SearchResult, len(verses))
	for i, v := range verses {
		results[i] = SearchResult{
			EntityType: EntityVerse,
			ID:         v.ID,
			Name:       v.Reference,
			Text:       v.Text,
			Metadata: ResultMeta{
				VerseNumber: v.Number,
				Reference:   v.Reference,
			},
		}
	}
	return results
}
