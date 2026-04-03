package healing

import (
	"context"
	"fmt"
	"math"
	"time"
)

// ErrorType is the high-level category of an error.
type ErrorType int

const (
	ToolExecutionError ErrorType = iota + 1
	TimeoutError
	PermissionError
	ResourceError
	SyntaxError
	RuntimeError
	NetworkError
	RateLimitError
)

func (e ErrorType) String() string {
	names := map[ErrorType]string{
		ToolExecutionError: "ToolExecutionError",
		TimeoutError:       "TimeoutError",
		PermissionError:    "PermissionError",
		ResourceError:      "ResourceError",
		SyntaxError:        "SyntaxError",
		RuntimeError:       "RuntimeError",
		NetworkError:       "NetworkError",
		RateLimitError:     "RateLimitError",
	}
	if n, ok := names[e]; ok {
		return n
	}
	return "UnknownError"
}

// Severity indicates how critical an error is.
type Severity int

const (
	Critical Severity = iota
	Major
	Minor
)

// ErrorDetector detects the type and severity from a raw error.
type ErrorDetector struct{}

// Detect analyzes the error and returns a best-effort ErrorType and Severity.
func (d *ErrorDetector) Detect(err error) (ErrorType, Severity) {
	if err == nil {
		return ToolExecutionError, Minor
	}
	s := fmt.Sprint(err)
	lower := toLowerASCII(s)

	if contains(lower, "timeout") {
		return TimeoutError, Critical
	}
	if contains(lower, "permission") {
		return PermissionError, Major
	}
	if contains(lower, "resource") {
		return ResourceError, Major
	}
	if contains(lower, "syntax") {
		return SyntaxError, Minor
	}
	if contains(lower, "runtime") {
		return RuntimeError, Major
	}
	if contains(lower, "network") {
		return NetworkError, Critical
	}
	if contains(lower, "rate limit") || contains(lower, "ratelimit") {
		return RateLimitError, Major
	}

	// Fallback conservative default
	return ToolExecutionError, Minor
}

// Helper: naive substring search (case-insensitive using ASCII).
func contains(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Helper: ASCII lowercase conversion (safe for typical error messages).
func toLowerASCII(s string) string {
	b := []byte(s)
	for i := range b {
		if b[i] >= 'A' && b[i] <= 'Z' {
			b[i] = b[i] + 32
		}
	}
	return string(b)
}

// Quick usage to ensure imports are used in this file during static checks.
var _ = context.Background()
var _ = time.Second
var _ = math.Max(0, 0)
