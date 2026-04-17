package embedding

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"lds-gpt/internal/utils/rate_limiter"
)

// defaultOllamaMaxConcurrent bounds in-flight requests to a remote Ollama
// server. Ollama serialises embed requests per model by default, so a low
// ceiling avoids piling up queue time; tune via WithMaxConcurrent if the
// server is backed by a batching runner.
const defaultOllamaMaxConcurrent = 8

// OllamaOption tweaks the client at construction time.
type OllamaOption func(*ollamaClient)

// WithMaxConcurrent overrides the default in-flight request ceiling.
func WithMaxConcurrent(n int) OllamaOption {
	return func(c *ollamaClient) { c.maxConcurrent = n }
}

// WithHTTPTimeout overrides the per-request timeout (default: 60s).
func WithHTTPTimeout(d time.Duration) OllamaOption {
	return func(c *ollamaClient) { c.http.Timeout = d }
}

type ollamaRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type ollamaResponse struct {
	Model      string      `json:"model"`
	Embeddings [][]float64 `json:"embeddings"`
}

type ollamaClient struct {
	*rate_limiter.Embeddable[[]float64]

	baseURL       string
	model         string
	maxConcurrent int
	http          *http.Client
}

// NewOllamaClient returns a Client that calls /api/embed on a remote Ollama
// server. baseURL may include a trailing slash (stripped) and should already
// include any scheme and port (e.g. "https://ollama.example.com:11434").
func NewOllamaClient(baseURL, model string, opts ...OllamaOption) Client {
	c := &ollamaClient{
		baseURL:       strings.TrimRight(baseURL, "/"),
		model:         model,
		maxConcurrent: defaultOllamaMaxConcurrent,
		http:          &http.Client{Timeout: 60 * time.Second},
	}
	for _, opt := range opts {
		opt(c)
	}
	c.Embeddable = rate_limiter.NewEmbeddable[[]float64](c.maxConcurrent)
	return c
}

// EmbedText embeds a single string via the shared batch code path. It goes
// through the rate-limiter pool so concurrent callers (e.g. multiple HTTP
// handlers in a future server) stay bounded without having to coordinate.
func (c *ollamaClient) EmbedText(ctx context.Context, text string) ([]float64, error) {
	return c.SubmitErr(func() ([]float64, error) {
		vecs, err := c.embed(ctx, []string{text})
		if err != nil {
			return nil, err
		}
		return vecs[0], nil
	})
}

// EmbedBatch sends all texts in a single POST to /api/embed. Callers (today
// only the Phase 6 dataloader) are responsible for bounding outer
// concurrency — hence no rate-limiter pool wrap here: that lives on the
// single-item path. Empty input short-circuits to avoid a pointless HTTP call.
func (c *ollamaClient) EmbedBatch(ctx context.Context, texts []string) ([][]float64, error) {
	if len(texts) == 0 {
		return nil, nil
	}
	return c.embed(ctx, texts)
}

func (c *ollamaClient) embed(ctx context.Context, texts []string) ([][]float64, error) {
	body, err := json.Marshal(ollamaRequest{Model: c.model, Input: texts})
	if err != nil {
		return nil, fmt.Errorf("marshal embed request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/embed", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build embed request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ollama embed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama embed: HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(b)))
	}

	var out ollamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decode embed response: %w", err)
	}
	if len(out.Embeddings) != len(texts) {
		return nil, fmt.Errorf("ollama embed: got %d embeddings for %d inputs", len(out.Embeddings), len(texts))
	}
	for i, v := range out.Embeddings {
		if len(v) == 0 {
			return nil, fmt.Errorf("ollama embed: empty embedding at index %d", i)
		}
	}
	return out.Embeddings, nil
}
