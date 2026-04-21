// Package multiagenttools implements tools for the multi-agent roundtable system.
package multiagenttools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/hermes-v2/aigo/internal/multiagent"
	"github.com/hermes-v2/aigo/internal/tools"
)

// RegisterMultiAgentTools registers roundtable tools in the registry.
func RegisterMultiAgentTools(reg *tools.Registry, brainFunc multiagent.BrainFunc) {
	rt := &roundtableState{brainFunc: brainFunc}
	reg.Register(&RoundtableStartTool{rt: rt})
	reg.Register(&RoundtableStatusTool{rt: rt})
	reg.Register(&RoundtableSummaryTool{rt: rt})
}

// roundtableState holds the shared state for roundtable tools.
type roundtableState struct {
	mu          sync.RWMutex
	brainFunc   multiagent.BrainFunc
	manager     *multiagent.MultiAgent
	messages    []string
	lastSummary string
	active      bool
	topic       string
	team        string
}

// captureSend returns a SendFunc that buffers all messages.
func (rt *roundtableState) captureSend() multiagent.SendFunc {
	return func(msg string) error {
		rt.mu.Lock()
		defer rt.mu.Unlock()
		rt.messages = append(rt.messages, msg)
		return nil
	}
}

// --- roundtable_start ---
type RoundtableStartTool struct {
	rt *roundtableState
}

func (t *RoundtableStartTool) Name() string        { return "roundtable_start" }
func (t *RoundtableStartTool) Description() string { return "Start a multi-agent roundtable discussion." }
func (t *RoundtableStartTool) Annotations() tools.Annotations {
	return tools.Annotations{Destructive: false, ReadOnly: false, SideEffects: []string{"network"}}
}
func (t *RoundtableStartTool) Schema() tools.ToolSchema {
	return tools.ToolSchema{
		Type: "function",
		Function: tools.ToolFunctionSchema{
			Name:        "roundtable_start",
			Description: "Start a multi-agent roundtable discussion on a topic. Agents with different perspectives will debate and collaborate.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"topic": map[string]string{
						"type":        "string",
						"description": "The topic or question for the team to discuss",
					},
					"team": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"tech", "debate", "creative", "business"},
						"description": "Team preset to use (tech, debate, creative, or business)",
					},
					"max_rounds": map[string]interface{}{
						"type":        "integer",
						"description": "Maximum number of discussion rounds (default 3)",
					},
				},
				"required": []string{"topic", "team"},
			},
		},
	}
}
func (t *RoundtableStartTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	topic, _ := args["topic"].(string)
	if topic == "" {
		return "", fmt.Errorf("topic is required")
	}

	team, _ := args["team"].(string)
	if team == "" {
		team = "debate"
	}

	maxRounds := 3
	if mr, ok := args["max_rounds"].(float64); ok && mr > 0 {
		maxRounds = int(mr)
	}

	// Check if a discussion is already active
	t.rt.mu.RLock()
	if t.rt.active {
		t.rt.mu.RUnlock()
		return "", fmt.Errorf("a roundtable discussion is already active on topic: %s", t.rt.topic)
	}
	t.rt.mu.RUnlock()

	// Reset state
	t.rt.mu.Lock()
	t.rt.active = true
	t.rt.topic = topic
	t.rt.team = team
	t.rt.messages = nil
	t.rt.lastSummary = ""
	t.rt.mu.Unlock()

	sendFn := t.rt.captureSend()
	mgr := multiagent.New(t.rt.brainFunc, sendFn)

	t.rt.mu.Lock()
	t.rt.manager = mgr
	t.rt.mu.Unlock()

	// Run the conversation (this blocks through all rounds)
	err := mgr.StartConversation(topic, team, maxRounds)

	t.rt.mu.Lock()
	t.rt.active = false
	t.rt.manager = nil
	messages := make([]string, len(t.rt.messages))
	copy(messages, t.rt.messages)
	t.rt.mu.Unlock()

	if err != nil {
		return fmt.Sprintf("⚠️ Roundtable error: %v\n\nPartial transcript:\n%s", err, strings.Join(messages, "\n")), nil
	}

	// Build transcript
	transcript := strings.Join(messages, "\n")
	return fmt.Sprintf("🎭 **Roundtable Complete**\n\n📋 Topic: %s | Team: %s | Rounds: %d\n\n%s", topic, team, maxRounds, transcript), nil
}

// --- roundtable_status ---
type RoundtableStatusTool struct {
	rt *roundtableState
}

func (t *RoundtableStatusTool) Name() string        { return "roundtable_status" }
func (t *RoundtableStatusTool) Description() string { return "Check if a roundtable discussion is active." }
func (t *RoundtableStatusTool) Annotations() tools.Annotations {
	return tools.Annotations{Destructive: false, ReadOnly: true}
}
func (t *RoundtableStatusTool) Schema() tools.ToolSchema {
	return tools.ToolSchema{
		Type: "function",
		Function: tools.ToolFunctionSchema{
			Name:        "roundtable_status",
			Description: "Check the current status of a roundtable discussion.",
			Parameters: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
	}
}
func (t *RoundtableStatusTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	t.rt.mu.RLock()
	defer t.rt.mu.RUnlock()

	result := map[string]interface{}{
		"active":        t.rt.active,
		"topic":         t.rt.topic,
		"team":          t.rt.team,
		"message_count": len(t.rt.messages),
	}

	if t.rt.manager != nil {
		status := t.rt.manager.GetStatus()
		for k, v := range status {
			result[k] = v
		}
	}

	b, _ := json.Marshal(result)
	return string(b), nil
}

// --- roundtable_summary ---
type RoundtableSummaryTool struct {
	rt *roundtableState
}

func (t *RoundtableSummaryTool) Name() string        { return "roundtable_summary" }
func (t *RoundtableSummaryTool) Description() string { return "Get a summary from the last roundtable discussion." }
func (t *RoundtableSummaryTool) Annotations() tools.Annotations {
	return tools.Annotations{Destructive: false, ReadOnly: true}
}
func (t *RoundtableSummaryTool) Schema() tools.ToolSchema {
	return tools.ToolSchema{
		Type: "function",
		Function: tools.ToolFunctionSchema{
			Name:        "roundtable_summary",
			Description: "Get the consensus summary from the last completed roundtable discussion.",
			Parameters: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
	}
}
func (t *RoundtableSummaryTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	t.rt.mu.RLock()
	defer t.rt.mu.RUnlock()

	if t.rt.active {
		return "", fmt.Errorf("roundtable is still active. Wait for it to complete")
	}

	if len(t.rt.messages) == 0 {
		return "No roundtable discussions have been run yet.", nil
	}

	return fmt.Sprintf("📋 Last roundtable: %s (team: %s)\n\n%s",
		t.rt.topic, t.rt.team, strings.Join(t.rt.messages, "\n")), nil
}
