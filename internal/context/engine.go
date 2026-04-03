package context

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/ahmad-ubaidillah/aigo/internal/memory"
	"github.com/ahmad-ubaidillah/aigo/pkg/types"
)

type L0Item struct {
	Summary    string
	Tokens     int
	Timestamp  time.Time
	Importance float64
}

type L1Item struct {
	Fact       string
	Category   string
	Tokens     int
	Confidence float64
	LastUsed   time.Time
}

type ToolEntry struct {
	Name          string
	InputSummary  string
	OutputSummary string
	Tokens        int
	Timestamp     time.Time
}

type SessionState struct {
	HotFiles       map[string]int // filepath -> access count
	ActiveErrors   []string       // last 5 errors
	LastCommands   []string       // last 20 commands
	InferredTask   string
	InferredDomain string
	// Legacy fields kept for compatibility
	LastError      string
	TaskGoal       string
	ToolHistory    []ToolEntry
	CommandEntries []CommandEntry
	TurnCount      int
	ErrorCount     int
}

type ContextEngine struct {
	db         *memory.SessionDB
	cfg        types.Config
	state      SessionState
	l0Items    []L0Item
	l1Items    []L1Item
	tokenCount int
}

func NewContextEngine(db *memory.SessionDB, cfg types.Config) *ContextEngine {
	if cfg.Memory.TokenBudget <= 0 {
		cfg.Memory.TokenBudget = 8000
	}
	if cfg.Memory.MaxL0Items <= 0 {
		cfg.Memory.MaxL0Items = 20
	}
	if cfg.Memory.MaxL1Items <= 0 {
		cfg.Memory.MaxL1Items = 50
	}

	return &ContextEngine{
		db:  db,
		cfg: cfg,
		state: SessionState{
			HotFiles:     make(map[string]int),
			ActiveErrors: make([]string, 0),
			LastCommands: make([]string, 0),
			ToolHistory:  make([]ToolEntry, 0),
		},
		l0Items: make([]L0Item, 0),
		l1Items: make([]L1Item, 0),
	}
}

func estimateTokens(text string) int {
	return int(math.Ceil(float64(len(text)) / 4.0))
}

func (e *ContextEngine) AddL0(summary string) {
	item := L0Item{
		Summary:    summary,
		Tokens:     estimateTokens(summary),
		Timestamp:  time.Now(),
		Importance: 1.0,
	}
	e.l0Items = append(e.l0Items, item)
	e.tokenCount += item.Tokens
	e.pruneL0()
}

func (e *ContextEngine) AddL1(fact, category string) {
	item := L1Item{
		Fact:       fact,
		Category:   category,
		Tokens:     estimateTokens(fact),
		Confidence: 1.0,
		LastUsed:   time.Now(),
	}
	e.l1Items = append(e.l1Items, item)
	e.tokenCount += item.Tokens
	e.pruneL1()
}

func (e *ContextEngine) pruneL0() {
	if len(e.l0Items) <= e.cfg.Memory.MaxL0Items {
		return
	}
	e.l0Items = e.l0Items[len(e.l0Items)-e.cfg.Memory.MaxL0Items:]
}

func (e *ContextEngine) pruneL1() {
	if len(e.l1Items) <= e.cfg.Memory.MaxL1Items {
		return
	}
	e.l1Items = e.l1Items[len(e.l1Items)-e.cfg.Memory.MaxL1Items:]
}

func (e *ContextEngine) pruneIfNeeded() {
	if !e.cfg.Memory.SmartPrune {
		return
	}

	budget := e.cfg.Memory.TokenBudget
	if e.tokenCount <= budget {
		return
	}

	e.smartPrune(budget)
}

func (e *ContextEngine) smartPrune(targetBudget int) {
	now := time.Now()

	for _, item := range e.l0Items {
		age := now.Sub(item.Timestamp).Minutes()
		item.Importance = 1.0 / (1.0 + age/60.0)
	}

	sort.Slice(e.l0Items, func(i, j int) bool {
		return e.l0Items[i].Importance > e.l0Items[j].Importance
	})

	var kept []L0Item
	var keptTokens int
	for _, item := range e.l0Items {
		if keptTokens+item.Tokens > targetBudget/2 {
			break
		}
		kept = append(kept, item)
		keptTokens += item.Tokens
	}

	e.l0Items = kept
	e.tokenCount = keptTokens

	for i := range e.l1Items {
		e.l1Items[i].LastUsed = now
	}

	relevance := make(map[string]float64)
	for _, item := range e.l1Items {
		relevance[item.Category] += item.Confidence
	}

	sort.Slice(e.l1Items, func(i, j int) bool {
		return e.l1Items[i].Confidence > e.l1Items[j].Confidence
	})

	kept = nil
	keptTokens = 0
	for _, item := range e.l0Items {
		if keptTokens+item.Tokens > targetBudget {
			break
		}
		kept = append(kept, item)
		keptTokens += item.Tokens
	}
}

