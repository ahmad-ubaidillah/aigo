package subagent

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// parseTasks extracts task definitions from LLM response.
func parseTasks(response string) ([]Task, error) {
	// Try to extract JSON array from response
	jsonStr := extractJSON(response, "[")
	if jsonStr == "" {
		return nil, fmt.Errorf("no JSON array found in response")
	}

	var tasks []Task
	if err := json.Unmarshal([]byte(jsonStr), &tasks); err != nil {
		return nil, fmt.Errorf("parse tasks: %w", err)
	}

	// Validate and fix
	for i := range tasks {
		if tasks[i].ID == "" {
			tasks[i].ID = fmt.Sprintf("t%d", i+1)
		}
		if tasks[i].Role == "" {
			tasks[i].Role = RoleBuilder
		}
		if tasks[i].Category == "" {
			tasks[i].Category = CategoryGeneral
		}
		if tasks[i].Priority == 0 {
			tasks[i].Priority = i + 1
		}
	}

	return tasks, nil
}

// parseIntent extracts intent analysis from LLM response.
func parseIntent(response string) (*IntentResult, error) {
	jsonStr := extractJSON(response, "{")
	if jsonStr == "" {
		return nil, fmt.Errorf("no JSON object found")
	}

	var result IntentResult
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("parse intent: %w", err)
	}

	return &result, nil
}

// extractJSON finds the first JSON object or array in a string.
func extractJSON(s string, startChar string) string {
	var start, depth int
	var inString bool
	var escape bool

	for i, c := range s {
		if escape {
			escape = false
			continue
		}
		if c == '\\' {
			escape = true
			continue
		}
		if c == '"' {
			inString = !inString
			continue
		}
		if inString {
			continue
		}

		if string(c) == startChar {
			if depth == 0 {
				start = i
			}
			depth++
		} else if (startChar == "[" && c == ']') || (startChar == "{" && c == '}') {
			depth--
			if depth == 0 {
				return s[start : i+1]
			}
		}
	}
	return ""
}

// truncateStr truncates a string to maxLen with ellipsis.
func truncateStr(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// parseTaskID extracts a task ID from a string like "t1" or "task-1".
var taskIDRegex = regexp.MustCompile(`t(\d+)`)

func extractTaskID(s string) string {
	matches := taskIDRegex.FindStringSubmatch(s)
	if len(matches) > 0 {
		return "t" + matches[1]
	}
	return ""
}

// sanitizeForPrompt cleans a string for use in a prompt.
func sanitizeForPrompt(s string) string {
	s = strings.TrimSpace(s)
	// Remove excessive whitespace
	spaceRegex := regexp.MustCompile(`\s+`)
	s = spaceRegex.ReplaceAllString(s, " ")
	return s
}
