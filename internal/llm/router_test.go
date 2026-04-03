package llm

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/ahmad-ubaidillah/aigo/pkg/types"
)

type mockProvider struct {
	shouldFail    bool
	failCount     int
	currentCalls  int
	completeResp  string
	chatResp      string
	systemResp    string
}

func (m *mockProvider) Complete(ctx context.Context, prompt string) (*Response, error) {
	m.currentCalls++
	if m.shouldFail && m.currentCalls <= m.failCount {
		return nil, fmt.Errorf("provider error")
	}
	return &Response{Content: m.completeResp}, nil
}

func (m *mockProvider) Chat(ctx context.Context, messages []Message) (*Response, error) {
	m.currentCalls++
	if m.shouldFail && m.currentCalls <= m.failCount {
		return nil, fmt.Errorf("provider error")
	}
	return &Response{Content: m.chatResp}, nil
}

func (m *mockProvider) CompleteWithSystem(ctx context.Context, systemPrompt, userPrompt string) (*Response, error) {
	m.currentCalls++
	if m.shouldFail && m.currentCalls <= m.failCount {
		return nil, fmt.Errorf("provider error")
	}
	return &Response{Content: m.systemResp}, nil
}

func TestNewLLMRouter_InitializesHealth(t *testing.T) {
	t.Parallel()

	providers := []NamedProvider{
		{Name: "openai", Client: &mockProvider{chatResp: "ok"}},
		{Name: "anthropic", Client: &mockProvider{chatResp: "ok"}},
	}
	r := NewLLMRouter(providers, nil)

	health := r.Health()
	if len(health) != 2 {
		t.Fatalf("expected 2 health entries, got %d", len(health))
	}
	if !health["openai"].IsHealthy {
		t.Error("expected openai to be healthy initially")
	}
	if !health["anthropic"].IsHealthy {
		t.Error("expected anthropic to be healthy initially")
	}
	if health["openai"].ConsecutiveFails != 0 {
		t.Errorf("expected 0 consecutive fails, got %d", health["openai"].ConsecutiveFails)
	}
}

