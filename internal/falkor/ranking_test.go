package falkor

import "testing"

// These tests are pure — no container needed. They mirror the LibSQL
// ranking_test.go but with string IDs.

func TestAssignSyntheticDistances(t *testing.T) {
	seedDistance := 0.10
	graph := []SearchResult{
		{EntityType: EntityVerse, ID: "v/ot/gen/1/1", Name: "Verse 1"},
		{EntityType: EntityVerse, ID: "v/ot/gen/1/2", Name: "Verse 2"},
	}
	scored := assignSyntheticDistances(graph, seedDistance)
	want := seedDistance + defaultHopPenalty
	for i, r := range scored {
		if r.Distance != want {
			t.Errorf("scored[%d].Distance = %f, want %f", i, r.Distance, want)
		}
	}
	for _, r := range graph {
		if r.Distance != 0 {
			t.Error("original slice was mutated")
		}
	}
}

func TestAssignSyntheticDistancesEmpty(t *testing.T) {
	if got := assignSyntheticDistances(nil, 0.10); len(got) != 0 {
		t.Errorf("expected 0 results for nil, got %d", len(got))
	}
}

func TestDeduplicateStage1Wins(t *testing.T) {
	stage1 := []SearchResult{
		{EntityType: EntityVerseGroup, ID: "vg/1", Distance: 0.10},
		{EntityType: EntityTopicalGuide, ID: "tg/Faith", Distance: 0.15},
	}
	graph := []SearchResult{
		{EntityType: EntityVerseGroup, ID: "vg/1", Distance: 0.20}, // dupe of stage 1
		{EntityType: EntityVerse, ID: "v/ot/gen/1/1", Distance: 0.18},
	}
	deduped := deduplicateResults(stage1, graph)
	if len(deduped) != 1 {
		t.Fatalf("expected 1 deduped result, got %d", len(deduped))
	}
	if got, want := deduped[0].ID, "v/ot/gen/1/1"; got != want {
		t.Errorf("deduped[0].ID = %q, want %q", got, want)
	}
}

func TestDeduplicateGraphKeepsLowerDistance(t *testing.T) {
	graph := []SearchResult{
		{EntityType: EntityVerse, ID: "v/1", Distance: 0.20},
		{EntityType: EntityVerse, ID: "v/1", Distance: 0.15}, // same entity, lower
		{EntityType: EntityVerse, ID: "v/2", Distance: 0.25},
	}
	deduped := deduplicateResults(nil, graph)
	if len(deduped) != 2 {
		t.Fatalf("expected 2 results, got %d", len(deduped))
	}
	for _, r := range deduped {
		if r.ID == "v/1" && r.Distance != 0.15 {
			t.Errorf("v/1 should have distance 0.15, got %f", r.Distance)
		}
	}
}

func TestDeduplicateEmptyInputs(t *testing.T) {
	if got := deduplicateResults(nil, nil); len(got) != 0 {
		t.Errorf("expected 0 results for nil inputs, got %d", len(got))
	}
	if got := deduplicateResults([]SearchResult{{ID: "x"}}, nil); len(got) != 0 {
		t.Errorf("expected 0 results with no graph results, got %d", len(got))
	}
}

func TestRankResults_VerseBonusBreaksTies(t *testing.T) {
	input := []SearchResult{
		{EntityType: EntityVerseGroup, ID: "vg/1", Distance: 0.32},
		{EntityType: EntityVerse, ID: "v/2", Distance: 0.35},
		{EntityType: EntityTopicalGuide, ID: "tg/3", Distance: 0.30},
	}
	ranked := rankResults(input)
	if len(ranked) != 3 {
		t.Fatalf("expected 3 results, got %d", len(ranked))
	}
	// verse_group (0.32) should be last; verse (0.30 after bonus) should
	// tie with TG (0.30) and beat the verse_group.
	if last := ranked[2]; last.EntityType != EntityVerseGroup {
		t.Errorf("last result type = %q, want %q", last.EntityType, EntityVerseGroup)
	}
}

func TestRankResults_DistantVerseDoesNotJump(t *testing.T) {
	input := []SearchResult{
		{EntityType: EntityVerseGroup, ID: "vg/1", Distance: 0.10},
		{EntityType: EntityVerse, ID: "v/2", Distance: 0.50},
	}
	ranked := rankResults(input)
	if ranked[0].ID != "vg/1" {
		t.Errorf("distant verse jumped rank; first ID=%q, want vg/1", ranked[0].ID)
	}
}

func TestRankResultsImmutability(t *testing.T) {
	original := []SearchResult{
		{EntityType: EntityVerseGroup, ID: "vg/1", Distance: 0.50},
		{EntityType: EntityVerse, ID: "v/2", Distance: 0.10},
	}
	ranked := rankResults(original)
	if ranked[0].ID != "v/2" {
		t.Errorf("ranked first ID = %q, want v/2", ranked[0].ID)
	}
	if original[0].ID != "vg/1" {
		t.Errorf("input slice was mutated; first ID = %q", original[0].ID)
	}
}

func TestRankResultsEmpty(t *testing.T) {
	if got := rankResults(nil); len(got) != 0 {
		t.Errorf("expected 0 for nil input, got %d", len(got))
	}
	if got := rankResults([]SearchResult{}); len(got) != 0 {
		t.Errorf("expected 0 for empty input, got %d", len(got))
	}
}
