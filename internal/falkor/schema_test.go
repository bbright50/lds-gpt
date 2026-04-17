package falkor

import (
	"context"
	"testing"
)

// TestMigrate_CreatesAllSixVectorIndexes confirms the generated CreateIndexes
// call issues the DDL FalkorDB needs before kNN queries can use the indexes.
// Running Migrate twice is a no-op — the generator's DDL is
// `already indexed`-tolerant. We assert the exact label/property pairs rather
// than a bare count so silent drift (a renamed @vector field, a missed label)
// fails loudly.
func TestMigrate_CreatesAllSixVectorIndexes(t *testing.T) {
	client := StartFalkorContainer(t)
	ctx := context.Background()

	if err := client.Migrate(ctx); err != nil {
		t.Fatalf("first Migrate: %v", err)
	}
	if err := client.Migrate(ctx); err != nil {
		t.Fatalf("second Migrate (should be idempotent): %v", err)
	}

	res, err := client.Raw().Query(`CALL db.indexes()`, nil, nil)
	if err != nil {
		t.Fatalf("introspect indexes: %v", err)
	}
	type key struct{ label, prop string }
	got := map[key]bool{}
	for res.Next() {
		rec := res.Record()
		label, _ := rec.Get("label")
		props, _ := rec.Get("properties")
		propsList, _ := props.([]interface{})
		for _, p := range propsList {
			got[key{label.(string), p.(string)}] = true
		}
	}
	want := []key{
		{"VerseGroup", "embedding"},
		{"Chapter", "summaryEmbedding"},
		{"TopicalGuideEntry", "embedding"},
		{"BibleDictEntry", "embedding"},
		{"IndexEntry", "embedding"},
		{"JSTPassage", "embedding"},
	}
	for _, w := range want {
		if !got[w] {
			t.Errorf("missing vector index for (%s, %s); have: %v", w.label, w.prop, got)
		}
	}
}

// TestVectorSimilarity_ViaRawClient is the critical Phase B validation. Stage
// 1 of DoContextualSearch runs kNN through the raw handle — go-ormql's
// FalkorDB driver has a known bug: its rewrite of @vector queries emits
// `CALL db.idx.vector.queryNodes($rw0, $rw1, $rw2, $rw3)` without wrapping
// `$rw3` in `vecf32(...)`, and FalkorDB ≥ 4.18 rejects a plain LIST<FLOAT>
// where a Vectorf32 is expected. We route Stage-1 through Client.Raw() and
// keep go-ormql for typed CRUD + graph traversal. See .claudemod/spec/search
// for the division of responsibilities.
func TestVectorSimilarity_ViaRawClient(t *testing.T) {
	client := StartFalkorContainer(t)
	ctx := context.Background()

	if err := client.Migrate(ctx); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	// Three directional vectors. v1 along axis 0 (identical to query), v2
	// a 45° mix of axes 0+1, v3 along axis 1 — so cosine similarity to the
	// query ≈ 1.0, 0.707, 0.001 respectively. Every element is a non-integer
	// float so falkordb-go's ToString serializes with decimal points (integer
	// stringification would type the param as LIST<INTEGER> and fail).
	// VerseGroup is the primary RAG retrieval unit — it owns the @vector
	// embedding in the schema, so kNN targets VerseGroup rather than Verse.
	seeds := []struct {
		ref string
		vec []float64
	}{
		{"g1", directionalVec(1024, []int{0}, 0.999)},
		{"g2", directionalVec(1024, []int{0, 1}, 0.707)},
		{"g3", directionalVec(1024, []int{1}, 0.999)},
	}
	for _, s := range seeds {
		if _, err := client.Raw().Query(
			`CREATE (g:VerseGroup {id: $ref, text: $ref, startVerseNumber: 1, endVerseNumber: 1, embedding: vecf32($vec)})`,
			map[string]interface{}{"ref": s.ref, "vec": toAnySlice(s.vec)},
			nil,
		); err != nil {
			t.Fatalf("seed %s: %v", s.ref, err)
		}
	}

	res, err := client.Raw().Query(
		`CALL db.idx.vector.queryNodes('VerseGroup', 'embedding', 3, vecf32($q))
		 YIELD node, score
		 RETURN node.id AS id, score
		 ORDER BY score ASC`,
		map[string]interface{}{"q": toAnySlice(queryVec())},
		nil,
	)
	if err != nil {
		t.Fatalf("kNN via raw: %v", err)
	}

	var got []string
	for res.Next() {
		rec := res.Record()
		id, _ := rec.GetByIndex(0)
		got = append(got, id.(string))
	}
	want := []string{"g1", "g2", "g3"}
	if len(got) != len(want) {
		t.Fatalf("got %d results, want %d", len(got), len(want))
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("rank[%d] = %q, want %q (full order: %v)", i, got[i], want[i], got)
		}
	}
}

