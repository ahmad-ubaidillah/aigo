package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// GLMClient implements LLMClient for GLM (z.ai) API.
// GLM uses OpenAI-compatible API format with JWT token authentication.
type GLMClient struct {
	apiKey     string
	model      string
	baseURL    string
	httpClient *http.Client
}

// NewGLMClient creates a new GLM client with JWT token authentication.
// The apiKey should be a JWT token generated from the GLM API key.
func NewGLMClient(apiKey, model string) *GLMClient {
	return &GLMClient{
		apiKey:     apiKey,
		model:      model,
		baseURL:    "https://open.bigmodel.cn/api/paas/v4",
		httpClient: &http.Client{},
	}
}

// NewGLMClientWithBaseURL creates a GLM client with a custom base URL.
func NewGLMClientWithBaseURL(apiKey, model, baseURL string) *GLMClient {
	return &GLMClient{
		apiKey:     apiKey,
		model:      model,
		baseURL:    baseURL,
		httpClient: &http.Client{},
	}
}

// glmToolFunction defines the function schema for tool calling.
type glmToolFunction struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Parameters  any    `json:"parameters"`
}

// glmToolDefinition represents a tool/function definition for the API request.
type glmToolDefinition struct {
	Type     string          `json:"type"`
	Function glmToolFunction `json:"function"`
}

// glmToolCall represents a tool call in the API response.
type glmToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

// glmMessage represents a message in GLM API format.
type glmMessage struct {
	Role       string        `json:"role"`
	Content    string        `json:"content"`
	ToolCalls  []glmToolCall `json:"tool_calls,omitempty"`
	ToolCallID string        `json:"tool_call_id,omitempty"`
}

// glmRequest represents the GLM API request structure.
type glmRequest struct {
	Model       string              `json:"model"`
	Messages    []glmMessage        `json:"messages"`
	Temperature float64             `json:"temperature,omitempty"`
	MaxTokens   int                 `json:"max_tokens,omitempty"`
	Stop        []string            `json:"stop,omitempty"`
	Stream      bool                `json:"stream"`
	Tools       []glmToolDefinition `json:"tools,omitempty"`
	ToolChoice  any                 `json:"tool_choice,omitempty"`
}

// glmResponse represents the GLM API response.
type glmResponse struct {
	Choices []struct {
		Message struct {
			Content   string        `json:"content"`
			ToolCalls []glmToolCall `json:"tool_calls,omitempty"`
		} `json:"message"`
		Delta struct {
			Content   string        `json:"content"`
			ToolCalls []glmToolCall `json:"tool_calls,omitempty"`
		} `json:"delta"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// Chat sends a chat request to GLM.
func (c *GLMClient) SetBaseURL(url string) {
	c.baseURL = url
}

func (c *GLMClient) Chat(ctx context.Context, messages []Message) (*Response, error) {
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

func (c *GLMClient) Complete(ctx context.Context, prompt string) (*Response, error) {
	messages := []Message{{Role: "user", Content: prompt}}
	return c.Chat(ctx, messages)
}

func (c *GLMClient) CompleteWithSystem(ctx context.Context, systemPrompt, userPrompt string) (*Response, error) {
	messages := []Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userPrompt},
	}
	return c.Chat(ctx, messages)
}

func (c *GLMClient) ChatWithOptions(ctx context.Context, messages []Message, opts ChatOptions) (*ChatResponse, error) {
	req := c.buildRequest(messages, opts, false)
	body, err := c.doRequest(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("glm chat request: %w", err)
	}
	return c.parseResponse(body)
}

// Stream sends a streaming chat request to GLM.
func (c *GLMClient) Stream(ctx context.Context, messages []Message, opts ChatOptions, callback func(delta string)) (*ChatResponse, error) {
	req := c.buildRequest(messages, opts, true)
	resp, err := c.doStreamRequest(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("glm stream request: %w", err)
	}
	defer resp.Body.Close()
	return c.parseStreamResponse(resp.Body, callback)
}

func (c *GLMClient) buildRequest(messages []Message, opts ChatOptions, stream bool) *glmRequest {
	msgs := make([]glmMessage, len(messages))
	for i, m := range messages {
		msgs[i] = glmMessage{Role: m.Role, Content: m.Content}
	}
	model := opts.Model
	if model == "" {
		model = c.model
	}
	return &glmRequest{
		Model:       model,
		Messages:    msgs,
		Temperature: opts.Temperature,
		MaxTokens:   opts.MaxTokens,
		Stop:        opts.StopSequences,
		Stream:      stream,
	}
}

func (c *GLMClient) doRequest(ctx context.Context, req *glmRequest) ([]byte, error) {
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
	
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()
	
	respBody, _ := io.ReadAll(resp.Body)
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("api error %d: %s", resp.StatusCode, string(respBody))
	}
	
	return respBody, nil
}

func (c *GLMClient) doStreamRequest(ctx context.Context, req *glmRequest) (*http.Response, error) {
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

func (c *GLMClient) parseResponse(body []byte) (*ChatResponse, error) {
	var resp glmResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}
	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}
	choice := resp.Choices[0]
	result := &ChatResponse{
		Content: choice.Message.Content,
		Usage: TokenUsage{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		},
	}
	if len(choice.Message.ToolCalls) > 0 {
		result.ToolCalls = make([]ToolCall, len(choice.Message.ToolCalls))
		for i, tc := range choice.Message.ToolCalls {
			result.ToolCalls[i] = ToolCall{
				ID:        tc.ID,
				Name:      tc.Function.Name,
				Arguments: tc.Function.Arguments,
			}
		}
	}
	return result, nil
}

func (c *GLMClient) parseStreamResponse(reader io.Reader, callback func(delta string)) (*ChatResponse, error) {
	var content strings.Builder
	var toolCalls []ToolCall
	scanner := bufio.NewScanner(reader)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "" || data == "[DONE]" {
			continue
		}
		var resp glmResponse
		if err := json.Unmarshal([]byte(data), &resp); err != nil {
			continue
		}
		if len(resp.Choices) == 0 {
			continue
		}
		delta := resp.Choices[0].Delta
		if delta.Content != "" {
			content.WriteString(delta.Content)
			if callback != nil {
				callback(delta.Content)
			}
		}
		if len(delta.ToolCalls) > 0 {
			for _, tc := range delta.ToolCalls {
				toolCalls = append(toolCalls, ToolCall{
					ID:        tc.ID,
					Name:      tc.Function.Name,
					Arguments: tc.Function.Arguments,
				})
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read stream: %w", err)
	}
	return &ChatResponse{
		Content:   content.String(),
		ToolCalls: toolCalls,
	}, nil
}