func (e *ContextEngine) RecordToolUse(name, input, output string) {
	entry := ToolEntry{
		Name:          name,
		InputSummary:  input,
		OutputSummary: output,
		Tokens:        estimateTokens(input + output),
		Timestamp:     time.Now(),
	}
	e.state.ToolHistory = append(e.state.ToolHistory, entry)
	e.tokenCount += entry.Tokens

	maxHistory := 20
	if len(e.state.ToolHistory) > maxHistory {
		removed := e.state.ToolHistory[:len(e.state.ToolHistory)-maxHistory]
		for _, r := range removed {
			e.tokenCount -= r.Tokens
		}
		e.state.ToolHistory = e.state.ToolHistory[len(e.state.ToolHistory)-maxHistory:]
	}
	e.pruneIfNeeded()
}

func (e *ContextEngine) RecordError(err string) {
	e.state.LastError = err
	e.state.ErrorCount++
	// Add to ActiveErrors, keep last 5
	e.state.ActiveErrors = append(e.state.ActiveErrors, err)
	if len(e.state.ActiveErrors) > 5 {
		e.state.ActiveErrors = e.state.ActiveErrors[len(e.state.ActiveErrors)-5:]
	}
}

func (e *ContextEngine) SetTaskGoal(goal string) {
	e.state.TaskGoal = goal
}

// RecordFileAccess increments the access count for a file path.
func (e *ContextEngine) RecordFileAccess(path string) {
	e.state.HotFiles[path]++
}

// RecordCommand adds a command to the LastCommands history, keeping last 20.
func (e *ContextEngine) RecordCommand(cmd string) {
	e.state.LastCommands = append(e.state.LastCommands, cmd)
	if len(e.state.LastCommands) > 20 {
		e.state.LastCommands = e.state.LastCommands[len(e.state.LastCommands)-20:]
	}
}

// InferTask analyzes LastCommands to detect the task type.
func (e *ContextEngine) InferTask() string {
	for _, cmd := range e.state.LastCommands {
		lower := strings.ToLower(cmd)
		if strings.HasPrefix(lower, "git ") {
			e.state.InferredTask = "version_control"
			return e.state.InferredTask
		}
		if strings.HasPrefix(lower, "go ") {
			e.state.InferredTask = "go_development"
			return e.state.InferredTask
		}
		if strings.Contains(lower, "npm ") || strings.Contains(lower, "yarn ") || strings.Contains(lower, "node ") {
			e.state.InferredTask = "node_development"
			return e.state.InferredTask
		}
		if strings.HasPrefix(lower, "docker ") {
			e.state.InferredTask = "containerization"
			return e.state.InferredTask
		}
		if strings.Contains(lower, "python") || strings.HasPrefix(lower, "pip ") {
			e.state.InferredTask = "python_development"
			return e.state.InferredTask
		}
		if strings.Contains(lower, "test") {
			e.state.InferredTask = "testing"
			return e.state.InferredTask
		}
		if strings.Contains(lower, "build") {
			e.state.InferredTask = "build"
			return e.state.InferredTask
		}
	}
	if e.state.InferredTask == "" {
		e.state.InferredTask = "unknown"
	}
	return e.state.InferredTask
}

