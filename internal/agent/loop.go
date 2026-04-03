// Package agent provides the autonomous agent loop for iterative task execution.
package agent

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"

	aigoctx "github.com/ahmad-ubaidillah/aigo/internal/context"
	"github.com/ahmad-ubaidillah/aigo/internal/intent"
	"github.com/ahmad-ubaidillah/aigo/internal/llm"
	"github.com/ahmad-ubaidillah/aigo/internal/memory"
	"github.com/ahmad-ubaidillah/aigo/internal/tools"
	"github.com/ahmad-ubaidillah/aigo/pkg/types"
)

const (
	defaultMaxIterations = 50
	defaultDoomThreshold = 3
	defaultTokenBudget   = 100000
	tokensPerChar        = 4
	doomHistorySize      = 10
)

// AgentLoop implements an autonomous loop that iteratively classifies, plans,
// executes tools, and self-corrects until the task is done or limits are hit.
type AgentLoop struct {
	registry      *tools.ToolRegistry
	checker       *tools.PermissionChecker
	ctxEngine     *aigoctx.ContextEngine
	db            *memory.SessionDB
	classifier    *intent.Classifier
	llmClient     llm.LLMClient
	maxIterations int
	doomThreshold int
	tokenBudget   int
}

// NewAgentLoop creates a new autonomous agent loop with the given dependencies.
func NewAgentLoop(
	registry *tools.ToolRegistry,
	checker *tools.PermissionChecker,
	ctxEngine *aigoctx.ContextEngine,
	db *memory.SessionDB,
	llmClient llm.LLMClient,
	maxIterations int,
) *AgentLoop {
	if maxIterations <= 0 {
		maxIterations = defaultMaxIterations
	}
	return &AgentLoop{
		registry:      registry,
		checker:       checker,
		ctxEngine:     ctxEngine,
		db:            db,
		classifier:    intent.NewClassifier(types.Config{}),
		llmClient:     llmClient,
		maxIterations: maxIterations,
		doomThreshold: defaultDoomThreshold,
		tokenBudget:   defaultTokenBudget,
	}
}

// Run executes the autonomous loop for a given task until completion or limits.
func (l *AgentLoop) Run(ctx context.Context, sessionID, task string) (*types.ToolResult, error) {
	if err := l.recordTaskStart(sessionID, task); err != nil {
		return nil, fmt.Errorf("record task start: %w", err)
	}

	l.ctxEngine.SetTaskGoal(task)

	var consecutiveErrors int
	doomTracker := newDoomTracker(doomHistorySize, l.doomThreshold)

	for iteration := 1; iteration <= l.maxIterations; iteration++ {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("context cancelled: %w", ctx.Err())
		default:
		}

		l.ctxEngine.IncrementTurns()
		classification := l.classifier.Classify(task)
		_ = l.ctxEngine.BuildPrompt(task)

		toolName, params := l.selectTool(classification, task, iteration)
		perm := l.checker.Check(toolName)

		if perm == tools.PermDeny {
			l.ctxEngine.AddL0(fmt.Sprintf("Tool %s denied by permissions", toolName))
			continue
		}

		if perm == tools.PermAsk {
			l.ctxEngine.AddL0(fmt.Sprintf("Tool %s requires user confirmation (skipped)", toolName))
			continue
		}

		callKey := makeToolCallKey(toolName, params)
		if doomTracker.isDoomed(callKey) {
			return &types.ToolResult{
				Success: false,
				Error:   fmt.Sprintf("doom loop detected: tool %s repeated %d times", toolName, l.doomThreshold),
			}, nil
		}
		doomTracker.record(callKey)

		result, err := l.registry.Execute(ctx, toolName, params)
		if err != nil {
			consecutiveErrors++
			l.ctxEngine.RecordError(err.Error())
			if consecutiveErrors >= maxConsecutiveErrors {
				return nil, fmt.Errorf("max consecutive errors (%d) reached: %w", maxConsecutiveErrors, err)
			}
			continue
		}
		consecutiveErrors = 0

		l.recordToolResult(toolName, result)

		if l.shouldCompact() {
			l.ctxEngine.Compress()
			l.ctxEngine.AddL0("[compaction triggered]")
		}

		if l.isTaskComplete(result) {
			_ = l.db.AddMessage(sessionID, "assistant", result.Output)
			_ = l.db.UpdateSessionActivity(sessionID)
			return result, nil
		}

		task = l.refineTask(task, result)
	}

	return &types.ToolResult{
		Success: false,
		Error:   fmt.Sprintf("max iterations (%d) reached without completion", l.maxIterations),
	}, nil
}

