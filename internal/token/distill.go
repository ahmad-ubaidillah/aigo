// Package token provides token efficiency utilities for reducing token usage
// while maintaining information quality.
package token

import (
	"regexp"
	"strings"

	"github.com/ahmad-ubaidillah/aigo/pkg/types"
)

// ContentType represents the type of content being processed.
type ContentType string

const (
	ContentTypeError    ContentType = "error"
	ContentTypeSuccess  ContentType = "success"
	ContentTypeProgress ContentType = "progress"
	ContentTypeInfo     ContentType = "info"
)

// Distiller implements a distillation pipeline that classifies, scores,
// and composes compact versions of LLM outputs.
type Distiller struct {
	cfg *types.Config
}

// NewDistiller creates a new Distiller instance.
func NewDistiller(cfg *types.Config) *Distiller {
	return &Distiller{cfg: cfg}
}

// Process takes a raw LLM output and returns a distilled (compact) version.
// Pipeline: Classify → Score → Compose
func (d *Distiller) Process(output string) string {
	if output == "" {
		return ""
	}

	// Step 1: Classify the content type
	contentType := d.classify(output)

	// Step 2: Score the importance (0.0 to 1.0)
	score := d.score(output, contentType)

	// Step 3: Compose compact version based on score
	return d.compose(output, contentType, score)
}

// classify determines the content type of the output.
func (d *Distiller) classify(output string) string {
	outputLower := strings.ToLower(output)

	// Error patterns
	errorPatterns := []string{
		"error:", "failed", "exception", "panic:", "fatal:",
		"cannot", "unable to", "unexpected", "invalid",
		"permission denied", "not found", "timeout",
	}
	for _, pattern := range errorPatterns {
		if strings.Contains(outputLower, pattern) {
			return string(ContentTypeError)
		}
	}

	// Success patterns
	successPatterns := []string{
		"successfully", "completed", "done", "finished",
		"created", "updated", "deleted", "installed",
		"✓", "✅", "success",
	}
	for _, pattern := range successPatterns {
		if strings.Contains(outputLower, pattern) {
			return string(ContentTypeSuccess)
		}
	}

	// Progress patterns
	progressPatterns := []string{
		"processing", "running", "loading", "downloading",
		"building", "compiling", "executing", "starting",
		"⏳", "⏱️", "...",
	}
	for _, pattern := range progressPatterns {
		if strings.Contains(outputLower, pattern) {
			return string(ContentTypeProgress)
		}
	}

	// Default to info
	return string(ContentTypeInfo)
}

// score calculates the importance score of the output.
// Higher score = more important = keep more details.
func (d *Distiller) score(output string, contentType string) float64 {
	baseScore := 0.5

	// Adjust based on content type
	switch contentType {
	case string(ContentTypeError):
		baseScore = 0.9 // Errors are important - keep details
	case string(ContentTypeSuccess):
		baseScore = 0.3 // Success messages can be heavily compressed
	case string(ContentTypeProgress):
		baseScore = 0.2 // Progress messages can be very compact
	case string(ContentTypeInfo):
		baseScore = 0.5 // Info is medium importance
	}

	// Boost score if contains important keywords
	importantKeywords := []string{
		"critical", "important", "warning", "security",
		"breaking", "deprecated", "required", "config",
	}
	outputLower := strings.ToLower(output)
	for _, keyword := range importantKeywords {
		if strings.Contains(outputLower, keyword) {
			baseScore += 0.1
		}
	}

	// Reduce score if contains repetitive content
	if d.isRepetitive(output) {
		baseScore -= 0.2
	}

	// Cap between 0.0 and 1.0
	if baseScore > 1.0 {
		baseScore = 1.0
	}
	if baseScore < 0.0 {
		baseScore = 0.0
	}

	return baseScore
}

// compose creates a compact version of the output based on score.
func (d *Distiller) compose(output string, contentType string, score float64) string {
	// High score (0.8-1.0): Keep most details, just trim whitespace
	if score >= 0.8 {
		return strings.TrimSpace(output)
	}

	// Medium score (0.5-0.8): Extract key lines
	if score >= 0.5 {
		return d.extractKeyLines(output, 5)
	}

	// Low score (0.0-0.5): Maximum compression
	return d.extractKeyLines(output, 2)
}

// extractKeyLines extracts the most important lines from output.
func (d *Distiller) extractKeyLines(output string, maxLines int) string {
	lines := strings.Split(output, "\n")
	if len(lines) <= maxLines {
		return strings.TrimSpace(output)
	}

	var keyLines []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Skip common noise patterns
		if d.isNoiseLine(line) {
			continue
		}

		keyLines = append(keyLines, line)
		if len(keyLines) >= maxLines {
			break
		}
	}

	if len(keyLines) == 0 {
		// Fallback to first N non-empty lines
		keyLines = []string{}
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" {
				keyLines = append(keyLines, line)
				if len(keyLines) >= maxLines {
					break
				}
			}
		}
	}

	return strings.Join(keyLines, "\n")
}

// isNoiseLine checks if a line is noise that can be skipped.
func (d *Distiller) isNoiseLine(line string) bool {
	noisePatterns := []string{
		`^\s*$`,          // Empty lines
		`^\s*#\s*$`,      // Just a hash
		`^\s*---+\s*$`,   // Separator lines
		`^\s*\*\*\*\s*$`, // Separator lines
		`^\s*===+\s*$`,   // Separator lines
	}

	for _, pattern := range noisePatterns {
		matched, _ := regexp.MatchString(pattern, line)
		if matched {
			return true
		}
	}

	return false
}

// isRepetitive checks if output contains repetitive patterns.
func (d *Distiller) isRepetitive(output string) bool {
	lines := strings.Split(output, "\n")
	if len(lines) < 3 {
		return false
	}

	// Check for repeated identical lines
	seen := make(map[string]int)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			seen[line]++
			if seen[line] >= 3 {
				return true
			}
		}
	}

	return false
}

// DistillBatch processes multiple outputs in batch.
func (d *Distiller) DistillBatch(outputs []string) []string {
	result := make([]string, len(outputs))
	for i, output := range outputs {
		result[i] = d.Process(output)
	}
	return result
}

// EstimateTokenSavings estimates the percentage of tokens saved by distillation.
func (d *Distiller) EstimateTokenSavings(original, distilled string) float64 {
	if len(original) == 0 {
		return 0.0
	}

	// Rough estimation: 1 token ≈ 4 characters
	originalTokens := float64(len(original)) / 4.0
	distilledTokens := float64(len(distilled)) / 4.0

	if originalTokens == 0 {
		return 0.0
	}

	savings := (originalTokens - distilledTokens) / originalTokens * 100
	if savings < 0 {
		savings = 0
	}

	return savings
}
