package context

import (
	"fmt"
	"strings"
	"testing"

	"github.com/ahmad-ubaidillah/aigo/internal/memory"
	"github.com/ahmad-ubaidillah/aigo/pkg/types"
)

func TestNewContextEngine(t *testing.T) {
	cfg := types.Config{
		Memory: types.MemoryConfig{
			MaxL0Items:   20,
			MaxL1Items:   50,
			AutoCompress: true,
		},
	}

	db := &memory.SessionDB{}
	engine := NewContextEngine(db, cfg)

	if engine == nil {
		t.Fatal("engine should not be nil")
	}

	if len(engine.l0Items) != 0 {
		t.Errorf("expected 0 L0 items, got %d", len(engine.l0Items))
	}

	if len(engine.l1Items) != 0 {
		t.Errorf("expected 0 L1 items, got %d", len(engine.l1Items))
	}

	if len(engine.state.ToolHistory) != 0 {
		t.Errorf("expected empty tool history, got %d items", len(engine.state.ToolHistory))
	}
}

func TestAddL0(t *testing.T) {
	cfg := types.Config{
		Memory: types.MemoryConfig{
			MaxL0Items: 3,
		},
	}

	db := &memory.SessionDB{}
	engine := NewContextEngine(db, cfg)

	engine.AddL0("First summary")
	engine.AddL0("Second summary")
	engine.AddL0("Third summary")

	if len(engine.l0Items) != 3 {
		t.Errorf("expected 3 L0 items, got %d", len(engine.l0Items))
	}

	// Test overflow - should keep only last 3
	engine.AddL0("Fourth summary")

	if len(engine.l0Items) != 3 {
		t.Errorf("expected 3 L0 items after overflow, got %d", len(engine.l0Items))
	}

	if engine.l0Items[0].Summary != "Second summary" {
		t.Errorf("expected first item to be 'Second summary', got %s", engine.l0Items[0].Summary)
	}

	if engine.l0Items[2].Summary != "Fourth summary" {
		t.Errorf("expected last item to be 'Fourth summary', got %s", engine.l0Items[2].Summary)
	}
}

func TestAddL1(t *testing.T) {
	cfg := types.Config{
		Memory: types.MemoryConfig{
			MaxL1Items: 3,
		},
	}

	db := &memory.SessionDB{}
	engine := NewContextEngine(db, cfg)

	engine.AddL1("Fact 1", "category1")
	engine.AddL1("Fact 2", "category2")
	engine.AddL1("Fact 3", "category3")

	if len(engine.l1Items) != 3 {
		t.Errorf("expected 3 L1 items, got %d", len(engine.l1Items))
	}

	// Test overflow - should keep only last 3
	engine.AddL1("Fact 4", "category4")

	if len(engine.l1Items) != 3 {
		t.Errorf("expected 3 L1 items after overflow, got %d", len(engine.l1Items))
	}

	if engine.l1Items[0].Fact != "Fact 2" {
		t.Errorf("expected first item to be 'Fact 2', got %s", engine.l1Items[0].Fact)
	}
}

func TestRecordToolUse(t *testing.T) {
	cfg := types.Config{}
	db := &memory.SessionDB{}
	engine := NewContextEngine(db, cfg)

	engine.RecordToolUse("read", "file.go", "success")
	engine.RecordToolUse("write", "file.go", "created")

	if len(engine.state.ToolHistory) != 2 {
		t.Errorf("expected 2 tool entries, got %d", len(engine.state.ToolHistory))
	}

	// Test tool history limit (20 items)
	for i := 0; i < 25; i++ {
		engine.RecordToolUse("test", "input", "output")
	}

	if len(engine.state.ToolHistory) != 20 {
		t.Errorf("expected 20 tool entries (limit), got %d", len(engine.state.ToolHistory))
	}
}

func TestRecordError(t *testing.T) {
	cfg := types.Config{}
	db := &memory.SessionDB{}
	engine := NewContextEngine(db, cfg)

	engine.RecordError("first error")

	if engine.state.LastError != "first error" {
		t.Errorf("expected 'first error', got %s", engine.state.LastError)
	}

	if engine.state.ErrorCount != 1 {
		t.Errorf("expected error count 1, got %d", engine.state.ErrorCount)
	}

	engine.RecordError("second error")

	if engine.state.ErrorCount != 2 {
		t.Errorf("expected error count 2, got %d", engine.state.ErrorCount)
	}
}

