package healing

import (
	"context"
	"fmt"
	"math"
	"time"
)

// RetryStrategy defines how delays are calculated between retries.
type RetryStrategy int

const (
	Immediate RetryStrategy = iota
	ExponentialBackoff
	LinearBackoff
	FixedDelay
)

// maxRetries for each error type governs how many times we are willing to retry.
func maxRetriesFor(t ErrorType) int {
	switch t {
	case NetworkError:
		return 5
	case RateLimitError:
		return 3
	case ToolExecutionError:
		return 3
	case SyntaxError:
		return 2
	case PermissionError:
		return 1
	default:
		return 1
	}
}

// RetryManager encapsulates a retry policy.
type RetryManager struct {
	Strategy   RetryStrategy
	BaseDelay  time.Duration
	MaxBackoff time.Duration
}

// ShouldRetry decides if another attempt should be made for the given type/attempt.
func (r *RetryManager) ShouldRetry(errorType ErrorType, attempt int) bool {
	max := maxRetriesFor(errorType)
	if max <= 0 {
		return false
	}
	return attempt <= max
}

// GetDelay returns the delay before the next retry for the given type/attempt.
func (r *RetryManager) GetDelay(errorType ErrorType, attempt int) time.Duration {
	base := r.BaseDelay
	if base <= 0 {
		base = 100 * time.Millisecond
	}
	var d time.Duration
	switch r.Strategy {
	case Immediate:
		d = 0
	case ExponentialBackoff:
		// 2^(attempt-1) * base
		mult := int64(1) << uint(attempt-1)
		d = time.Duration(mult) * base
	case LinearBackoff:
		d = time.Duration(attempt) * base
	case FixedDelay:
		d = base
	default:
		d = base
	}

	// Apply max backoff cap if configured.
	if r.MaxBackoff > 0 {
		d = time.Duration(math.Min(float64(d), float64(r.MaxBackoff)))
	}
	// Small normalization to avoid negative durations.
	if d < 0 {
		d = 0
	}
	return d
}

// Simple convenience for common default policy.
func NewDefaultRetryManager() *RetryManager {
	return &RetryManager{Strategy: ExponentialBackoff, BaseDelay: 100 * time.Millisecond, MaxBackoff: 5 * time.Second}
}

// Usage helper to ensure imports are utilized in static checks.
var _ = context.Background()
var _ = fmt.Sprintf("%d", 1)
var _ = time.Second
