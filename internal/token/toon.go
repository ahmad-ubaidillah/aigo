package token

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/ahmad-ubaidillah/aigo/pkg/types"
)

// ToonFormatter implements the Toon format for compact output representation.
// Toon format reduces token usage by using compact symbols and structured formatting.
// Reference: https://github.com/toon-format/toon
type ToonFormatter struct {
	cfg *types.Config
}

// NewToonFormatter creates a new ToonFormatter instance.
func NewToonFormatter(cfg *types.Config) *ToonFormatter {
	return &ToonFormatter{cfg: cfg}
}

// Format converts output to Toon format for token-efficient representation.
func (t *ToonFormatter) Format(output string) string {
	if output == "" {
		return ""
	}

	// Try different format strategies based on content
	if t.isJSON(output) {
		return t.formatJSON(output)
	}

	if t.isError(output) {
		return t.formatError(output)
	}

	if t.isLog(output) {
		return t.formatLog(output)
	}

	// Default: compact text format
	return t.formatText(output)
}

// isJSON checks if output is JSON format.
func (t *ToonFormatter) isJSON(output string) bool {
	output = strings.TrimSpace(output)
	return strings.HasPrefix(output, "{") || strings.HasPrefix(output, "[")
}

// isError checks if output is an error message.
func (t *ToonFormatter) isError(output string) bool {
	outputLower := strings.ToLower(output)
	errorIndicators := []string{"error:", "failed", "exception", "panic", "fatal:"}
	for _, indicator := range errorIndicators {
		if strings.Contains(outputLower, indicator) {
			return true
		}
	}
	return false
}

