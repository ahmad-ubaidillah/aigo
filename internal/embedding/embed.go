package embedding

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strings"
	"sync"
)

type Embedder interface {
	Embed(text string) ([]float64, error)
	EmbedBatch(texts []string) ([][]float64, error)
}

type OpenAIEmbedder struct {
	apiKey string
	model  string
	client *http.Client
	mu     sync.RWMutex
	cache  map[string][]float64
}

func NewOpenAIEmbedder(apiKey, model string) *OpenAIEmbedder {
	if model == "" {
		model = "text-embedding-ada-002"
	}
	return &OpenAIEmbedder{
		apiKey: apiKey,
		model:  model,
		client: &http.Client{},
		cache:  make(map[string][]float64),
	}
}

func (e *OpenAIEmbedder) Embed(text string) ([]float64, error) {
	e.mu.RLock()
	if vec, ok := e.cache[text]; ok {
		e.mu.RUnlock()
		return vec, nil
	}
	e.mu.RUnlock()

	body := map[string]any{"model": e.model, "input": text}
	data, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", "https://api.openai.com/v1/embeddings", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+e.apiKey)

	resp, err := e.client.Do(req)
	if err != nil {
		return localEmbed(text), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return localEmbed(text), nil
	}

	var result openAIEmbedResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return localEmbed(text), nil
	}
	if len(result.Data) == 0 {
		return localEmbed(text), nil
	}

	vec := result.Data[0].Embedding
	e.mu.Lock()
	e.cache[text] = vec
	e.mu.Unlock()
	return vec, nil
}

func (e *OpenAIEmbedder) EmbedBatch(texts []string) ([][]float64, error) {
	results := make([][]float64, len(texts))
	uncached := make([]string, 0)
	uncachedIdx := make([]int, 0)

	e.mu.RLock()
	for i, text := range texts {
		if vec, ok := e.cache[text]; ok {
			results[i] = vec
		} else {
			uncached = append(uncached, text)
			uncachedIdx = append(uncachedIdx, i)
		}
	}
	e.mu.RUnlock()

	if len(uncached) == 0 {
		return results, nil
	}

	body := map[string]any{"model": e.model, "input": uncached}
	data, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", "https://api.openai.com/v1/embeddings", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+e.apiKey)

	resp, err := e.client.Do(req)
	if err != nil {
		for i, t := range texts {
			if results[i] == nil {
				results[i] = localEmbed(t)
			}
		}
		return results, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		for i, t := range texts {
			if results[i] == nil {
				results[i] = localEmbed(t)
			}
		}
		return results, nil
	}

	var result openAIEmbedResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		for i, t := range texts {
			if results[i] == nil {
				results[i] = localEmbed(t)
			}
		}
		return results, nil
	}

	e.mu.Lock()
	for i, d := range result.Data {
		idx := uncachedIdx[i]
		results[idx] = d.Embedding
		e.cache[uncached[i]] = d.Embedding
	}
	e.mu.Unlock()

	return results, nil
}

type openAIEmbedResponse struct {
	Object string `json:"object"`
	Data   []struct {
		Embedding []float64 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
	Model string `json:"model"`
	Usage struct {
		PromptTokens int `json:"prompt_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
}

func (e *OpenAIEmbedder) CacheSize() int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return len(e.cache)
}

func (e *OpenAIEmbedder) ClearCache() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.cache = make(map[string][]float64)
}

type LocalEmbedder struct {
	dim int
}

func NewLocalEmbedder(dim int) *LocalEmbedder {
	if dim <= 0 {
		dim = 128
	}
	return &LocalEmbedder{dim: dim}
}

func (e *LocalEmbedder) Embed(text string) ([]float64, error) {
	return localEmbedDim(text, e.dim), nil
}

func (e *LocalEmbedder) EmbedBatch(texts []string) ([][]float64, error) {
	result := make([][]float64, len(texts))
	for i, t := range texts {
		result[i] = localEmbedDim(t, e.dim)
	}
	return result, nil
}

func localEmbed(text string) []float64 {
	return localEmbedDim(text, 128)
}

func localEmbedDim(text string, dim int) []float64 {
	vec := make([]float64, dim)
	hash := sha256.Sum256([]byte(text))
	words := strings.Fields(text)
	for i := 0; i < dim; i++ {
		b1 := hash[i%len(hash)]
		b2 := hash[(i+1)%len(hash)]
		val := float64(b1)/255.0 - 0.5
		if i < len(words) {
			w := words[i]
			for _, c := range w {
				val += float64(c) / 1000.0
			}
		}
		vec[i] = val * math.Cos(float64(b2)*math.Pi/128.0)
	}
	norm := 0.0
	for _, v := range vec {
		norm += v * v
	}
	norm = math.Sqrt(norm)
	if norm > 0 {
		for i := range vec {
			vec[i] /= norm
		}
	}
	return vec
}

type EmbeddingCache struct {
	cache map[string][]float64
}

func NewEmbeddingCache() *EmbeddingCache {
	return &EmbeddingCache{cache: make(map[string][]float64)}
}

func (c *EmbeddingCache) Get(text string) ([]float64, bool) {
	key := cacheKey(text)
	v, ok := c.cache[key]
	return v, ok
}

func (c *EmbeddingCache) Set(text string, embedding []float64) {
	key := cacheKey(text)
	c.cache[key] = embedding
}

func (c *EmbeddingCache) CachedEmbed(embedder Embedder, text string) ([]float64, error) {
	if v, ok := c.Get(text); ok {
		return v, nil
	}
	vec, err := embedder.Embed(text)
	if err != nil {
		return nil, err
	}
	c.Set(text, vec)
	return vec, nil
}

func cacheKey(text string) string {
	h := sha256.Sum256([]byte(text))
	return hex.EncodeToString(h[:8])
}

type CachedEmbedder struct {
	embedder Embedder
	cache    *EmbeddingCache
}

func NewCachedEmbedder(embedder Embedder) *CachedEmbedder {
	return &CachedEmbedder{
		embedder: embedder,
		cache:    NewEmbeddingCache(),
	}
}

func (c *CachedEmbedder) Embed(text string) ([]float64, error) {
	return c.cache.CachedEmbed(c.embedder, text)
}

func (c *CachedEmbedder) EmbedBatch(texts []string) ([][]float64, error) {
	result := make([][]float64, len(texts))
	for i, t := range texts {
		v, err := c.Embed(t)
		if err != nil {
			return nil, fmt.Errorf("embed text %d: %w", i, err)
		}
		result[i] = v
	}
	return result, nil
}