// TestGraphTraversal_ViaGeneratedClient validates that go-ormql's typed
// Execute path works for non-vector queries — this is the path Stage 2 of
// DoContextualSearch will use for 1-hop graph expansion. A bug here would
// mean we can't lean on the generated client for relationship traversal.
func TestGraphTraversal_ViaGeneratedClient(t *testing.T) {
	client := StartFalkorContainer(t)
	ctx := context.Background()

	if err := client.Migrate(ctx); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	// Seed via raw Cypher (workaround for @vector non-null — matches what
	// Phase 1 of the production loader will do).
	seedQ := `
		CREATE (c:Chapter {id: 'ch-1', number: 1, summaryEmbedding: vecf32($e1)})
		CREATE (v1:Verse {id: 'v-1', number: 1, reference: 'Ch 1:1', text: 'alpha', embedding: vecf32($e1)})
		CREATE (v2:Verse {id: 'v-2', number: 2, reference: 'Ch 1:2', text: 'beta',  embedding: vecf32($e2)})
		CREATE (c)-[:HAS_VERSE]->(v1)
		CREATE (c)-[:HAS_VERSE]->(v2)
	`
	if _, err := client.Raw().Query(seedQ, map[string]interface{}{
		"e1": toAnySlice(directionalVec(1024, []int{0}, 0.999)),
		"e2": toAnySlice(directionalVec(1024, []int{1}, 0.999)),
	}, nil); err != nil {
		t.Fatalf("seed: %v", err)
	}

	result, err := client.GraphQL().Execute(ctx, `
		query {
			chapters {
				id
				number
				versesConnection { edges { node { id reference text } } }
			}
		}`, nil)
	if err != nil {
		t.Fatalf("chapters query: %v", err)
	}

	var decoded struct {
		Chapters []struct {
			Id                string `json:"id"`
			Number            int    `json:"number"`
			VersesConnection struct {
				Edges []struct {
					Node struct {
						Id        string `json:"id"`
						Reference string `json:"reference"`
						Text      string `json:"text"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"versesConnection"`
		} `json:"chapters"`
	}
	if err := result.Decode(&decoded); err != nil {
		t.Fatalf("decode: %v (raw: %+v)", err, result.Data())
	}

	if got, want := len(decoded.Chapters), 1; got != want {
		t.Fatalf("chapters = %d, want %d (data: %+v)", got, want, result.Data())
	}
	if got, want := len(decoded.Chapters[0].VersesConnection.Edges), 2; got != want {
		t.Fatalf("chapter[0] verses = %d, want %d", got, want)
	}
}

// TestGeneratedClient_VectorQuery_KnownBug pins the current go-ormql FalkorDB
// driver limitation so a regression is visible if we ever try to route kNN
// through the typed client again. Skipped today; re-enable once upstream
// fixes rewriteVectorQuery to emit vecf32(...) around the vector parameter.
func TestGeneratedClient_VectorQuery_KnownBug(t *testing.T) {
	t.Skip("go-ormql FalkorDB driver does not wrap the vector param in " +
		"vecf32(...) when rewriting @vector queries; Stage 1 kNN uses " +
		"Client.Raw() instead. Remove this skip when upstream ships a fix.")
}

// directionalVec builds a `dim`-length vector that is `magnitude` along each
// axis in `axes` and a small non-zero baseline (0.001) everywhere else. Every
// element is a non-integer float so params serialize with decimal points.
func directionalVec(dim int, axes []int, magnitude float64) []float64 {
	v := make([]float64, dim)
	for i := range v {
		v[i] = 0.001
	}
	for _, a := range axes {
		v[a] = magnitude
	}
	return v
}

// queryVec is the Stage-1 query used by TestVectorSimilarity_ViaRawClient.
// Pulled out so the expected ranking is self-documenting.
func queryVec() []float64 { return directionalVec(1024, []int{0}, 0.999) }

// toAnySlice converts []float64 to []interface{} because falkordb-go's
// BuildParamsHeader only stringifies []interface{} / []string / scalars —
// not []float64.
func toAnySlice(v []float64) []interface{} {
	out := make([]interface{}, len(v))
	for i, x := range v {
		out[i] = x
	}
	return out
}
