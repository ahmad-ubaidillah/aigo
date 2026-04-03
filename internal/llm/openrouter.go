package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// OpenRouterClient implements LLMClient for OpenRouter API (OpenAI-compatible).
type OpenRouterClient struct {
	apiKey     string
	model      string
	baseURL    string
	httpClient *http.Client
}

// NewOpenRouterClient creates a new OpenRouter client.
func NewOpenRouterClient(apiKey, model string) *OpenRouterClient {
	return &OpenRouterClient{
		apiKey:     apiKey,
		model:      model,
		baseURL:    "https://openrouter.ai/api/v1",
		httpClient: &http.Client{},
	}
}

// openRouterRequest represents the OpenRouter API request (OpenAI-compatible).
type openRouterRequest struct {
	Model       string              `json:"model"`
	Messages    []openRouterMessage `json:"messages"`
	Temperature float64             `json:"temperature,omitempty"`
	MaxTokens   int                 `json:"max_tokens,omitempty"`
	Stop        []string            `json:"stop,omitempty"`
	Stream      bool                `json:"stream"`
}

// openRouterMessage represents a message in OpenRouter format.
type openRouterMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// openRouterResponse represents the OpenRouter API response.
type openRouterResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// Chat sends a chat request to OpenRouter.
func (c *OpenRouterClient) Chat(ctx context.Context, messages []Message) (*Response, error) {
	resp, err := c.ChatWithOptions(ctx, messages, ChatOptions{})
	if err != nil {
		return nil, err
	}
	return &Response{
		Content: resp.Content,
		Usage: Usage{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		},
	}, nil
}

func (c *OpenRouterClient) ChatWithOptions(ctx context.Context, messages []Message, opts ChatOptions) (*ChatResponse, error) {
	req := c.buildRequest(messages, opts, false)
	body, err := c.doRequest(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("openrouter chat request: %w", err)
	}
	return c.parseResponse(body)
}

// Stream sends a streaming chat request to OpenRouter.
func (c *OpenRouterClient) Stream(ctx context.Context, messages []Message, opts ChatOptions, callback func(delta string)) (*ChatResponse, error) {
	req := c.buildRequest(messages, opts, true)
	resp, err := c.doStreamRequest(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("openrouter stream request: %w", err)
	}
	defer resp.Body.Close()
	return c.parseStreamResponse(resp.Body, callback)
}

func (c *OpenRouterClient) Complete(ctx context.Context, prompt string) (*Response, error) {
	messages := []Message{{Role: "user", Content: prompt}}
	resp, err := c.ChatWithOptions(ctx, messages, ChatOptions{Model: c.model})
	if err != nil {
		return nil, err
	}
	return &Response{Content: resp.Content, Usage: Usage(resp.Usage)}, nil
}

func (c *OpenRouterClient) CompleteWithSystem(ctx context.Context, systemPrompt, userPrompt string) (*Response, error) {
	messages := []Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userPrompt},
	}
	resp, err := c.ChatWithOptions(ctx, messages, ChatOptions{Model: c.model})
	if err != nil {
		return nil, err
	}
	return &Response{Content: resp.Content, Usage: Usage(resp.Usage)}, nil
}

func (c *OpenRouterClient) buildRequest(messages []Message, opts ChatOptions, stream bool) *openRouterRequest {
	msgs := make([]openRouterMessage, len(messages))
	for i, m := range messages {
		msgs[i] = openRouterMessage{Role: m.Role, Content: m.Content}
	}
	model := opts.Model
	if model == "" {
		model = c.model
	}
	return &openRouterRequest{
		Model:       model,
		Messages:    msgs,
		Temperature: opts.Temperature,
		MaxTokens:   opts.MaxTokens,
		Stop:        opts.StopSequences,
		Stream:      stream,
	}
}

func (c *OpenRouterClient) doRequest(ctx context.Context, req *openRouterRequest) ([]byte, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("HTTP-Referer", "https://github.com/ahmad-ubaidillah/aigo")
	httpReq.Header.Set("X-Title", "Aigo")
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("api error %d: %s", resp.StatusCode, string(respBody))
	}
	return io.ReadAll(resp.Body)
}

func (c *OpenRouterClient) SetBaseURL(url string) {
	c.baseURL = url
}

func (c *OpenRouterClient) doStreamRequest(ctx context.Context, req *openRouterRequest) (*http.Response, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("HTTP-Referer", "https://github.com/ahmad-ubaidillah/aigo")
	httpReq.Header.Set("X-Title", "Aigo")
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("api error %d: %s", resp.StatusCode, string(respBody))
	}
	return resp, nil
}

func (c *OpenRouterClient) parseResponse(body []byte) (*ChatResponse, error) {
	var resp openRouterResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}
	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}
	return &ChatResponse{
		Content: resp.Choices[0].Message.Content,
		Usage: TokenUsage{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		},
	}, nil
}

func (c *OpenRouterClient) parseStreamResponse(reader io.Reader, callback func(delta string)) (*ChatResponse, error) {
	var content strings.Builder
	buf := make([]byte, 4096)
	for {
		n, err := reader.Read(buf)
		if n > 0 {
			chunks := strings.Split(string(buf[:n]), "data: ")
			for _, chunk := range chunks {
				chunk = strings.TrimSpace(chunk)
				if chunk == "" || chunk == "[DONE]" {
					continue
				}
				var resp openRouterResponse
				if err := json.Unmarshal([]byte(chunk), &resp); err != nil {
					continue
				}
				if len(resp.Choices) > 0 && resp.Choices[0].Delta.Content != "" {
					delta := resp.Choices[0].Delta.Content
					content.WriteString(delta)
					if callback != nil {
						callback(delta)
					}
				}
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read stream: %w", err)
		}
	}
	return &ChatResponse{Content: content.String()}, nil
}
