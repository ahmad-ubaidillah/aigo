package healing

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ahmad-ubaidillah/aigo/internal/llm"
)

// RootCause represents the identified root cause of an error.
type RootCause struct {
	Type        string
	Description string
	File        string
	Line        int
	Suggestion  string
	Confidence  float64
}

// RootCauseAnalyzer performs root cause analysis on errors.
type RootCauseAnalyzer struct {
	client llm.LLMClient
	model  string
}

// NewRootCauseAnalyzer creates a new root cause analyzer.
func NewRootCauseAnalyzer(client llm.LLMClient, model string) *RootCauseAnalyzer {
	if model == "" {
		model = "gpt-4o-mini"
	}
	return &RootCauseAnalyzer{client: client, model: model}
}

// AnalyzeRootCause analyzes a traceback to identify the root cause.
func (a *RootCauseAnalyzer) AnalyzeRootCause(traceback string) (*RootCause, error) {
	if a.client == nil {
		return a.analyzeHeuristic(traceback), nil
	}

	prompt := fmt.Sprintf(`Analyze this error and identify the root cause.
Return: type, description, file, line, suggestion, confidence (0.0-1.0).

Error:
%s`, traceback)

	messages := []llm.Message{
		{Role: "system", Content: "You are an expert debugger. Identify the exact root cause of errors."},
		{Role: "user", Content: prompt},
	}

	resp, err := a.client.Chat(context.Background(), messages)
	if err != nil {
		return a.analyzeHeuristic(traceback), nil
	}

	return a.parseLLMResponse(resp.Content), nil
}

// IdentifyCauseType identifies the type of error from a traceback.
func (a *RootCauseAnalyzer) IdentifyCauseType(traceback string) string {
	lower := strings.ToLower(traceback)
	switch {
	case strings.Contains(lower, "import") && strings.Contains(lower, "not found"):
		return "MissingImport"
	case strings.Contains(lower, "type") && strings.Contains(lower, "mismatch"):
		return "TypeMismatch"
	case strings.Contains(lower, "null") || strings.Contains(lower, "nil") || strings.Contains(lower, "undefined"):
		return "NullPointer"
	case strings.Contains(lower, "index") && strings.Contains(lower, "out of range"):
		return "IndexOutOfRange"
	case strings.Contains(lower, "permission") || strings.Contains(lower, "denied"):
		return "PermissionDenied"
	case strings.Contains(lower, "timeout"):
		return "Timeout"
	case strings.Contains(lower, "connection") || strings.Contains(lower, "refused"):
		return "ConnectionFailed"
	case strings.Contains(lower, "syntax"):
		return "SyntaxError"
	default:
		return "Unknown"
	}
}

// SuggestFix returns a suggested fix for the error type.
func (a *RootCauseAnalyzer) SuggestFix(errorType string, traceback string) string {
	switch errorType {
	case "MissingImport":
		return "Add the missing import statement at the top of the file. Check the module path is correct."
	case "TypeMismatch":
		return "Ensure variable types match the expected type. Use type assertions or conversions as needed."
	case "NullPointer":
		return "Add nil checks before accessing object properties. Initialize variables before use."
	case "IndexOutOfRange":
		return "Check array bounds before accessing elements. Use len() to verify size."
	case "PermissionDenied":
		return "Check file permissions or API credentials. Ensure the process has required access."
	case "Timeout":
		return "Increase timeout value or optimize the operation. Check for network issues."
	case "ConnectionFailed":
		return "Verify the service is running and accessible. Check network configuration."
	case "SyntaxError":
		return "Fix the syntax error at the indicated line. Check for missing brackets, quotes, or semicolons."
	default:
		return "Review the error message and traceback. Check the indicated file and line number."
	}
}