// isLog checks if output is log output.
func (t *ToonFormatter) isLog(output string) bool {
	// Match common log patterns: timestamp + level
	logPattern := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}[T\s]\d{2}:\d{2}:\d{2}`)
	return logPattern.MatchString(output)
}

// formatJSON compacts JSON output by removing whitespace and using short keys.
func (t *ToonFormatter) formatJSON(output string) string {
	var data interface{}
	if err := json.Unmarshal([]byte(output), &data); err != nil {
		// Not valid JSON, return as-is
		return output
	}

	// Compact JSON (remove whitespace)
	compacted, err := json.Marshal(data)
	if err != nil {
		return output
	}

	return string(compacted)
}

// formatError compresses error messages using Toon symbols.
func (t *ToonFormatter) formatError(output string) string {
	lines := strings.Split(output, "\n")
	var compressed []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Replace common error prefixes with symbols
		line = strings.ReplaceAll(line, "Error:", "✗")
		line = strings.ReplaceAll(line, "error:", "✗")
		line = strings.ReplaceAll(line, "Warning:", "⚠")
		line = strings.ReplaceAll(line, "warning:", "⚠")
		line = strings.ReplaceAll(line, "Failed", "✗")
		line = strings.ReplaceAll(line, "failed", "✗")
		line = strings.ReplaceAll(line, "Success", "✓")
		line = strings.ReplaceAll(line, "success", "✓")

		compressed = append(compressed, line)
	}

	return strings.Join(compressed, " ")
}

// formatLog compresses log output using Toon format.
func (t *ToonFormatter) formatLog(output string) string {
	lines := strings.Split(output, "\n")
	var compressed []string

	// Log level symbols
	levelSymbols := map[string]string{
		"DEBUG":   "🔍",
		"INFO":    "ℹ",
		"WARN":    "⚠",
		"WARNING": "⚠",
		"ERROR":   "✗",
		"FATAL":   "☠",
	}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Extract and replace log level
		for level, symbol := range levelSymbols {
			if strings.Contains(line, level) {
				line = strings.Replace(line, level, symbol, 1)
				break
			}
		}

		// Remove timestamp (common patterns)
		tsPattern := regexp.MustCompile(`\d{4}-\d{2}-\d{2}[T\s]\d{2}:\d{2}:\d{2}(\.\d+)?(Z|[+-]\d{2}:?\d{2})?\s*`)
		line = tsPattern.ReplaceAllString(line, "")

		compressed = append(compressed, line)
	}

	return strings.Join(compressed, " | ")
}

// formatText compacts plain text output.
func (t *ToonFormatter) formatText(output string) string {
	// Remove excessive whitespace
	output = strings.TrimSpace(output)
	spacePattern := regexp.MustCompile(`\s+`)
	output = spacePattern.ReplaceAllString(output, " ")

	// Remove duplicate lines
	lines := strings.Split(output, "\n")
	seen := make(map[string]bool)
	var unique []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !seen[line] {
			seen[line] = true
			unique = append(unique, line)
		}
	}

	return strings.Join(unique, " ")
}

// FormatBatch formats multiple outputs in batch.
func (t *ToonFormatter) FormatBatch(outputs []string) []string {
	result := make([]string, len(outputs))
	for i, output := range outputs {
		result[i] = t.Format(output)
	}
	return result
}

// EstimateCompression estimates the compression ratio achieved by Toon formatting.
func (t *ToonFormatter) EstimateCompression(original, formatted string) float64 {
	if len(original) == 0 {
		return 0.0
	}

	compression := float64(len(original)-len(formatted)) / float64(len(original)) * 100
	if compression < 0 {
		compression = 0
	}

	return compression
}

// ToonSymbol represents a Toon format symbol mapping.
type ToonSymbol struct {
	Symbol   string
	Meaning  string
	Category string
}

// GetToonSymbols returns common Toon symbols and their meanings.
func GetToonSymbols() []ToonSymbol {
	return []ToonSymbol{
		// Status symbols
		{"✓", "Success/Complete", "status"},
		{"✗", "Error/Failed", "status"},
		{"⚠", "Warning", "status"},
		{"ℹ", "Info", "status"},
		{"☠", "Fatal", "status"},

		// Action symbols
		{"→", "Arrow/Result", "action"},
		{"←", "Return/Back", "action"},
		{"↓", "Download/Input", "action"},
		{"↑", "Upload/Output", "action"},
		{"⟳", "Refresh/Retry", "action"},

		// Object symbols
		{"📁", "File/Folder", "object"},
		{"📝", "Document/Text", "object"},
		{"🔧", "Config/Settings", "object"},
		{"🔒", "Security/Lock", "object"},
		{"💾", "Database/Storage", "object"},

		// Process symbols
		{"⏳", "Processing/Loading", "process"},
		{"⏱", "Timing/Duration", "process"},
		{"🔍", "Search/Debug", "process"},
		{"⚡", "Fast/Performance", "process"},
		{"🚀", "Launch/Deploy", "process"},
	}
}

// CompactWithSymbols replaces common words with Toon symbols.
func (t *ToonFormatter) CompactWithSymbols(text string) string {
	replacements := map[string]string{
		"success":     "✓",
		"failed":      "✗",
		"error":       "✗",
		"warning":     "⚠",
		"info":        "ℹ",
		"loading":     "⏳",
		"processing":  "⏳",
		"complete":    "✓",
		"downloading": "↓",
		"uploading":   "↑",
		"file":        "📁",
		"folder":      "📁",
		"config":      "🔧",
		"database":    "💾",
		"search":      "🔍",
	}

	textLower := strings.ToLower(text)
	for word, symbol := range replacements {
		textLower = strings.ReplaceAll(textLower, word, symbol)
	}

	return textLower
}

// Example demonstrates Toon format usage.
func Example() {
	cfg := &types.Config{}
	formatter := NewToonFormatter(cfg)

	// Example 1: JSON
	jsonInput := `{"status": "success", "message": "Task completed", "data": {"id": 123}}`
	fmt.Println("Original JSON:", jsonInput)
	fmt.Println("Toon format:", formatter.Format(jsonInput))

	// Example 2: Error
	errorInput := "Error: Failed to connect to database\nConnection timeout after 30s"
	fmt.Println("\nOriginal error:", errorInput)
	fmt.Println("Toon format:", formatter.Format(errorInput))

	// Example 3: Log
	logInput := "2024-01-15 10:30:45 INFO Server started on port 8080\n2024-01-15 10:30:46 INFO Database connected successfully"
	fmt.Println("\nOriginal log:", logInput)
	fmt.Println("Toon format:", formatter.Format(logInput))
}
