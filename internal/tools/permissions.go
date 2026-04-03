// Package tools provides permission checking for tool execution.
package tools

import (
	"fmt"
	"path"
	"sync"
)

// PermissionLevel represents the access level for a tool.
type PermissionLevel int

const (
	// PermAllow grants automatic execution permission.
	PermAllow PermissionLevel = iota
	// PermAsk requires user confirmation before execution.
	PermAsk
	// PermDeny blocks tool execution.
	PermDeny
)

// String returns the string representation of the permission level.
func (p PermissionLevel) String() string {
	switch p {
	case PermAllow:
		return "allow"
	case PermAsk:
		return "ask"
	case PermDeny:
		return "deny"
	default:
		return "unknown"
	}
}

// PermissionRule defines a permission rule for tool matching.
type PermissionRule struct {
	Tool    string          // tool name or pattern (e.g., "bash", "file_*", "*")
	Pattern string          // additional pattern for parameters (reserved for future use)
	Level   PermissionLevel // permission level for matching tools
}

// PermissionChecker manages permission rules and checks tool access.
type PermissionChecker struct {
	mu    sync.RWMutex
	rules []PermissionRule
}

// NewPermissionChecker creates a new permission checker with no rules.
func NewPermissionChecker() *PermissionChecker {
	return &PermissionChecker{
		rules: make([]PermissionRule, 0),
	}
}

// AddRule appends a new permission rule to the checker.
func (pc *PermissionChecker) AddRule(rule PermissionRule) {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	pc.rules = append(pc.rules, rule)
}

// Check determines the permission level for a given tool name.
// Rules are evaluated in order; first match wins.
// Returns PermAllow if no rules match (default allow).
func (pc *PermissionChecker) Check(toolName string) PermissionLevel {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	for _, rule := range pc.rules {
		if matches, err := path.Match(rule.Tool, toolName); err == nil && matches {
			return rule.Level
		}
	}

	return PermAllow
}

// Clear removes all permission rules.
func (pc *PermissionChecker) Clear() {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	pc.rules = make([]PermissionRule, 0)
}

// Rules returns a copy of all current permission rules.
func (pc *PermissionChecker) Rules() []PermissionRule {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	result := make([]PermissionRule, len(pc.rules))
	copy(result, pc.rules)
	return result
}

// SetRules replaces all permission rules with the provided slice.
func (pc *PermissionChecker) SetRules(rules []PermissionRule) {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	pc.rules = rules
}

// ParsePermissionLevel converts a string to PermissionLevel.
func ParsePermissionLevel(s string) (PermissionLevel, error) {
	switch s {
	case "allow":
		return PermAllow, nil
	case "ask":
		return PermAsk, nil
	case "deny":
		return PermDeny, nil
	default:
		return PermAllow, fmt.Errorf("unknown permission level: %q", s)
	}
}