// InferDomain detects the project type from files in the working directory.
func (e *ContextEngine) InferDomain() string {
	// Check hot files for domain hints
	for path := range e.state.HotFiles {
		ext := strings.ToLower(path)
		if strings.HasSuffix(ext, ".go") {
			e.state.InferredDomain = "go"
			return e.state.InferredDomain
		}
		if strings.HasSuffix(ext, ".ts") || strings.HasSuffix(ext, ".js") || strings.HasSuffix(ext, ".tsx") || strings.HasSuffix(ext, ".jsx") {
			e.state.InferredDomain = "node"
			return e.state.InferredDomain
		}
		if strings.HasSuffix(ext, ".py") {
			e.state.InferredDomain = "python"
			return e.state.InferredDomain
		}
		if strings.HasSuffix(ext, ".rs") {
			e.state.InferredDomain = "rust"
			return e.state.InferredDomain
		}
		if strings.HasSuffix(ext, ".java") {
			e.state.InferredDomain = "java"
			return e.state.InferredDomain
		}
		if strings.HasSuffix(ext, ".rb") {
			e.state.InferredDomain = "ruby"
			return e.state.InferredDomain
		}
		// Check for project files
		if strings.Contains(path, "go.mod") {
			e.state.InferredDomain = "go"
			return e.state.InferredDomain
		}
		if strings.Contains(path, "package.json") {
			e.state.InferredDomain = "node"
			return e.state.InferredDomain
		}
		if strings.Contains(path, "requirements.txt") || strings.Contains(path, "pyproject.toml") {
			e.state.InferredDomain = "python"
			return e.state.InferredDomain
		}
		if strings.Contains(path, "Cargo.toml") {
			e.state.InferredDomain = "rust"
			return e.state.InferredDomain
		}
		if strings.Contains(path, "pom.xml") || strings.Contains(path, "build.gradle") {
			e.state.InferredDomain = "java"
			return e.state.InferredDomain
		}
	}

	if e.state.InferredDomain == "" {
		e.state.InferredDomain = "unknown"
	}
	return e.state.InferredDomain
}

// ContextBoost returns a boost score for a specific file path.
// Returns +0.1 if path is in HotFiles (access count > 2)
// Returns +0.25 if path has recent errors
func (e *ContextEngine) ContextBoost(path string) float64 {
	var boost float64

	// Check if path is a hot file (access count > 2)
	if count, exists := e.state.HotFiles[path]; exists && count > 2 {
		boost += 0.1
	}

	// Check if path has recent errors
	for _, err := range e.state.ActiveErrors {
		if strings.Contains(err, path) {
			boost += 0.25
			break
		}
	}

	return boost
}

// AddHotFile adds a file to the hot files list (legacy compatibility).
func (e *ContextEngine) AddHotFile(path string) {
	e.state.HotFiles[path]++
}

func (e *ContextEngine) GetTokenCount() int {
	return e.tokenCount
}

func (e *ContextEngine) GetBudgetUsage() float64 {
	budget := e.cfg.Memory.TokenBudget
	if budget <= 0 {
		return 0
	}
	return float64(e.tokenCount) / float64(budget)
}

// AddActiveError appends an error message to the ActiveErrors log,
// keeping only the most recent 5 errors.
func (e *ContextEngine) AddActiveError(err string) {
	e.state.ActiveErrors = append(e.state.ActiveErrors, err)
	if len(e.state.ActiveErrors) > 5 {
		e.state.ActiveErrors = e.state.ActiveErrors[len(e.state.ActiveErrors)-5:]
	}
}

// AddLastCommand records a user/system command in the history,
// keeping the last 20 commands.
func (e *ContextEngine) AddLastCommand(cmd string) {
	e.state.LastCommands = append(e.state.LastCommands, cmd)
	if len(e.state.LastCommands) > 20 {
		e.state.LastCommands = e.state.LastCommands[len(e.state.LastCommands)-20:]
	}
}

// SetInferredTask stores an auto-detected task for the session.
func (e *ContextEngine) SetInferredTask(task string) {
	e.state.InferredTask = task
}

// SetInferredDomain stores an inferred project type/domain for the session.
func (e *ContextEngine) SetInferredDomain(domain string) {
	e.state.InferredDomain = domain
}

// ContextBoostScore returns a computed boost score based on context.
// Caps at 2.0.
func (e *ContextEngine) ContextBoostScore() float64 {
	score := 0.0
	// hot files contribution: +0.1 per file, capped at 0.5 for 5+ files
	if len(e.state.HotFiles) > 0 {
		boost := 0.1 * float64(len(e.state.HotFiles))
		if boost > 0.5 {
			boost = 0.5
		}
		score += boost
	}
	// active errors contribution: +0.25 per error, capped at 0.75 for 3+ errors
	if len(e.state.ActiveErrors) > 0 {
		boost := 0.25 * float64(len(e.state.ActiveErrors))
		if boost > 0.75 {
			boost = 0.75
		}
		score += boost
	}
	// inferred task/domain
	if e.state.InferredTask != "" {
		score += 0.2
	}
	if e.state.InferredDomain != "" {
		score += 0.1
	}
	if score > 2.0 {
		score = 2.0
	}
	return score
}

