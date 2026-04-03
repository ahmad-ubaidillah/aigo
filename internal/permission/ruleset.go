package permission

import (
	"fmt"
	"path"
	"sync"
)

const (
	PermAllow PermissionLevel = iota
	PermAsk
	PermDeny
)

type PermissionLevel int

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

type PermissionRule struct {
	Tool    string
	Pattern string
	Level   PermissionLevel
}

type Ruleset struct {
	rules []PermissionRule
	mu    sync.RWMutex
}

func NewRuleset() *Ruleset {
	return &Ruleset{rules: make([]PermissionRule, 0)}
}

func (r *Ruleset) AddRule(tool, pattern string, level PermissionLevel) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.rules = append(r.rules, PermissionRule{
		Tool:    tool,
		Pattern: pattern,
		Level:   level,
	})
}

func (r *Ruleset) Check(toolName string) PermissionLevel {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, rule := range r.rules {
		if matches, err := path.Match(rule.Tool, toolName); err == nil && matches {
			return rule.Level
		}
	}
	return PermAllow
}

func (r *Ruleset) Allow(tool string) { r.AddRule(tool, "*", PermAllow) }
func (r *Ruleset) Deny(tool string)  { r.AddRule(tool, "*", PermDeny) }
func (r *Ruleset) Ask(tool string)   { r.AddRule(tool, "*", PermAsk) }
func (r *Ruleset) Clear()            { r.mu.Lock(); defer r.mu.Unlock(); r.rules = make([]PermissionRule, 0) }
func (r *Ruleset) Rules() []PermissionRule {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]PermissionRule, len(r.rules))
	copy(out, r.rules)
	return out
}
func (r *Ruleset) ParseLevel(s string) (PermissionLevel, error) {
	switch s {
	case "allow":
		return PermAllow, nil
	case "ask":
		return PermAsk, nil
	case "deny":
		return PermDeny, nil
	default:
		return PermAllow, fmt.Errorf("unknown level: %s", s)
	}
}