func TestSetTaskGoal(t *testing.T) {
	cfg := types.Config{}
	db := &memory.SessionDB{}
	engine := NewContextEngine(db, cfg)

	goal := "Fix the authentication bug"
	engine.SetTaskGoal(goal)

	if engine.state.TaskGoal != goal {
		t.Errorf("expected goal %q, got %q", goal, engine.state.TaskGoal)
	}
}

func TestAddHotFile(t *testing.T) {
	cfg := types.Config{}
	db := &memory.SessionDB{}
	engine := NewContextEngine(db, cfg)

	engine.AddHotFile("/path/to/file1.go")
	engine.AddHotFile("/path/to/file2.go")

	if len(engine.state.HotFiles) != 2 {
		t.Errorf("expected 2 hot files, got %d", len(engine.state.HotFiles))
	}

	// Test deduplication
	engine.AddHotFile("/path/to/file1.go")

	if len(engine.state.HotFiles) != 2 {
		t.Errorf("expected 2 hot files (no duplicate), got %d", len(engine.state.HotFiles))
	}
}

func TestBuildPrompt(t *testing.T) {
	cfg := types.Config{}
	db := &memory.SessionDB{}
	engine := NewContextEngine(db, cfg)

	// Setup state
	engine.AddL0("Turn 1: Fixed bug")
	engine.AddL0("Turn 2: Added tests")
	engine.AddL1("Uses OAuth2", "auth")
	engine.AddL1("PostgreSQL database", "database")
	engine.SetTaskGoal("Implement user authentication")
	engine.AddHotFile("/src/auth.go")
	engine.AddHotFile("/src/user.go")
	engine.RecordError("connection timeout")
	engine.RecordToolUse("read", "auth.go", "200 lines")

	task := "Add auth password reset functionality"
	prompt := engine.BuildPrompt(task)

	// Verify prompt contains all sections
	if !strings.Contains(prompt, "## L0 Recent Turn Summaries") {
		t.Error("prompt missing L0 section")
	}

	if !strings.Contains(prompt, "Turn 1: Fixed bug") {
		t.Error("prompt missing L0 item")
	}

	if !strings.Contains(prompt, "## L1 Relevant Facts") {
		t.Error("prompt missing L1 section")
	}

	if !strings.Contains(prompt, "Uses OAuth2") {
		t.Error("prompt missing L1 item")
	}

	if !strings.Contains(prompt, "## Session State") {
		t.Error("prompt missing session state section")
	}

	if !strings.Contains(prompt, "Implement user authentication") {
		t.Error("prompt missing task goal")
	}

	if !strings.Contains(prompt, "connection timeout") {
		t.Error("prompt missing last error")
	}

	if !strings.Contains(prompt, "/src/auth.go") {
		t.Error("prompt missing hot file")
	}

	if !strings.Contains(prompt, "## Tool History") {
		t.Error("prompt missing tool history section")
	}

	if !strings.Contains(prompt, "## Current Task") {
		t.Error("prompt missing current task section")
	}

	if !strings.Contains(prompt, task) {
		t.Error("prompt missing current task")
	}
}

func TestBuildPromptWithCategoryFilter(t *testing.T) {
	cfg := types.Config{}
	db := &memory.SessionDB{}
	engine := NewContextEngine(db, cfg)

	engine.AddL1("Uses OAuth2", "auth")
	engine.AddL1("PostgreSQL database", "database")
	engine.AddL1("Redis cache", "cache")

	// Task mentioning 'auth' should include auth-related L1 items
	prompt := engine.BuildPrompt("Fix auth bug")

	if !strings.Contains(prompt, "Uses OAuth2") {
		t.Error("expected auth-related L1 item in prompt")
	}
}

