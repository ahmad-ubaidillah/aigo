package llm

import (
	"testing"
)

func TestMessage(t *testing.T) {
	t.Parallel()
	msg := Message{Role: "user", Content: "Hello"}
	if msg.Role != "user" {
		t.Errorf("expected user, got %s", msg.Role)
	}
	if msg.Content != "Hello" {
		t.Errorf("expected Hello, got %s", msg.Content)
	}
}

func TestResponse(t *testing.T) {
	t.Parallel()
	resp := Response{
		Content: "test response",
		Usage:   Usage{PromptTokens: 10, CompletionTokens: 5, TotalTokens: 15},
	}
	if resp.Content != "test response" {
		t.Errorf("expected test response, got %s", resp.Content)
	}
	if resp.Usage.TotalTokens != 15 {
		t.Errorf("expected 15 tokens, got %d", resp.Usage.TotalTokens)
	}
}
