package llm

import (
	"testing"

	"github.com/ahmad-ubaidillah/aigo/pkg/types"
)

func TestNewProvider_OpenAI(t *testing.T) {
	t.Parallel()

	cfg := types.ProviderConfig{
		Name:    "openai",
		APIKey:  "sk-test",
		Model:   "gpt-4o",
		Enabled: true,
	}
	client, err := NewProvider(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client for valid OpenAI config")
	}

	cfg.APIKey = ""
	client, err = NewProvider(cfg)
	if err != nil {
		t.Fatalf("unexpected error for empty API key: %v", err)
	}
	if client != nil {
		t.Fatal("expected nil client for empty API key")
	}

	cfg.APIKey = "sk-test"
	cfg.Enabled = false
	client, err = NewProvider(cfg)
	if err != nil {
		t.Fatalf("unexpected error for disabled provider: %v", err)
	}
	if client != nil {
		t.Fatal("expected nil client for disabled provider")
	}
}

func TestNewProvider_OpenAI_CustomBaseURL(t *testing.T) {
	t.Parallel()

	cfg := types.ProviderConfig{
		Name:    "openai",
		APIKey:  "sk-test",
		Model:   "gpt-4o",
		BaseURL: "https://custom.openai.proxy/v1",
		Enabled: true,
	}
	client, err := NewProvider(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}
	oa, ok := client.(*OpenAIClient)
	if !ok {
		t.Fatal("expected *OpenAIClient")
	}
	if oa.baseURL != "https://custom.openai.proxy/v1" {
		t.Errorf("expected baseURL to be custom, got %s", oa.baseURL)
	}
}

func TestNewProvider_Anthropic(t *testing.T) {
	t.Parallel()

	cfg := types.ProviderConfig{
		Name:    "anthropic",
		APIKey:  "sk-ant-test",
		Model:   "claude-sonnet-4-20250514",
		Enabled: true,
	}
	client, err := NewProvider(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client for valid Anthropic config")
	}

	cfg.APIKey = ""
	client, err = NewProvider(cfg)
	if err != nil {
		t.Fatalf("unexpected error for empty API key: %v", err)
	}
	if client != nil {
		t.Fatal("expected nil client for empty API key")
	}
}

func TestNewProvider_OpenRouter(t *testing.T) {
	t.Parallel()

	cfg := types.ProviderConfig{
		Name:    "openrouter",
		APIKey:  "sk-or-test",
		Model:   "openai/gpt-4o",
		Enabled: true,
	}
	client, err := NewProvider(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client for valid OpenRouter config")
	}

	cfg.APIKey = ""
	client, err = NewProvider(cfg)
	if err != nil {
		t.Fatalf("unexpected error for empty API key: %v", err)
	}
	if client != nil {
		t.Fatal("expected nil client for empty API key")
	}
}

func TestNewProvider_OpenRouter_CustomBaseURL(t *testing.T) {
	t.Parallel()

	cfg := types.ProviderConfig{
		Name:    "openrouter",
		APIKey:  "sk-or-test",
		Model:   "openai/gpt-4o",
		BaseURL: "https://custom.openrouter.proxy/v1",
		Enabled: true,
	}
	client, err := NewProvider(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}
	or, ok := client.(*OpenRouterClient)
	if !ok {
		t.Fatal("expected *OpenRouterClient")
	}
	if or.baseURL != "https://custom.openrouter.proxy/v1" {
		t.Errorf("expected baseURL to be custom, got %s", or.baseURL)
	}
}

func TestNewProvider_GLM(t *testing.T) {
	t.Parallel()

	cfg := types.ProviderConfig{
		Name:    "glm",
		APIKey:  "glm-test",
		Model:   "glm-4-plus",
		Enabled: true,
	}
	client, err := NewProvider(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client for valid GLM config")
	}

	cfg.BaseURL = "https://custom.glm.proxy/v4"
	client, err = NewProvider(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client for GLM with custom base URL")
	}

	cfg.APIKey = ""
	client, err = NewProvider(cfg)
	if err != nil {
		t.Fatalf("unexpected error for empty API key: %v", err)
	}
	if client != nil {
		t.Fatal("expected nil client for empty API key")
	}
}