// GetSessionState returns a copy of the current session state.
func (e *ContextEngine) GetSessionState() SessionState {
	return e.state
}

func (e *ContextEngine) BuildPrompt(task string) string {
	var b strings.Builder

	tokenBudget := e.cfg.Memory.TokenBudget
	availableTokens := tokenBudget - 500

	b.WriteString("## L0 Recent Turn Summaries\n")
	for _, item := range e.l0Items {
		if availableTokens <= 0 {
			break
		}
		line := fmt.Sprintf("- %s\n", item.Summary)
		lineTokens := estimateTokens(line)
		if availableTokens < lineTokens {
			break
		}
		b.WriteString(line)
		availableTokens -= lineTokens
	}

	b.WriteString("\n## L1 Relevant Facts\n")
	taskLower := strings.ToLower(task)
	for _, item := range e.l1Items {
		if availableTokens <= 0 {
			break
		}
		relevant := task == "" ||
			strings.Contains(taskLower, strings.ToLower(item.Category)) ||
			strings.Contains(taskLower, strings.ToLower(item.Fact))
		if !relevant && item.Confidence < 0.5 {
			continue
		}
		line := fmt.Sprintf("- [%s] %s\n", item.Category, item.Fact)
		lineTokens := estimateTokens(line)
		if availableTokens < lineTokens {
			break
		}
		b.WriteString(line)
		availableTokens -= lineTokens
	}

	b.WriteString("\n## Session State\n")
	b.WriteString(fmt.Sprintf("Task Goal: %s\n", e.state.TaskGoal))
	b.WriteString(fmt.Sprintf("Token Budget: %d/%d (%.1f%%)\n",
		e.tokenCount, tokenBudget, e.GetBudgetUsage()*100))
	if e.state.LastError != "" {
		b.WriteString(fmt.Sprintf("Last Error: %s\n", e.state.LastError))
	}
	if len(e.state.HotFiles) > 0 {
		b.WriteString("Hot Files:\n")
		for f, count := range e.state.HotFiles {
			b.WriteString(fmt.Sprintf("  - %s (accessed %d times)\n", f, count))
		}
	}

	// Optional: Active errors section
	if len(e.state.ActiveErrors) > 0 {
		b.WriteString("Active Errors:\n")
		for _, er := range e.state.ActiveErrors {
			b.WriteString(fmt.Sprintf("  - %s\n", er))
		}
	}

	// Optional: Last commands section
	if len(e.state.LastCommands) > 0 {
		b.WriteString("Last Commands:\n")
		for _, c := range e.state.LastCommands {
			b.WriteString(fmt.Sprintf("  - %s\n", c))
		}
	}

	// Optional: Inferred task/domain
	if e.state.InferredTask != "" {
		b.WriteString(fmt.Sprintf("Inferred Task: %s\n", e.state.InferredTask))
	}
	if e.state.InferredDomain != "" {
		b.WriteString(fmt.Sprintf("Inferred Domain: %s\n", e.state.InferredDomain))
	}

	// Context boost score
	b.WriteString(fmt.Sprintf("Context Boost Score: %.2f\n", e.ContextBoostScore()))

	b.WriteString("\n## Tool History (Last 5)\n")
	start := 0
	history := e.state.ToolHistory
	if len(history) > 5 {
		start = len(history) - 5
	}
	for i := start; i < len(history); i++ {
		entry := history[i]
		b.WriteString(fmt.Sprintf("- %s: input=%s, output=%s\n",
			entry.Name, entry.InputSummary, entry.OutputSummary))
	}

	b.WriteString(fmt.Sprintf("\n## Current Task\n%s\n", task))

	return b.String()
}

func (e *ContextEngine) Compress() {
	if len(e.l0Items) == 0 {
		return
	}

	var combined strings.Builder
	for _, item := range e.l0Items {
		combined.WriteString(item.Summary)
		combined.WriteString("; ")
	}

	e.l0Items = []L0Item{
		{
			Summary:    fmt.Sprintf("Compressed history: %s", combined.String()),
			Tokens:     estimateTokens(combined.String()),
			Timestamp:  time.Now(),
			Importance: 0.5,
		},
	}

	e.tokenCount = e.l0Items[0].Tokens

	maxL0 := e.cfg.Memory.MaxL0Items
	if maxL0 <= 0 {
		maxL0 = 20
	}
	if len(e.l0Items) > maxL0 {
		e.l0Items = e.l0Items[len(e.l0Items)-maxL0:]
	}

	maxL1 := 5
	if maxL1 <= 0 {
		maxL1 = 5
	}
	if len(e.l1Items) > maxL1 {
		e.l1Items = e.l1Items[len(e.l1Items)-maxL1:]
	}
}

