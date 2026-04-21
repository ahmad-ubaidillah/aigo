// Package providers streaming support via SSE.
package providers

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

// StreamChunk represents a single streaming chunk.
type StreamChunk struct {
	Delta      string     `json:"delta"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	FinishReason string   `json:"finish_reason,omitempty"`
	Usage      *Usage     `json:"usage,omitempty"`
	Done       bool       `json:"done"`
}

// StreamCallback is called for each chunk during streaming.
type StreamCallback func(chunk StreamChunk)

// ChatStream sends a streaming request and calls onChunk for each SSE event.
func (p *OpenAIProvider) ChatStream(ctx context.Context, messages []Message, tools []ToolDef, onChunk StreamCallback) (*Response, error) {
	body := map[string]interface{}{
		"model":    p.model,
		"messages": messages,
		"stream":   true,
	}
	if len(tools) > 0 {
		body["tools"] = tools
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat/completions", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("Accept", "text/event-stream")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var fullContent strings.Builder
	var allToolCalls []ToolCall
	var finalUsage Usage
	var finishReason string

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, ":") {
			continue
		}

		// Parse SSE data
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")

		// End of stream
		if data == "[DONE]" {
			if onChunk != nil {
				onChunk(StreamChunk{Done: true})
			}
			break
		}

		var event struct {
			Choices []struct {
				Delta struct {
					Content   string     `json:"content"`
					ToolCalls []ToolCall `json:"tool_calls"`
				} `json:"delta"`
				FinishReason string `json:"finish_reason"`
			} `json:"choices"`
			Usage Usage `json:"usage"`
		}

		if err := json.Unmarshal([]byte(data), &event); err != nil {
			continue // Skip malformed chunks
		}

		if len(event.Choices) > 0 {
			choice := event.Choices[0]

			if choice.Delta.Content != "" {
				fullContent.WriteString(choice.Delta.Content)
				if onChunk != nil {
					onChunk(StreamChunk{Delta: choice.Delta.Content})
				}
			}

			if len(choice.Delta.ToolCalls) > 0 {
				allToolCalls = append(allToolCalls, choice.Delta.ToolCalls...)
			}

			if choice.FinishReason != "" {
				finishReason = choice.FinishReason
			}
		}

		if event.Usage.TotalTokens > 0 {
			finalUsage = event.Usage
		}
	}

	return &Response{
		Content:      fullContent.String(),
		ToolCalls:    allToolCalls,
		FinishReason: finishReason,
		Usage:        finalUsage,
		Model:        p.model,
	}, nil
}
