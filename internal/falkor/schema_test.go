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

// TestVectorSimilarity_ViaTypedClient exercises Stage-1 kNN through the
// go-ormql generated client (verseGroupsSimilar), which compiles down to
// `CALL db.idx.vector.queryNodes(..., vecf32(...))` after the fork-patched
// driver rewrite. Asserts cosine-distance ordering for a known directional
// seed set.
func TestVectorSimilarity_ViaTypedClient(t *testing.T) {
	client := StartFalkorContainer(t)
	ctx := context.Background()

	if err := client.Migrate(ctx); err != nil {
		t.Fatalf("migrate: %v", err)
	}

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

	var out struct {
		Hits []struct {
			Score float64 `json:"score"`
			Node  struct {
				Id string `json:"id"`
			} `json:"node"`
		} `json:"verseGroupsSimilar"`
	}
	if err := execQuery(ctx, client.GraphQL(), `
		query ($vec: [Float!]!, $first: Int) {
		  verseGroupsSimilar(vector: $vec, first: $first) {
		    score
		    node { id }
		  }
		}`, map[string]any{"vec": toAnySlice(queryVec()), "first": 3}, &out); err != nil {
		t.Fatalf("verseGroupsSimilar: %v", err)
	}

	got := make([]string, 0, len(out.Hits))
	for _, h := range out.Hits {
		got = append(got, h.Node.Id)
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

// TestNestedConnect_CreatesStubForMissingTarget pins a known caveat of the
// fork's nested-connect template: because `FOREACH` can't MATCH, the
// template uses `MERGE (target:<Label> {id: ...})` to resolve the target
// node. If the id doesn't refer to an existing node, MERGE creates a bare
// stub with only the id property set (no name, no other fields).
//
// Our loader never hits this because parents always pre-exist by the time
// their children reference them. But if a future caller (e.g. an HTTP
// write endpoint) ever exposes `connect` to untrusted input, a bad id
// would silently create a garbage node — at that point the caller is
// responsible for validating ids before the typed mutation runs, or
// switching to raw-Cypher MATCH semantics.
func TestNestedConnect_CreatesStubForMissingTarget(t *testing.T) {
	client := StartFalkorContainer(t)
	ctx := context.Background()
	if err := client.Migrate(ctx); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	// Create a Book with a connect pointing at a Volume that DOESN'T exist.
	_, err := client.GraphQL().Execute(ctx, `
		mutation ($input: [BookCreateInput!]!) {
		  createBooks(input: $input) { books { id } }
		}`, map[string]any{
		"input": []any{map[string]any{
			"id":      "book/orphan",
			"name":    "Orphan Book",
			"slug":    "orphan",
			"urlPath": "nowhere/orphan",
			"volume":  map[string]any{"connect": []any{map[string]any{"where": map[string]any{"id": "vol/does-not-exist"}}}},
		}},
	})
	if err != nil {
		t.Fatalf("createBooks: %v", err)
	}

	// The FOREACH+MERGE template has materialised a stub Volume.
	res, err := client.Raw().Query(
		`MATCH (v:Volume {id: 'vol/does-not-exist'}) RETURN v.id AS id, v.name AS name, v.abbreviation AS abbr`,
		nil, nil,
	)
	if err != nil {
		t.Fatalf("introspect stub: %v", err)
	}
	if !res.Next() {
		t.Fatal("expected a stub Volume for the missing connect target; none was created " +
			"— the template semantics have drifted, investigate before ignoring this")
	}
	rec := res.Record()
	id, _ := rec.Get("id")
	name, _ := rec.Get("name")
	abbr, _ := rec.Get("abbr")
	if id != "vol/does-not-exist" {
		t.Errorf("stub id = %v, want 'vol/does-not-exist'", id)
	}
	// The stub has NO other properties — a real Volume created via Phase 1
	// would have name + abbreviation set.
	if name != nil || abbr != nil {
		t.Errorf("stub unexpectedly carries other properties (name=%v, abbr=%v) — "+
			"MERGE may have matched a real node instead of creating a stub", name, abbr)
	}

	// The CONTAINS edge still gets created between the stub and the Book,
	// so from the graph's perspective the connection is real. This is the
	// reason stub creation is a correctness hazard in user-facing contexts
	// and not just a debug annoyance.
	res, err = client.Raw().Query(
		`MATCH (:Volume {id: 'vol/does-not-exist'})-[:CONTAINS]->(:Book {id: 'book/orphan'})
		 RETURN count(*) AS n`,
		nil, nil,
	)
	if err != nil {
		t.Fatalf("edge count: %v", err)
	}
	if !res.Next() {
		t.Fatal("edge count query returned no rows")
	}
	n, _ := res.Record().GetByIndex(0)
	if n.(int64) != 1 {
		t.Errorf("expected 1 CONTAINS edge between stub and Book, got %v", n)
	}
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
