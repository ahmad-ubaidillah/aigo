// Package tools provides file operation tools for Aigo's tool system.
package tools

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ahmad-ubaidillah/aigo/pkg/types"
)

const (
	maxReadSize  = 50 * 1024 // 50KB
	maxGrepLines = 100
)

// helper: getStringParam extracts a string parameter from params map.
func getStringParam(params map[string]any, key string) (string, error) {
	val, exists := params[key]
	if !exists {
		return "", fmt.Errorf("missing required parameter: %s", key)
	}
	str, ok := val.(string)
	if !ok {
		return "", fmt.Errorf("parameter %s must be a string", key)
	}
	return str, nil
}

// ============================================================
// ReadTool - reads file content
// ============================================================

// ReadTool reads file content from the filesystem.
type ReadTool struct{}

func (t *ReadTool) Name() string { return "read" }

func (t *ReadTool) Description() string {
	return "Reads file content from the filesystem. Returns error if file doesn't exist or exceeds 50KB."
}

func (t *ReadTool) Schema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"path": map[string]any{
				"type":        "string",
				"description": "The absolute path to the file to read",
			},
		},
		"required": []string{"path"},
	}
}

func (t *ReadTool) Execute(ctx context.Context, params map[string]any) (*types.ToolResult, error) {
	path, err := getStringParam(params, "path")
	if err != nil {
		return &types.ToolResult{Success: false, Error: err.Error()}, nil
	}

	info, err := os.Stat(path)
	if err != nil {
		return &types.ToolResult{Success: false, Error: fmt.Sprintf("file not found: %s", path)}, nil
	}

	if info.Size() > maxReadSize {
		return &types.ToolResult{
			Success: false,
			Error:   fmt.Sprintf("file too large: %d bytes (max %d)", info.Size(), maxReadSize),
		}, nil
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return &types.ToolResult{Success: false, Error: fmt.Sprintf("failed to read file: %v", err)}, nil
	}

	return &types.ToolResult{
		Success: true,
		Output:  string(content),
		Metadata: map[string]string{
			"path": path,
			"size": fmt.Sprintf("%d", len(content)),
		},
	}, nil
}

// ============================================================
// WriteTool - writes content to file
// ============================================================

// WriteTool writes content to a file, creating parent directories if needed.
type WriteTool struct{}

func (t *WriteTool) Name() string { return "write" }

func (t *WriteTool) Description() string {
	return "Writes content to a file. Creates parent directories if they don't exist."
}

func (t *WriteTool) Schema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"path": map[string]any{
				"type":        "string",
				"description": "The absolute path to the file to write",
			},
			"content": map[string]any{
				"type":        "string",
				"description": "The content to write to the file",
			},
		},
		"required": []string{"path", "content"},
	}
}

func (t *WriteTool) Execute(ctx context.Context, params map[string]any) (*types.ToolResult, error) {
	path, err := getStringParam(params, "path")
	if err != nil {
		return &types.ToolResult{Success: false, Error: err.Error()}, nil
	}

	content, err := getStringParam(params, "content")
	if err != nil {
		return &types.ToolResult{Success: false, Error: err.Error()}, nil
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return &types.ToolResult{
			Success: false,
			Error:   fmt.Sprintf("failed to create directories: %v", err),
		}, nil
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return &types.ToolResult{
			Success: false,
			Error:   fmt.Sprintf("failed to write file: %v", err),
		}, nil
	}

	return &types.ToolResult{
		Success: true,
		Output:  fmt.Sprintf("Successfully wrote %d bytes to %s", len(content), path),
		Metadata: map[string]string{
			"path": path,
			"size": fmt.Sprintf("%d", len(content)),
		},
	}, nil
}

// ============================================================
// EditTool - replaces first occurrence of string in file
// ============================================================

// EditTool performs string replacement in a file.
type EditTool struct{}

func (t *EditTool) Name() string { return "edit" }

func (t *EditTool) Description() string {
	return "Edits a file by replacing the first occurrence of old_string with new_string. " +
		"Returns error if old_string not found or appears multiple times."
}

func (t *EditTool) Schema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"path": map[string]any{
				"type":        "string",
				"description": "The absolute path to the file to edit",
			},
			"old_string": map[string]any{
				"type":        "string",
				"description": "The text to search for (must match exactly)",
			},
			"new_string": map[string]any{
				"type":        "string",
				"description": "The text to replace old_string with",
			},
		},
		"required": []string{"path", "old_string", "new_string"},
	}
}

