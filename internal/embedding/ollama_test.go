package embedding

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestOllamaEmbedBatch_SendsInputArray pins the wire format: a single POST
// to /api/embed carries every chunk as `input: [...]` and the response is
// fanned back out positionally. Guards against a regression where someone
// reverts `Input` to `string` (losing the batch amortisation) or slices up
// the request per-chunk (losing the round-trip saving).
func TestOllamaEmbedBatch_SendsInputArray(t *testing.T) {
	var got ollamaRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/embed" {
			t.Errorf("unexpected path %q", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(body, &got); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		embeddings := make([][]float64, len(got.Input))
		for i := range got.Input {
			embeddings[i] = []float64{float64(i) + 0.5}
		}
		_ = json.NewEncoder(w).Encode(ollamaResponse{
			Model:      got.Model,
			Embeddings: embeddings,
		})
	}))
	defer server.Close()

	c := NewOllamaClient(server.URL, "test-model")
	got.Input = nil

	inputs := []string{"alpha", "beta", "gamma"}
	vecs, err := c.EmbedBatch(context.Background(), inputs)
	if err != nil {
		t.Fatalf("EmbedBatch: %v", err)
	}

	if got.Model != "test-model" {
		t.Errorf("model = %q, want test-model", got.Model)
	}
	if len(got.Input) != len(inputs) {
		t.Fatalf("server saw %d inputs, want %d", len(got.Input), len(inputs))
	}
	for i, text := range inputs {
		if got.Input[i] != text {
			t.Errorf("input[%d] = %q, want %q", i, got.Input[i], text)
		}
	}
	for i, want := range []float64{0.5, 1.5, 2.5} {
		if vecs[i][0] != want {
			t.Errorf("vec[%d][0] = %v, want %v", i, vecs[i][0], want)
		}
	}
}

// TestOllamaEmbedText_WrapsBatch confirms the single-item path delegates
// through EmbedBatch (wire format: a 1-element input array) rather than
// living as its own code path.
func TestOllamaEmbedText_WrapsBatch(t *testing.T) {
	var got ollamaRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&got)
		_ = json.NewEncoder(w).Encode(ollamaResponse{
			Embeddings: [][]float64{{0.42}},
		})
	}))
	defer server.Close()

	c := NewOllamaClient(server.URL, "m")
	v, err := c.EmbedText(context.Background(), "solo")
	if err != nil {
		t.Fatalf("EmbedText: %v", err)
	}
	if len(got.Input) != 1 || got.Input[0] != "solo" {
		t.Errorf("server saw input %v, want [\"solo\"]", got.Input)
	}
	if v[0] != 0.42 {
		t.Errorf("vec[0] = %v, want 0.42", v[0])
	}
}

// TestOllamaEmbedBatch_EmptyInputShortCircuits confirms we don't fire an
// HTTP request when there's nothing to embed.
func TestOllamaEmbedBatch_EmptyInputShortCircuits(t *testing.T) {
	var called bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	c := NewOllamaClient(server.URL, "m")
	vecs, err := c.EmbedBatch(context.Background(), nil)
	if err != nil {
		t.Fatalf("EmbedBatch: %v", err)
	}
	if vecs != nil {
		t.Errorf("vecs = %v, want nil for empty input", vecs)
	}
	if called {
		t.Errorf("HTTP endpoint unexpectedly called for empty input")
	}
}
