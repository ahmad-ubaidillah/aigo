package tools

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ahmad-ubaidillah/aigo/pkg/types"
)

func TestToolRegistry_RegisterAndGet(t *testing.T) {
	t.Parallel()

	r := NewRegistry()
	r.Register(&ReadTool{})

	tool := r.Get("read")
	if tool == nil {
		t.Fatal("expected read tool")
	}
	if tool.Name() != "read" {
		t.Errorf("expected read, got %s", tool.Name())
	}
}

func TestToolRegistry_GetNotFound(t *testing.T) {
	t.Parallel()

	r := NewRegistry()
	if r.Get("nonexistent") != nil {
		t.Error("expected nil for unknown tool")
	}
}

func TestToolRegistry_List(t *testing.T) {
	t.Parallel()

	r := NewRegistry()
	r.Register(&ReadTool{})
	r.Register(&WriteTool{})

	tools := r.List()
	if len(tools) != 2 {
		t.Errorf("expected 2 tools, got %d", len(tools))
	}
}

func TestToolRegistry_Execute(t *testing.T) {
	t.Parallel()

	r := NewRegistry()
	r.Register(&ReadTool{})

	result, err := r.Execute(context.Background(), "read", map[string]any{"path": "/nonexistent"})
	if err != nil {
		t.Fatal(err)
	}
	if result.Success {
		t.Error("expected failure for nonexistent file")
	}
}

func TestToolRegistry_ExecuteNotFound(t *testing.T) {
	t.Parallel()

	r := NewRegistry()
	result, err := r.Execute(context.Background(), "nonexistent", nil)
	if err == nil {
		t.Error("expected error for unknown tool")
	}
	if result != nil {
		t.Error("expected nil result")
	}
}

func TestOutputTruncate(t *testing.T) {
	t.Parallel()

	result := &types.ToolResult{Output: "a"}
	OutputTruncate(result, 10)
	if result.Output != "a" {
		t.Errorf("expected no truncation, got %q", result.Output)
	}

	result = &types.ToolResult{Output: "abcdefghij12345"}
	OutputTruncate(result, 10)
	if len(result.Output) <= 10 {
		t.Errorf("expected truncation message, got %d chars", len(result.Output))
	}
}

func TestPermissionChecker_DefaultAllow(t *testing.T) {
	t.Parallel()

	c := NewPermissionChecker()
	if c.Check("bash") != PermAllow {
		t.Error("expected default allow")
	}
}

func TestPermissionChecker_Deny(t *testing.T) {
	t.Parallel()

	c := NewPermissionChecker()
	c.AddRule(PermissionRule{Tool: "bash", Pattern: "*", Level: PermDeny})
	if c.Check("bash") != PermDeny {
		t.Error("expected deny")
	}
}

func TestPermissionChecker_Wildcard(t *testing.T) {
	t.Parallel()

	c := NewPermissionChecker()
	c.AddRule(PermissionRule{Tool: "web*", Pattern: "*", Level: PermDeny})
	if c.Check("webfetch") != PermDeny {
		t.Error("expected wildcard deny")
	}
	if c.Check("bash") != PermAllow {
		t.Error("expected allow for non-matching")
	}
}

func TestPermissionChecker_Ask(t *testing.T) {
	t.Parallel()

	c := NewPermissionChecker()
	c.AddRule(PermissionRule{Tool: "edit", Pattern: "*", Level: PermAsk})
	if c.Check("edit") != PermAsk {
		t.Error("expected ask")
	}
}

func TestReadTool(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	os.WriteFile(path, []byte("hello"), 0644)

	tool := &ReadTool{}
	result, err := tool.Execute(context.Background(), map[string]any{"path": path})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Success {
		t.Errorf("expected success: %s", result.Error)
	}
	if result.Output != "hello" {
		t.Errorf("expected hello, got %s", result.Output)
	}
}

func TestReadTool_NotFound(t *testing.T) {
	t.Parallel()

	tool := &ReadTool{}
	result, err := tool.Execute(context.Background(), map[string]any{"path": "/nonexistent"})
	if err != nil {
		t.Fatal(err)
	}
	if result.Success {
		t.Error("expected failure")
	}
}

func TestReadTool_MissingParam(t *testing.T) {
	t.Parallel()

	tool := &ReadTool{}
	result, err := tool.Execute(context.Background(), map[string]any{})
	if err != nil {
		t.Fatal(err)
	}
	if result.Success {
		t.Error("expected failure")
	}
}