func (t *EditTool) Execute(ctx context.Context, params map[string]any) (*types.ToolResult, error) {
	path, err := getStringParam(params, "path")
	if err != nil {
		return &types.ToolResult{Success: false, Error: err.Error()}, nil
	}

	oldStr, err := getStringParam(params, "old_string")
	if err != nil {
		return &types.ToolResult{Success: false, Error: err.Error()}, nil
	}

	newStr, err := getStringParam(params, "new_string")
	if err != nil {
		return &types.ToolResult{Success: false, Error: err.Error()}, nil
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return &types.ToolResult{Success: false, Error: fmt.Sprintf("failed to read file: %v", err)}, nil
	}

	contentStr := string(content)
	count := strings.Count(contentStr, oldStr)

	if count == 0 {
		return &types.ToolResult{
			Success: false,
			Error:   fmt.Sprintf("old_string not found in file: %s", path),
		}, nil
	}

	if count > 1 {
		return &types.ToolResult{
			Success: false,
			Error:   fmt.Sprintf("old_string found %d times (expected exactly 1)", count),
		}, nil
	}

	newContent := strings.Replace(contentStr, oldStr, newStr, 1)
	if err := os.WriteFile(path, []byte(newContent), 0644); err != nil {
		return &types.ToolResult{
			Success: false,
			Error:   fmt.Sprintf("failed to write file: %v", err),
		}, nil
	}

	return &types.ToolResult{
		Success: true,
		Output:  fmt.Sprintf("Successfully edited %s", path),
		Metadata: map[string]string{
			"path":       path,
			"replaced":   "1",
			"old_length": fmt.Sprintf("%d", len(contentStr)),
			"new_length": fmt.Sprintf("%d", len(newContent)),
		},
	}, nil
}

// ============================================================
// GlobTool - finds files matching pattern
// ============================================================

// GlobTool finds files matching a glob pattern.
type GlobTool struct{}

func (t *GlobTool) Name() string { return "glob" }

func (t *GlobTool) Description() string {
	return "Finds files matching a glob pattern (e.g., *.go, **/*.ts). Returns matching paths."
}

func (t *GlobTool) Schema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"pattern": map[string]any{
				"type":        "string",
				"description": "The glob pattern to match files against",
			},
		},
		"required": []string{"pattern"},
	}
}

func (t *GlobTool) Execute(ctx context.Context, params map[string]any) (*types.ToolResult, error) {
	pattern, err := getStringParam(params, "pattern")
	if err != nil {
		return &types.ToolResult{Success: false, Error: err.Error()}, nil
	}

	matches, err := filepath.Glob(pattern)
	if err != nil {
		return &types.ToolResult{
			Success: false,
			Error:   fmt.Sprintf("invalid glob pattern: %v", err),
		}, nil
	}

	if len(matches) == 0 {
		return &types.ToolResult{
			Success: true,
			Output:  "No files found matching pattern",
			Metadata: map[string]string{
				"pattern": pattern,
				"count":   "0",
			},
		}, nil
	}

	return &types.ToolResult{
		Success: true,
		Output:  strings.Join(matches, "\n"),
		Metadata: map[string]string{
			"pattern": pattern,
			"count":   fmt.Sprintf("%d", len(matches)),
		},
	}, nil
}

// ============================================================
// GrepTool - searches files for regex pattern
// ============================================================

// GrepTool searches files for lines matching a regex pattern.
type GrepTool struct{}

func (t *GrepTool) Name() string { return "grep" }

func (t *GrepTool) Description() string {
	return "Searches files for lines matching a regex pattern. Returns file:line:content format. " +
		"Searches .go, .ts, .js, .py, .md, .txt, .yaml, .yml, .json files. Max 100 results."
}

func (t *GrepTool) Schema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"pattern": map[string]any{
				"type":        "string",
				"description": "The regex pattern to search for",
			},
			"path": map[string]any{
				"type":        "string",
				"description": "The directory to search in (defaults to current directory)",
			},
		},
		"required": []string{"pattern"},
	}
}

// allowedExtensions contains file extensions to search in.
var allowedExtensions = map[string]bool{
	".go": true, ".ts": true, ".js": true, ".py": true, ".md": true,
	".txt": true, ".yaml": true, ".yml": true, ".json": true,
}

func (t *GrepTool) Execute(ctx context.Context, params map[string]any) (*types.ToolResult, error) {
	pattern, err := getStringParam(params, "pattern")
	if err != nil {
		return &types.ToolResult{Success: false, Error: err.Error()}, nil
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return &types.ToolResult{
			Success: false,
			Error:   fmt.Sprintf("invalid regex pattern: %v", err),
		}, nil
	}

	searchPath := "."
	if p, exists := params["path"]; exists {
		if str, ok := p.(string); ok && str != "" {
			searchPath = str
		}
	}

	var results []string
	err = filepath.WalkDir(searchPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // skip files/dirs we can't access
		}
		if d.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if !allowedExtensions[ext] {
			return nil
		}
		file, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer file.Close()
		scanner := bufio.NewScanner(file)
		lineNum := 0
		for scanner.Scan() && len(results) < maxGrepLines {
			lineNum++
			line := scanner.Text()
			if re.MatchString(line) {
				results = append(results, fmt.Sprintf("%s:%d:%s", path, lineNum, line))
			}
		}
		return nil
	})
	if err != nil {
		return &types.ToolResult{
			Success: false,
			Error:   fmt.Sprintf("error walking directory: %v", err),
		}, nil
	}

	if len(results) == 0 {
		return &types.ToolResult{
			Success: true,
			Output:  "No matches found",
			Metadata: map[string]string{
				"pattern": pattern,
				"path":    searchPath,
				"count":   "0",
			},
		}, nil
	}

	return &types.ToolResult{
		Success: true,
		Output:  strings.Join(results, "\n"),
		Metadata: map[string]string{
			"pattern": pattern,
			"path":    searchPath,
			"count":   fmt.Sprintf("%d", len(results)),
		},
	}, nil
}
