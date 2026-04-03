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

// AnthropicClient implements LLMClient for Anthropic API.
type AnthropicClient struct {
	apiKey     string
	model      string
	baseURL    string
	httpClient *http.Client
}

// NewAnthropicClient creates a new Anthropic client.
func NewAnthropicClient(apiKey, model string) *AnthropicClient {
	return &AnthropicClient{
		apiKey:     apiKey,
		model:      model,
		baseURL:    "https://api.anthropic.com/v1",
		httpClient: &http.Client{},
	}
}

// anthropicRequest represents the Anthropic API request structure.
type anthropicRequest struct {
	Model         string             `json:"model"`
	Messages      []anthropicMessage `json:"messages"`
	System        string             `json:"system,omitempty"`
	MaxTokens     int                `json:"max_tokens"`
	Temperature   float64            `json:"temperature,omitempty"`
	StopSequences []string           `json:"stop_sequences,omitempty"`
	Stream        bool               `json:"stream"`
}

// anthropicMessage represents a message in Anthropic format.
type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// anthropicResponse represents the Anthropic API response.
type anthropicResponse struct {
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
	Usage struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
	StopReason string `json:"stop_reason"`
}

// anthropicStreamEvent represents a streaming event.
type anthropicStreamEvent struct {
	Type  string `json:"type"`
	Delta struct {
		Text string `json:"text"`
	} `json:"delta"`
	Message *anthropicResponse `json:"message,omitempty"`
}

// Chat sends a chat request to Anthropic.
func (c *AnthropicClient) Chat(ctx context.Context, messages []Message) (*Response, error) {
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

func (c *AnthropicClient) ChatWithOptions(ctx context.Context, messages []Message, opts ChatOptions) (*ChatResponse, error) {
	req := c.buildRequest(messages, opts, false)
	body, err := c.doRequest(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("anthropic chat request: %w", err)
	}
	return c.parseResponse(body)
}

func (c *AnthropicClient) Complete(ctx context.Context, prompt string) (*Response, error) {
	messages := []Message{{Role: "user", Content: prompt}}
	resp, err := c.ChatWithOptions(ctx, messages, ChatOptions{Model: c.model, MaxTokens: 4096})
	if err != nil {
		return nil, err
	}
	return &Response{Content: resp.Content, Usage: Usage(resp.Usage)}, nil
}

func (c *AnthropicClient) CompleteWithSystem(ctx context.Context, systemPrompt, userPrompt string) (*Response, error) {
	messages := []Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userPrompt},
	}
	resp, err := c.ChatWithOptions(ctx, messages, ChatOptions{Model: c.model, MaxTokens: 4096})
	if err != nil {
		return nil, err
	}
	return &Response{Content: resp.Content, Usage: Usage(resp.Usage)}, nil
}

func (c *AnthropicClient) Stream(ctx context.Context, messages []Message, opts ChatOptions, callback func(delta string)) (*ChatResponse, error) {
	req := c.buildRequest(messages, opts, true)
	resp, err := c.doStreamRequest(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("anthropic stream request: %w", err)
	}
	defer resp.Body.Close()
	return c.parseStreamResponse(resp.Body, callback)
}

func (c *AnthropicClient) buildRequest(messages []Message, opts ChatOptions, stream bool) *anthropicRequest {
	var system string
	msgs := make([]anthropicMessage, 0, len(messages))
	for _, m := range messages {
		if m.Role == "system" {
			system = m.Content
			continue
		}
		msgs = append(msgs, anthropicMessage{Role: m.Role, Content: m.Content})
	}
	model := opts.Model
	if model == "" {
		model = c.model
	}
	maxTokens := opts.MaxTokens
	if maxTokens == 0 {
		maxTokens = 4096
	}
	return &anthropicRequest{
		Model:         model,
		Messages:      msgs,
		System:        system,
		MaxTokens:     maxTokens,
		Temperature:   opts.Temperature,
		StopSequences: opts.StopSequences,
		Stream:        stream,
	}
}

func (c *AnthropicClient) doRequest(ctx context.Context, req *anthropicRequest) ([]byte, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/messages", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")
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

func (c *AnthropicClient) doStreamRequest(ctx context.Context, req *anthropicRequest) (*http.Response, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/messages", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")
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

func (c *AnthropicClient) parseResponse(body []byte) (*ChatResponse, error) {
	var resp anthropicResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}
	if len(resp.Content) == 0 {
		return nil, fmt.Errorf("no content in response")
	}
	return &ChatResponse{
		Content: resp.Content[0].Text,
		Usage: TokenUsage{
			PromptTokens:     resp.Usage.InputTokens,
			CompletionTokens: resp.Usage.OutputTokens,
			TotalTokens:      resp.Usage.InputTokens + resp.Usage.OutputTokens,
		},
	}, nil
}

func (c *AnthropicClient) parseStreamResponse(reader io.Reader, callback func(delta string)) (*ChatResponse, error) {
	var content strings.Builder
	var usage TokenUsage
	buf := make([]byte, 4096)
	for {
		n, err := reader.Read(buf)
		if n > 0 {
			lines := strings.Split(string(buf[:n]), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if !strings.HasPrefix(line, "data: ") {
					continue
				}
				data := strings.TrimPrefix(line, "data: ")
				if data == "" {
					continue
				}
				var event anthropicStreamEvent
				if err := json.Unmarshal([]byte(data), &event); err != nil {
					continue
				}
				if event.Type == "content_block_delta" && event.Delta.Text != "" {
					content.WriteString(event.Delta.Text)
					if callback != nil {
						callback(event.Delta.Text)
					}
				}
				if event.Type == "message_delta" && event.Message != nil {
					usage.PromptTokens = event.Message.Usage.InputTokens
					usage.CompletionTokens = event.Message.Usage.OutputTokens
					usage.TotalTokens = usage.PromptTokens + usage.CompletionTokens
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
	return &ChatResponse{Content: content.String(), Usage: usage}, nil
}
