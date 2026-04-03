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

type OpenAIClient struct {
	apiKey     string
	model      string
	baseURL    string
	httpClient *http.Client
}

func NewOpenAIClient(apiKey, model string) *OpenAIClient {
	return &OpenAIClient{
		apiKey:     apiKey,
		model:      model,
		baseURL:    "https://api.openai.com/v1",
		httpClient: &http.Client{},
	}
}

func (c *OpenAIClient) SetBaseURL(url string) {
	c.baseURL = url
}

type openAIRequest struct {
	Model       string          `json:"model"`
	Messages    []openAIMessage `json:"messages"`
	Temperature float64         `json:"temperature,omitempty"`
	MaxTokens   int             `json:"max_tokens,omitempty"`
	Stop        []string        `json:"stop,omitempty"`
	Stream      bool            `json:"stream"`
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIResponse struct {
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

func (c *OpenAIClient) Chat(ctx context.Context, messages []Message) (*Response, error) {
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

func (c *OpenAIClient) Complete(ctx context.Context, prompt string) (*Response, error) {
	messages := []Message{{Role: "user", Content: prompt}}
	return c.Chat(ctx, messages)
}

func (c *OpenAIClient) CompleteWithSystem(ctx context.Context, systemPrompt, userPrompt string) (*Response, error) {
	messages := []Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userPrompt},
	}
	return c.Chat(ctx, messages)
}

func (c *OpenAIClient) ChatWithOptions(ctx context.Context, messages []Message, opts ChatOptions) (*ChatResponse, error) {
	req := c.buildRequest(messages, opts, false)
	body, err := c.doRequest(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("openai chat request: %w", err)
	}
	return c.parseResponse(body)
}

func (c *OpenAIClient) Stream(ctx context.Context, messages []Message, opts ChatOptions, callback func(delta string)) (*ChatResponse, error) {
	req := c.buildRequest(messages, opts, true)
	resp, err := c.doStreamRequest(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("openai stream request: %w", err)
	}
	defer resp.Body.Close()
	return c.parseStreamResponse(resp.Body, callback)
}

func (c *OpenAIClient) buildRequest(messages []Message, opts ChatOptions, stream bool) *openAIRequest {
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

func (c *OpenAIClient) doRequest(ctx context.Context, req *openAIRequest) ([]byte, error) {
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
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("api error %d: %s", resp.StatusCode, string(respBody))
	}
	return io.ReadAll(resp.Body)
}

func (c *OpenAIClient) doStreamRequest(ctx context.Context, req *openAIRequest) (*http.Response, error) {
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

func (c *OpenAIClient) parseResponse(body []byte) (*ChatResponse, error) {
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

func (c *OpenAIClient) parseStreamResponse(reader io.Reader, callback func(delta string)) (*ChatResponse, error) {
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
