package healing

import (
	"context"
	"fmt"
)

// RecoveryAction describes a recommended action to recover from a given error type.
type RecoveryAction struct {
	ActionType  string
	Description string
}

// GetRecoveryActions returns a set of actions to recover from the given error type.
func GetRecoveryActions(errorType ErrorType) []RecoveryAction {
	switch errorType {
	case ToolExecutionError:
		return []RecoveryAction{{ActionType: "RetryTool", Description: "Retry the tool operation after verification"}, {ActionType: "Log", Description: "Log the failure and continue"}}
	case TimeoutError:
		return []RecoveryAction{{ActionType: "IncreaseTimeout", Description: "Extend timeout window"}, {ActionType: "Retry", Description: "Retry after delay"}}
	case PermissionError:
		return []RecoveryAction{{ActionType: "Escalate", Description: "Request elevated permissions"}}
	case ResourceError:
		return []RecoveryAction{{ActionType: "AllocateResource", Description: "Acquire missing resources"}}
	case SyntaxError:
		return []RecoveryAction{{ActionType: "ValidateSyntax", Description: "Run linters/formatter"}}
	case RuntimeError:
		return []RecoveryAction{{ActionType: "Restart", Description: "Restart affected service"}}
	case NetworkError:
		return []RecoveryAction{{ActionType: "CheckNetwork", Description: "Verify connectivity"}}
	case RateLimitError:
		return []RecoveryAction{{ActionType: "Backoff", Description: "Back off and retry respecting limits"}}
	default:
		return []RecoveryAction{{ActionType: "Retry", Description: "Generic retry"}}
	}
}

// Ensure imports are used in a minimal way for static checks during development.
var _ = context.Background()
var _ = fmt.Sprintf("%d", 1)
