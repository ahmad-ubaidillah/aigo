package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type Provider interface {
	Name() string
	Chat(ctx context.Context, messages []Message, tools []ToolDef) (*Response, error)
	GetModel() string
}

type Message struct {
	Role         string     `json:"role"`
	Content      string     `json:"content,omitempty"`
	ToolCalls    []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID   string     `json:"tool_call_id,omitempty"`
	CacheControl *CacheControl `json:"cache_control,omitempty"`
}

type CacheControl struct {
	Type string `json:"type"`
}

type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function FunctionCall `json:"function"`
}

type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type ToolDef struct {
	Type     string       `json:"type"`
	Function ToolFunction `json:"function"`
}

type ToolFunction struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Parameters  interface{} `json:"parameters"`
}

type Response struct {
	Content      string     `json:"content"`
	ToolCalls    []ToolCall `json:"tool_calls,omitempty"`
	FinishReason string     `json:"finish_reason"`
	Usage       Usage      `json:"usage"`
	Model       string     `json:"model"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type ProviderInfo struct {
	BaseURL   string
	APIKeyEnv string
	APIMode  string
}

var ProviderRegistry = map[string]ProviderInfo{
	"nous":          {BaseURL: "https://inference-api.nousresearch.com/v1", APIKeyEnv: "NOUS_API_KEY", APIMode: "openai"},
	"openai":        {BaseURL: "https://api.openai.com/v1", APIKeyEnv: "OPENAI_API_KEY", APIMode: "openai"},
	"anthropic":     {BaseURL: "https://api.anthropic.com", APIKeyEnv: "ANTHROPIC_API_KEY", APIMode: "anthropic"},
	"google":        {BaseURL: "https://generativelanguage.googleapis.com/v1beta", APIKeyEnv: "GOOGLE_API_KEY", APIMode: "google"},
	"gemini":       {BaseURL: "https://generativelanguage.googleapis.com/v1beta", APIKeyEnv: "GEMINI_API_KEY", APIMode: "google"},
	"deepseek":      {BaseURL: "https://api.deepseek.com/v1", APIKeyEnv: "DEEPSEEK_API_KEY", APIMode: "openai"},
	"xai":          {BaseURL: "https://api.x.ai/v1", APIKeyEnv: "XAI_API_KEY", APIMode: "openai"},
	"nvidia":       {BaseURL: "https://integrate.api.nvidia.com/v1", APIKeyEnv: "NVIDIA_API_KEY", APIMode: "openai"},
	"kimi":         {BaseURL: "https://api.moonshot.ai/v1", APIKeyEnv: "KIMI_API_KEY", APIMode: "openai"},
	"kimi-coding":   {BaseURL: "https://api.moonshot.ai/v1", APIKeyEnv: "KIMI_API_KEY", APIMode: "openai"},
	"kimi-coding-cn": {BaseURL: "https://api.moonshot.cn/v1", APIKeyEnv: "KIMI_CN_API_KEY", APIMode: "openai"},
	"glm":          {BaseURL: "https://api.z.ai/api/paas/v4", APIKeyEnv: "GLM_API_KEY", APIMode: "openai"},
	"z":            {BaseURL: "https://api.z.ai/api/paas/v4", APIKeyEnv: "ZAI_API_KEY", APIMode: "openai"},
	"minimax":      {BaseURL: "https://api.minimax.io/anthropic", APIKeyEnv: "MINIMAX_API_KEY", APIMode: "anthropic"},
	"minimax-cn":   {BaseURL: "https://api.minimaxi.com/anthropic", APIKeyEnv: "MINIMAX_CN_API_KEY", APIMode: "anthropic"},
	"qwen":         {BaseURL: "https://portal.qwen.ai/v1", APIKeyEnv: "QWEN_API_KEY", APIMode: "openai"},
	"qwen-oauth":   {BaseURL: "https://portal.qwen.ai/v1", APIKeyEnv: "QWEN_OAUTH_TOKEN", APIMode: "oauth"},
	"alibaba":      {BaseURL: "https://dashscope-intl.aliyuncs.com/compatible-mode/v1", APIKeyEnv: "DASHSCOPE_API_KEY", APIMode: "openai"},
	"xiaomi":       {BaseURL: "https://api.xiaomimimo.com/v1", APIKeyEnv: "XIAOMI_API_KEY", APIMode: "openai"},
	"mimo":         {BaseURL: "https://api.xiaomimimo.com/v1", APIKeyEnv: "XIAOMI_API_KEY", APIMode: "openai"},
	"openrouter":   {BaseURL: "https://openrouter.ai/api/v1", APIKeyEnv: "OPENROUTER_API_KEY", APIMode: "openai"},
	"huggingface":  {BaseURL: "https://router.huggingface.co/v1", APIKeyEnv: "HF_TOKEN", APIMode: "openai"},
	"ollama":       {BaseURL: "http://localhost:11434/v1", APIKeyEnv: "OLLAMA_API_KEY", APIMode: "openai"},
	"ollama-cloud": {BaseURL: "https://ollama.com/v1", APIKeyEnv: "OLLAMA_API_KEY", APIMode: "openai"},
	"local":        {BaseURL: "http://localhost:11434/v1", APIKeyEnv: "", APIMode: "openai"},
	"copilot":      {BaseURL: "https://api.githubcopilot.com", APIKeyEnv: "GITHUB_TOKEN", APIMode: "openai"},
	"azure":        {BaseURL: "", APIKeyEnv: "AZURE_OPENAI_API_KEY", APIMode: "azure"},
	"aws":          {BaseURL: "https://bedrock-runtime.us-east-1.amazonaws.com", APIKeyEnv: "AWS_ACCESS_KEY_ID", APIMode: "aws"},
	"bedrock":      {BaseURL: "https://bedrock-runtime.us-east-1.amazonaws.com", APIKeyEnv: "AWS_ACCESS_KEY_ID", APIMode: "aws"},
	"arcee":       {BaseURL: "https://api.arcee.ai/api/v1", APIKeyEnv: "ARCEEAI_API_KEY", APIMode: "openai"},
	"ai-gateway":  {BaseURL: "https://ai-gateway.vercel.sh/v1", APIKeyEnv: "AI_GATEWAY_API_KEY", APIMode: "openai"},
	"opencode-zen": {BaseURL: "https://opencode.ai/zen/v1", APIKeyEnv: "OPENCODE_ZEN_API_KEY", APIMode: "openai"},
	"opencode-go":  {BaseURL: "https://opencode.ai/zen/go/v1", APIKeyEnv: "OPENCODE_GO_API_KEY", APIMode: "openai"},
	"kilocode":    {BaseURL: "https://api.kilo.ai/api/gateway", APIKeyEnv: "KILOCODE_API_KEY", APIMode: "openai"},
	"groq":        {BaseURL: "https://api.groq.com/openai/v1", APIKeyEnv: "GROQ_API_KEY", APIMode: "openai"},
	"perplexity":   {BaseURL: "https://api.perplexity.ai", APIKeyEnv: "PERPLEXITY_API_KEY", APIMode: "openai"},
}

func GetProvider(name string) (ProviderInfo, bool) {
	info, ok := ProviderRegistry[name]
	return info, ok
}

func ListProviders() []string {
	names := make([]string, 0, len(ProviderRegistry))
	for name := range ProviderRegistry {
		names = append(names, name)
	}
	return names
}

type OpenAIProvider struct {
	name    string
	baseURL string
	apiKey string
	model  string
	client *http.Client
	apiMode string
}

func NewProvider(name, model string) *OpenAIProvider {
	info, ok := ProviderRegistry[name]
	if !ok {
		info = ProviderInfo{BaseURL: "https://api.openai.com/v1", APIMode: "openai"}
	}
	if model == "" {
		model = getDefaultModel(name)
	}
	apiKey := os.Getenv(info.APIKeyEnv)
	if apiKey == "" {
		apiKey = os.Getenv("OPENAI_API_KEY")
	}
	return &OpenAIProvider{
		name:    name,
		baseURL: info.BaseURL,
		apiKey:  apiKey,
		model:  model,
		client: &http.Client{Timeout: 120 * time.Second},
		apiMode: info.APIMode,
	}
}

// NewProviderWithAPIKey creates a provider with an explicit API key (overrides env var lookup).
func NewProviderWithAPIKey(name, model, apiKey string) *OpenAIProvider {
	info, ok := ProviderRegistry[name]
	if !ok {
		info = ProviderInfo{BaseURL: "https://api.openai.com/v1", APIMode: "openai"}
	}
	if model == "" {
		model = getDefaultModel(name)
	}
	return &OpenAIProvider{
		name:    name,
		baseURL: info.BaseURL,
		apiKey:  apiKey,
		model:  model,
		client: &http.Client{Timeout: 120 * time.Second},
		apiMode: info.APIMode,
	}
}

func getDefaultModel(provider string) string {
	models := map[string]string{
		"nous":          "xiaomi/mimo-v2-pro",
		"openai":       "gpt-4o",
		"anthropic":    "claude-3-5-sonnet-20241022",
		"google":       "gemini-2.0-flash",
		"gemini":      "gemini-2.0-flash",
		"deepseek":     "deepseek-chat",
		"xai":        "grok-2",
		"nvidia":      "nvidia/nemotron-70b",
		"kimi":       "kimi-k2",
		"glm":         "glm-4-flash",
		"z":          "glm-4-flash",
		"minimax":     "abab6.5s-chat",
		"qwen":        "qwen2.5-coder-32b",
		"alibaba":     "qwen-turbo",
		"xiaomi":      "xiaomi/mimo-v2-pro",
		"openrouter":  "openai/gpt-4o",
		"huggingface": "facebook/bart-large",
		"ollama":     "llama3.1",
		"copilot":    "gpt-4o",
		"arcee":      "arcee-ai/diva",
		"groq":       "llama-3.1-70b-versatile",
		"perplexity": "llama-3.1-sonar-small-online",
	}
	if m, ok := models[provider]; ok {
		return m
	}
	return "gpt-4o"
}

func (p *OpenAIProvider) Name() string    { return p.name }
func (p *OpenAIProvider) GetModel() string { return p.model }

func (p *OpenAIProvider) Chat(ctx context.Context, messages []Message, tools []ToolDef) (*Response, error) {
	if p.apiKey == "" {
		return nil, fmt.Errorf("no API key for provider: %s", p.name)
	}

	body := map[string]interface{}{"model": p.model, "messages": messages}
	if len(tools) > 0 {
		body["tools"] = tools
	}

	// Add prompt caching metadata for supported models (e.g., o1, o1-mini)
	if isCachedModel(p.model) {
		body["metadata"] = map[string]interface{}{
			"cache_tokens": calculateCacheTokens(messages),
		}
	}

	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat/completions", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	// Add Anthropic-style cache control header for supported providers
	if isCachedModel(p.model) {
		req.Header.Set("anthropic-beta", "prompt-caching-2024-07-31")
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request: %w", err)
	}
	defer resp.Body.Close()
	bodyBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API %d: %s", resp.StatusCode, string(bodyBytes))
	}
	var result struct {
		Choices []struct {
			Message struct {
				Content   string     `json:"content"`
				ToolCalls []ToolCall `json:"tool_calls"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
		Usage Usage `json:"usage"`
		Model string `json:"model"`
	}
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}
	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("no choices")
	}
	choice := result.Choices[0]
	return &Response{
		Content:      choice.Message.Content,
		ToolCalls:    choice.Message.ToolCalls,
		FinishReason: choice.FinishReason,
		Usage:      result.Usage,
		Model:      result.Model,
	}, nil
}

type ProviderManager struct {
	providers map[string]*OpenAIProvider
	default_  string
}

func NewProviderManager() *ProviderManager {
	return &ProviderManager{providers: make(map[string]*OpenAIProvider)}
}

func (pm *ProviderManager) Register(name, model string) {
	pm.providers[name] = NewProvider(name, model)
}

// RegisterWithAPIKey registers a provider with an explicit API key (overrides env var).
func (pm *ProviderManager) RegisterWithAPIKey(name, model, apiKey string) {
	pm.providers[name] = NewProviderWithAPIKey(name, model, apiKey)
}

func (pm *ProviderManager) SetDefault(name string) {
	pm.default_ = name
}

func (pm *ProviderManager) Get(name string) (*OpenAIProvider, error) {
	if name == "" {
		name = pm.default_
	}
	if p, ok := pm.providers[name]; ok {
		return p, nil
	}
	if p, ok := pm.providers[pm.default_]; ok {
		return p, nil
	}
	return nil, fmt.Errorf("provider not found: %s", name)
}

func (pm *ProviderManager) List() []string {
	names := make([]string, 0, len(pm.providers))
	for name := range pm.providers {
		names = append(names, name)
	}
	return names
}