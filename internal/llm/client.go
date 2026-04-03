// Package llm provides interfaces for LLM clients
package llm

import (
	"context"
)

// Message represents a message in a conversation
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Response represents a response from the LLM
type Response struct {
	Content string `json:"content"`
	Usage   Usage  `json:"usage,omitempty"`
}

// Usage represents token usage information
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ChatOptions contains options for chat requests
type ChatOptions struct {
	Model         string    `json:"model,omitempty"`
	Temperature   float64   `json:"temperature,omitempty"`
	MaxTokens     int       `json:"max_tokens,omitempty"`
	StopSequences []string  `json:"stop_sequences,omitempty"`
	Tools         []ToolDef `json:"tools,omitempty"`
}

// ChatResponse represents a response from a chat request
type ChatResponse struct {
	Content   string     `json:"content"`
	Usage     TokenUsage `json:"usage,omitempty"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

// TokenUsage represents token usage information
type TokenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ToolCall represents a tool call in a response
type ToolCall struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// ToolDef represents a tool definition for function calling
type ToolDef struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Parameters  any    `json:"parameters"`
}

// LLMClient defines the interface for LLM clients
type LLMClient interface {
	Complete(ctx context.Context, prompt string) (*Response, error)
	Chat(ctx context.Context, messages []Message) (*Response, error)
	CompleteWithSystem(ctx context.Context, systemPrompt, userPrompt string) (*Response, error)
}

// ExtendedLLMClient defines an extended interface with options support
type ExtendedLLMClient interface {
	LLMClient
	// ChatWithOptions sends a chat request with options
	ChatWithOptions(ctx context.Context, messages []Message, opts ChatOptions) (*ChatResponse, error)
}

// Streamer defines the interface for streaming LLM clients
type Streamer interface {
	// Stream sends a streaming chat request
	Stream(ctx context.Context, messages []Message, opts ChatOptions, callback func(delta string)) (*ChatResponse, error)
}

// Chatter defines the interface for chat with options
type Chatter interface {
	// ChatWithOptions sends a chat request with options
	ChatWithOptions(ctx context.Context, messages []Message, opts ChatOptions) (*ChatResponse, error)
}
