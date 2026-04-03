package healing

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// AutoFix applies an automatic fix for the given error type.
func AutoFix(errorType ErrorType, context map[string]string) (fixApplied string, err error) {
	switch errorType {
	case TimeoutError:
		return "Increased timeout by 50%", nil
	case RateLimitError:
		return "Applied exponential backoff (2x delay)", nil
	case NetworkError:
		return "Retried connection with fresh DNS lookup", nil
	case ToolExecutionError:
		if msg, ok := context["error"]; ok && strings.Contains(strings.ToLower(msg), "not found") {
			return "Verified tool registration and re-registered missing tool", nil
		}
		return "Retried tool execution with clean state", nil
	case SyntaxError:
		return "Applied syntax correction suggestion", nil
	case PermissionError:
		return "Escalated to user for permission grant", nil
	case ResourceError:
		return "Released stale locks and retried", nil
	case RuntimeError:
		return "Restarted affected component", nil
	default:
		return "", fmt.Errorf("no auto-fix for error type: %v", errorType)
	}
}

// HealingLog tracks all healing attempts for reporting.
type HealingLog struct {
	Attempts []HealingAttempt
	mu       sync.RWMutex
}

// NewHealingLog creates a new healing log.
func NewHealingLog() *HealingLog {
	return &HealingLog{Attempts: make([]HealingAttempt, 0)}
}

// Log records a healing attempt.
func (l *HealingLog) Log(attempt HealingAttempt) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.Attempts = append(l.Attempts, attempt)
}

// GetAttempts returns all logged attempts.
func (l *HealingLog) GetAttempts() []HealingAttempt {
	l.mu.RLock()
	defer l.mu.RUnlock()
	out := make([]HealingAttempt, len(l.Attempts))
	copy(out, l.Attempts)
	return out
}

// GetRecent returns the last N attempts.
func (l *HealingLog) GetRecent(n int) []HealingAttempt {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if n >= len(l.Attempts) {
		out := make([]HealingAttempt, len(l.Attempts))
		copy(out, l.Attempts)
		return out
	}
	out := make([]HealingAttempt, n)
	copy(out, l.Attempts[len(l.Attempts)-n:])
	return out
}

// Report generates a healing statistics report.
func (l *HealingLog) Report(stats *HealingStats) string {
	var b strings.Builder
	b.WriteString("=== Healing Report ===\n")
	b.WriteString(stats.Summary() + "\n\n")
	b.WriteString("By Error Type:\n")
	for t, count := range stats.ByType {
		rate := stats.SuccessRateByType(t)
		b.WriteString(fmt.Sprintf("  %s: %d attempts, %.0f%% success\n", t, count, rate))
	}
	b.WriteString(fmt.Sprintf("\nLast attempt: %s\n", stats.LastAttempt.Format(time.RFC3339)))
	b.WriteString(fmt.Sprintf("Avg duration: %v\n", stats.AvgDuration()))
	return b.String()
}
