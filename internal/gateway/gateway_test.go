package gateway

import (
	"context"
	"testing"
)

func TestGatewayRegisterAndChannels(t *testing.T) {
	// Create a minimal mock agent
	gw := &Gateway{
		channels: make(map[string]Channel),
	}

	mock := &mockChannel{name: "test"}
	gw.Register(mock)

	if len(gw.channels) != 1 {
		t.Errorf("expected 1 channel, got %d", len(gw.channels))
	}
}

func TestGatewaySendMessage(t *testing.T) {
	gw := &Gateway{
		channels: make(map[string]Channel),
	}

	mock := &mockChannel{name: "test"}
	gw.channels["test"] = mock

	err := gw.sendResponse(Message{Channel: "test", ChatID: "123", Text: "hello"}, "response")
	if err != nil {
		t.Fatal(err)
	}
	if mock.lastMessage != "response" {
		t.Errorf("expected 'response', got '%s'", mock.lastMessage)
	}
}

func TestGatewaySendUnknownChannel(t *testing.T) {
	gw := &Gateway{
		channels: make(map[string]Channel),
	}

	err := gw.sendResponse(Message{Channel: "unknown", ChatID: "123"}, "test")
	if err == nil {
		t.Error("expected error for unknown channel")
	}
}

func TestTruncate(t *testing.T) {
	if truncate("hello", 10) != "hello" {
		t.Error("short string should not be truncated")
	}
	if truncate("hello world this is long", 10) != "hello worl..." {
		t.Errorf("expected truncation, got '%s'", truncate("hello world this is long", 10))
	}
}

// mockChannel is a test channel implementation.
type mockChannel struct {
	name        string
	lastMessage string
}

func (m *mockChannel) Name() string { return m.name }
func (m *mockChannel) Start(ctx context.Context, onMessage func(Message) error) error {
	<-ctx.Done()
	return nil
}
func (m *mockChannel) Send(chatID string, text string) error {
	m.lastMessage = text
	return nil
}
func (m *mockChannel) Stop() error { return nil }
