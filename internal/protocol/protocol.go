// Package protocol implements the AIGO Protocol for structured agent output.
// Similar to Golem's GOLEM_PROTOCOL but simpler and more flexible.
//
// The protocol uses structured markers in AI responses:
//   [AIGO_MEMORY] - Information to remember
//   [AIGO_ACTION] - Actions to execute (JSON)
//   [AIGO_REPLY]  - Response to user
//   [AIGO_DIARY]  - Diary entry
//   [AIGO_LEARN]  - New learning/knowledge
//
// If no markers present, entire response is treated as AIGO_REPLY.
package protocol

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// ParsedResponse represents a parsed protocol response.
type ParsedResponse struct {
	Memory   []string         `json:"memory"`
	Actions  []Action         `json:"actions"`
	Reply    string           `json:"reply"`
	Diary    string           `json:"diary,omitempty"`
	Learn    []Learning       `json:"learn,omitempty"`
}

// Action represents a structured action from the protocol.
type Action struct {
	Type    string            `json:"type"`
	Params  map[string]string `json:"params,omitempty"`
	Content string            `json:"content,omitempty"`
}

// Learning represents a new piece of knowledge.
type Learning struct {
	Topic   string `json:"topic"`
	Content string `json:"content"`
}

var (
	memoryRegex  = regexp.MustCompile(`(?is)\[AIGO_MEMORY\](.*?)(?=\[AIGO_|$)`)
	actionRegex  = regexp.MustCompile(`(?is)\[AIGO_ACTION\](.*?)(?=\[AIGO_|$)`)
	replyRegex   = regexp.MustCompile(`(?is)\[AIGO_REPLY\](.*?)(?=\[AIGO_|$)`)
	diaryRegex   = regexp.MustCompile(`(?is)\[AIGO_DIARY\](.*?)(?=\[AIGO_|$)`)
	learnRegex   = regexp.MustCompile(`(?is)\[AIGO_LEARN\](.*?)(?=\[AIGO_|$)`)
)

// Parse extracts structured data from AI response.
func Parse(raw string) *ParsedResponse {
	result := &ParsedResponse{}

	// Extract memory
	if match := memoryRegex.FindStringSubmatch(raw); len(match) > 1 {
		lines := strings.Split(strings.TrimSpace(match[1]), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" && !strings.HasPrefix(line, "[") && len(line) > 3 {
				result.Memory = append(result.Memory, line)
			}
		}
	}

	// Extract actions
	if match := actionRegex.FindStringSubmatch(raw); len(match) > 1 {
		jsonMatches := regexp.MustCompile(`\{[^{}]*\}`).FindAllString(match[1], -1)
		for _, jsonStr := range jsonMatches {
			var action Action
			if err := json.Unmarshal([]byte(jsonStr), &action); err == nil {
				result.Actions = append(result.Actions, action)
			}
		}
	}

	// Extract diary
	if match := diaryRegex.FindStringSubmatch(raw); len(match) > 1 {
		result.Diary = strings.TrimSpace(match[1])
	}

	// Extract learnings
	if match := learnRegex.FindStringSubmatch(raw); len(match) > 1 {
		lines := strings.Split(strings.TrimSpace(match[1]), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "[") {
				continue
			}
			// Format: "topic: content" or just "content"
			if idx := strings.Index(line, ": "); idx > 0 && idx < 50 {
				result.Learn = append(result.Learn, Learning{
					Topic:   line[:idx],
					Content: line[idx+2:],
				})
			} else {
				result.Learn = append(result.Learn, Learning{
					Topic:   "general",
					Content: line,
				})
			}
		}
	}

	// Extract reply
	if match := replyRegex.FindStringSubmatch(raw); len(match) > 1 {
		result.Reply = strings.TrimSpace(match[1])
	} else {
		// No markers - use entire response as reply
		reply := raw
		// Strip any protocol markers
		reply = memoryRegex.ReplaceAllString(reply, "")
		reply = actionRegex.ReplaceAllString(reply, "")
		reply = diaryRegex.ReplaceAllString(reply, "")
		reply = learnRegex.ReplaceAllString(reply, "")
		result.Reply = strings.TrimSpace(reply)
	}

	return result
}

// BuildSystemInstruction returns the protocol instruction for system prompt injection.
func BuildSystemInstruction() string {
	return `## Output Protocol

When responding, use these structured markers for clarity:

[AIGO_MEMORY]
📌 Important facts to remember (one per line)

[AIGO_REPLY]
Your actual response to the user

[AIGO_ACTION]
{"type": "search", "params": {"query": "something"}}

[AIGO_DIARY]
Personal diary entry (for self-reflection)

[AIGO_LEARN]
topic: new knowledge or learning

You can use any combination of markers. If you don't use any, your entire response becomes AIGO_REPLY.
For Indonesian users, mix Indonesian with English technical terms naturally.`
}

// ParseActions validates action types.
func ParseActions(actions []Action) []Action {
	allowedTypes := map[string]bool{
		"search":   true,
		"fetch":    true,
		"calculate": true,
		"translate": true,
		"remember": true,
	}

	var valid []Action
	for _, a := range actions {
		if allowedTypes[a.Type] {
			valid = append(valid, a)
		}
	}
	return valid
}

// FormatMemoryForPrompt formats parsed memories for prompt injection.
func FormatMemoryForPrompt(memories []string) string {
	if len(memories) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("\n[Memory Notes]\n")
	for _, m := range memories {
		sb.WriteString(fmt.Sprintf("• %s\n", m))
	}
	return sb.String()
}
