package providers

import (
	"context"
	"testing"
)

func TestProviderManager(t *testing.T) {
	pm := NewProviderManager()

	pm.Register("openai", "gpt-4")
	pm.Register("nous", "hermes-3")
	pm.SetDefault("openai")

	p, err := pm.Get("nous")
	if err != nil {
		t.Fatal(err)
	}
	if p.Name() != "nous" {
		t.Errorf("expected 'nous', got '%s'", p.Name())
	}
	if p.GetModel() != "hermes-3" {
		t.Errorf("expected 'hermes-3', got '%s'", p.GetModel())
	}

	p, err = pm.Get("")
	if err != nil {
		t.Fatal(err)
	}
	if p.Name() != "openai" {
		t.Errorf("expected default 'openai', got '%s'", p.Name())
	}

	names := pm.List()
	if len(names) != 2 {
		t.Errorf("expected 2 providers, got %d", len(names))
	}
}

func TestProviderManagerNotFound(t *testing.T) {
	pm := NewProviderManager()
	_, err := pm.Get("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent provider")
	}
}

func TestNewProvider(t *testing.T) {
	p := NewProvider("test", "model-1")
	if p.Name() != "test" {
		t.Errorf("expected 'test', got '%s'", p.Name())
	}
	if p.GetModel() != "model-1" {
		t.Errorf("expected 'model-1', got '%s'", p.GetModel())
	}
}

func TestProviderRegistry(t *testing.T) {
	providers := ListProviders()
	if len(providers) < 30 {
		t.Errorf("expected at least 30 providers, got %d", len(providers))
	}

	info, ok := GetProvider("openai")
	if !ok {
		t.Error("expected openai provider")
	}
	if info.BaseURL == "" {
		t.Error("expected base URL")
	}
}

func TestMessageStruct(t *testing.T) {
	msg := Message{
		Role:    "user",
		Content: "hello",
	}
	if msg.Role != "user" {
		t.Error("wrong role")
	}
	if msg.Content != "hello" {
		t.Error("wrong content")
	}
}

func TestToolDefStruct(t *testing.T) {
	tool := ToolDef{
		Type: "function",
		Function: ToolFunction{
			Name:        "test_tool",
			Description: "A test tool",
			Parameters: map[string]interface{}{
				"type": "object",
			},
		},
	}
	if tool.Function.Name != "test_tool" {
		t.Error("wrong tool name")
	}
}

func TestProviderChatBadURL(t *testing.T) {
	p := NewProvider("bad", "model")
	_, err := p.Chat(context.Background(), []Message{{Role: "user", Content: "hi"}}, nil)
	if err == nil {
		t.Error("expected error for bad URL")
	}
}