func (e *ContextEngine) IncrementTurns() {
	e.state.TurnCount++
	if e.cfg.Memory.AutoCompress && e.state.TurnCount%10 == 0 {
		e.Compress()
	}
}

func (e *ContextEngine) GetContextSummary() string {
	return fmt.Sprintf("L0=%d items (%d tokens), L1=%d items, Budget=%.1f%%",
		len(e.l0Items), e.tokenCount, len(e.l1Items), e.GetBudgetUsage()*100)
}

type CommandEntry struct {
	Command  string        `json:"command"`
	ExitCode int           `json:"exit_code"`
	Duration time.Duration `json:"duration"`
}

func (e *ContextEngine) TrackFileAccess(path string) {
	e.state.HotFiles[path]++
	if len(e.state.HotFiles) > 50 {
		sorted := make([]string, 0, len(e.state.HotFiles))
		for k := range e.state.HotFiles {
			sorted = append(sorted, k)
		}
		sort.Slice(sorted, func(i, j int) bool {
			return e.state.HotFiles[sorted[i]] > e.state.HotFiles[sorted[j]]
		})
		keep := make(map[string]int)
		for _, k := range sorted[:25] {
			keep[k] = e.state.HotFiles[k]
		}
		e.state.HotFiles = keep
	}
}

func (e *ContextEngine) AddCommand(cmd string, exitCode int, duration time.Duration) {
	entry := CommandEntry{Command: cmd, ExitCode: exitCode, Duration: duration}
	e.state.CommandEntries = append(e.state.CommandEntries, entry)
	if len(e.state.CommandEntries) > 20 {
		e.state.CommandEntries = e.state.CommandEntries[len(e.state.CommandEntries)-20:]
	}
}

func (e *ContextEngine) InferTaskFromCommands() string {
	for _, c := range e.state.CommandEntries {
		lower := strings.ToLower(c.Command)
		if strings.HasPrefix(lower, "git ") {
			return "version_control"
		}
		if strings.HasPrefix(lower, "go ") {
			return "go_development"
		}
		if strings.Contains(lower, "npm ") || strings.Contains(lower, "yarn ") {
			return "package_management"
		}
		if strings.HasPrefix(lower, "docker ") {
			return "containerization"
		}
		if strings.Contains(lower, "test") {
			return "testing"
		}
	}
	return "unknown"
}

func (e *ContextEngine) GetHotFiles() []string {
	type kv struct {
		K string
		V int
	}
	var sorted []kv
	for k, v := range e.state.HotFiles {
		sorted = append(sorted, kv{k, v})
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].V > sorted[j].V
	})
	result := make([]string, len(sorted))
	for i, kv := range sorted {
		result[i] = kv.K
	}
	return result
}

func (e *ContextEngine) GetCommandHistory() []CommandEntry {
	out := make([]CommandEntry, len(e.state.CommandEntries))
	copy(out, e.state.CommandEntries)
	return out
}

func (e *ContextEngine) SaveState(sessionID string) error {
	data, err := json.Marshal(e.state)
	if err != nil {
		return fmt.Errorf("marshal state: %w", err)
	}
	_, err = e.db.DB().Exec(
		`INSERT OR REPLACE INTO session_state (session_id, state_json) VALUES (?, ?)`,
		sessionID, string(data),
	)
	if err != nil {
		return fmt.Errorf("save state: %w", err)
	}
	return nil
}

func (e *ContextEngine) LoadState(sessionID string) error {
	var stateJSON string
	err := e.db.DB().QueryRow(
		`SELECT state_json FROM session_state WHERE session_id = ?`,
		sessionID,
	).Scan(&stateJSON)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return fmt.Errorf("load state: %w", err)
	}
	return json.Unmarshal([]byte(stateJSON), &e.state)
}

func InitSessionStateTable(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS session_state (
		session_id TEXT PRIMARY KEY,
		state_json TEXT NOT NULL
	)`)
	return err
}
