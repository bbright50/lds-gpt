package falkor

import (
	"context"
	"testing"
)

// TestDoContextualSearch_Stage1Ordering seeds three VerseGroup nodes with
// embeddings of known angular relationship to the query, runs the full
// pipeline, and asserts the ranking collapses to the expected order. This
// is the end-to-end acceptance test for Phase E.
func TestDoContextualSearch_Stage1Ordering(t *testing.T) {
	client := StartFalkorContainer(t)
	ctx := context.Background()
	if err := client.Migrate(ctx); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	seeds := []struct {
		id  string
		vec []float64
	}{
		{"vg/a", directionalFloat64(1024, []int{0}, 0.999)},    // identical to query
		{"vg/b", directionalFloat64(1024, []int{0, 1}, 0.707)}, // 45° mix
		{"vg/c", directionalFloat64(1024, []int{1}, 0.999)},    // orthogonal
	}
	for _, s := range seeds {
		vec := toAnyFloats(s.vec)
		if _, err := client.Raw().Query(
			`CREATE (:VerseGroup {id: $id, text: $id, startVerseNumber: 1, endVerseNumber: 1, embedding: vecf32($vec)})`,
			map[string]interface{}{"id": s.id, "vec": vec}, nil,
		); err != nil {
			t.Fatalf("seed %s: %v", s.id, err)
		}
	}

	query := directionalFloat64To32(1024, []int{0}, 0.999)
	results, err := client.DoContextualSearch(ctx, query, WithKNN(10))
	if err != nil {
		t.Fatalf("DoContextualSearch: %v", err)
	}

	if len(results) == 0 {
		t.Fatal("expected at least one result")
	}

	// The top result must be the VerseGroup identical to the query.
	if got, want := results[0].ID, "vg/a"; got != want {
		t.Errorf("top result ID = %q, want %q (full: %+v)", got, want, results)
	}
	// Cosine distance on identical vectors should be ~0.
	if results[0].Distance > 0.01 {
		t.Errorf("top result distance = %f, want near 0", results[0].Distance)
	}
	// Every result must be a VerseGroup (we didn't seed any other embedded
	// node types).
	for _, r := range results {
		if r.EntityType != EntityVerseGroup {
			t.Errorf("unexpected entity type %q in results", r.EntityType)
		}
	}
}

// TestDoContextualSearch_Stage2GraphExpansion proves that a Stage 1 hit on a
// VerseGroup pulls in its INCLUDES verses via the graph expansion step.
func TestDoContextualSearch_Stage2GraphExpansion(t *testing.T) {
	client := StartFalkorContainer(t)
	ctx := context.Background()
	if err := client.Migrate(ctx); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	queryVec := directionalFloat64(1024, []int{0}, 0.999)
	seedQ := `
		CREATE (g:VerseGroup {id: 'vg/1', text: 'group', startVerseNumber: 1, endVerseNumber: 2, embedding: vecf32($vec)})
		CREATE (v1:Verse {id: 'v/ot/gen/1/1', number: 1, reference: 'Gen. 1:1', text: 'alpha', translationNotes: '', alternateReadings: '', explanatoryNotes: ''})
		CREATE (v2:Verse {id: 'v/ot/gen/1/2', number: 2, reference: 'Gen. 1:2', text: 'beta',  translationNotes: '', alternateReadings: '', explanatoryNotes: ''})
		CREATE (g)-[:INCLUDES]->(v1)
		CREATE (g)-[:INCLUDES]->(v2)
	`
	if _, err := client.Raw().Query(seedQ, map[string]interface{}{"vec": toAnyFloats(queryVec)}, nil); err != nil {
		t.Fatalf("seed: %v", err)
	}

	query := directionalFloat64To32(1024, []int{0}, 0.999)
	results, err := client.DoContextualSearch(ctx, query, WithKNN(10))
	if err != nil {
		t.Fatalf("DoContextualSearch: %v", err)
	}

	// Expect: VerseGroup vg/1 (Stage 1) plus Verses v/ot/gen/1/1 and
	// v/ot/gen/1/2 (Stage 2 via INCLUDES).
	seen := map[string]bool{}
	for _, r := range results {
		seen[r.ID] = true
	}
	for _, id := range []string{"vg/1", "v/ot/gen/1/1", "v/ot/gen/1/2"} {
		if !seen[id] {
			t.Errorf("expected %q in results; got %+v", id, results)
		}
	}

	// Graph hits should carry synthetic distance = stage1 + hopPenalty
	// (greater than the stage1 distance).
	var stage1Distance float64
	for _, r := range results {
		if r.ID == "vg/1" {
			stage1Distance = r.Distance
		}
	}
	for _, r := range results {
		if r.EntityType == EntityVerse {
			if r.Distance <= stage1Distance {
				t.Errorf("verse %s distance (%f) must exceed stage1 seed distance (%f) by the hop penalty",
					r.ID, r.Distance, stage1Distance)
			}
		}
	}
}

// TestDoContextualSearch_EmptyEmbeddingErrors covers the input-validation path.
func TestDoContextualSearch_InputValidation(t *testing.T) {
	client := StartFalkorContainer(t)
	ctx := context.Background()
	if err := client.Migrate(ctx); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	if _, err := client.DoContextualSearch(ctx, nil); err == nil {
		t.Error("expected error for nil embedding")
	}
	if _, err := client.DoContextualSearch(ctx, []float32{0.1, 0.2}, WithKNN(0)); err == nil {
		t.Error("expected error for kNN=0")
	}
	if _, err := client.DoContextualSearch(ctx, []float32{0.1, 0.2}, WithKNN(-5)); err == nil {
		t.Error("expected error for negative kNN")
	}
}

// --- test helpers ---

// directionalFloat64 mirrors the test helper in schema_test but returns
// []float64 for direct use in Cypher params (via toAnyFloats).
func directionalFloat64(dim int, axes []int, magnitude float64) []float64 {
	v := make([]float64, dim)
	for i := range v {
		v[i] = 0.001
	}
	for _, a := range axes {
		v[a] = magnitude
	}
	return v
}

func directionalFloat64To32(dim int, axes []int, magnitude float64) []float32 {
	src := directionalFloat64(dim, axes, magnitude)
	out := make([]float32, len(src))
	for i, x := range src {
		out[i] = float32(x)
	}
	return out
}

// toAnyFloats packs a []float64 into []interface{} of float64 — needed
// because falkordb-go's ToString does not accept a bare []float64.
func toAnyFloats(v []float64) []interface{} {
	out := make([]interface{}, len(v))
	for i, x := range v {
		out[i] = x
	}
	return out
}