func TestWriteTool(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "test.txt")

	tool := &WriteTool{}
	result, err := tool.Execute(context.Background(), map[string]any{
		"path":    path,
		"content": "world",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Success {
		t.Errorf("expected success: %s", result.Error)
	}

	data, _ := os.ReadFile(path)
	if string(data) != "world" {
		t.Errorf("expected world, got %s", string(data))
	}
}

func TestEditTool(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	os.WriteFile(path, []byte("hello world"), 0644)

	tool := &EditTool{}
	result, err := tool.Execute(context.Background(), map[string]any{
		"path":       path,
		"old_string": "world",
		"new_string": "go",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Success {
		t.Errorf("expected success: %s", result.Error)
	}

	data, _ := os.ReadFile(path)
	if string(data) != "hello go" {
		t.Errorf("expected 'hello go', got %s", string(data))
	}
}

func TestEditTool_NotFound(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	os.WriteFile(path, []byte("hello"), 0644)

	tool := &EditTool{}
	result, err := tool.Execute(context.Background(), map[string]any{
		"path":       path,
		"old_string": "nonexistent",
		"new_string": "go",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Success {
		t.Error("expected failure")
	}
}

func TestGlobTool(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.go"), []byte(""), 0644)
	os.WriteFile(filepath.Join(dir, "b.go"), []byte(""), 0644)

	tool := &GlobTool{}
	result, err := tool.Execute(context.Background(), map[string]any{
		"pattern": filepath.Join(dir, "*.go"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Success {
		t.Errorf("expected success: %s", result.Error)
	}
}

func TestGrepTool(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "test.go"), []byte("func main() {}"), 0644)

	tool := &GrepTool{}
	result, err := tool.Execute(context.Background(), map[string]any{
		"pattern": "func main",
		"path":    dir,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Success {
		t.Errorf("expected success: %s", result.Error)
	}
}

func TestBashTool(t *testing.T) {
	t.Parallel()

	tool := &BashTool{}
	result, err := tool.Execute(context.Background(), map[string]any{
		"command": "echo hello",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Success {
		t.Errorf("expected success: %s", result.Error)
	}
}

func TestBashTool_Failure(t *testing.T) {
	t.Parallel()

	tool := &BashTool{}
	result, err := tool.Execute(context.Background(), map[string]any{
		"command": "exit 1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Success {
		t.Error("expected failure")
	}
}

func TestTaskTool(t *testing.T) {
	t.Parallel()

	tool := &TaskTool{}
	result, err := tool.Execute(context.Background(), map[string]any{
		"description": "test task",
		"category":    "quick",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Success {
		t.Errorf("expected success: %s", result.Error)
	}
}

func TestWebSearchTool(t *testing.T) {
	t.Parallel()

	t.Skip("websearch requires network access")
}

func TestWebSearchTool_MissingQuery(t *testing.T) {
	t.Parallel()

	tool := &WebSearchTool{}
	result, err := tool.Execute(context.Background(), map[string]any{})
	if err != nil {
		t.Fatal(err)
	}
	if result.Success {
		t.Error("expected failure for missing query")
	}
}

func TestWebSearchTool_InvalidQuery(t *testing.T) {
	t.Parallel()

	tool := &WebSearchTool{}
	result, err := tool.Execute(context.Background(), map[string]any{"query": 123})
	if err != nil {
		t.Fatal(err)
	}
	if result.Success {
		t.Error("expected failure for invalid query type")
	}
}

func TestTodoTool(t *testing.T) {
	t.Parallel()

	tool := &TodoTool{}

	result, _ := tool.Execute(context.Background(), map[string]any{
		"action":  "add",
		"content": "test todo",
	})
	if !result.Success {
		t.Errorf("add failed: %s", result.Error)
	}

	result, _ = tool.Execute(context.Background(), map[string]any{"action": "list"})
	if !result.Success {
		t.Errorf("list failed: %s", result.Error)
	}

	result, _ = tool.Execute(context.Background(), map[string]any{
		"action": "complete",
		"index":  1,
	})
	if !result.Success {
		t.Errorf("complete failed: %s", result.Error)
	}

	result, _ = tool.Execute(context.Background(), map[string]any{"action": "invalid"})
	if result.Success {
		t.Error("expected failure for invalid action")
	}
}

func TestDelegateTool(t *testing.T) {
	t.Parallel()

	tool := NewDelegateTool()

	id := tool.SpawnChild("parent-1", "test task", "quick", 1)
	if id == "" {
		t.Fatal("expected child ID")
	}

	child, err := tool.GetChild(id)
	if err != nil {
		t.Fatal(err)
	}
	if child.Description != "test task" {
		t.Errorf("expected test task, got %s", child.Description)
	}

	children := tool.ListChildren("parent-1")
	if len(children) != 1 {
		t.Errorf("expected 1 child, got %d", len(children))
	}

	err = tool.UpdateChild(id, "done", "result")
	if err != nil {
		t.Fatal(err)
	}

	if tool.MaxDepth() != 2 {
		t.Errorf("expected max depth 2, got %d", tool.MaxDepth())
	}
}

func TestDelegateTool_Execute(t *testing.T) {
	t.Parallel()

	tool := NewDelegateTool()
	result, err := tool.Execute(context.Background(), map[string]any{
		"description": "spawn test",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Success {
		t.Errorf("expected success: %s", result.Error)
	}
}

func TestDelegateTool_GetChildNotFound(t *testing.T) {
	t.Parallel()

	tool := NewDelegateTool()
	_, err := tool.GetChild("nonexistent")
	if err == nil {
		t.Error("expected error")
	}
}

func TestDelegateTool_UpdateChildNotFound(t *testing.T) {
	t.Parallel()

	tool := NewDelegateTool()
	err := tool.UpdateChild("nonexistent", "done", "")
	if err == nil {
		t.Error("expected error")
	}
}

func TestDelegateTool_Name(t *testing.T) {
	t.Parallel()

	tool := NewDelegateTool()
	if tool.Name() != "delegate" {
		t.Errorf("expected delegate, got %s", tool.Name())
	}
}

func TestDelegateTool_Description(t *testing.T) {
	t.Parallel()

	tool := NewDelegateTool()
	if tool.Description() == "" {
		t.Error("expected description")
	}
}

func TestDelegateTool_Schema(t *testing.T) {
	t.Parallel()

	tool := NewDelegateTool()
	schema := tool.Schema()
	if schema["type"] != "object" {
		t.Errorf("expected object type, got %v", schema["type"])
	}
}

func TestDelegateTool_MaxDepthExceeded(t *testing.T) {
	t.Parallel()

	tool := NewDelegateTool()
	id := tool.SpawnChild("", "parent", "", 1)
	result, err := tool.Execute(context.Background(), map[string]any{
		"description": "child",
		"session_id":  id,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Success {
		t.Errorf("expected success, got %s", result.Error)
	}

	result, err = tool.Execute(context.Background(), map[string]any{
		"description": "grandchild",
		"session_id":  result.Metadata["child_id"],
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Success {
		t.Error("expected failure when depth exceeds max")
	}
}

func TestDelegateTool_MissingDescription(t *testing.T) {
	t.Parallel()

	tool := NewDelegateTool()
	result, err := tool.Execute(context.Background(), map[string]any{})
	if err != nil {
		t.Fatal(err)
	}
	if result.Success {
		t.Error("expected failure for missing description")
	}
}

func TestBashTool_Name(t *testing.T) {
	t.Parallel()

	tool := &BashTool{}
	if tool.Name() != "bash" {
		t.Errorf("expected bash, got %s", tool.Name())
	}
}

func TestBashTool_Description(t *testing.T) {
	t.Parallel()

	tool := &BashTool{}
	if tool.Description() == "" {
		t.Error("expected description")
	}
}

func TestBashTool_Schema(t *testing.T) {
	t.Parallel()

	tool := &BashTool{}
	schema := tool.Schema()
	if schema["type"] != "object" {
		t.Errorf("expected object type, got %v", schema["type"])
	}
}

func TestBashTool_MissingCommand(t *testing.T) {
	t.Parallel()

	tool := &BashTool{}
	result, err := tool.Execute(context.Background(), map[string]any{})
	if err != nil {
		t.Fatal(err)
	}
	if result.Success {
		t.Error("expected failure for missing command")
	}
}

func TestBashTool_InvalidCommandType(t *testing.T) {
	t.Parallel()

	tool := &BashTool{}
	result, err := tool.Execute(context.Background(), map[string]any{"command": 123})
	if err != nil {
		t.Fatal(err)
	}
	if result.Success {
		t.Error("expected failure for invalid command type")
	}
}

func TestBashTool_CustomTimeout(t *testing.T) {
	t.Parallel()

	tool := &BashTool{}
	result, err := tool.Execute(context.Background(), map[string]any{
		"command": "echo ok",
		"timeout": 5,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Success {
		t.Errorf("expected success: %s", result.Error)
	}
}

func TestTaskTool_Name(t *testing.T) {
	t.Parallel()

	tool := &TaskTool{}
	if tool.Name() != "task" {
		t.Errorf("expected task, got %s", tool.Name())
	}
}

func TestTaskTool_Description(t *testing.T) {
	t.Parallel()

	tool := &TaskTool{}
	if tool.Description() == "" {
		t.Error("expected description")
	}
}

func TestTaskTool_Schema(t *testing.T) {
	t.Parallel()

	tool := &TaskTool{}
	schema := tool.Schema()
	if schema["type"] != "object" {
		t.Errorf("expected object type, got %v", schema["type"])
	}
}

func TestTaskTool_MissingDescription(t *testing.T) {
	t.Parallel()

	tool := &TaskTool{}
	result, err := tool.Execute(context.Background(), map[string]any{})
	if err != nil {
		t.Fatal(err)
	}
	if result.Success {
		t.Error("expected failure for missing description")
	}
}

func TestTaskTool_InvalidDescriptionType(t *testing.T) {
	t.Parallel()

	tool := &TaskTool{}
	result, err := tool.Execute(context.Background(), map[string]any{"description": 123})
	if err != nil {
		t.Fatal(err)
	}
	if result.Success {
		t.Error("expected failure for invalid description type")
	}
}

func TestTaskTool_WithSessionID(t *testing.T) {
	t.Parallel()

	tool := &TaskTool{}
	result, err := tool.Execute(context.Background(), map[string]any{
		"description": "test",
		"session_id":  "parent-1",
		"category":    "quick",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Success {
		t.Errorf("expected success: %s", result.Error)
	}
	if result.Metadata["session_id"] != "parent-1" {
		t.Errorf("expected parent-1, got %s", result.Metadata["session_id"])
	}
}

func TestReadTool_Description(t *testing.T) {
	t.Parallel()

	tool := &ReadTool{}
	if tool.Description() == "" {
		t.Error("expected description")
	}
}

func TestReadTool_Schema(t *testing.T) {
	t.Parallel()

	tool := &ReadTool{}
	schema := tool.Schema()
	if schema["type"] != "object" {
		t.Errorf("expected object type, got %v", schema["type"])
	}
}

func TestWriteTool_Name(t *testing.T) {
	t.Parallel()

	tool := &WriteTool{}
	if tool.Name() != "write" {
		t.Errorf("expected write, got %s", tool.Name())
	}
}

func TestWriteTool_Description(t *testing.T) {
	t.Parallel()

	tool := &WriteTool{}
	if tool.Description() == "" {
		t.Error("expected description")
	}
}

func TestWriteTool_Schema(t *testing.T) {
	t.Parallel()

	tool := &WriteTool{}
	schema := tool.Schema()
	if schema["type"] != "object" {
		t.Errorf("expected object type, got %v", schema["type"])
	}
}

func TestWriteTool_MissingPath(t *testing.T) {
	t.Parallel()

	tool := &WriteTool{}
	result, err := tool.Execute(context.Background(), map[string]any{"content": "x"})
	if err != nil {
		t.Fatal(err)
	}
	if result.Success {
		t.Error("expected failure for missing path")
	}
}

func TestWriteTool_MissingContent(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	tool := &WriteTool{}
	result, err := tool.Execute(context.Background(), map[string]any{"path": dir + "/f.txt"})
	if err != nil {
		t.Fatal(err)
	}
	if result.Success {
		t.Error("expected failure for missing content")
	}
}

func TestEditTool_Name(t *testing.T) {
	t.Parallel()

	tool := &EditTool{}
	if tool.Name() != "edit" {
		t.Errorf("expected edit, got %s", tool.Name())
	}
}

func TestEditTool_Description(t *testing.T) {
	t.Parallel()

	tool := &EditTool{}
	if tool.Description() == "" {
		t.Error("expected description")
	}
}

func TestEditTool_Schema(t *testing.T) {
	t.Parallel()

	tool := &EditTool{}
	schema := tool.Schema()
	if schema["type"] != "object" {
		t.Errorf("expected object type, got %v", schema["type"])
	}
}

func TestEditTool_MissingParams(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := dir + "/f.txt"
	os.WriteFile(path, []byte("hello"), 0644)

	tool := &EditTool{}
	result, err := tool.Execute(context.Background(), map[string]any{"path": path})
	if err != nil {
		t.Fatal(err)
	}
	if result.Success {
		t.Error("expected failure for missing old_string")
	}
}

func TestEditTool_MultipleOccurrences(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := dir + "/f.txt"
	os.WriteFile(path, []byte("hello hello"), 0644)

	tool := &EditTool{}
	result, err := tool.Execute(context.Background(), map[string]any{
		"path":       path,
		"old_string": "hello",
		"new_string": "hi",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Success {
		t.Error("expected failure for multiple occurrences")
	}
}

func TestGlobTool_Name(t *testing.T) {
	t.Parallel()

	tool := &GlobTool{}
	if tool.Name() != "glob" {
		t.Errorf("expected glob, got %s", tool.Name())
	}
}

func TestGlobTool_Description(t *testing.T) {
	t.Parallel()

	tool := &GlobTool{}
	if tool.Description() == "" {
		t.Error("expected description")
	}
}

func TestGlobTool_Schema(t *testing.T) {
	t.Parallel()

	tool := &GlobTool{}
	schema := tool.Schema()
	if schema["type"] != "object" {
		t.Errorf("expected object type, got %v", schema["type"])
	}
}

func TestGlobTool_MissingPattern(t *testing.T) {
	t.Parallel()

	tool := &GlobTool{}
	result, err := tool.Execute(context.Background(), map[string]any{})
	if err != nil {
		t.Fatal(err)
	}
	if result.Success {
		t.Error("expected failure for missing pattern")
	}
}

func TestGlobTool_NoMatches(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	tool := &GlobTool{}
	result, err := tool.Execute(context.Background(), map[string]any{
		"pattern": dir + "/nonexistent*",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Success {
		t.Errorf("expected success: %s", result.Error)
	}
}

func TestGrepTool_Name(t *testing.T) {
	t.Parallel()

	tool := &GrepTool{}
	if tool.Name() != "grep" {
		t.Errorf("expected grep, got %s", tool.Name())
	}
}

func TestGrepTool_Description(t *testing.T) {
	t.Parallel()

	tool := &GrepTool{}
	if tool.Description() == "" {
		t.Error("expected description")
	}
}

func TestGrepTool_Schema(t *testing.T) {
	t.Parallel()

	tool := &GrepTool{}
	schema := tool.Schema()
	if schema["type"] != "object" {
		t.Errorf("expected object type, got %v", schema["type"])
	}
}

func TestGrepTool_MissingPattern(t *testing.T) {
	t.Parallel()

	tool := &GrepTool{}
	result, err := tool.Execute(context.Background(), map[string]any{})
	if err != nil {
		t.Fatal(err)
	}
	if result.Success {
		t.Error("expected failure for missing pattern")
	}
}

func TestGrepTool_InvalidRegex(t *testing.T) {
	t.Parallel()

	tool := &GrepTool{}
	result, err := tool.Execute(context.Background(), map[string]any{"pattern": "[invalid"})
	if err != nil {
		t.Fatal(err)
	}
	if result.Success {
		t.Error("expected failure for invalid regex")
	}
}

func TestGrepTool_NoMatches(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	os.WriteFile(dir+"/f.go", []byte("package main"), 0644)

	tool := &GrepTool{}
	result, err := tool.Execute(context.Background(), map[string]any{
		"pattern": "nonexistent_pattern_xyz",
		"path":    dir,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Success {
		t.Errorf("expected success: %s", result.Error)
	}
}

func TestWebFetchTool_Name(t *testing.T) {
	t.Parallel()

	tool := &WebFetchTool{}
	if tool.Name() != "webfetch" {
		t.Errorf("expected webfetch, got %s", tool.Name())
	}
}

func TestWebFetchTool_Description(t *testing.T) {
	t.Parallel()

	tool := &WebFetchTool{}
	if tool.Description() == "" {
		t.Error("expected description")
	}
}

func TestWebFetchTool_Schema(t *testing.T) {
	t.Parallel()

	tool := &WebFetchTool{}
	schema := tool.Schema()
	if schema["type"] != "object" {
		t.Errorf("expected object type, got %v", schema["type"])
	}
}

func TestWebFetchTool_MissingURL(t *testing.T) {
	t.Parallel()

	tool := &WebFetchTool{}
	result, err := tool.Execute(context.Background(), map[string]any{})
	if err != nil {
		t.Fatal(err)
	}
	if result.Success {
		t.Error("expected failure for missing url")
	}
}

func TestWebFetchTool_InvalidURL(t *testing.T) {
	t.Parallel()

	tool := &WebFetchTool{}
	result, err := tool.Execute(context.Background(), map[string]any{"url": "not-a-url"})
	if err != nil {
		t.Fatal(err)
	}
	if result.Success {
		t.Error("expected failure for invalid url")
	}
}

func TestWebFetchTool_EmptyURL(t *testing.T) {
	t.Parallel()

	tool := &WebFetchTool{}
	result, err := tool.Execute(context.Background(), map[string]any{"url": "  "})
	if err != nil {
		t.Fatal(err)
	}
	if result.Success {
		t.Error("expected failure for empty url")
	}
}

func TestWebSearchTool_Name(t *testing.T) {
	t.Parallel()

	tool := &WebSearchTool{}
	if tool.Name() != "websearch" {
		t.Errorf("expected websearch, got %s", tool.Name())
	}
}

func TestWebSearchTool_Description(t *testing.T) {
	t.Parallel()

	tool := &WebSearchTool{}
	if tool.Description() == "" {
		t.Error("expected description")
	}
}

func TestWebSearchTool_Schema(t *testing.T) {
	t.Parallel()

	tool := &WebSearchTool{}
	schema := tool.Schema()
	if schema["type"] != "object" {
		t.Errorf("expected object type, got %v", schema["type"])
	}
}

func TestWebSearchTool_WithNumResults_SKIP(t *testing.T) {
	t.Skip("requires network")
	t.Parallel()

	tool := &WebSearchTool{}
	result, err := tool.Execute(context.Background(), map[string]any{
		"query":       "test",
		"num_results": 10,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Success {
		t.Errorf("expected success: %s", result.Error)
	}
	if !strings.Contains(result.Output, "results=10") {
		t.Errorf("expected results=10, got %s", result.Output)
	}
}

func TestTodoTool_Name(t *testing.T) {
	t.Parallel()

	tool := &TodoTool{}
	if tool.Name() != "todo" {
		t.Errorf("expected todo, got %s", tool.Name())
	}
}

func TestTodoTool_Description(t *testing.T) {
	t.Parallel()

	tool := &TodoTool{}
	if tool.Description() == "" {
		t.Error("expected description")
	}
}

func TestTodoTool_Schema(t *testing.T) {
	t.Parallel()

	tool := &TodoTool{}
	schema := tool.Schema()
	if schema["type"] != "object" {
		t.Errorf("expected object type, got %v", schema["type"])
	}
}

func TestTodoTool_MissingAction(t *testing.T) {
	t.Parallel()

	tool := &TodoTool{}
	result, err := tool.Execute(context.Background(), map[string]any{})
	if err != nil {
		t.Fatal(err)
	}
	if result.Success {
		t.Error("expected failure for missing action")
	}
}

func TestTodoTool_AddMissingContent(t *testing.T) {
	t.Parallel()

	tool := &TodoTool{}
	result, err := tool.Execute(context.Background(), map[string]any{"action": "add"})
	if err != nil {
		t.Fatal(err)
	}
	if result.Success {
		t.Error("expected failure for missing content")
	}
}

func TestTodoTool_AddInvalidContent(t *testing.T) {
	t.Parallel()

	tool := &TodoTool{}
	result, err := tool.Execute(context.Background(), map[string]any{"action": "add", "content": 123})
	if err != nil {
		t.Fatal(err)
	}
	if result.Success {
		t.Error("expected failure for invalid content type")
	}
}

func TestTodoTool_CompleteMissingIndex(t *testing.T) {
	t.Parallel()

	tool := &TodoTool{}
	result, err := tool.Execute(context.Background(), map[string]any{"action": "complete"})
	if err != nil {
		t.Fatal(err)
	}
	if result.Success {
		t.Error("expected failure for missing index")
	}
}

func TestTodoTool_CompleteInvalidIndex(t *testing.T) {
	t.Parallel()

	tool := &TodoTool{}
	result, err := tool.Execute(context.Background(), map[string]any{"action": "complete", "index": 999})
	if err != nil {
		t.Fatal(err)
	}
	if result.Success {
		t.Error("expected failure for invalid index")
	}
}

func TestTodoTool_CancelMissingIndex(t *testing.T) {
	t.Parallel()

	tool := &TodoTool{}
	result, err := tool.Execute(context.Background(), map[string]any{"action": "cancel"})
	if err != nil {
		t.Fatal(err)
	}
	if result.Success {
		t.Error("expected failure for missing index")
	}
}

func TestTodoTool_CancelInvalidIndex(t *testing.T) {
	t.Parallel()

	tool := &TodoTool{}
	result, err := tool.Execute(context.Background(), map[string]any{"action": "cancel", "index": 999})
	if err != nil {
		t.Fatal(err)
	}
	if result.Success {
		t.Error("expected failure for invalid index")
	}
}

func TestPermissionChecker_String(t *testing.T) {
	t.Parallel()

	if PermAllow.String() != "allow" {
		t.Errorf("expected allow, got %s", PermAllow.String())
	}
	if PermAsk.String() != "ask" {
		t.Errorf("expected ask, got %s", PermAsk.String())
	}
	if PermDeny.String() != "deny" {
		t.Errorf("expected deny, got %s", PermDeny.String())
	}
}

func TestPermissionChecker_Clear(t *testing.T) {
	t.Parallel()

	c := NewPermissionChecker()
	c.AddRule(PermissionRule{Tool: "bash", Pattern: "*", Level: PermDeny})
	c.Clear()
	if c.Check("bash") != PermAllow {
		t.Error("expected allow after clear")
	}
}

func TestPermissionChecker_Rules(t *testing.T) {
	t.Parallel()

	c := NewPermissionChecker()
	c.AddRule(PermissionRule{Tool: "bash", Pattern: "*", Level: PermDeny})
	rules := c.Rules()
	if len(rules) != 1 {
		t.Errorf("expected 1 rule, got %d", len(rules))
	}
}

func TestPermissionChecker_SetRules(t *testing.T) {
	t.Parallel()

	c := NewPermissionChecker()
	c.SetRules([]PermissionRule{
		{Tool: "bash", Pattern: "*", Level: PermDeny},
		{Tool: "read", Pattern: "*", Level: PermAllow},
	})
	if c.Check("bash") != PermDeny {
		t.Error("expected deny")
	}
	if c.Check("read") != PermAllow {
		t.Error("expected allow")
	}
}

func TestParsePermissionLevel(t *testing.T) {
	t.Parallel()

	level, err := ParsePermissionLevel("allow")
	if err != nil || level != PermAllow {
		t.Errorf("expected allow, got %v, %v", level, err)
	}

	level, err = ParsePermissionLevel("deny")
	if err != nil || level != PermDeny {
		t.Errorf("expected deny, got %v, %v", level, err)
	}

	level, err = ParsePermissionLevel("ask")
	if err != nil || level != PermAsk {
		t.Errorf("expected ask, got %v, %v", level, err)
	}

	_, err = ParsePermissionLevel("invalid")
	if err == nil {
		t.Error("expected error for invalid level")
	}
}

func TestToolRegistry_RegisterDuplicate(t *testing.T) {
	t.Parallel()

	r := NewRegistry()
	r.Register(&ReadTool{})
	r.Register(&ReadTool{})
	tools := r.List()
	if len(tools) != 1 {
		t.Errorf("expected 1 tool after duplicate register, got %d", len(tools))
	}
}

func TestOutputTruncate_ExactSize(t *testing.T) {
	t.Parallel()

	result := &types.ToolResult{Output: "1234567890"}
	OutputTruncate(result, 10)
	if result.Output != "1234567890" {
		t.Errorf("expected no truncation, got %q", result.Output)
	}
}

func TestGetStringParam_InvalidType(t *testing.T) {
	t.Parallel()

	tool := &ReadTool{}
	result, err := tool.Execute(context.Background(), map[string]any{"path": 123})
	if err != nil {
		t.Fatal(err)
	}
	if result.Success {
		t.Error("expected failure for invalid path type")
	}
}

func TestWriteTool_MkdirAllFailure(t *testing.T) {
	t.Parallel()

	tool := &WriteTool{}
	result, err := tool.Execute(context.Background(), map[string]any{
		"path":    "/proc/1/fake/file.txt",
		"content": "test",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Success {
		t.Error("expected failure for unwritable path")
	}
}

func TestReadTool_FileTooLarge(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := dir + "/large.txt"
	large := make([]byte, 60*1024)
	os.WriteFile(path, large, 0644)

	tool := &ReadTool{}
	result, err := tool.Execute(context.Background(), map[string]any{"path": path})
	if err != nil {
		t.Fatal(err)
	}
	if result.Success {
		t.Error("expected failure for large file")
	}
}

func TestReadTool_InvalidPathType(t *testing.T) {
	t.Parallel()

	tool := &ReadTool{}
	result, err := tool.Execute(context.Background(), map[string]any{"path": 123})
	if err != nil {
		t.Fatal(err)
	}
	if result.Success {
		t.Error("expected failure for invalid path type")
	}
}

func TestGrepTool_UnreadableDir(t *testing.T) {
	t.Parallel()

	tool := &GrepTool{}
	result, err := tool.Execute(context.Background(), map[string]any{
		"pattern": "test",
		"path":    "/proc/1/root",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Success {
		t.Errorf("expected success (graceful skip): %s", result.Error)
	}
}

func TestWebSearchTool_NumResultsAsString_SKIP(t *testing.T) {
	t.Skip("requires network")
	t.Parallel()

	tool := &WebSearchTool{}
	result, err := tool.Execute(context.Background(), map[string]any{
		"query":       "test",
		"num_results": "7",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Success {
		t.Errorf("expected success: %s", result.Error)
	}
	if !strings.Contains(result.Output, "results=7") {
		t.Errorf("expected results=7, got %s", result.Output)
	}
}

func TestWebSearchTool_NumResultsAsFloat_SKIP(t *testing.T) {
	t.Skip("requires network")
	t.Parallel()

	tool := &WebSearchTool{}
	result, err := tool.Execute(context.Background(), map[string]any{
		"query":       "test",
		"num_results": float64(3),
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Success {
		t.Errorf("expected success: %s", result.Error)
	}
}

func TestWebSearchTool_NumResultsAsInt64_SKIP(t *testing.T) {
	t.Skip("requires network")
	t.Parallel()

	tool := &WebSearchTool{}
	result, err := tool.Execute(context.Background(), map[string]any{
		"query":       "test",
		"num_results": int64(4),
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Success {
		t.Errorf("expected success: %s", result.Error)
	}
}

func TestWebSearchTool_NumResultsAsFloat32_SKIP(t *testing.T) {
	t.Skip("requires network")
	t.Parallel()

	tool := &WebSearchTool{}
	result, err := tool.Execute(context.Background(), map[string]any{
		"query":       "test",
		"num_results": float32(6),
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Success {
		t.Errorf("expected success: %s", result.Error)
	}
}

func TestTodoTool_Cancel(t *testing.T) {
	t.Parallel()

	tool := &TodoTool{}
	tool.Execute(context.Background(), map[string]any{"action": "add", "content": "cancel me"})
	result, err := tool.Execute(context.Background(), map[string]any{"action": "cancel", "index": 1})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Success {
		t.Errorf("expected success: %s", result.Error)
	}
}

func TestTodoTool_CompleteInvalidIndexType(t *testing.T) {
	t.Parallel()

	tool := &TodoTool{}
	result, err := tool.Execute(context.Background(), map[string]any{"action": "complete", "index": "abc"})
	if err != nil {
		t.Fatal(err)
	}
	if result.Success {
		t.Error("expected failure for invalid index type")
	}
}

func TestTodoTool_CancelInvalidIndexType(t *testing.T) {
	t.Parallel()

	tool := &TodoTool{}
	result, err := tool.Execute(context.Background(), map[string]any{"action": "cancel", "index": "abc"})
	if err != nil {
		t.Fatal(err)
	}
	if result.Success {
		t.Error("expected failure for invalid index type")
	}
}

func TestPermissionLevel_String_Unknown(t *testing.T) {
	t.Parallel()

	level := PermissionLevel(99)
	if level.String() != "unknown" {
		t.Errorf("expected unknown, got %s", level.String())
	}
}

func TestToolRegistry_RegisterNilTool(t *testing.T) {
	t.Parallel()

	r := NewRegistry()
	err := r.Register(nil)
	if err == nil {
		t.Error("expected error for nil tool")
	}
}

func TestOutputTruncate_NilResult(t *testing.T) {
	t.Parallel()

	OutputTruncate(nil, 10)
}

func TestOutputTruncate_OneByteOver(t *testing.T) {
	t.Parallel()

	result := &types.ToolResult{Output: strings.Repeat("a", 11)}
	OutputTruncate(result, 10)
	if !strings.Contains(result.Output, "omitted") {
		t.Errorf("expected truncation notice, got %q", result.Output)
	}
	if len(result.Output) > 50 {
		t.Errorf("output too long after truncation: %d chars", len(result.Output))
	}
}

func TestWriteTool_WriteToExistingFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "existing.txt")
	os.WriteFile(path, []byte("old"), 0644)

	tool := &WriteTool{}
	result, err := tool.Execute(context.Background(), map[string]any{
		"path":    path,
		"content": "new",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Success {
		t.Errorf("expected success: %s", result.Error)
	}

	data, _ := os.ReadFile(path)
	if string(data) != "new" {
		t.Errorf("expected 'new', got %s", string(data))
	}
}

func TestGrepTool_DefaultPath(t *testing.T) {
	t.Parallel()

	tool := &GrepTool{}
	result, err := tool.Execute(context.Background(), map[string]any{
		"pattern": "nonexistent_pattern_xyz_12345",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Success {
		t.Errorf("expected success: %s", result.Error)
	}
}

func TestReadTool_ReadEmptyFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "empty.txt")
	os.WriteFile(path, []byte(""), 0644)

	tool := &ReadTool{}
	result, err := tool.Execute(context.Background(), map[string]any{"path": path})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Success {
		t.Errorf("expected success: %s", result.Error)
	}
}

func TestWebFetchTool_InvalidHostname(t *testing.T) {
	t.Parallel()

	tool := &WebFetchTool{}
	result, err := tool.Execute(context.Background(), map[string]any{"url": "http://[::1]:namedport"})
	if err != nil {
		t.Fatal(err)
	}
	if result.Success {
		t.Error("expected failure for invalid hostname")
	}
}

func TestWebSearchTool_NumResultsAsInt32_SKIP(t *testing.T) {
	t.Skip("requires network")
	t.Parallel()

	tool := &WebSearchTool{}
	result, err := tool.Execute(context.Background(), map[string]any{
		"query":       "test",
		"num_results": int32(8),
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Success {
		t.Errorf("expected success: %s", result.Error)
	}
}

func TestBashTool_OutputTruncation(t *testing.T) {
	t.Parallel()

	tool := &BashTool{}
	large := strings.Repeat("x", 110000)
	result, err := tool.Execute(context.Background(), map[string]any{
		"command": "printf '%s' " + large,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Output) > 102400+50 {
		t.Errorf("expected truncated output, got %d chars", len(result.Output))
	}
}

func TestDelegateTool_ListChildrenNoParent(t *testing.T) {
	t.Parallel()

	tool := NewDelegateTool()
	tool.SpawnChild("parent-1", "task", "", 1)
	children := tool.ListChildren("nonexistent-parent")
	if len(children) != 0 {
		t.Errorf("expected 0 children, got %d", len(children))
	}
}

func TestWebFetchTool_Success(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<html><body>Hello World</body></html>"))
	}))
	defer server.Close()

	tool := &WebFetchTool{}
	result, err := tool.Execute(context.Background(), map[string]any{
		"url": server.URL,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Success {
		t.Errorf("expected success: %s", result.Error)
	}
	if !strings.Contains(result.Output, "Hello World") {
		t.Errorf("expected 'Hello World', got %s", result.Output)
	}
}

func TestWebFetchTool_HTTPError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("not found"))
	}))
	defer server.Close()

	tool := &WebFetchTool{}
	result, err := tool.Execute(context.Background(), map[string]any{
		"url": server.URL,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Success {
		t.Error("expected failure for 404")
	}
	if !strings.Contains(result.Error, "404") {
		t.Errorf("expected 404 in error, got %s", result.Error)
	}
}

func TestWebFetchTool_ConnectionError(t *testing.T) {
	t.Parallel()

	tool := &WebFetchTool{}
	result, err := tool.Execute(context.Background(), map[string]any{
		"url": "http://127.0.0.1:1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Success {
		t.Error("expected failure for connection error")
	}
}

func TestToInt_Int32(t *testing.T) {
	t.Parallel()

	val, err := toInt(int32(42))
	if err != nil {
		t.Fatal(err)
	}
	if val != 42 {
		t.Errorf("expected 42, got %d", val)
	}
}

func TestToInt_Float32(t *testing.T) {
	t.Parallel()

	val, err := toInt(float32(3.14))
	if err != nil {
		t.Fatal(err)
	}
	if val != 3 {
		t.Errorf("expected 3, got %d", val)
	}
}

func TestToInt_String(t *testing.T) {
	t.Parallel()

	val, err := toInt("99")
	if err != nil {
		t.Fatal(err)
	}
	if val != 99 {
		t.Errorf("expected 99, got %d", val)
	}
}

func TestToInt_StringInvalid(t *testing.T) {
	t.Parallel()

	_, err := toInt("not-a-number")
	if err == nil {
		t.Error("expected error for invalid string")
	}
}

func TestToInt_Default(t *testing.T) {
	t.Parallel()

	_, err := toInt(true)
	if err == nil {
		t.Error("expected error for unsupported type")
	}
}

func TestToInt_Int(t *testing.T) {
	t.Parallel()

	val, err := toInt(7)
	if err != nil {
		t.Fatal(err)
	}
	if val != 7 {
		t.Errorf("expected 7, got %d", val)
	}
}

func TestToInt_Int64(t *testing.T) {
	t.Parallel()

	val, err := toInt(int64(100))
	if err != nil {
		t.Fatal(err)
	}
	if val != 100 {
		t.Errorf("expected 100, got %d", val)
	}
}

func TestToInt_Float64(t *testing.T) {
	t.Parallel()

	val, err := toInt(float64(5.9))
	if err != nil {
		t.Fatal(err)
	}
	if val != 5 {
		t.Errorf("expected 5, got %d", val)
	}
}

func TestTodoTool_CompleteWithInt32Index(t *testing.T) {
	t.Parallel()

	tool := &TodoTool{}
	tool.Execute(context.Background(), map[string]any{"action": "add", "content": "test"})
	result, err := tool.Execute(context.Background(), map[string]any{
		"action": "complete",
		"index":  int32(1),
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Success {
		t.Errorf("expected success: %s", result.Error)
	}
}

func TestTodoTool_CancelWithInt32Index(t *testing.T) {
	t.Parallel()

	tool := &TodoTool{}
	tool.Execute(context.Background(), map[string]any{"action": "add", "content": "test"})
	result, err := tool.Execute(context.Background(), map[string]any{
		"action": "cancel",
		"index":  int32(1),
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Success {
		t.Errorf("expected success: %s", result.Error)
	}
}

func TestBashTool_Float64Timeout(t *testing.T) {
	t.Parallel()

	tool := &BashTool{}
	result, err := tool.Execute(context.Background(), map[string]any{
		"command": "echo ok",
		"timeout": float64(5),
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Success {
		t.Errorf("expected success: %s", result.Error)
	}
}

func TestBashTool_Int32Timeout(t *testing.T) {
	t.Parallel()

	tool := &BashTool{}
	result, err := tool.Execute(context.Background(), map[string]any{
		"command": "echo ok",
		"timeout": int32(5),
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Success {
		t.Errorf("expected success: %s", result.Error)
	}
}

func TestBashTool_StringTimeout(t *testing.T) {
	t.Parallel()

	tool := &BashTool{}
	result, err := tool.Execute(context.Background(), map[string]any{
		"command": "echo ok",
		"timeout": "5",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Success {
		t.Errorf("expected success: %s", result.Error)
	}
}

func TestDelegateTool_EmptyDescription(t *testing.T) {
	t.Parallel()

	tool := NewDelegateTool()
	result, err := tool.Execute(context.Background(), map[string]any{
		"description": "",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Success {
		t.Error("expected failure for empty description")
	}
}

func TestDelegateTool_NonStringDescription(t *testing.T) {
	t.Parallel()

	tool := NewDelegateTool()
	result, err := tool.Execute(context.Background(), map[string]any{
		"description": 123,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Success {
		t.Error("expected failure for non-string description")
	}
}

func TestWebSearchTool_Int32NumResults_SKIP(t *testing.T) {
	t.Skip("requires network")
	t.Parallel()

	tool := &WebSearchTool{}
	result, err := tool.Execute(context.Background(), map[string]any{
		"query":       "test",
		"num_results": int32(5),
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Success {
		t.Errorf("expected success: %s", result.Error)
	}
}

func TestWebSearchTool_Float32NumResults_SKIP(t *testing.T) {
	t.Skip("requires network")
	t.Parallel()

	tool := &WebSearchTool{}
	result, err := tool.Execute(context.Background(), map[string]any{
		"query":       "test",
		"num_results": float32(5),
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Success {
		t.Errorf("expected success: %s", result.Error)
	}
}

func TestWebSearchTool_IntNumResults_SKIP(t *testing.T) {
	t.Skip("requires network")
	t.Parallel()

	tool := &WebSearchTool{}
	result, err := tool.Execute(context.Background(), map[string]any{
		"query":       "test",
		"num_results": 7,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Success {
		t.Errorf("expected success: %s", result.Error)
	}
	if !strings.Contains(result.Output, "results=7") {
		t.Errorf("expected results=7, got %s", result.Output)
	}
}

func TestWebSearchTool_Int64NumResults(t *testing.T) {
	t.Parallel()
	t.Skip("requires network")

	tool := &WebSearchTool{}
	result, err := tool.Execute(context.Background(), map[string]any{
		"query":       "test",
		"num_results": int64(9),
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Success {
		t.Errorf("expected success: %s", result.Error)
	}
}

func TestWebSearchTool_StringNumResults_SKIP(t *testing.T) {
	t.Skip("requires network")
	t.Parallel()

	tool := &WebSearchTool{}
	result, err := tool.Execute(context.Background(), map[string]any{
		"query":       "test",
		"num_results": "11",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Success {
		t.Errorf("expected success: %s", result.Error)
	}
}

func TestWebSearchTool_BoolNumResults_SKIP(t *testing.T) {
	t.Skip("requires network")
	t.Parallel()

	tool := &WebSearchTool{}
	result, err := tool.Execute(context.Background(), map[string]any{
		"query":       "test",
		"num_results": true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Success {
		t.Errorf("expected success: %s", result.Error)
	}
}

func TestTodoTool_CompleteWithFloat32Index(t *testing.T) {
	t.Parallel()

	tool := &TodoTool{}
	tool.Execute(context.Background(), map[string]any{"action": "add", "content": "test"})
	result, err := tool.Execute(context.Background(), map[string]any{
		"action": "complete",
		"index":  float32(1),
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Success {
		t.Errorf("expected success: %s", result.Error)
	}
}

func TestTodoTool_CancelWithFloat32Index(t *testing.T) {
	t.Parallel()

	tool := &TodoTool{}
	tool.Execute(context.Background(), map[string]any{"action": "add", "content": "test"})
	result, err := tool.Execute(context.Background(), map[string]any{
		"action": "cancel",
		"index":  float32(1),
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Success {
		t.Errorf("expected success: %s", result.Error)
	}
}

func TestTodoTool_CompleteWithStringIndex(t *testing.T) {
	t.Parallel()

	tool := &TodoTool{}
	tool.Execute(context.Background(), map[string]any{"action": "add", "content": "test"})
	result, err := tool.Execute(context.Background(), map[string]any{
		"action": "complete",
		"index":  "1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Success {
		t.Errorf("expected success: %s", result.Error)
	}
}

func TestTodoTool_CancelWithStringIndex(t *testing.T) {
	t.Parallel()

	tool := &TodoTool{}
	tool.Execute(context.Background(), map[string]any{"action": "add", "content": "test"})
	result, err := tool.Execute(context.Background(), map[string]any{
		"action": "cancel",
		"index":  "1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Success {
		t.Errorf("expected success: %s", result.Error)
	}
}

func TestTodoTool_CompleteWithIntIndex(t *testing.T) {
	t.Parallel()

	tool := &TodoTool{}
	tool.Execute(context.Background(), map[string]any{"action": "add", "content": "test"})
	result, err := tool.Execute(context.Background(), map[string]any{
		"action": "complete",
		"index":  1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Success {
		t.Errorf("expected success: %s", result.Error)
	}
}

func TestTodoTool_CancelWithIntIndex(t *testing.T) {
	t.Parallel()

	tool := &TodoTool{}
	tool.Execute(context.Background(), map[string]any{"action": "add", "content": "test"})
	result, err := tool.Execute(context.Background(), map[string]any{
		"action": "cancel",
		"index":  1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Success {
		t.Errorf("expected success: %s", result.Error)
	}
}

func TestTodoTool_CompleteWithInt64Index(t *testing.T) {
	t.Parallel()

	tool := &TodoTool{}
	tool.Execute(context.Background(), map[string]any{"action": "add", "content": "test"})
	result, err := tool.Execute(context.Background(), map[string]any{
		"action": "complete",
		"index":  int64(1),
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Success {
		t.Errorf("expected success: %s", result.Error)
	}
}

func TestTodoTool_CancelWithInt64Index(t *testing.T) {
	t.Parallel()

	tool := &TodoTool{}
	tool.Execute(context.Background(), map[string]any{"action": "add", "content": "test"})
	result, err := tool.Execute(context.Background(), map[string]any{
		"action": "cancel",
		"index":  int64(1),
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Success {
		t.Errorf("expected success: %s", result.Error)
	}
}

func TestTodoTool_CompleteWithFloat64Index(t *testing.T) {
	t.Parallel()

	tool := &TodoTool{}
	tool.Execute(context.Background(), map[string]any{"action": "add", "content": "test"})
	result, err := tool.Execute(context.Background(), map[string]any{
		"action": "complete",
		"index":  float64(1),
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Success {
		t.Errorf("expected success: %s", result.Error)
	}
}

func TestTodoTool_CancelWithFloat64Index(t *testing.T) {
	t.Parallel()

	tool := &TodoTool{}
	tool.Execute(context.Background(), map[string]any{"action": "add", "content": "test"})
	result, err := tool.Execute(context.Background(), map[string]any{
		"action": "cancel",
		"index":  float64(1),
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Success {
		t.Errorf("expected success: %s", result.Error)
	}
}

func TestTodoTool_CompleteWithBoolIndex(t *testing.T) {
	t.Parallel()

	tool := &TodoTool{}
	tool.Execute(context.Background(), map[string]any{"action": "add", "content": "test"})
	result, err := tool.Execute(context.Background(), map[string]any{
		"action": "complete",
		"index":  true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Success {
		t.Error("expected failure for bool index")
	}
}

func TestTodoTool_CancelWithBoolIndex(t *testing.T) {
	t.Parallel()

	tool := &TodoTool{}
	tool.Execute(context.Background(), map[string]any{"action": "add", "content": "test"})
	result, err := tool.Execute(context.Background(), map[string]any{
		"action": "cancel",
		"index":  true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Success {
		t.Error("expected failure for bool index")
	}
}

func TestReadTool_IsDir(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	tool := &ReadTool{}
	result, err := tool.Execute(context.Background(), map[string]any{"path": dir})
	if err != nil {
		t.Fatal(err)
	}
	if result.Success {
		t.Error("expected failure for directory")
	}
}

func TestWriteTool_InvalidPathType(t *testing.T) {
	t.Parallel()

	tool := &WriteTool{}
	result, err := tool.Execute(context.Background(), map[string]any{
		"path":    123,
		"content": "test",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Success {
		t.Error("expected failure for invalid path type")
	}
}

func TestWriteTool_InvalidContentType(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	tool := &WriteTool{}
	result, err := tool.Execute(context.Background(), map[string]any{
		"path":    dir + "/f.txt",
		"content": 123,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Success {
		t.Error("expected failure for invalid content type")
	}
}

func TestEditTool_InvalidPathType(t *testing.T) {
	t.Parallel()

	tool := &EditTool{}
	result, err := tool.Execute(context.Background(), map[string]any{
		"path":       123,
		"old_string": "x",
		"new_string": "y",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Success {
		t.Error("expected failure for invalid path type")
	}
}

func TestEditTool_MissingOldString(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := dir + "/f.txt"
	os.WriteFile(path, []byte("hello"), 0644)

	tool := &EditTool{}
	result, err := tool.Execute(context.Background(), map[string]any{
		"path":       path,
		"new_string": "y",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Success {
		t.Error("expected failure for missing old_string")
	}
}

func TestEditTool_InvalidOldStringType(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := dir + "/f.txt"
	os.WriteFile(path, []byte("hello"), 0644)

	tool := &EditTool{}
	result, err := tool.Execute(context.Background(), map[string]any{
		"path":       path,
		"old_string": 123,
		"new_string": "y",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Success {
		t.Error("expected failure for invalid old_string type")
	}
}

func TestEditTool_MissingNewString(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := dir + "/f.txt"
	os.WriteFile(path, []byte("hello"), 0644)

	tool := &EditTool{}
	result, err := tool.Execute(context.Background(), map[string]any{
		"path":       path,
		"old_string": "hello",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Success {
		t.Error("expected failure for missing new_string")
	}
}

func TestEditTool_InvalidNewStringType(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := dir + "/f.txt"
	os.WriteFile(path, []byte("hello"), 0644)

	tool := &EditTool{}
	result, err := tool.Execute(context.Background(), map[string]any{
		"path":       path,
		"old_string": "hello",
		"new_string": 123,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Success {
		t.Error("expected failure for invalid new_string type")
	}
}

func TestGlobTool_InvalidPatternType(t *testing.T) {
	t.Parallel()

	tool := &GlobTool{}
	result, err := tool.Execute(context.Background(), map[string]any{
		"pattern": 123,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Success {
		t.Error("expected failure for invalid pattern type")
	}
}

func TestGrepTool_InvalidPathType(t *testing.T) {
	t.Parallel()

	tool := &GrepTool{}
	result, err := tool.Execute(context.Background(), map[string]any{
		"pattern": "nonexistent_xyz_12345",
		"path":    123,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Success {
		t.Errorf("expected success (falls back to default path): %s", result.Error)
	}
}

func TestWriteTool_ReadOnlyDir(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	os.Chmod(dir, 0555)
	defer os.Chmod(dir, 0755)

	tool := &WriteTool{}
	result, err := tool.Execute(context.Background(), map[string]any{
		"path":    filepath.Join(dir, "test.txt"),
		"content": "test",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Success {
		t.Error("expected failure for read-only dir")
	}
}

func TestEditTool_ReadError(t *testing.T) {
	t.Parallel()

	tool := &EditTool{}
	result, err := tool.Execute(context.Background(), map[string]any{
		"path":       "/proc/1/environ",
		"old_string": "x",
		"new_string": "y",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Success {
		t.Error("expected failure for unreadable file")
	}
}

func TestEditTool_WriteError(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "readonly.txt")
	os.WriteFile(path, []byte("hello world"), 0444)

	tool := &EditTool{}
	result, err := tool.Execute(context.Background(), map[string]any{
		"path":       path,
		"old_string": "world",
		"new_string": "go",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Success {
		t.Error("expected failure for read-only file")
	}
}

func TestGlobTool_InvalidPatternError(t *testing.T) {
	t.Parallel()

	tool := &GlobTool{}
	result, err := tool.Execute(context.Background(), map[string]any{
		"pattern": "[invalid",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Success {
		t.Error("expected failure for invalid glob pattern")
	}
}

func TestGrepTool_NonDirPath(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	filePath := filepath.Join(dir, "test.go")
	os.WriteFile(filePath, []byte("package main"), 0644)

	tool := &GrepTool{}
	result, err := tool.Execute(context.Background(), map[string]any{
		"pattern": "package",
		"path":    filePath,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Success {
		t.Errorf("expected success: %s", result.Error)
	}
}

func TestGrepTool_WalkDirError(t *testing.T) {
	t.Parallel()

	tool := &GrepTool{}
	result, err := tool.Execute(context.Background(), map[string]any{
		"pattern": "test",
		"path":    "/proc/1/root",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Success {
		t.Errorf("expected success (graceful skip): %s", result.Error)
	}
}

func TestWebFetchTool_ReadError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hj, ok := w.(http.Hijacker)
		if !ok {
			w.WriteHeader(http.StatusOK)
			return
		}
		conn, _, _ := hj.Hijack()
		conn.Close()
	}))
	defer server.Close()

	tool := &WebFetchTool{}
	result, err := tool.Execute(context.Background(), map[string]any{
		"url": server.URL,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Success {
		t.Error("expected failure for read error")
	}
}
