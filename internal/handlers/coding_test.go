package handlers

import (
	"context"
	"testing"

	"github.com/ahmad-ubaidillah/aigo/internal/tools"
	"github.com/ahmad-ubaidillah/aigo/pkg/types"
)

type mockInstaller struct {
	available  bool
	path       string
	checkErr   error
	installErr error
}

func (m *mockInstaller) CheckOpenCode() (bool, string, error) {
	return m.available, m.path, m.checkErr
}

func (m *mockInstaller) InstallOpenCode(ctx context.Context, path string) error {
	return m.installErr
}

type mockBashTool struct {
	executeResult *types.ToolResult
	executeErr    error
	lastParams    map[string]any
}

func (m *mockBashTool) Name() string { return "bash" }
func (m *mockBashTool) Description() string { return "Execute a bash command" }
func (m *mockBashTool) Schema() map[string]any {
	return map[string]any{"type": "object"}
}
func (m *mockBashTool) Execute(ctx context.Context, params map[string]any) (*types.ToolResult, error) {
	m.lastParams = params
	return m.executeResult, m.executeErr
}

func TestCodingHandler_CanHandle(t *testing.T) {
	t.Parallel()

	registry := tools.NewRegistry()
	installer := &mockInstaller{}
	h := NewCodingHandler(registry, installer)

	if !h.CanHandle(types.IntentCoding) {
		t.Error("expected CanHandle to return true for IntentCoding")
	}
	if h.CanHandle(types.IntentFile) {
		t.Error("expected CanHandle to return false for IntentFile")
	}
	if h.CanHandle(types.IntentWeb) {
		t.Error("expected CanHandle to return false for IntentWeb")
	}
	if h.CanHandle(types.IntentGeneral) {
		t.Error("expected CanHandle to return false for IntentGeneral")
	}
}

func TestCodingHandler_Execute_CheckFailed(t *testing.T) {
	t.Parallel()

	registry := tools.NewRegistry()
	bashTool := &mockBashTool{
		executeResult: &types.ToolResult{Success: true, Output: "bash output"},
	}
	registry.Register(bashTool)

	installer := &mockInstaller{
		checkErr: &testErr{msg: "check failed"},
	}
	h := NewCodingHandler(registry, installer)

	result, err := h.Execute(context.Background(), &types.Task{
		Description: "fix the bug",
		SessionID:   "test-1",
	}, "/tmp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Native fallback succeeds via bash, so Success is true
	if !result.Success {
		t.Error("expected success from native fallback")
	}
	// Output should be prefixed with [Native Fallback]
	if result.Output == "" {
		t.Error("expected non-empty output from native fallback")
	}
}

func TestCodingHandler_Execute_InstallFailed(t *testing.T) {
	t.Parallel()

	registry := tools.NewRegistry()
	bashTool := &mockBashTool{
		executeResult: &types.ToolResult{Success: true, Output: "bash output"},
	}
	registry.Register(bashTool)

	installer := &mockInstaller{
		available:  false,
		path:       "/tmp/opencode",
		installErr: &testErr{msg: "install failed"},
	}
	h := NewCodingHandler(registry, installer)

	result, err := h.Execute(context.Background(), &types.Task{
		Description: "fix the bug",
		SessionID:   "test-1",
	}, "/tmp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Native fallback succeeds via bash, so Success is true
	if !result.Success {
		t.Error("expected success from native fallback")
	}
}

func TestCodingHandler_Execute_NoFallbackAvailable(t *testing.T) {
	t.Parallel()

	registry := tools.NewRegistry()
	// No bash tool registered

	installer := &mockInstaller{
		checkErr: &testErr{msg: "not found"},
	}
	h := NewCodingHandler(registry, installer)

	result, err := h.Execute(context.Background(), &types.Task{
		Description: "fix the bug",
		SessionID:   "test-1",
	}, "/tmp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Success {
		t.Error("expected failure when no fallback available")
	}
	if result.Error == "" {
		t.Error("expected error message when no fallback available")
	}
}

func TestCodingHandler_NativeFallback_OutputPrefixed(t *testing.T) {
	t.Parallel()

	registry := tools.NewRegistry()
	bashTool := &mockBashTool{
		executeResult: &types.ToolResult{Success: true, Output: "raw output"},
	}
	registry.Register(bashTool)

	installer := &mockInstaller{
		checkErr: &testErr{msg: "not available"},
	}
	h := NewCodingHandler(registry, installer)

	result, err := h.Execute(context.Background(), &types.Task{
		Description: "fix the bug",
		SessionID:   "test-1",
	}, "/tmp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "[Native Fallback] OpenCode check failed: not available"
	if len(result.Output) < len(expected) {
		t.Errorf("expected output to start with %q, got %q", expected, result.Output)
	}
}

type testErr struct {
	msg string
}

func (e *testErr) Error() string { return e.msg }
