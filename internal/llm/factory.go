package llm

import (
	"fmt"
	"sort"
	"time"

	"github.com/ahmad-ubaidillah/aigo/pkg/types"
)

type NamedProvider struct {
	Name    string
	Client  LLMClient
	Model   string
	Timeout time.Duration
}

func NewProvider(cfg types.ProviderConfig) (LLMClient, error) {
	if !cfg.Enabled {
		return nil, nil
	}

	switch cfg.Name {
	case "openai":
		if cfg.APIKey == "" {
			return nil, nil
		}
		client := NewOpenAIClient(cfg.APIKey, cfg.Model)
		if cfg.BaseURL != "" {
			client.SetBaseURL(cfg.BaseURL)
		}
		return client, nil
	case "anthropic":
		if cfg.APIKey == "" {
			return nil, nil
		}
		return NewAnthropicClient(cfg.APIKey, cfg.Model), nil
	case "openrouter":
		if cfg.APIKey == "" {
			return nil, nil
		}
		client := NewOpenRouterClient(cfg.APIKey, cfg.Model)
		if cfg.BaseURL != "" {
			client.SetBaseURL(cfg.BaseURL)
		}
		return client, nil
	case "glm":
		if cfg.APIKey == "" {
			return nil, nil
		}
		if cfg.BaseURL != "" {
			return NewGLMClientWithBaseURL(cfg.APIKey, cfg.Model, cfg.BaseURL), nil
		}
		return NewGLMClient(cfg.APIKey, cfg.Model), nil
	case "local":
		return NewLocalClient(cfg.Model, cfg.BaseURL), nil
	case "custom":
		if cfg.BaseURL == "" {
			return nil, fmt.Errorf("custom provider requires base_url")
		}
		client := NewOpenAIClient(cfg.APIKey, cfg.Model)
		client.SetBaseURL(cfg.BaseURL)
		return client, nil
	default:
		return nil, fmt.Errorf("unknown provider: %s", cfg.Name)
	}
}

func NewProviders(cfg types.LLMConfig) ([]NamedProvider, error) {
	sorted := make([]types.ProviderConfig, len(cfg.Providers))
	copy(sorted, cfg.Providers)

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Priority < sorted[j].Priority
	})

	var result []NamedProvider
	for _, p := range sorted {
		client, err := NewProvider(p)
		if err != nil {
			return nil, fmt.Errorf("create provider %s: %w", p.Name, err)
		}
		if client == nil {
			continue
		}

		timeout := time.Duration(p.Timeout) * time.Second
		if p.Timeout == 0 {
			timeout = 30 * time.Second
		}

		result = append(result, NamedProvider{
			Name:    p.Name,
			Client:  client,
			Model:   p.Model,
			Timeout: timeout,
		})
	}

	if len(cfg.Fallback) > 0 {
		result = sortByFallback(result, cfg.Fallback)
	}

	return result, nil
}

func sortByFallback(providers []NamedProvider, fallback []string) []NamedProvider {
	providerMap := make(map[string]NamedProvider)
	for _, p := range providers {
		providerMap[p.Name] = p
	}

	var result []NamedProvider
	seen := make(map[string]bool)

	for _, name := range fallback {
		if p, ok := providerMap[name]; ok && !seen[name] {
			result = append(result, p)
			seen[name] = true
		}
	}

	for _, p := range providers {
		if !seen[p.Name] {
			result = append(result, p)
		}
	}

	return result
}
