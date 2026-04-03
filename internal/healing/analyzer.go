package healing

import (
	"context"
	"fmt"
	"math"
	"time"
)

// ErrorAnalysis holds structured information about an error.
type ErrorAnalysis struct {
	Type         ErrorType
	Location     string
	Context      string
	RootCause    string
	SuggestedFix string
}

// ErrorAnalyzer analyzes an error and produces a structured analysis.
type ErrorAnalyzer struct {
	Detector *ErrorDetector
}

// Analyze inspects the error and returns a detailed analysis.
func (a *ErrorAnalyzer) Analyze(err error) (*ErrorAnalysis, error) {
	if err == nil {
		return &ErrorAnalysis{Type: ToolExecutionError, RootCause: ""}, nil
	}
	et, _ := a.Detector.Detect(err)
	analysis := &ErrorAnalysis{
		Type:         et,
		RootCause:    err.Error(),
		Location:     extractLocationFromError(err),
		Context:      fmt.Sprintf("Detected error type %v", et),
		SuggestedFix: suggestedFixForType(et),
	}
	if analysis.Location == "" {
		analysis.Location = "unknown"
	}
	return analysis, nil
}

// extractLocationFromError attempts to pull location info if implemented.
func extractLocationFromError(err error) string {
	type l interface{ Location() string }
	if x, ok := err.(l); ok {
		if loc := x.Location(); loc != "" {
			return loc
		}
	}
	return ""
}

func suggestedFixForType(t ErrorType) string {
	switch t {
	case TimeoutError:
		return "Increase timeout and retry with backoff."
	case PermissionError:
		return "Request elevated permissions or adjust access rights."
	case ResourceError:
		return "Allocate required resources before retry."
	case SyntaxError:
		return "Fix syntax and run lint/tests."
	case RuntimeError:
		return "Analyze runtime environment and restart if needed."
	case NetworkError:
		return "Check network, retry after short delay."
	case RateLimitError:
		return "Respect rate limits and back off before retrying."
	default:
		return "Retry operation with standard retry policy."
	}
}

// Usage helpers to keep imports referenced in a meaningful way.
var _ = context.Background()
var _ = time.Now()
var _ = math.Max(0, 0)