// recordTaskStart records the initial task in the session.
func (l *AgentLoop) recordTaskStart(sessionID, task string) error {
	if err := l.db.AddMessage(sessionID, "user", task); err != nil {
		return err
	}
	return l.db.UpdateSessionActivity(sessionID)
}

// selectTool chooses the next tool based on classification and current state.
func (l *AgentLoop) selectTool(classification intent.Classification, task string, iteration int) (string, map[string]any) {
	toolName := l.mapIntentToTool(classification.Intent)
	params := l.buildToolParams(toolName, task, classification)

	if l.registry.Get(toolName) == nil {
		toolName = "bash"
		params = map[string]any{"command": task}
	}

	return toolName, params
}

// selectToolWithLLM uses the LLM to choose the next tool based on the task and context.
// Falls back to hardcoded selectTool if LLM is unavailable or returns invalid response.
func (l *AgentLoop) selectToolWithLLM(ctx context.Context, task string, classification intent.Classification, iteration int) (string, map[string]any) {
	if l.llmClient == nil {
		return l.selectTool(classification, task, iteration)
	}

	availableTools := l.registry.List()
	contextSummary := l.ctxEngine.BuildPrompt(task)

	prompt := BuildToolSelectionPrompt(task, availableTools, contextSummary)

	messages := []llm.Message{
		{Role: "system", Content: "You are a tool selection assistant. Your job is to choose the best tool for the given task."},
		{Role: "user", Content: prompt},
	}

	resp, err := l.llmClient.Chat(ctx, messages)
	if err != nil {
		return l.selectTool(classification, task, iteration)
	}

	toolName, params, err := ParseLLMToolResponse(resp.Content)
	if err != nil {
		return l.selectTool(classification, task, iteration)
	}

	if l.registry.Get(toolName) == nil {
		return l.selectTool(classification, task, iteration)
	}

	return toolName, params
}

// mapIntentToTool maps an intent category to a tool name.
func (l *AgentLoop) mapIntentToTool(intentType string) string {
	mapping := map[string]string{
		types.IntentCoding:     "bash",
		types.IntentWeb:        "websearch",
		types.IntentFile:       "read",
		types.IntentGateway:    "bash",
		types.IntentMemory:     "bash",
		types.IntentAutomation: "bash",
		types.IntentSkill:      "bash",
		types.IntentResearch:   "websearch",
		types.IntentHTTPCall:   "webfetch",
		types.IntentBrowser:    "bash",
		types.IntentPython:     "bash",
		types.IntentGeneral:    "bash",
	}

	if tool, ok := mapping[intentType]; ok {
		return tool
	}
	return "bash"
}

// buildToolParams constructs parameters for the selected tool.
func (l *AgentLoop) buildToolParams(toolName, task string, classification intent.Classification) map[string]any {
	switch toolName {
	case "bash":
		return map[string]any{"command": task}
	case "websearch":
		return map[string]any{"query": task, "num_results": 5}
	case "webfetch":
		return map[string]any{"url": extractURL(task)}
	case "read":
		return map[string]any{"path": extractPath(task, classification.Workspace)}
	case "write":
		return map[string]any{"path": extractPath(task, classification.Workspace), "content": task}
	case "glob":
		return map[string]any{"pattern": extractPattern(task)}
	case "grep":
		return map[string]any{"pattern": task}
	default:
		return map[string]any{"input": task}
	}
}

