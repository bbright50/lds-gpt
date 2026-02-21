package libsql

import "sort"

// assignSyntheticDistances returns a new slice where each result's Distance
// is set to seedDistance + defaultHopPenalty. The original slice is not modified.
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

// deduplicateResults removes graph-traversed results that duplicate Stage 1 results
// (same EntityType + ID). Among duplicate graph results, keeps the lower distance.
// Returns the deduplicated graph results only (Stage 1 results are unchanged).
func deduplicateResults(stage1, graphResults []SearchResult) []SearchResult {
	if len(graphResults) == 0 {
		return nil
	}

	type key struct {
		entityType EntityType
		id         int
	}

	// Index Stage 1 results for quick lookup.
	stage1Set := make(map[key]struct{}, len(stage1))
	for _, r := range stage1 {
		stage1Set[key{r.EntityType, r.ID}] = struct{}{}
	}

	// Deduplicate graph results: discard if in Stage 1, keep lowest distance among graph dupes.
	best := make(map[key]SearchResult)
	for _, r := range graphResults {
		k := key{r.EntityType, r.ID}

		// Discard if it duplicates a Stage 1 result.
		if _, exists := stage1Set[k]; exists {
			continue
		}

		// Keep the one with lower distance among graph dupes.
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
// EntityVerse gets typeBonus = defaultVerseBonus; all others get 0.
// Returns a new sorted slice (lower rankScore = higher rank). Original is not modified.
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
		items[i] = scored{
			result:    r,
			rankScore: r.Distance - bonus,
		}
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
