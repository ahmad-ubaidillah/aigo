// Package distill implements OMNI-style token optimization through content classification and scoring.
package distill

import (
	"encoding/json"
	"strings"
)

// ContentType represents the type of content being analyzed.
type ContentType int

const (
	ContentUnknown ContentType = iota
	ContentGitDiff
	ContentBuildOutput
	ContentTestOutput
	ContentInfraOutput
	ContentLogOutput
	ContentTabularData
	ContentStructuredData
)

// Classifier detects the type of content in a text string.
type Classifier struct{}

// NewClassifier creates a new Classifier instance.
func NewClassifier() *Classifier {
	return &Classifier{}
}

// Classify analyzes text and returns its ContentType.
func (c *Classifier) Classify(text string) ContentType {
	text = strings.TrimSpace(text)
	if text == "" {
		return ContentUnknown
	}

	if c.isGitDiff(text) {
		return ContentGitDiff
	}
	if c.isStructuredData(text) {
		return ContentStructuredData
	}
	if c.isTestOutput(text) {
		return ContentTestOutput
	}
	if c.isBuildOutput(text) {
		return ContentBuildOutput
	}
	if c.isInfraOutput(text) {
		return ContentInfraOutput
	}
	if c.isLogOutput(text) {
		return ContentLogOutput
	}
	if c.isTabularData(text) {
		return ContentTabularData
	}

	return ContentUnknown
}

// isGitDiff detects git diff output.
func (c *Classifier) isGitDiff(text string) bool {
	return strings.Contains(text, "diff --git") ||
		(strings.Contains(text, "+++") && strings.Contains(text, "---"))
}

// isBuildOutput detects build/compilation output.
func (c *Classifier) isBuildOutput(text string) bool {
	buildPatterns := []string{
		"go build",
		"error:",
		"warning:",
		"cannot find package",
		"undefined:",
		"compilation",
		"build failed",
	}
	textLower := strings.ToLower(text)
	for _, pattern := range buildPatterns {
		if strings.Contains(textLower, strings.ToLower(pattern)) {
			return true
		}
	}
	return false
}

// isTestOutput detects test runner output.
func (c *Classifier) isTestOutput(text string) bool {
	testPatterns := []string{
		"=== RUN",
		"--- PASS",
		"--- FAIL",
		"FAIL\t",
		"ok\t",
		"PASS",
		"test suite",
		"testing",
	}
	for _, pattern := range testPatterns {
		if strings.Contains(text, pattern) {
			return true
		}
	}
	return false
}

// isInfraOutput detects infrastructure tool output.
func (c *Classifier) isInfraOutput(text string) bool {
	infraTools := []string{"kubectl", "terraform", "docker", "aws ", "gcloud ", "az "}
	actionVerbs := []string{"apply", "create", "delete", "get", "list", "deploy", "run", "build", "push", "pull"}

	textLower := strings.ToLower(text)
	hasTool := false
	for _, tool := range infraTools {
		if strings.Contains(textLower, strings.ToLower(tool)) {
			hasTool = true
			break
		}
	}
	if !hasTool {
		return false
	}

	for _, verb := range actionVerbs {
		if strings.Contains(textLower, verb) {
			return true
		}
	}
	return false
}

// isLogOutput detects log output with timestamps or log levels.
func (c *Classifier) isLogOutput(text string) bool {
	lines := strings.Split(text, "\n")
	if len(lines) < 2 {
		return false
	}

	logLevelPatterns := []string{"[INFO]", "[ERROR]", "[WARN]", "[DEBUG]", "[FATAL]", "[TRACE]", "INFO:", "ERROR:", "WARN:"}
	timestampPatterns := []string{"2024-", "2025-", "2026-", "T", "::"}

	matchedLines := 0
	for i, line := range lines {
		if i >= 10 {
			break
		}
		hasLogLevel := false
		for _, pattern := range logLevelPatterns {
			if strings.Contains(line, pattern) {
				hasLogLevel = true
				break
			}
		}
		hasTimestamp := false
		for _, pattern := range timestampPatterns {
			if strings.Contains(line, pattern) {
				hasTimestamp = true
				break
			}
		}
		if hasLogLevel || hasTimestamp {
			matchedLines++
		}
	}

	return matchedLines >= 2
}

// isTabularData detects tabular data with consistent separators.
func (c *Classifier) isTabularData(text string) bool {
	lines := strings.Split(text, "\n")
	if len(lines) < 3 {
		return false
	}

	pipeCount := 0
	tabCount := 0

	for i, line := range lines {
		if i >= 10 {
			break
		}
		pipeLines := strings.Count(line, "|")
		tabLines := strings.Count(line, "\t")

		if pipeLines >= 2 {
			pipeCount++
		}
		if tabLines >= 2 {
			tabCount++
		}
	}

	return pipeCount >= 2 || tabCount >= 2
}

// isStructuredData detects JSON structured data.
func (c *Classifier) isStructuredData(text string) bool {
	trimmed := strings.TrimSpace(text)
	if !strings.HasPrefix(trimmed, "{") && !strings.HasPrefix(trimmed, "[") {
		return false
	}
	return json.Valid([]byte(trimmed))
}
