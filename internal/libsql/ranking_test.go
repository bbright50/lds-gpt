package libsql

import "testing"

// --- Task 5: Synthetic distance scoring and deduplication ---

func TestAssignSyntheticDistances(t *testing.T) {
	seedDistance := 0.10

	graphResults := []SearchResult{
		{EntityType: EntityVerse, ID: 1, Name: "Verse 1"},
		{EntityType: EntityVerse, ID: 2, Name: "Verse 2"},
	}

	scored := assignSyntheticDistances(graphResults, seedDistance)

	expectedDistance := seedDistance + defaultHopPenalty // 0.10 + 0.05 = 0.15

	for i, r := range scored {
		if r.Distance != expectedDistance {
			t.Errorf("scored[%d].Distance = %f, want %f", i, r.Distance, expectedDistance)
		}
	}

	// Verify original slice is not modified (immutability).
	for _, r := range graphResults {
		if r.Distance != 0 {
			t.Error("original slice was mutated")
		}
	}
}

func TestAssignSyntheticDistancesEmpty(t *testing.T) {
	scored := assignSyntheticDistances(nil, 0.10)
	if len(scored) != 0 {
		t.Errorf("expected 0 results for nil input, got %d", len(scored))
	}
}

func TestDeduplicateStage1Wins(t *testing.T) {
	stage1 := []SearchResult{
		{EntityType: EntityVerseGroup, ID: 1, Distance: 0.10},
		{EntityType: EntityTopicalGuide, ID: 2, Distance: 0.15},
	}

	graphResults := []SearchResult{
		// Duplicates stage1[0] — should be discarded.
		{EntityType: EntityVerseGroup, ID: 1, Distance: 0.20},
		// Unique — should be kept.
		{EntityType: EntityVerse, ID: 3, Distance: 0.18},
	}

	deduped := deduplicateResults(stage1, graphResults)

	// Should only contain the unique graph result.
	if len(deduped) != 1 {
		t.Fatalf("expected 1 deduplicated result, got %d", len(deduped))
	}
	if deduped[0].ID != 3 {
		t.Errorf("expected ID 3, got %d", deduped[0].ID)
	}
}

func TestDeduplicateGraphKeepsLowerDistance(t *testing.T) {
	stage1 := []SearchResult{} // no stage1 results for this test

	graphResults := []SearchResult{
		{EntityType: EntityVerse, ID: 1, Distance: 0.20},
		{EntityType: EntityVerse, ID: 1, Distance: 0.15}, // same entity, lower distance
		{EntityType: EntityVerse, ID: 2, Distance: 0.25},
	}

	deduped := deduplicateResults(stage1, graphResults)

	// Should have 2 unique results.
	if len(deduped) != 2 {
		t.Fatalf("expected 2 results, got %d", len(deduped))
	}

	// Verse ID=1 should have the lower distance (0.15).
	for _, r := range deduped {
		if r.ID == 1 && r.Distance != 0.15 {
			t.Errorf("verse ID=1 should have distance 0.15, got %f", r.Distance)
		}
	}
}

func TestDeduplicateEmptyInputs(t *testing.T) {
	deduped := deduplicateResults(nil, nil)
	if len(deduped) != 0 {
		t.Errorf("expected 0 results for nil inputs, got %d", len(deduped))
	}

	deduped = deduplicateResults([]SearchResult{{ID: 1}}, nil)
	if len(deduped) != 0 {
		t.Errorf("expected 0 results with no graph results, got %d", len(deduped))
	}
}

// --- Task 6: Heuristic re-ranking ---

func TestRankResults(t *testing.T) {
	input := []SearchResult{
		{EntityType: EntityVerseGroup, ID: 1, Distance: 0.32},
		{EntityType: EntityVerse, ID: 2, Distance: 0.35},
		{EntityType: EntityTopicalGuide, ID: 3, Distance: 0.30},
	}

	ranked := rankResults(input)

	// Expected rank scores:
	// EntityVerseGroup (ID=1): 0.32 - 0.00 = 0.32
	// EntityVerse      (ID=2): 0.35 - 0.05 = 0.30  <-- best
	// EntityTopicalGuide(ID=3): 0.30 - 0.00 = 0.30  <-- tied with verse
	// After ranking: verse(0.30) or TG(0.30) first, then verse_group(0.32)

	if len(ranked) != 3 {
		t.Fatalf("expected 3 results, got %d", len(ranked))
	}

	// The verse_group with distance 0.32 should be last (rankScore 0.32).
	last := ranked[2]
	if last.EntityType != EntityVerseGroup || last.ID != 1 {
		t.Errorf("expected verse_group (ID=1) to be last, got %q (ID=%d)", last.EntityType, last.ID)
	}

	// The verse (ID=2) with distance 0.35 should beat verse_group (ID=1) with distance 0.32.
	// Verify verse appears before verse_group.
	verseIdx := -1
	vgIdx := -1
	for i, r := range ranked {
		if r.EntityType == EntityVerse && r.ID == 2 {
			verseIdx = i
		}
		if r.EntityType == EntityVerseGroup && r.ID == 1 {
			vgIdx = i
		}
	}
	if verseIdx > vgIdx {
		t.Errorf("verse (rankScore 0.30) should appear before verse_group (rankScore 0.32), but verse at %d, vg at %d", verseIdx, vgIdx)
	}
}

func TestRankResultsDistantVerseDontJump(t *testing.T) {
	input := []SearchResult{
		{EntityType: EntityVerseGroup, ID: 1, Distance: 0.10},
		{EntityType: EntityVerse, ID: 2, Distance: 0.50}, // distant verse
	}

	ranked := rankResults(input)

	if len(ranked) != 2 {
		t.Fatalf("expected 2 results, got %d", len(ranked))
	}

	// EntityVerseGroup (ID=1): 0.10 - 0.00 = 0.10
	// EntityVerse      (ID=2): 0.50 - 0.05 = 0.45
	// Verse_group should still be first.
	if ranked[0].ID != 1 {
		t.Errorf("distant verse should not jump over close verse_group; first result ID=%d, expected 1", ranked[0].ID)
	}
}

func TestRankResultsImmutability(t *testing.T) {
	original := []SearchResult{
		{EntityType: EntityVerseGroup, ID: 1, Distance: 0.50},
		{EntityType: EntityVerse, ID: 2, Distance: 0.10},
	}

	ranked := rankResults(original)

	if len(ranked) != 2 {
		t.Fatalf("expected 2 results, got %d", len(ranked))
	}

	// Original should be unchanged.
	if original[0].ID != 1 {
		t.Errorf("original was mutated: first ID = %d, want 1", original[0].ID)
	}

	// Ranked should be reordered.
	if ranked[0].ID != 2 {
		t.Errorf("ranked should have verse (ID=2) first, got ID=%d", ranked[0].ID)
	}
}

func TestRankResultsEmpty(t *testing.T) {
	ranked := rankResults(nil)
	if len(ranked) != 0 {
		t.Errorf("expected 0 results for nil input, got %d", len(ranked))
	}

	ranked = rankResults([]SearchResult{})
	if len(ranked) != 0 {
		t.Errorf("expected 0 results for empty input, got %d", len(ranked))
	}
}
