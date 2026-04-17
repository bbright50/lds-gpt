package falkor

import (
	"context"
	"testing"

	"github.com/FalkorDB/falkordb-go/v2"
)

// TestSpike_VectorIndexOnPlainList_VsVecf32 answers the FalkorDB behavioral
// question blocking Phase-D option B: does `CREATE VECTOR INDEX` include
// nodes whose property is a plain LIST<FLOAT>, or only nodes whose property
// is a VectorF32 (produced by vecf32(...))?
//
// Three nodes are inserted:
//
//   - g1: vecf32([0.999, 0.001, 0.001, ...]) — proper vectorf32, identical
//     to the query. Control: this MUST be found by kNN.
//   - g2: plain [0.9, 0.001, ...] list (no vecf32 wrap). If the index is
//     lenient, this appears in kNN results. If strict, it does not.
//   - g3: created as plain list (like Phase 1 would), then upgraded via
//     `SET embedding = vecf32(...)` in a second step. This is what Phase
//     6 would do.
//
// Success criteria for option B:
//   - g2 is NOT in kNN results (strict index).
//   - g3 IS in kNN results after the SET (vecf32 promotion works).
//   - g1 ranks first (sanity).
func TestSpike_VectorIndexOnPlainList_VsVecf32(t *testing.T) {
	client := StartFalkorContainer(t)
	ctx := context.Background()
	if err := client.Migrate(ctx); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	g := client.Raw()

	// g1: proper vecf32, identical to the query direction.
	if _, err := g.Query(
		`CREATE (:VerseGroup {id: 'g1', text: 'g1', startVerseNumber: 1, endVerseNumber: 1, embedding: vecf32($vec)})`,
		map[string]interface{}{"vec": toAnySlice(directionalVec(1024, []int{0}, 0.999))}, nil,
	); err != nil {
		t.Fatalf("create g1 (vecf32): %v", err)
	}

	// g2: plain list — NO vecf32 wrap. If index is lenient, this becomes
	// a kNN hit and option B is unsafe.
	if _, err := g.Query(
		`CREATE (:VerseGroup {id: 'g2', text: 'g2', startVerseNumber: 2, endVerseNumber: 2, embedding: $vec})`,
		map[string]interface{}{"vec": toAnySlice(directionalVec(1024, []int{0}, 0.99))}, nil,
	); err != nil {
		t.Fatalf("create g2 (plain list): %v", err)
	}

	// g3: plain list first (this is what Phase 1 would do under option B),
	// then upgraded via SET vecf32 (what Phase 6 would do).
	if _, err := g.Query(
		`CREATE (:VerseGroup {id: 'g3', text: 'g3', startVerseNumber: 3, endVerseNumber: 3, embedding: $vec})`,
		map[string]interface{}{"vec": toAnySlice(directionalVec(1024, []int{0}, 0.95))}, nil,
	); err != nil {
		t.Fatalf("create g3 (plain list, pre-upgrade): %v", err)
	}

	// Sanity-check kNN BEFORE the g3 upgrade. Record the IDs that come back.
	before := knnIDs(t, g, directionalVec(1024, []int{0}, 0.999))
	t.Logf("before g3 upgrade, kNN returned: %v", before)

	// Now upgrade g3 to a vecf32-typed property. Simulates Phase 6.
	if _, err := g.Query(
		`MATCH (n:VerseGroup {id: 'g3'})
		 SET n.embedding = vecf32($vec)`,
		map[string]interface{}{"vec": toAnySlice(directionalVec(1024, []int{0}, 0.95))}, nil,
	); err != nil {
		t.Fatalf("SET vecf32 on g3: %v", err)
	}

	after := knnIDs(t, g, directionalVec(1024, []int{0}, 0.999))
	t.Logf("after g3 upgrade, kNN returned: %v", after)

	// --- Assertions ---

	// (1) g1 (proper vecf32 from creation) must always be in kNN results.
	if !contains(before, "g1") || !contains(after, "g1") {
		t.Errorf("g1 (vecf32 at create) missing from kNN results: before=%v after=%v", before, after)
	}

	// (2) g2 (plain list, never upgraded) — the key question. If it appears,
	// the index is lenient and option B (placeholder-at-create) would
	// pollute the index with pre-embedding nodes.
	g2InBefore := contains(before, "g2")
	g2InAfter := contains(after, "g2")
	t.Logf("g2 (plain list, never upgraded) in kNN results: before=%v after=%v", g2InBefore, g2InAfter)
	if g2InBefore || g2InAfter {
		t.Errorf(
			"FAIL option B: FalkorDB's vector index is LENIENT — it includes plain-list nodes.\n"+
				"Inserting a placeholder at Phase 1 would pollute kNN with pre-embedding nodes.\n"+
				"Fall back to option A (drop @vector, manage indexes via hand-written DDL).\n"+
				"g2 showed up: before=%v after=%v",
			g2InBefore, g2InAfter,
		)
	}

	// (3) g3 must appear in kNN AFTER the SET vecf32 upgrade.
	if !contains(after, "g3") {
		t.Errorf(
			"SET vecf32 promotion did not add g3 to the index — Phase 6's "+
				"upgrade path fails. kNN after upgrade: %v",
			after,
		)
	}

	// (4) Bonus check: g3 should NOT be in the before set (its plain-list
	// state should not be indexed — same question as (2) for a different
	// node). If this passes but (2) also passes, the index is strict and
	// option B is green.
	if contains(before, "g3") {
		t.Errorf("g3 appeared in kNN before its SET upgrade — plain-list nodes are being indexed; option B unsafe. before=%v", before)
	}
}

func knnIDs(t *testing.T, g *falkordb.Graph, queryVec []float64) []string {
	t.Helper()
	res, err := g.Query(
		`CALL db.idx.vector.queryNodes('VerseGroup', 'embedding', 10, vecf32($q))
		 YIELD node, score
		 RETURN node.id AS id
		 ORDER BY score ASC`,
		map[string]interface{}{"q": toAnySlice(queryVec)}, nil,
	)
	if err != nil {
		t.Fatalf("kNN query: %v", err)
	}
	var out []string
	for res.Next() {
		id, _ := res.Record().Get("id")
		out = append(out, id.(string))
	}
	return out
}

func contains(xs []string, s string) bool {
	for _, x := range xs {
		if x == s {
			return true
		}
	}
	return false
}