func TestLLMRouter_Chat_SingleProvider(t *testing.T) {
	t.Parallel()

	mp := &mockProvider{chatResp: "hello from openai"}
	providers := []NamedProvider{{Name: "openai", Client: mp}}
	r := NewLLMRouter(providers, nil)

	resp, err := r.Chat(context.Background(), []Message{{Role: "user", Content: "hi"}}, ChatOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Content != "hello from openai" {
		t.Errorf("expected 'hello from openai', got %s", resp.Content)
	}
}

func TestLLMRouter_Chat_Fallback(t *testing.T) {
	t.Parallel()

	failProvider := &mockProvider{shouldFail: true, failCount: 10, chatResp: "should not reach"}
	okProvider := &mockProvider{chatResp: "hello from anthropic"}

	providers := []NamedProvider{
		{Name: "openai", Client: failProvider},
		{Name: "anthropic", Client: okProvider},
	}
	r := NewLLMRouter(providers, nil)

	resp, err := r.Chat(context.Background(), []Message{{Role: "user", Content: "hi"}}, ChatOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Content != "hello from anthropic" {
		t.Errorf("expected fallback to anthropic, got %s", resp.Content)
	}
}

func TestLLMRouter_Chat_AllFail(t *testing.T) {
	t.Parallel()

	providers := []NamedProvider{
		{Name: "openai", Client: &mockProvider{shouldFail: true, failCount: 10}},
		{Name: "anthropic", Client: &mockProvider{shouldFail: true, failCount: 10}},
	}
	r := NewLLMRouter(providers, nil)

	_, err := r.Chat(context.Background(), []Message{{Role: "user", Content: "hi"}}, ChatOptions{})
	if err == nil {
		t.Fatal("expected error when all providers fail")
	}
}

func TestLLMRouter_MarkSuccess_ResetsFailures(t *testing.T) {
	t.Parallel()

	mp := &mockProvider{chatResp: "ok"}
	providers := []NamedProvider{{Name: "openai", Client: mp}}
	r := NewLLMRouter(providers, nil)

	r.markFailure("openai", fmt.Errorf("error 1"))
	r.markFailure("openai", fmt.Errorf("error 2"))
	r.markSuccess("openai")

	health := r.Health()
	if health["openai"].ConsecutiveFails != 0 {
		t.Errorf("expected 0 consecutive fails after success, got %d", health["openai"].ConsecutiveFails)
	}
	if !health["openai"].IsHealthy {
		t.Error("expected healthy after success")
	}
}

func TestLLMRouter_MarkFailure_UnhealthyAfter3(t *testing.T) {
	t.Parallel()

	mp := &mockProvider{chatResp: "ok"}
	providers := []NamedProvider{{Name: "openai", Client: mp}}
	r := NewLLMRouter(providers, nil)

	r.markFailure("openai", fmt.Errorf("error 1"))
	r.markFailure("openai", fmt.Errorf("error 2"))

	health := r.Health()
	if !health["openai"].IsHealthy {
		t.Error("expected healthy after 2 failures (not yet 3)")
	}

	r.markFailure("openai", fmt.Errorf("error 3"))
	health = r.Health()
	if health["openai"].IsHealthy {
		t.Error("expected unhealthy after 3 failures")
	}
	if health["openai"].ConsecutiveFails != 3 {
		t.Errorf("expected 3 consecutive fails, got %d", health["openai"].ConsecutiveFails)
	}
}

func TestLLMRouter_OrderedProviders_SkipsUnhealthy(t *testing.T) {
	t.Parallel()

	mp1 := &mockProvider{chatResp: "ok1"}
	mp2 := &mockProvider{chatResp: "ok2"}
	providers := []NamedProvider{
		{Name: "openai", Client: mp1},
		{Name: "anthropic", Client: mp2},
	}
	r := NewLLMRouter(providers, nil)

	// Mark openai as unhealthy
	r.markFailure("openai", fmt.Errorf("e1"))
	r.markFailure("openai", fmt.Errorf("e2"))
	r.markFailure("openai", fmt.Errorf("e3"))

	ordered := r.orderedProviders()
	if len(ordered) != 1 {
		t.Fatalf("expected 1 provider (skipping unhealthy), got %d", len(ordered))
	}
	if ordered[0].Name != "anthropic" {
		t.Errorf("expected anthropic, got %s", ordered[0].Name)
	}
}

func TestLLMRouter_OrderedProviders_LastResort(t *testing.T) {
	t.Parallel()

	mp1 := &mockProvider{chatResp: "ok1"}
	mp2 := &mockProvider{chatResp: "ok2"}
	mp3 := &mockProvider{chatResp: "ok3"}
	providers := []NamedProvider{
		{Name: "openai", Client: mp1},
		{Name: "anthropic", Client: mp2},
		{Name: "glm", Client: mp3},
	}
	r := NewLLMRouter(providers, []string{"glm"})

	// Mark openai and anthropic as unhealthy, glm stays healthy
	r.markFailure("openai", fmt.Errorf("e1"))
	r.markFailure("openai", fmt.Errorf("e2"))
	r.markFailure("openai", fmt.Errorf("e3"))
	r.markFailure("anthropic", fmt.Errorf("e1"))
	r.markFailure("anthropic", fmt.Errorf("e2"))
	r.markFailure("anthropic", fmt.Errorf("e3"))

	ordered := r.orderedProviders()
	// glm is healthy -> added in first loop
	// openai, anthropic are unhealthy -> skipped in first loop
	// Fallback: glm already seen
	// Last resort: only adds providers NOT seen. But all were marked seen in first loop.
	// So only healthy providers are returned.
	if len(ordered) != 1 {
		t.Fatalf("expected 1 provider (only healthy glm), got %d", len(ordered))
	}
	if ordered[0].Name != "glm" {
		t.Errorf("expected glm, got %s", ordered[0].Name)
	}
}

func TestLLMRouter_Complete(t *testing.T) {
	t.Parallel()

	mp := &mockProvider{chatResp: "completed"}
	providers := []NamedProvider{{Name: "openai", Client: mp}}
	r := NewLLMRouter(providers, nil)

	resp, err := r.Complete(context.Background(), "test prompt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Content != "completed" {
		t.Errorf("expected 'completed', got %s", resp.Content)
	}
}

func TestLLMRouter_CompleteWithSystem(t *testing.T) {
	t.Parallel()

	mp := &mockProvider{chatResp: "system response"}
	providers := []NamedProvider{{Name: "openai", Client: mp}}
	r := NewLLMRouter(providers, nil)

	resp, err := r.CompleteWithSystem(context.Background(), "system", "user")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Content != "system response" {
		t.Errorf("expected 'system response', got %s", resp.Content)
	}
}

func TestLLMRouter_ContextCancellation(t *testing.T) {
	t.Parallel()

	mp := &mockProvider{shouldFail: true, failCount: 10}
	providers := []NamedProvider{{Name: "openai", Client: mp}}
	r := NewLLMRouter(providers, nil)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := r.Chat(ctx, []Message{{Role: "user", Content: "hi"}}, ChatOptions{})
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}

func TestNewRouterFromConfig_ValidConfig(t *testing.T) {
	t.Parallel()

	cfg := types.LLMConfig{
		Providers: []types.ProviderConfig{
			{Name: "openai", APIKey: "sk-test", Model: "gpt-4o", Enabled: true, Priority: 1},
		},
		Fallback: []string{"openai"},
	}
	r, err := NewRouterFromConfig(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r == nil {
		t.Fatal("expected non-nil router")
	}
}

func TestNewRouterFromConfig_NoProviders(t *testing.T) {
	t.Parallel()

	cfg := types.LLMConfig{
		Providers: []types.ProviderConfig{},
	}
	_, err := NewRouterFromConfig(cfg)
	if err == nil {
		t.Fatal("expected error for empty providers")
	}
}

func TestLLMRouter_Health_ReturnsCopy(t *testing.T) {
	t.Parallel()

	mp := &mockProvider{chatResp: "ok"}
	providers := []NamedProvider{{Name: "openai", Client: mp}}
	r := NewLLMRouter(providers, nil)

	h1 := r.Health()
	h1["openai"].IsHealthy = false

	h2 := r.Health()
	if h2["openai"].IsHealthy {
		// Good — Health() returns a copy
	} else {
		t.Error("expected Health() to return a copy, original should still be healthy")
	}
}

func TestLLMRouter_Timeout_Default(t *testing.T) {
	t.Parallel()

	mp := &mockProvider{chatResp: "ok"}
	providers := []NamedProvider{
		{Name: "openai", Client: mp, Timeout: 0},
	}
	r := NewLLMRouter(providers, nil)

	resp, err := r.Chat(context.Background(), []Message{{Role: "user", Content: "hi"}}, ChatOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Content != "ok" {
		t.Errorf("expected 'ok', got %s", resp.Content)
	}
}

func TestLLMRouter_Timeout_Custom(t *testing.T) {
	t.Parallel()

	mp := &mockProvider{chatResp: "ok"}
	providers := []NamedProvider{
		{Name: "openai", Client: mp, Timeout: 5 * time.Second},
	}
	r := NewLLMRouter(providers, nil)

	resp, err := r.Chat(context.Background(), []Message{{Role: "user", Content: "hi"}}, ChatOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Content != "ok" {
		t.Errorf("expected 'ok', got %s", resp.Content)
	}
}

func TestLLMRouter_CheckHealth(t *testing.T) {
	t.Parallel()

	mp := &mockProvider{chatResp: "pong"}
	providers := []NamedProvider{{Name: "openai", Client: mp}}
	r := NewLLMRouter(providers, nil)

	results := r.CheckHealth(context.Background())
	if len(results) != 1 {
		t.Fatalf("expected 1 health result, got %d", len(results))
	}
	if !results["openai"] {
		t.Error("expected openai to be healthy")
	}
}