// recordToolResult stores the tool execution result in context.
func (l *AgentLoop) recordToolResult(toolName string, result *types.ToolResult) {
	summary := truncate(result.Output, 200)
	if !result.Success {
		summary = "ERROR: " + result.Error
	}
	l.ctxEngine.AddL0(fmt.Sprintf("%s: %s", toolName, summary))
	l.ctxEngine.RecordToolUse(toolName, truncate(result.Output, 100), summary)
}

// shouldCompact checks if context compaction is needed.
func (l *AgentLoop) shouldCompact() bool {
	return l.ctxEngine.GetTokenCount() > l.tokenBudget
}

// isTaskComplete determines if the task has been successfully completed.
func (l *AgentLoop) isTaskComplete(result *types.ToolResult) bool {
	if !result.Success {
		return false
	}
	output := strings.ToLower(result.Output)
	completeIndicators := []string{
		"task complete", "done", "finished", "successfully",
		"completed", "all tests pass", "build succeeded",
	}
	for _, indicator := range completeIndicators {
		if strings.Contains(output, indicator) {
			return true
		}
	}
	return false
}

// refineTask updates the task based on the current result for the next iteration.
func (l *AgentLoop) refineTask(originalTask string, result *types.ToolResult) string {
	if result.Success && len(result.Output) > 0 {
		return fmt.Sprintf("Continue working on: %s. Last output: %s",
			originalTask, truncate(result.Output, 100))
	}
	return originalTask
}

// doomTracker tracks tool calls to detect repeated patterns.
type doomTracker struct {
	history   []string
	maxSize   int
	threshold int
}

func newDoomTracker(maxSize, threshold int) *doomTracker {
	return &doomTracker{
		history:   make([]string, 0, maxSize),
		maxSize:   maxSize,
		threshold: threshold,
	}
}

func (d *doomTracker) record(key string) {
	d.history = append(d.history, key)
	if len(d.history) > d.maxSize {
		d.history = d.history[1:]
	}
}

func (d *doomTracker) isDoomed(key string) bool {
	count := 0
	for _, k := range d.history {
		if k == key {
			count++
		}
	}
	return count >= d.threshold
}

// makeToolCallKey creates a unique key for a tool call (name + sorted params).
func makeToolCallKey(toolName string, params map[string]any) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var b strings.Builder
	b.WriteString(toolName)
	b.WriteString("|")
	for _, k := range keys {
		b.WriteString(k)
		b.WriteString("=")
		b.WriteString(fmt.Sprintf("%v", params[k]))
		b.WriteString(";")
	}

	hash := sha256.Sum256([]byte(b.String()))
	return hex.EncodeToString(hash[:8])
}

// extractURL attempts to extract a URL from the task text.
func extractURL(task string) string {
	words := strings.Fields(task)
	for _, w := range words {
		if strings.HasPrefix(w, "http://") || strings.HasPrefix(w, "https://") {
			return strings.TrimRight(w, ".,;:!?)")
		}
	}
	return ""
}

// extractPath attempts to extract a file path from the task.
func extractPath(task, workspace string) string {
	if workspace != "" {
		return workspace
	}
	words := strings.Fields(task)
	for _, w := range words {
		if strings.HasPrefix(w, "/") || strings.HasPrefix(w, "./") || strings.HasPrefix(w, "~/") {
			return strings.TrimRight(w, ".,;:!?)")
		}
	}
	return "."
}

// extractPattern attempts to extract a glob pattern from the task.
func extractPattern(task string) string {
	words := strings.Fields(task)
	for _, w := range words {
		if strings.Contains(w, "*") || strings.Contains(w, "?") {
			return strings.TrimRight(w, ".,;:!?)")
		}
	}
	return "*"
}
