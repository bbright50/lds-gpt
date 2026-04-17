package falkor

import (
	"context"
	"fmt"
	"reflect"
	"testing"
)

func TestClient_ConnectAndPing(t *testing.T) {
	client := StartFalkorContainer(t)
	if err := client.Ping(context.Background()); err != nil {
		t.Fatalf("ping: %v", err)
	}
}

func TestClient_CreateAndReadNode(t *testing.T) {
	client := StartFalkorContainer(t)
	graph := client.Raw()

	if _, err := graph.Query(
		`CREATE (v:Verse {reference: "John 3:16", text: "For God so loved the world"})`,
		nil, nil,
	); err != nil {
		t.Fatalf("create: %v", err)
	}

	res, err := graph.Query(
		`MATCH (v:Verse) RETURN v.reference, v.text`,
		nil, nil,
	)
	if err != nil {
		t.Fatalf("match: %v", err)
	}
	if !res.Next() {
		t.Fatal("expected one record, got none")
	}
	rec := res.Record()
	ref, err := rec.GetByIndex(0)
	if err != nil {
		t.Fatalf("GetByIndex(0): %v", err)
	}
	if ref != "John 3:16" {
		t.Errorf("reference = %v, want John 3:16", ref)
	}
}

// TestClient_VectorIndexRoundTrip exercises the full vector pipeline against
// a real FalkorDB container:
//   - CREATE VECTOR INDEX (the DDL Phase C will automate)
//   - insert nodes with vecf32(...) embeddings
//   - CALL db.idx.vector.queryNodes (the procedure Phase E will invoke once
//     per entity-type during Stage 1 of DoContextualSearch)
//
// If any step fails, the migration plan is not viable as drafted.
func TestClient_VectorIndexRoundTrip(t *testing.T) {
	client := StartFalkorContainer(t)
	graph := client.Raw()

	if _, err := graph.Query(
		`CREATE VECTOR INDEX FOR (v:Verse) ON (v.embedding) OPTIONS {dimension: 3, similarityFunction: 'cosine'}`,
		nil, nil,
	); err != nil {
		t.Fatalf("create vector index: %v", err)
	}

	// Three nodes whose cosine similarity to [1,0,0] strictly orders v1 > v2 > v3.
	seeds := []struct {
		name string
		vec  string
	}{
		{"v1", "[1.0, 0.0, 0.0]"}, // cos = 1.0
		{"v2", "[0.9, 0.1, 0.0]"}, // cos ≈ 0.994
		{"v3", "[0.0, 1.0, 0.0]"}, // cos = 0.0
	}
	for _, s := range seeds {
		q := fmt.Sprintf(
			`CREATE (v:Verse {name: "%s", embedding: vecf32(%s)})`,
			s.name, s.vec,
		)
		if _, err := graph.Query(q, nil, nil); err != nil {
			t.Fatalf("insert %s: %v", s.name, err)
		}
	}

	res, err := graph.Query(
		`CALL db.idx.vector.queryNodes('Verse', 'embedding', 3, vecf32([1.0, 0.0, 0.0]))
		 YIELD node, score
		 RETURN node.name, score`,
		nil, nil,
	)
	if err != nil {
		t.Fatalf("knn query: %v", err)
	}

	var got []string
	for res.Next() {
		rec := res.Record()
		name, err := rec.GetByIndex(0)
		if err != nil {
			t.Fatalf("GetByIndex(0): %v", err)
		}
		got = append(got, name.(string))
	}

	want := []string{"v1", "v2", "v3"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("order = %v, want %v", got, want)
	}
}

func TestNewClient_RejectsEmptyConfig(t *testing.T) {
	cases := []struct {
		name string
		cfg  Config
	}{
		{"empty URL", Config{URL: "", GraphName: "g"}},
		{"empty GraphName", Config{URL: "redis://localhost:6379", GraphName: ""}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := NewClient(tc.cfg); err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}

func TestClose_NilClientIsSafe(t *testing.T) {
	var c *Client
	if err := c.Close(); err != nil {
		t.Errorf("Close on nil client: %v", err)
	}
}
