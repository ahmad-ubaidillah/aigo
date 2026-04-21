// Package tools internal implementations.
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// execCommand runs a shell command.
func execCommand(ctx context.Context, command string) (string, error) {
	cmd := exec.CommandContext(ctx, "bash", "-c", command)
	cmd.Env = os.Environ()
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// readFile reads a file's contents.
func readFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// writeFile writes content to a file.
func writeFile(path, content string) (string, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return "", err
	}
	return fmt.Sprintf("Wrote %d bytes to %s", len(content), path), nil
}

// searchFiles searches for a pattern in files.
func searchFiles(pattern, searchPath string) (string, error) {
	var results []string
	re, err := regexp.Compile(pattern)
	if err != nil {
		// Fallback to literal search
		re = regexp.MustCompile(regexp.QuoteMeta(pattern))
	}

	err = filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		// Skip binary files and large files
		if info.Size() > 1_000_000 {
			return nil
		}
		ext := filepath.Ext(path)
		if ext == ".exe" || ext == ".bin" || ext == ".so" || ext == ".dylib" {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		lines := strings.Split(string(data), "\n")
		for i, line := range lines {
			if re.MatchString(line) {
				results = append(results, fmt.Sprintf("%s:%d: %s", path, i+1, strings.TrimSpace(line)))
				if len(results) >= 50 {
					return filepath.SkipAll
				}
			}
		}
		return nil
	})

	if err != nil && len(results) == 0 {
		return "", err
	}
	if len(results) == 0 {
		return "No matches found.", nil
	}
	return strings.Join(results, "\n"), nil
}

// kvExecute handles KV store operations.
func kvExecute(storagePath, action, key, value string) (string, error) {
	if err := os.MkdirAll(storagePath, 0755); err != nil {
		return "", fmt.Errorf("create kv dir: %w", err)
	}

	// Validate key
	if action != "list" {
		if key == "" {
			return "", fmt.Errorf("key is required")
		}
		if strings.Contains(key, "..") || strings.Contains(key, "/") || strings.Contains(key, "\\") {
			return "", fmt.Errorf("invalid key: cannot contain path separators or '..'")
		}
	}

	kvPath := filepath.Join(storagePath, key)

	switch action {
	case "get":
		data, err := os.ReadFile(kvPath)
		if err != nil {
			return "", fmt.Errorf("key '%s' not found", key)
		}
		return string(data), nil

	case "set":
		if value == "" {
			return "", fmt.Errorf("value is required for set")
		}
		if err := os.WriteFile(kvPath, []byte(value), 0644); err != nil {
			return "", err
		}
		b, _ := json.Marshal(map[string]interface{}{"ok": true, "key": key, "size": len(value)})
		return string(b), nil

	case "list":
		entries, err := os.ReadDir(storagePath)
		if err != nil {
			return "[]", nil
		}
		keys := make([]string, 0, len(entries))
		for _, e := range entries {
			if !e.IsDir() {
				keys = append(keys, e.Name())
			}
		}
		b, _ := json.Marshal(map[string]interface{}{"keys": keys, "count": len(keys)})
		return string(b), nil

	case "delete":
		if err := os.Remove(kvPath); err != nil {
			return "", fmt.Errorf("key '%s' not found", key)
		}
		b, _ := json.Marshal(map[string]interface{}{"ok": true, "deleted": key})
		return string(b), nil

	default:
		return "", fmt.Errorf("unknown action: %s (use get/set/list/delete)", action)
	}
}

// currentTimeISO returns the current time in ISO-8601 format.
func currentTimeISO() string {
	return time.Now().UTC().Format(time.RFC3339)
}

// jsonMarshalToString is a helper to marshal JSON to string.
func jsonMarshalToString(v interface{}) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
