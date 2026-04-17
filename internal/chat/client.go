// Package chat is the generation tier — it wraps Ollama's /api/chat endpoint
// for grounded-answer synthesis on top of retrieval. The underlying HTTP path
// is identical to internal/embedding's, just a different endpoint + model, so
// both tiers share the same OLLAMA_URL but use their own model env var.
//
// Two abstractions live here:
//
//   - Client: stateless POST /api/chat wrapper. One shot, full messages in,
//     assistant reply out. Reuse across conversations is fine.
//   - Session: stateful caller-local history wrapper around a Client. Use it
//     when you want multi-turn context without hand-rolling the slice
//     append dance on every turn.
package chat

import "context"

// Role values follow Ollama's /api/chat wire format verbatim.
const (
	RoleSystem    = "system"
	RoleUser      = "user"
	RoleAssistant = "assistant"
)

// Message is one entry in a chat. Content is plain text; tool/image fields
// are intentionally not modelled until we need them.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Client is the low-level, stateless chat caller. Implementations MUST NOT
// mutate the input `messages` slice — Session relies on that guarantee.
//
//go:generate mockgen -source=client.go -destination=mocks/mock_chat_client.go -package=mocks
type Client interface {
	Complete(ctx context.Context, messages []Message) (Message, error)
}
