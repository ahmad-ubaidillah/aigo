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

// LocalClient implements LLMClient for local llama.cpp server.
// Uses OpenAI-compatible API format (llama.cpp supports this natively).
type LocalClient struct {
	model      string
	baseURL    string
	httpClient *http.Client
}

// NewLocalClient creates a client for a local llama.cpp server.
// baseURL defaults to http://localhost:8080/v1 if empty.
// No API key required — pass empty string.
func NewLocalClient(model, baseURL string) *LocalClient {
	if baseURL == "" {
		baseURL = "http://localhost:8080/v1"
	}
	return &LocalClient{
		model:      model,
		baseURL:    baseURL,
		httpClient: &http.Client{},
	}
}

// Chat sends a chat request to the local llama.cpp server.
func (c *LocalClient) Chat(ctx context.Context, messages []Message) (*Response, error) {
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

func (c *LocalClient) Complete(ctx context.Context, prompt string) (*Response, error) {
	messages := []Message{{Role: "user", Content: prompt}}
	return c.Chat(ctx, messages)
}

func (c *LocalClient) CompleteWithSystem(ctx context.Context, systemPrompt, userPrompt string) (*Response, error) {
	messages := []Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userPrompt},
	}
	return c.Chat(ctx, messages)
}

func (c *LocalClient) ChatWithOptions(ctx context.Context, messages []Message, opts ChatOptions) (*ChatResponse, error) {
	req := c.buildRequest(messages, opts, false)
	body, err := c.doRequest(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("local llm chat request: %w", err)
	}
	return c.parseResponse(body)
}

// Stream sends a streaming chat request to the local llama.cpp server.
func (c *LocalClient) Stream(ctx context.Context, messages []Message, opts ChatOptions, callback func(delta string)) (*ChatResponse, error) {
	req := c.buildRequest(messages, opts, true)
	resp, err := c.doStreamRequest(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("local llm stream request: %w", err)
	}
	defer resp.Body.Close()
	return c.parseStreamResponse(resp.Body, callback)
}

func (c *LocalClient) buildRequest(messages []Message, opts ChatOptions, stream bool) *openAIRequest {
	msgs := make([]openAIMessage, len(messages))
	for i, m := range messages {
		msgs[i] = openAIMessage{Role: m.Role, Content: m.Content}
	}
	model := opts.Model
	if model == "" {
		model = c.model
	}
	return &openAIRequest{
		Model:       model,
		Messages:    msgs,
		Temperature: opts.Temperature,
		MaxTokens:   opts.MaxTokens,
		Stop:        opts.StopSequences,
		Stream:      stream,
	}
}

func (c *LocalClient) doRequest(ctx context.Context, req *openAIRequest) ([]byte, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	// llama.cpp server does not require API key, but set it if present
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("local llm error %d: %s", resp.StatusCode, string(respBody))
	}
	return io.ReadAll(resp.Body)
}

func (c *LocalClient) doStreamRequest(ctx context.Context, req *openAIRequest) (*http.Response, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("local llm error %d: %s", resp.StatusCode, string(respBody))
	}
	return resp, nil
}

func (c *LocalClient) parseResponse(body []byte) (*ChatResponse, error) {
	var resp openAIResponse
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

func (c *LocalClient) parseStreamResponse(reader io.Reader, callback func(delta string)) (*ChatResponse, error) {
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
				var resp openAIResponse
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
