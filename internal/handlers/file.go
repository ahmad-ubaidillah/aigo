package handlers

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ahmad-ubaidillah/aigo/pkg/types"
)

const maxReadSize = 10 * 1024

type FileHandler struct{}

func (h *FileHandler) CanHandle(intent string) bool {
	return intent == types.IntentFile
}

func (h *FileHandler) Read(path string) (*types.ToolResult, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open file %q: %w", path, err)
	}
	defer file.Close()

	content, err := io.ReadAll(io.LimitReader(file, maxReadSize))
	if err != nil {
		return nil, fmt.Errorf("read file %q: %w", path, err)
	}

	return &types.ToolResult{
		Success: true,
		Output:  string(content),
	}, nil
}

func (h *FileHandler) Write(path, content string) (*types.ToolResult, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("create dirs for %q: %w", path, err)
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return nil, fmt.Errorf("write file %q: %w", path, err)
	}

	return &types.ToolResult{
		Success: true,
		Output:  fmt.Sprintf("wrote %d bytes to %s", len(content), path),
	}, nil
}

func (h *FileHandler) List(dir string) (*types.ToolResult, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("list dir %q: %w", dir, err)
	}

	var lines []string
	for _, entry := range entries {
		entryType := "file"
		if entry.IsDir() {
			entryType = "dir"
		}
		lines = append(lines, fmt.Sprintf("[%s] %s", entryType, entry.Name()))
	}

	return &types.ToolResult{
		Success: true,
		Output:  strings.Join(lines, "\n"),
	}, nil
}

func (h *FileHandler) Copy(src, dst string) (*types.ToolResult, error) {
	data, err := os.ReadFile(src)
	if err != nil {
		return nil, fmt.Errorf("read source %q: %w", src, err)
	}

	if err := os.WriteFile(dst, data, 0644); err != nil {
		return nil, fmt.Errorf("write dest %q: %w", dst, err)
	}

	return &types.ToolResult{
		Success: true,
		Output:  fmt.Sprintf("copied %s to %s", src, dst),
	}, nil
}

func (h *FileHandler) Move(src, dst string) (*types.ToolResult, error) {
	if err := os.Rename(src, dst); err != nil {
		return nil, fmt.Errorf("move %q to %q: %w", src, dst, err)
	}

	return &types.ToolResult{
		Success: true,
		Output:  fmt.Sprintf("moved %s to %s", src, dst),
	}, nil
}

func (h *FileHandler) Delete(path string) (*types.ToolResult, error) {
	if err := os.Remove(path); err != nil {
		return nil, fmt.Errorf("delete %q: %w", path, err)
	}

	return &types.ToolResult{
		Success: true,
		Output:  fmt.Sprintf("deleted %s", path),
	}, nil
}

func (h *FileHandler) Search(dir, pattern string) (*types.ToolResult, error) {
	fullPattern := filepath.Join(dir, pattern)
	matches, err := filepath.Glob(fullPattern)
	if err != nil {
		return nil, fmt.Errorf("glob pattern %q: %w", pattern, err)
	}

	return &types.ToolResult{
		Success: true,
		Output:  strings.Join(matches, "\n"),
	}, nil
}

func (h *FileHandler) Execute(ctx context.Context, task *types.Task, workspace string) (*types.ToolResult, error) {
	desc := strings.TrimSpace(task.Description)
	parts := strings.SplitN(desc, " ", 2)
	if len(parts) < 2 {
		return &types.ToolResult{
			Success: false,
			Error:   "unknown file command format",
		}, nil
	}

	command := parts[0]
	args := parts[1]

	switch command {
	case "read":
		return h.Read(args)
	case "write":
		writeParts := strings.SplitN(args, " ", 2)
		if len(writeParts) < 2 {
			return &types.ToolResult{
				Success: false,
				Error:   "write requires path and content",
			}, nil
		}
		return h.Write(writeParts[0], writeParts[1])
	case "list":
		return h.List(args)
	case "copy":
		copyParts := strings.SplitN(args, " ", 2)
		if len(copyParts) < 2 {
			return &types.ToolResult{
				Success: false,
				Error:   "copy requires source and destination",
			}, nil
		}
		return h.Copy(copyParts[0], copyParts[1])
	case "move":
		moveParts := strings.SplitN(args, " ", 2)
		if len(moveParts) < 2 {
			return &types.ToolResult{
				Success: false,
				Error:   "move requires source and destination",
			}, nil
		}
		return h.Move(moveParts[0], moveParts[1])
	case "delete":
		return h.Delete(args)
	case "search":
		searchParts := strings.SplitN(args, " ", 2)
		if len(searchParts) < 2 {
			return &types.ToolResult{
				Success: false,
				Error:   "search requires directory and pattern",
			}, nil
		}
		return h.Search(searchParts[0], searchParts[1])
	default:
		return &types.ToolResult{
			Success: false,
			Error:   fmt.Sprintf("unknown file command: %s", command),
		}, nil
	}
}
