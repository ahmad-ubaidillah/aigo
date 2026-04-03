package llm

import (
	"context"
	"testing"
)

type mockClient struct {
	completeCalled bool
	chatCalled     bool
}

func (m *mockClient) Complete(ctx context.Context, prompt string) (*Response, error) {
	m.completeCalled = true
	return &Response{Content: "mock response"}, nil
}

func (m *mockClient) Chat(ctx context.Context, messages []Message) (*Response, error) {
	m.chatCalled = true
	return &Response{Content: "mock chat"}, nil
}

func (m *mockClient) CompleteWithSystem(ctx context.Context, systemPrompt, userPrompt string) (*Response, error) {
	return &Response{Content: systemPrompt + ": " + userPrompt}, nil
}

func TestMockClient_Complete(t *testing.T) {
	t.Parallel()

	c := &mockClient{}
	resp, err := c.Complete(context.Background(), "test")
	if err != nil {
		t.Fatal(err)
	}
	if !c.completeCalled {
		t.Error("expected complete called")
	}
	if resp.Content != "mock response" {
		t.Errorf("expected 'mock response', got %s", resp.Content)
	}
}

func TestMockClient_Chat(t *testing.T) {
	t.Parallel()

	c := &mockClient{}
	resp, err := c.Chat(context.Background(), []Message{{Role: "user", Content: "hi"}})
	if err != nil {
		t.Fatal(err)
	}
	if !c.chatCalled {
		t.Error("expected chat called")
	}
	if resp.Content != "mock chat" {
		t.Errorf("expected 'mock chat', got %s", resp.Content)
	}
}

func TestMockClient_CompleteWithSystem(t *testing.T) {
	t.Parallel()

	c := &mockClient{}
	resp, err := c.CompleteWithSystem(context.Background(), "sys", "user")
	if err != nil {
		t.Fatal(err)
	}
	if resp.Content != "sys: user" {
		t.Errorf("expected 'sys: user', got %s", resp.Content)
	}
}

func TestChatOptions(t *testing.T) {
	t.Parallel()

	opts := ChatOptions{
		Model:         "gpt-4",
		Temperature:   0.7,
		MaxTokens:     1000,
		StopSequences: []string{"\n"},
	}
	if opts.Model != "gpt-4" {
		t.Errorf("expected gpt-4, got %s", opts.Model)
	}
	if opts.Temperature != 0.7 {
		t.Errorf("expected 0.7, got %f", opts.Temperature)
	}
}

func TestChatResponse(t *testing.T) {
	t.Parallel()

	resp := ChatResponse{
		Content: "hello",
		Usage: TokenUsage{
			PromptTokens:     10,
			CompletionTokens: 5,
			TotalTokens:      15,
		},
	}
	if resp.Usage.TotalTokens != 15 {
		t.Errorf("expected 15, got %d", resp.Usage.TotalTokens)
	}
}

func TestToolCall(t *testing.T) {
	t.Parallel()

	tc := ToolCall{
		ID:        "call-1",
		Name:      "bash",
		Arguments: `{"command": "ls"}`,
	}
	if tc.Name != "bash" {
		t.Errorf("expected bash, got %s", tc.Name)
	}
}

func TestToolDef(t *testing.T) {
	t.Parallel()

	td := ToolDef{
		Name:        "bash",
		Description: "Execute a command",
		Parameters:  map[string]any{"type": "object"},
	}
	if td.Name != "bash" {
		t.Errorf("expected bash, got %s", td.Name)
	}
}