func TestNewProvider_Local(t *testing.T) {
	t.Parallel()

	cfg := types.ProviderConfig{
		Name:    "local",
		Model:   "llama-3",
		BaseURL: "http://localhost:8080/v1",
		Enabled: true,
	}
	client, err := NewProvider(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client for local provider")
	}
}

func TestNewProvider_Custom(t *testing.T) {
	t.Parallel()

	cfg := types.ProviderConfig{
		Name:    "custom",
		APIKey:  "sk-custom",
		BaseURL: "https://my-provider.example.com/v1",
		Model:   "my-model",
		Enabled: true,
	}
	client, err := NewProvider(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client for valid custom config")
	}

	cfg.BaseURL = ""
	_, err = NewProvider(cfg)
	if err == nil {
		t.Fatal("expected error for custom provider without base_url")
	}
}

func TestNewProvider_Unknown(t *testing.T) {
	t.Parallel()

	cfg := types.ProviderConfig{
		Name:    "unknown",
		APIKey:  "sk-test",
		Enabled: true,
	}
	_, err := NewProvider(cfg)
	if err == nil {
		t.Fatal("expected error for unknown provider")
	}
}

func TestNewProviders_SortedByPriority(t *testing.T) {
	t.Parallel()

	cfg := types.LLMConfig{
		Providers: []types.ProviderConfig{
			{Name: "openai", APIKey: "sk-test", Model: "gpt-4o", Enabled: true, Priority: 2},
			{Name: "anthropic", APIKey: "sk-ant", Model: "claude", Enabled: true, Priority: 1},
			{Name: "local", Model: "llama", Enabled: true, Priority: 3},
		},
	}
	providers, err := NewProviders(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(providers) != 3 {
		t.Fatalf("expected 3 providers, got %d", len(providers))
	}
	if providers[0].Name != "anthropic" {
		t.Errorf("expected first provider to be anthropic, got %s", providers[0].Name)
	}
	if providers[1].Name != "openai" {
		t.Errorf("expected second provider to be openai, got %s", providers[1].Name)
	}
	if providers[2].Name != "local" {
		t.Errorf("expected third provider to be local, got %s", providers[2].Name)
	}
}

func TestNewProviders_SkipsDisabledAndUnconfigured(t *testing.T) {
	t.Parallel()

	cfg := types.LLMConfig{
		Providers: []types.ProviderConfig{
			{Name: "openai", APIKey: "sk-test", Model: "gpt-4o", Enabled: true, Priority: 1},
			{Name: "anthropic", APIKey: "", Model: "claude", Enabled: true, Priority: 2},
			{Name: "local", Model: "llama", Enabled: false, Priority: 3},
		},
	}
	providers, err := NewProviders(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(providers) != 1 {
		t.Fatalf("expected 1 provider, got %d", len(providers))
	}
	if providers[0].Name != "openai" {
		t.Errorf("expected first provider to be openai, got %s", providers[0].Name)
	}
}

func TestSortByFallback(t *testing.T) {
	t.Parallel()

	providers := []NamedProvider{
		{Name: "openai", Client: &OpenAIClient{}},
		{Name: "anthropic", Client: &AnthropicClient{}},
		{Name: "local", Client: &LocalClient{}},
	}

	fallback := []string{"local", "openai"}
	result := sortByFallback(providers, fallback)

	if len(result) != 3 {
		t.Fatalf("expected 3 providers, got %d", len(result))
	}
	if result[0].Name != "local" {
		t.Errorf("expected first to be local, got %s", result[0].Name)
	}
	if result[1].Name != "openai" {
		t.Errorf("expected second to be openai, got %s", result[1].Name)
	}
	if result[2].Name != "anthropic" {
		t.Errorf("expected third to be anthropic, got %s", result[2].Name)
	}
}

func TestSortByFallback_PartialMatch(t *testing.T) {
	t.Parallel()

	providers := []NamedProvider{
		{Name: "openai", Client: &OpenAIClient{}},
		{Name: "glm", Client: &GLMClient{}},
	}

	fallback := []string{"openai", "nonexistent"}
	result := sortByFallback(providers, fallback)

	if len(result) != 2 {
		t.Fatalf("expected 1 provider, got %d", len(result))
	}
	if result[0].Name != "openai" {
		t.Errorf("expected first to be openai, got %s", result[0].Name)
	}
	if result[1].Name != "glm" {
		t.Errorf("expected second to be glm, got %s", result[1].Name)
	}
}
