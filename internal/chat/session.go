package chat

import (
	"context"
	"errors"
	"sync"
)

// Session is a stateful multi-turn conversation on top of a Client. Each
// call to Send appends a user message, sends the full history, and appends
// the assistant reply — so the next Send carries everything before it as
// context, which is how Ollama (and every other chat API) expects history
// to be represented.
//
// Sessions are safe for concurrent Send calls but serialise them — Ollama
// keeps no server-side state, so two in-flight Sends on the same Session
// would both send overlapping histories and mix replies. If you need
// true concurrency, create one Session per logical conversation.
type Session struct {
	client   Client
	mu       sync.Mutex
	messages []Message
}

// NewSession returns a new stateful conversation. `systemPrompt` is
// optional; pass "" to start without one. The system message is locked in
// as the first entry and is kept even across Reset().
func NewSession(client Client, systemPrompt string) *Session {
	s := &Session{client: client}
	if systemPrompt != "" {
		s.messages = append(s.messages, Message{Role: RoleSystem, Content: systemPrompt})
	}
	return s
}

// Send appends `userMessage` to the history, asks the Client to complete,
// appends the assistant reply, and returns its text. If the Client errors,
// the user message is left in the history (so a retry can resend). To drop
// it, call PopLast() before retrying.
func (s *Session) Send(ctx context.Context, userMessage string) (string, error) {
	if userMessage == "" {
		return "", errors.New("chat: user message must not be empty")
	}

	s.mu.Lock()
	s.messages = append(s.messages, Message{Role: RoleUser, Content: userMessage})
	// Pass a copy so the Client implementation can't corrupt our history
	// via aliasing shenanigans.
	snapshot := append([]Message(nil), s.messages...)
	s.mu.Unlock()

	reply, err := s.client.Complete(ctx, snapshot)
	if err != nil {
		return "", err
	}

	s.mu.Lock()
	s.messages = append(s.messages, reply)
	s.mu.Unlock()
	return reply.Content, nil
}

// History returns a copy of the current conversation (system + user +
// assistant turns in order). Safe to mutate — callers do not share memory
// with the Session.
func (s *Session) History() []Message {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]Message(nil), s.messages...)
}

// Reset clears the conversation back to the system prompt (if any). Used
// when the caller wants a fresh turn of topic without rebuilding the
// Session from scratch.
func (s *Session) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.messages) > 0 && s.messages[0].Role == RoleSystem {
		s.messages = s.messages[:1]
	} else {
		s.messages = s.messages[:0]
	}
}

// PopLast drops the most recent message — useful for recovering from a
// failed Send where the user message was recorded but never replied to.
func (s *Session) PopLast() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.messages) > 0 {
		s.messages = s.messages[:len(s.messages)-1]
	}
}