func (a *RootCauseAnalyzer) analyzeHeuristic(traceback string) *RootCause {
	errorType := a.IdentifyCauseType(traceback)
	suggestion := a.SuggestFix(errorType, traceback)

	confidence := 0.5
	if errorType != "Unknown" {
		confidence = 0.7
	}

	file := ""
	line := 0
	if idx := strings.Index(traceback, ".go:"); idx > 0 {
		rest := traceback[idx+4:]
		if endIdx := strings.IndexAny(rest, ": \n"); endIdx > 0 {
			fmt.Sscanf(rest[:endIdx], "%d", &line)
			fileStart := strings.LastIndex(traceback[:idx], " ")
			if fileStart < 0 {
				fileStart = 0
			}
			file = traceback[fileStart:idx]
		}
	}

	return &RootCause{
		Type:        errorType,
		Description: traceback,
		File:        file,
		Line:        line,
		Suggestion:  suggestion,
		Confidence:  confidence,
	}
}

func (a *RootCauseAnalyzer) parseLLMResponse(content string) *RootCause {
	lines := strings.Split(content, "\n")
	result := &RootCause{Confidence: 0.8}
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "type:") {
			result.Type = strings.TrimSpace(strings.TrimPrefix(line, "type:"))
		} else if strings.HasPrefix(line, "description:") {
			result.Description = strings.TrimSpace(strings.TrimPrefix(line, "description:"))
		} else if strings.HasPrefix(line, "file:") {
			result.File = strings.TrimSpace(strings.TrimPrefix(line, "file:"))
		} else if strings.HasPrefix(line, "line:") {
			fmt.Sscanf(strings.TrimSpace(strings.TrimPrefix(line, "line:")), "%d", &result.Line)
		} else if strings.HasPrefix(line, "suggestion:") {
			result.Suggestion = strings.TrimSpace(strings.TrimPrefix(line, "suggestion:"))
		}
	}
	if result.Type == "" {
		result.Type = a.IdentifyCauseType(content)
	}
	if result.Suggestion == "" {
		result.Suggestion = a.SuggestFix(result.Type, content)
	}
	return result
}

// HealingAttempt tracks a single healing attempt.
type HealingAttempt struct {
	Timestamp time.Time
	ErrorType string
	RootCause string
	Action    string
	Success   bool
	Duration  time.Duration
}

// HealingStats tracks healing statistics.
type HealingStats struct {
	Attempts      int
	Successes     int
	Failures      int
	ByType        map[string]int
	BySuccess     map[string]int
	LastAttempt   time.Time
	totalDuration time.Duration
}

// NewHealingStats creates a new healing stats tracker.
func NewHealingStats() *HealingStats {
	return &HealingStats{
		ByType:    make(map[string]int),
		BySuccess: make(map[string]int),
	}
}

// RecordAttempt records a healing attempt.
func (s *HealingStats) RecordAttempt(attempt HealingAttempt) {
	s.Attempts++
	s.ByType[attempt.ErrorType]++
	s.totalDuration += attempt.Duration
	if attempt.Success {
		s.Successes++
		s.BySuccess[attempt.ErrorType]++
	} else {
		s.Failures++
	}
	s.LastAttempt = attempt.Timestamp
}

// SuccessRate returns the overall success rate.
func (s *HealingStats) SuccessRate() float64 {
	if s.Attempts == 0 {
		return 0
	}
	return float64(s.Successes) / float64(s.Attempts) * 100
}

// SuccessRateByType returns the success rate for a specific error type.
func (s *HealingStats) SuccessRateByType(errorType string) float64 {
	total := s.ByType[errorType]
	if total == 0 {
		return 0
	}
	success := s.BySuccess[errorType]
	return float64(success) / float64(total) * 100
}

// AvgDuration returns the average healing attempt duration.
func (s *HealingStats) AvgDuration() time.Duration {
	if s.Attempts == 0 {
		return 0
	}
	return s.totalDuration / time.Duration(s.Attempts)
}

// Summary returns a human-readable summary.
func (s *HealingStats) Summary() string {
	return fmt.Sprintf("Healing: %d attempts, %d successes (%.1f%%), avg %.0fms",
		s.Attempts, s.Successes, s.SuccessRate(), s.AvgDuration().Seconds()*1000)
}
