package agent

import (
	"context"
	"testing"

	"github.com/ahmad-ubaidillah/aigo/internal/memory"
	"github.com/ahmad-ubaidillah/aigo/internal/opencode"
	"github.com/ahmad-ubaidillah/aigo/pkg/types"
)

func TestNewRouter(t *testing.T) {
	db := &memory.SessionDB{}
	cfg := types.Config{}
	ocClient := &opencode.Client{}

	router := NewRouter(db, cfg, ocClient)

	if router == nil {
		t.Fatal("router should not be nil")
	}

	if len(router.handlers) == 0 {
		t.Errorf("expected default handlers to be registered, got %d", len(router.handlers))
	}
}

func TestRegisterHandler(t *testing.T) {
	db := &memory.SessionDB{}
	cfg := types.Config{}
	ocClient := &opencode.Client{}
	router := NewRouter(db, cfg, ocClient)

	mockHandler := &MockHandler{}
	router.RegisterHandler("test_intent", mockHandler)

	if router.handlers["test_intent"] != mockHandler {
		t.Error("handler not registered correctly")
	}
}

func TestRouteWithValidIntent(t *testing.T) {
	db := &memory.SessionDB{}
	cfg := types.Config{}
	ocClient := &opencode.Client{}
	router := NewRouter(db, cfg, ocClient)

	classification := Classification{
		Intent:     types.IntentGeneral,
		Confidence: 0.9,
		Workspace:  "/tmp/test",
		SessionID:  "test-session",
	}

	result, err := router.Route(classification, "test task description")

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if result == nil {
		t.Error("result should not be nil")
	}
}

func TestRouteWithInvalidIntent(t *testing.T) {
	db := &memory.SessionDB{}
	cfg := types.Config{}
	ocClient := &opencode.Client{}
	router := NewRouter(db, cfg, ocClient)

	classification := Classification{
		Intent:     "invalid_intent",
		Confidence: 0.9,
		Workspace:  "/tmp/test",
		SessionID:  "test-session",
	}

	result, err := router.Route(classification, "test task")

	if err == nil {
		t.Error("expected error for invalid intent")
	}

	if result != nil {
		t.Error("result should be nil for error case")
	}
}

func TestRouteCreatesTask(t *testing.T) {
	db := &memory.SessionDB{}
	cfg := types.Config{}
	ocClient := &opencode.Client{}
	router := NewRouter(db, cfg, ocClient)

	classification := Classification{
		Intent:     types.IntentGeneral,
		Confidence: 0.9,
		Workspace:  "/workspace/test",
		SessionID:  "session-123",
	}

	taskDesc := "Test task description"

	// Mock handler to capture the task
	mockHandler := &taskCapturingHandler{}
	router.RegisterHandler(types.IntentGeneral, mockHandler)

	_, err := router.Route(classification, taskDesc)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if mockHandler.capturedTask == nil {
		t.Error("task was not passed to handler")
	}

	if mockHandler.capturedTask.Description != taskDesc {
		t.Errorf("task description mismatch: expected %q, got %q", taskDesc, mockHandler.capturedTask.Description)
	}

	if mockHandler.capturedTask.SessionID != classification.SessionID {
		t.Errorf("task session ID mismatch: expected %q, got %q", classification.SessionID, mockHandler.capturedTask.SessionID)
	}

	if mockHandler.capturedTask.Workspace != classification.Workspace {
		t.Errorf("task workspace mismatch: expected %q, got %q", classification.Workspace, mockHandler.capturedTask.Workspace)
	}
}

func TestDefaultHandlerRegistration(t *testing.T) {
	db := &memory.SessionDB{}
	cfg := types.Config{}
	ocClient := &opencode.Client{}
	router := NewRouter(db, cfg, ocClient)

	requiredHandlers := []string{
		types.IntentCoding,
		types.IntentWeb,
		types.IntentFile,
		types.IntentGateway,
		types.IntentMemory,
		types.IntentAutomation,
		types.IntentGeneral,
	}

	for _, intent := range requiredHandlers {
		if _, exists := router.handlers[intent]; !exists {
			t.Errorf("missing default handler for intent: %s", intent)
		}
	}
}

func TestCodingHandlerCanHandle(t *testing.T) {
	ocClient := &opencode.Client{}
	handler := &codingHandler{client: ocClient}

	if !handler.CanHandle(types.IntentCoding) {
		t.Error("coding handler should handle coding intent")
	}

	if handler.CanHandle(types.IntentWeb) {
		t.Error("coding handler should not handle web intent")
	}
}

func TestCodingHandlerExecuteWithNilClient(t *testing.T) {
	handler := &codingHandler{client: nil}

	task := &types.Task{
		Description: "test task",
	}

	result, err := handler.Execute(context.Background(), task, "")

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("result should not be nil")
	}

	if result.Success {
		t.Error("expected failure with nil client")
	}

	if result.Error == "" {
		t.Error("expected error message")
	}
}

// Mock handlers for testing

type MockHandler struct{}

func (h *MockHandler) CanHandle(intent string) bool {
	return intent == "test_intent"
}

func (h *MockHandler) Execute(ctx context.Context, task *types.Task, workspace string) (*types.ToolResult, error) {
	return &types.ToolResult{
		Success: true,
		Output:  "mock output",
	}, nil
}

type taskCapturingHandler struct {
	capturedTask *types.Task
}

func (h *taskCapturingHandler) CanHandle(intent string) bool {
	return intent == types.IntentGeneral
}

func (h *taskCapturingHandler) Execute(ctx context.Context, task *types.Task, workspace string) (*types.ToolResult, error) {
	h.capturedTask = task
	return &types.ToolResult{
		Success: true,
		Output:  "captured",
	}, nil
}
