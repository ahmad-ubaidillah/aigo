package tools

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRegistryRegisterAndGet(t *testing.T) {
	reg := NewRegistry()
	reg.Register(&TerminalTool{})
	reg.Register(&ReadFileTool{})

	if reg.Count() != 2 {
		t.Errorf("expected 2 tools, got %d", reg.Count())
	}

	tool, ok := reg.Get("terminal")
	if !ok {
		t.Error("expected to find 'terminal' tool")
	}
	if tool.Name() != "terminal" {
		t.Errorf("expected 'terminal', got '%s'", tool.Name())
	}

	_, ok = reg.Get("nonexistent")
	if ok {
		t.Error("expected not to find 'nonexistent' tool")
	}
}

func TestRegistrySchemas(t *testing.T) {
	reg := NewRegistry()
	reg.Register(&TerminalTool{})
	reg.Register(&GetCurrentTimeTool{})

	schemas := reg.Schemas()
	if len(schemas) != 2 {
		t.Errorf("expected 2 schemas, got %d", len(schemas))
	}
}

func TestRegistryList(t *testing.T) {
	reg := NewRegistry()
	reg.Register(&TerminalTool{})
	reg.Register(&ReadFileTool{})
	reg.Register(&KVTool{storagePath: "/tmp"})

	names := reg.List()
	if len(names) != 3 {
		t.Errorf("expected 3 names, got %d", len(names))
	}
}

func TestRegistryExecute(t *testing.T) {
	reg := NewRegistry()
	reg.Register(&GetCurrentTimeTool{})

	result, err := reg.Execute(context.Background(), "get_current_time", map[string]interface{}{})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result, "time") {
		t.Errorf("expected time in result, got: %s", result)
	}
}

func TestRegistryExecuteNotFound(t *testing.T) {
	reg := NewRegistry()
	_, err := reg.Execute(context.Background(), "nonexistent", map[string]interface{}{})
	if err == nil {
		t.Error("expected error for nonexistent tool")
	}
}

func TestTerminalTool(t *testing.T) {
	tool := &TerminalTool{}

	if tool.Name() != "terminal" {
		t.Error("wrong name")
	}
	if tool.Annotations().Destructive != true {
		t.Error("terminal should be destructive")
	}

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"command": "echo hello",
	})
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(result) != "hello" {
		t.Errorf("expected 'hello', got '%s'", result)
	}
}

func TestTerminalToolMissingCommand(t *testing.T) {
	tool := &TerminalTool{}
	_, err := tool.Execute(context.Background(), map[string]interface{}{})
	if err == nil {
		t.Error("expected error for missing command")
	}
}

func TestReadFileTool(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(testFile, []byte("hello world"), 0644)

	tool := &ReadFileTool{}
	if tool.Annotations().ReadOnly != true {
		t.Error("read_file should be readOnly")
	}

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"path": testFile,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result != "hello world" {
		t.Errorf("expected 'hello world', got '%s'", result)
	}
}

func TestWriteFileTool(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "subdir", "test.txt")

	tool := &WriteFileTool{}
	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"path":    testFile,
		"content": "test content",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result, "Wrote") {
		t.Error("expected 'Wrote' in result")
	}

	// Verify file exists
	data, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "test content" {
		t.Error("file content mismatch")
	}
}

func TestSearchFilesTool(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "a.txt"), []byte("hello world\ntest line\n"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "b.txt"), []byte("another file\n"), 0644)

	tool := &SearchFilesTool{}
	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"pattern": "hello",
		"path":    tmpDir,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result, "hello world") {
		t.Errorf("expected match, got: %s", result)
	}
}

func TestKVTool(t *testing.T) {
	tmpDir := t.TempDir()
	tool := NewKVTool(tmpDir)

	// Set
	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"action": "set",
		"key":    "testkey",
		"value":  "testvalue",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result, "ok") {
		t.Errorf("expected ok, got: %s", result)
	}

	// Get
	result, err = tool.Execute(context.Background(), map[string]interface{}{
		"action": "get",
		"key":    "testkey",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result != "testvalue" {
		t.Errorf("expected 'testvalue', got '%s'", result)
	}

	// List
	result, err = tool.Execute(context.Background(), map[string]interface{}{
		"action": "list",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result, "testkey") {
		t.Errorf("expected testkey in list, got: %s", result)
	}

	// Delete
	result, err = tool.Execute(context.Background(), map[string]interface{}{
		"action": "delete",
		"key":    "testkey",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result, "deleted") {
		t.Errorf("expected deleted, got: %s", result)
	}
}

func TestKVToolPathTraversal(t *testing.T) {
	tmpDir := t.TempDir()
	tool := NewKVTool(tmpDir)

	_, err := tool.Execute(context.Background(), map[string]interface{}{
		"action": "set",
		"key":    "../../../etc/passwd",
		"value":  "hack",
	})
	if err == nil {
		t.Error("expected error for path traversal")
	}
}

func TestGetCurrentTimeTool(t *testing.T) {
	tool := &GetCurrentTimeTool{}
	result, err := tool.Execute(context.Background(), map[string]interface{}{})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result, "time") {
		t.Error("expected 'time' in result")
	}
}
