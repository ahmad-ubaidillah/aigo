// Package agent — Context compression.
// Inspired by NekoClaw's CJK-aware token estimation and tool output truncation.
package agent

import (
	"strconv"
	"unicode/utf8"

	"github.com/hermes-v2/aigo/internal/providers"
)

// Compressor handles context window management.
type Compressor struct {
	maxTokens    int
	reserveTokens int // Reserve for response
}

// NewCompressor creates a context compressor.
func NewCompressor(maxTokens, reserveTokens int) *Compressor {
	return &Compressor{
		maxTokens:     maxTokens,
		reserveTokens: reserveTokens,
	}
}

// EstimateTokens estimates token count (CJK-aware).
// ~4 chars/token for English, ~2 chars/token for CJK.
func EstimateTokens(text string) int {
	if text == "" {
		return 0
	}

	cjkCount := 0
	latinCount := 0

	for _, r := range text {
		if isCJK(r) {
			cjkCount++
		} else {
			latinCount++
		}
	}

	// CJK: ~1 token per char, Latin: ~4 chars per token
	return cjkCount + (latinCount / 4) + 1
}

func isCJK(r rune) bool {
	return (r >= 0x4E00 && r <= 0x9FFF) ||   // CJK Unified
		(r >= 0x3040 && r <= 0x309F) ||   // Hiragana
		(r >= 0x30A0 && r <= 0x30FF) ||   // Katakana
		(r >= 0xAC00 && r <= 0xD7AF)      // Hangul
}

// NeedsCompression checks if messages exceed the context window.
func (c *Compressor) NeedsCompression(messages []providers.Message) bool {
	total := 0
	for _, msg := range messages {
		total += EstimateTokens(msg.Content)
		for _, tc := range msg.ToolCalls {
			total += EstimateTokens(tc.Function.Name)
			total += EstimateTokens(tc.Function.Arguments)
		}
	}
	return total > (c.maxTokens - c.reserveTokens)
}

// Compress reduces context by truncating older messages.
// Strategy: Keep system prompt + last N messages, truncate middle tool results.
func (c *Compressor) Compress(messages []providers.Message) []providers.Message {
	if !c.NeedsCompression(messages) {
		return messages
	}

	if len(messages) <= 3 {
		return messages
	}

	// Keep: system prompt (index 0), last user message, last 4 tool interactions
	result := []providers.Message{messages[0]} // system prompt

	// Find the last user message
	lastUserIdx := -1
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == "user" {
			lastUserIdx = i
			break
		}
	}

	// Keep recent messages (last 4 interactions = ~8 messages)
	keepFrom := len(messages) - 8
	if keepFrom < 1 {
		keepFrom = 1
	}

	// Ensure we include the last user message
	if lastUserIdx > 0 && lastUserIdx < keepFrom {
		// Add a summary placeholder
		result = append(result, providers.Message{
			Role:    "assistant",
			Content: "[Context compressed — previous messages summarized]",
		})
		keepFrom = lastUserIdx
	}

	for i := keepFrom; i < len(messages); i++ {
		msg := messages[i]
		// Truncate long tool results aggressively to save tokens
		if msg.Role == "tool" && len(msg.Content) > 800 {
			msg.Content = TruncateToolOutput(msg.Content, 800)
		}
		result = append(result, msg)
	}

	return result
}

// TruncateToolOutput keeps head + tail of long tool output.
// Inspired by NekoClaw's head+tail truncation.
func TruncateToolOutput(output string, maxLen int) string {
	if len(output) <= maxLen {
		return output
	}

	headLen := maxLen * 60 / 100 // 60% head
	tailLen := maxLen * 40 / 100 // 40% tail

	head := output[:headLen]
	tail := output[len(output)-tailLen:]

	// Ensure we don't break UTF-8
	for !utf8.ValidString(head) {
		head = head[:len(head)-1]
	}
	for !utf8.ValidString(tail) {
		tail = tail[1:]
	}

	return head + "\n\n... [truncated " + strconv.Itoa(len(output)-headLen-tailLen) + " chars] ...\n\n" + tail
}

// CompactToolOutput compresses successful tool output into a minimal format.
// This saves tokens by avoiding full tool output in the LLM context.
func CompactToolOutput(toolName, output string) string {
	// Errors are kept as-is (the LLM needs to see errors to fix them)
	if len(output) >= 6 && output[:6] == "Error:" {
		return output
	}
	if len(output) > 500 {
		output = TruncateToolOutput(output, 500)
	}
	return output
}
