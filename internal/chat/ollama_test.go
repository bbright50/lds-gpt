package chat

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestOllamaComplete_WireFormat pins the exact POST /api/chat payload:
// model, messages (array of role+content), stream=false. Guards against a
// regression where a refactor e.g. drops the messages array, sends a
// combined prompt string, or flips stream to true (which would change the
// response decoding path).
func TestOllamaComplete_WireFormat(t *testing.T) {
	var got ollamaChatRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/chat" {
			t.Errorf("path = %q, want /api/chat", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(body, &got); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		_ = json.NewEncoder(w).Encode(ollamaChatResponse{
			Model:   got.Model,
			Message: Message{Role: RoleAssistant, Content: "pong"},
			Done:    true,
		})
	}))
	defer server.Close()

	c := NewOllamaClient(server.URL, "test-model")
	messages := []Message{
		{Role: RoleSystem, Content: "be concise"},
		{Role: RoleUser, Content: "ping"},
	}
	reply, err := c.Complete(context.Background(), messages)
	if err != nil {
		t.Fatalf("Complete: %v", err)
	}

	if got.Model != "test-model" {
		t.Errorf("model = %q, want test-model", got.Model)
	}
	if got.Stream {
		t.Errorf("stream = true, want false")
	}
	if len(got.Messages) != 2 {
		t.Fatalf("server saw %d messages, want 2", len(got.Messages))
	}
	if got.Messages[0].Role != RoleSystem || got.Messages[0].Content != "be concise" {
		t.Errorf("message[0] = %+v, want system/be concise", got.Messages[0])
	}
	if got.Messages[1].Role != RoleUser || got.Messages[1].Content != "ping" {
		t.Errorf("message[1] = %+v, want user/ping", got.Messages[1])
	}
	if reply.Content != "pong" || reply.Role != RoleAssistant {
		t.Errorf("reply = %+v, want assistant/pong", reply)
	}
}

// TestSessionSend_GrowsHistoryAcrossTurns is the core of the stateful API:
// each Send must carry every prior message to the server so the model has
// the context it needs for multi-turn conversation.
func TestSessionSend_GrowsHistoryAcrossTurns(t *testing.T) {
	var turns [][]Message
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req ollamaChatRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		turns = append(turns, append([]Message(nil), req.Messages...))
		_ = json.NewEncoder(w).Encode(ollamaChatResponse{
			Message: Message{Role: RoleAssistant, Content: "reply-" + req.Messages[len(req.Messages)-1].Content},
		})
	}))
	defer server.Close()

	sess := NewSession(NewOllamaClient(server.URL, "m"), "you are helpful")

	if _, err := sess.Send(context.Background(), "hello"); err != nil {
		t.Fatalf("Send 1: %v", err)
	}
	if _, err := sess.Send(context.Background(), "how are you"); err != nil {
		t.Fatalf("Send 2: %v", err)
	}
	if _, err := sess.Send(context.Background(), "bye"); err != nil {
		t.Fatalf("Send 3: %v", err)
	}

	// Turn 1 carries: system + user(hello). Turn 2: system + user + assistant +
	// user(how are you). Turn 3: previous + assistant + user(bye).
	wantLens := []int{2, 4, 6}
	for i, got := range turns {
		if len(got) != wantLens[i] {
			t.Errorf("turn %d sent %d messages, want %d (full: %+v)", i+1, len(got), wantLens[i], got)
		}
	}
	if turns[1][0].Role != RoleSystem || turns[1][0].Content != "you are helpful" {
		t.Errorf("system prompt missing on turn 2: %+v", turns[1][0])
	}
	if turns[1][2].Role != RoleAssistant || turns[1][2].Content != "reply-hello" {
		t.Errorf("turn-2 assistant echo wrong: %+v", turns[1][2])
	}

	// History() should include every exchange, in order.
	h := sess.History()
	if len(h) != 7 { // system + 3 × (user + assistant)
		t.Fatalf("history len = %d, want 7", len(h))
	}
}

// TestSessionReset_KeepsSystemPrompt makes sure Reset wipes the turns but
// preserves the system message — otherwise a long-running app would lose
// its guardrails after every topic change.
func TestSessionReset_KeepsSystemPrompt(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(ollamaChatResponse{
			Message: Message{Role: RoleAssistant, Content: "ack"},
		})
	}))
	defer server.Close()

	sess := NewSession(NewOllamaClient(server.URL, "m"), "stay on task")
	_, _ = sess.Send(context.Background(), "first")
	_, _ = sess.Send(context.Background(), "second")
	sess.Reset()

	h := sess.History()
	if len(h) != 1 || h[0].Role != RoleSystem {
		t.Errorf("after Reset history = %+v, want [system only]", h)
	}
}

// TestSessionSend_EmptyInputRejected ensures we don't send a no-op to the
// model (and don't pollute the history slice with an empty user message).
func TestSessionSend_EmptyInputRejected(t *testing.T) {
	sess := NewSession(NewOllamaClient("http://unused", "m"), "")
	_, err := sess.Send(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty user message")
	}
	if len(sess.History()) != 0 {
		t.Errorf("empty message leaked into history: %+v", sess.History())
	}
}
