package falkor

import "sort"

// assignSyntheticDistances returns a copy of results with every Distance set
// to seedDistance + defaultHopPenalty. The input is not modified.
func assignSyntheticDistances(results []SearchResult, seedDistance float64) []SearchResult {
	if len(results) == 0 {
		return nil
	}
	scored := make([]SearchResult, len(results))
	copy(scored, results)
	distance := seedDistance + defaultHopPenalty
	for i := range scored {
		scored[i].Distance = distance
	}
	return scored
}

// deduplicateResults drops graph-traversed results that duplicate Stage 1
// hits (same EntityType + ID). Among graph duplicates, keeps the lowest
// Distance copy. Stage 1 results are not modified; only the deduplicated
// Stage 2 set is returned.
func deduplicateResults(stage1, graphResults []SearchResult) []SearchResult {
	if len(graphResults) == 0 {
		return nil
	}
	type key struct {
		entityType EntityType
		id         string
	}

	stage1Set := make(map[key]struct{}, len(stage1))
	for _, r := range stage1 {
		stage1Set[key{r.EntityType, r.ID}] = struct{}{}
	}

	best := make(map[key]SearchResult)
	for _, r := range graphResults {
		k := key{r.EntityType, r.ID}
		if _, exists := stage1Set[k]; exists {
			continue
		}
		if existing, exists := best[k]; exists {
			if r.Distance < existing.Distance {
				best[k] = r
			}
		} else {
			best[k] = r
		}
	}

	deduped := make([]SearchResult, 0, len(best))
	for _, r := range best {
		deduped = append(deduped, r)
	}
	return deduped
}

// rankResults applies heuristic re-ranking: rankScore = distance - typeBonus.
// EntityVerse gets typeBonus = defaultVerseBonus; all other types get 0.
// Returns a new sorted slice (lower rankScore = higher rank).
func rankResults(results []SearchResult) []SearchResult {
	if len(results) == 0 {
		return nil
	}
	type scored struct {
		result    SearchResult
		rankScore float64
	}
	items := make([]scored, len(results))
	for i, r := range results {
		bonus := 0.0
		if r.EntityType == EntityVerse {
			bonus = defaultVerseBonus
		}
		items[i] = scored{result: r, rankScore: r.Distance - bonus}
	}
	sort.SliceStable(items, func(i, j int) bool {
		return items[i].rankScore < items[j].rankScore
	})
	ranked := make([]SearchResult, len(items))
	for i, item := range items {
		ranked[i] = item.result
	}
	return ranked
}
