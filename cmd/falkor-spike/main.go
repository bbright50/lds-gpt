// Phase A manual spike — mirrors TestClient_VectorIndexRoundTrip but runs
// against an externally-managed FalkorDB (e.g. `docker run -p 6379:6379
// falkordb/falkordb:latest`) so you can eyeball the API surface without
// booting a testcontainer.
//
// Retire once Phase B is under way — the generated go-ormql client will
// replace hand-rolled Cypher in every real caller.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"lds-gpt/internal/falkor"
)

func main() {
	url := os.Getenv("FALKORDB_URL")
	if url == "" {
		url = "redis://localhost:6379"
	}
	graphName := os.Getenv("FALKORDB_GRAPH")
	if graphName == "" {
		graphName = "falkor-spike"
	}

	client, err := falkor.NewClient(falkor.Config{URL: url, GraphName: graphName})
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer client.Close()

	if err := client.Ping(context.Background()); err != nil {
		log.Fatalf("ping: %v", err)
	}
	fmt.Println("connected to", url, "graph:", graphName)

	graph := client.Raw()

	// Clear any prior state so the spike is idempotent.
	_ = graph.Delete()

	if _, err := graph.Query(
		`CREATE VECTOR INDEX FOR (v:Verse) ON (v.embedding) OPTIONS {dimension: 3, similarityFunction: 'cosine'}`,
		nil, nil,
	); err != nil {
		log.Fatalf("create index: %v", err)
	}

	seeds := []struct {
		name string
		vec  string
	}{
		{"v1", "[1.0, 0.0, 0.0]"},
		{"v2", "[0.9, 0.1, 0.0]"},
		{"v3", "[0.0, 1.0, 0.0]"},
	}
	for _, s := range seeds {
		q := fmt.Sprintf(`CREATE (v:Verse {name: "%s", embedding: vecf32(%s)})`, s.name, s.vec)
		if _, err := graph.Query(q, nil, nil); err != nil {
			log.Fatalf("insert %s: %v", s.name, err)
		}
	}

	res, err := graph.Query(
		`CALL db.idx.vector.queryNodes('Verse', 'embedding', 3, vecf32([1.0, 0.0, 0.0]))
		 YIELD node, score
		 RETURN node.name, score`,
		nil, nil,
	)
	if err != nil {
		log.Fatalf("knn: %v", err)
	}

	fmt.Println("\nkNN ranking (query = [1,0,0]):")
	for res.Next() {
		rec := res.Record()
		name, _ := rec.GetByIndex(0)
		score, _ := rec.GetByIndex(1)
		fmt.Printf("  %-4s score=%v\n", name, score)
	}
}