func TestCompress(t *testing.T) {
	cfg := types.Config{
		Memory: types.MemoryConfig{
			MaxL0Items: 50,
		},
	}
	db := &memory.SessionDB{}
	engine := NewContextEngine(db, cfg)

	// Add multiple L0 items
	engine.AddL0("Turn 1")
	engine.AddL0("Turn 2")
	engine.AddL0("Turn 3")
	engine.AddL1("Fact 1", "cat1")
	engine.AddL1("Fact 2", "cat2")
	engine.AddL1("Fact 3", "cat3")
	engine.AddL1("Fact 4", "cat4")
	engine.AddL1("Fact 5", "cat5")
	engine.AddL1("Fact 6", "cat6")

	// Compress
	engine.Compress()

	// L0 should be compressed to 1 item
	if len(engine.l0Items) != 1 {
		t.Errorf("expected 1 L0 item after compression, got %d", len(engine.l0Items))
	}

	if !strings.Contains(engine.l0Items[0].Summary, "Compressed history:") {
		t.Errorf("expected compressed summary, got %s", engine.l0Items[0].Summary)
	}

	// L1 should be limited to 5 items
	if len(engine.l1Items) != 5 {
		t.Errorf("expected 5 L1 items after compression, got %d", len(engine.l1Items))
	}
}

func TestCompressWithEmptyL0(t *testing.T) {
	cfg := types.Config{}
	db := &memory.SessionDB{}
	engine := NewContextEngine(db, cfg)

	// Compress with no L0 items should not panic
	engine.Compress()

	if len(engine.l0Items) != 0 {
		t.Errorf("expected 0 L0 items, got %d", len(engine.l0Items))
	}
}

func TestIncrementTurns(t *testing.T) {
	cfg := types.Config{
		Memory: types.MemoryConfig{
			AutoCompress: true,
		},
	}
	db := &memory.SessionDB{}
	engine := NewContextEngine(db, cfg)

	// Add some L0 items to test auto-compression
	for i := 0; i < 5; i++ {
		engine.AddL0("Turn")
	}

	// Increment 9 turns (no compression yet)
	for i := 0; i < 9; i++ {
		engine.IncrementTurns()
	}

	if engine.state.TurnCount != 9 {
		t.Errorf("expected turn count 9, got %d", engine.state.TurnCount)
	}

	if len(engine.l0Items) != 5 {
		t.Errorf("expected 5 L0 items (no compression yet), got %d", len(engine.l0Items))
	}

	// 10th turn should trigger compression
	engine.IncrementTurns()

	if engine.state.TurnCount != 10 {
		t.Errorf("expected turn count 10, got %d", engine.state.TurnCount)
	}

	// Should be compressed
	if len(engine.l0Items) != 1 {
		t.Errorf("expected 1 L0 item after auto-compression, got %d", len(engine.l0Items))
	}
}

func TestIncrementTurnsWithoutAutoCompress(t *testing.T) {
	cfg := types.Config{
		Memory: types.MemoryConfig{
			AutoCompress: false,
		},
	}
	db := &memory.SessionDB{}
	engine := NewContextEngine(db, cfg)

	for i := 0; i < 5; i++ {
		engine.AddL0("Turn")
	}

	// Increment 10 turns
	for i := 0; i < 10; i++ {
		engine.IncrementTurns()
	}

	// Should NOT be compressed
	if len(engine.l0Items) != 5 {
		t.Errorf("expected 5 L0 items (no auto-compression), got %d", len(engine.l0Items))
	}
}

func TestDefaultMaxItems(t *testing.T) {
	// Test with empty config (should use defaults)
	cfg := types.Config{Memory: types.MemoryConfig{}}
	db := &memory.SessionDB{}
	engine := NewContextEngine(db, cfg)

	// Add more than default L0 items (default 20)
	for i := 0; i < 60; i++ {
		engine.AddL0(fmt.Sprintf("Summary %d", i))
	}
	if len(engine.l0Items) != 20 {
		t.Errorf("expected 20 L0 items (default limit), got %d", len(engine.l0Items))
	}

	// Add more than default L1 items (default 50)
	for i := 0; i < 60; i++ {
		engine.AddL1(fmt.Sprintf("Fact %d", i), "cat1")
	}
	if len(engine.l1Items) != 50 {
		t.Errorf("expected 50 L1 items (default limit), got %d", len(engine.l1Items))
	}
}
