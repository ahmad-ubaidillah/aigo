package healing

import (
	"context"
	"fmt"
	"time"
)

// HealingLoop orchestrates detection, analysis, and retry decisions.
type HealingLoop struct {
	Detector     *ErrorDetector
	Analyzer     *ErrorAnalyzer
	RetryManager *RetryManager
	Stats        *HealingStats
	Log          *HealingLog
}

// NewHealingLoop creates a fully configured healing loop.
func NewHealingLoop() *HealingLoop {
	return &HealingLoop{
		Detector:     &ErrorDetector{},
		Analyzer:     &ErrorAnalyzer{Detector: &ErrorDetector{}},
		RetryManager: NewDefaultRetryManager(),
		Stats:        NewHealingStats(),
		Log:          NewHealingLog(),
	}
}

// Execute runs the provided function with automatic self-healing.
// It returns nil on success, or a non-nil error if unrecoverable.
func (h *HealingLoop) Execute(ctx context.Context, fn func() error) error {
	var consecutive int

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		start := time.Now()
		err := fn()
		duration := time.Since(start)

		if err != nil {
			consecutive++
			if h.Detector == nil || h.Analyzer == nil || h.RetryManager == nil {
				return fmt.Errorf("healing loop misconfigured: missing components: %w", err)
			}
			et, _ := h.Detector.Detect(err)
			analysis, _ := h.Analyzer.Analyze(err)

			h.Stats.RecordAttempt(HealingAttempt{
				Timestamp: time.Now(),
				ErrorType: et.String(),
				RootCause: analysis.RootCause,
				Action:    "retry",
				Success:   false,
				Duration:  duration,
			})
			h.Log.Log(HealingAttempt{
				Timestamp: time.Now(),
				ErrorType: et.String(),
				RootCause: analysis.RootCause,
				Action:    "retry",
				Success:   false,
				Duration:  duration,
			})

			if consecutive >= 3 {
				actions := GetRecoveryActions(et)
				return fmt.Errorf("escalated after %d errors: type=%s, actions=%v, root=%s",
					consecutive, et.String(), actions, analysis.RootCause)
			}
			if h.RetryManager.ShouldRetry(et, consecutive) {
				delay := h.RetryManager.GetDelay(et, consecutive)
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(delay):
					continue
				}
			}
			actions := GetRecoveryActions(et)
			return fmt.Errorf("not retrying after %d attempts: type=%s, actions=%v, root=%s",
				consecutive, et.String(), actions, analysis.RootCause)
		}

		if consecutive > 0 {
			h.Stats.RecordAttempt(HealingAttempt{
				Timestamp: time.Now(),
				ErrorType: "recovered",
				Action:    "success",
				Success:   true,
				Duration:  duration,
			})
			consecutive = 0
		}
		return nil
	}
}

// GetReport returns the current healing statistics report.
func (h *HealingLoop) GetReport() string {
	if h.Stats == nil {
		return "No stats available"
	}
	if h.Log == nil {
		return h.Stats.Summary()
	}
	return h.Log.Report(h.Stats)
}
