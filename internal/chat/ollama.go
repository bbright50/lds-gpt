package chat

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// defaultChatTimeout is long because large generative models (gemma3:27b,
// llama3:70b, etc.) can take many seconds per turn. Per-turn streaming is
// not yet modelled; until it is, the whole response has to arrive inside
// one HTTP read.
const defaultChatTimeout = 5 * time.Minute

// OllamaOption tweaks the client at construction time.
type OllamaOption func(*ollamaClient)

// WithHTTPTimeout overrides the per-request timeout (default: 5 min).
func WithHTTPTimeout(d time.Duration) OllamaOption {
	return func(c *ollamaClient) { c.http.Timeout = d }
}

// WithOptions layers in Ollama's per-model `options` map (temperature,
// num_ctx, top_p, etc.). The keys pass through to the server verbatim —
// see https://github.com/ollama/ollama/blob/main/docs/modelfile.md for the
// authoritative list.
func WithOptions(opts map[string]any) OllamaOption {
	return func(c *ollamaClient) {
		if c.options == nil {
			c.options = map[string]any{}
		}
		for k, v := range opts {
			c.options[k] = v
		}
	}
}

type ollamaChatRequest struct {
	Model    string         `json:"model"`
	Messages []Message      `json:"messages"`
	Stream   bool           `json:"stream"`
	Options  map[string]any `json:"options,omitempty"`
}

type ollamaChatResponse struct {
	Model   string  `json:"model"`
	Message Message `json:"message"`
	Done    bool    `json:"done"`
}

type ollamaClient struct {
	baseURL string
	model   string
	http    *http.Client
	options map[string]any
}

// NewOllamaClient returns a Client that calls /api/chat on a remote Ollama
// server. baseURL may include a trailing slash (stripped). Typical use:
//
//	cli := chat.NewOllamaClient(cfg.OllamaURL, cfg.OllamaChatModel)
//	sess := chat.NewSession(cli, "You are a helpful assistant.")
//	reply, _ := sess.Send(ctx, "What's the capital of France?")
func NewOllamaClient(baseURL, model string, opts ...OllamaOption) Client {
	c := &ollamaClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		model:   model,
		http:    &http.Client{Timeout: defaultChatTimeout},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func (c *ollamaClient) Complete(ctx context.Context, messages []Message) (Message, error) {
	if len(messages) == 0 {
		return Message{}, fmt.Errorf("chat: messages must not be empty")
	}
	payload := ollamaChatRequest{
		Model:    c.model,
		Messages: messages,
		Stream:   false,
		Options:  c.options,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return Message{}, fmt.Errorf("marshal chat request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/chat", bytes.NewReader(body))
	if err != nil {
		return Message{}, fmt.Errorf("build chat request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return Message{}, fmt.Errorf("ollama chat: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return Message{}, fmt.Errorf("ollama chat: HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(b)))
	}

	var out ollamaChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return Message{}, fmt.Errorf("decode chat response: %w", err)
	}
	if out.Message.Content == "" {
		return Message{}, fmt.Errorf("ollama chat: empty message in response")
	}
	if out.Message.Role == "" {
		out.Message.Role = RoleAssistant
	}
	return out.Message, nil
}
