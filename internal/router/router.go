// Package router implements semantic routing for model selection.
//
// Inspired by OMO's category-based routing:
//   - visual-engineering → frontend, UI/UX, design
//   - deep → autonomous research + execution
//   - quick → single-file changes, typos
//   - ultrabrain → hard logic, architecture decisions
//
// The router analyzes query complexity and category, then selects
// the best available model from the configured provider.
package router

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

// Route defines a routing rule: category → model.
type Route struct {
	Category  string `json:"category" yaml:"category"`
	Model     string `json:"model" yaml:"model"`
	Provider  string `json:"provider,omitempty" yaml:"provider,omitempty"`
	MaxTokens int    `json:"max_tokens,omitempty" yaml:"max_tokens,omitempty"`
	Desc      string `json:"description,omitempty" yaml:"description,omitempty"`
}

// Config is the router configuration.
type Config struct {
	Enabled       bool    `json:"enabled" yaml:"enabled"`
	DefaultModel  string  `json:"default_model" yaml:"default_model"`
	CheapModel    string  `json:"cheap_model,omitempty" yaml:"cheap_model,omitempty"`
	Routes        []Route `json:"routes" yaml:"routes"`
	AutoClassify  bool    `json:"auto_classify" yaml:"auto_classify"`
}

// Router manages semantic model routing.
type Router struct {
	mu     sync.RWMutex
	config Config
	stats  RouteStats
}

// RouteStats tracks routing decisions.
type RouteStats struct {
	mu     sync.RWMutex
	Counts map[string]int `json:"counts"`
	Total  int            `json:"total"`
}

// RouteResult is the output of a routing decision.
type RouteResult struct {
	Category   string `json:"category"`
	Model      string `json:"model"`
	Provider   string `json:"provider,omitempty"`
	Confidence float64 `json:"confidence"`
	Reason     string  `json:"reason"`
}

// New creates a new router with the given config.
func New(cfg Config) *Router {
	if !cfg.Enabled {
		cfg.Enabled = true
	}
	if cfg.DefaultModel == "" {
		cfg.DefaultModel = "auto"
	}
	if len(cfg.Routes) == 0 {
		cfg.Routes = defaultRoutes()
	}

	return &Router{
		config: cfg,
		stats: RouteStats{
			Counts: make(map[string]int),
		},
	}
}

// Route selects the best model for a given category.
func (r *Router) Route(category string) RouteResult {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Normalize category
	category = strings.ToLower(strings.TrimSpace(category))

	// Find matching route
	for _, route := range r.config.Routes {
		if strings.ToLower(route.Category) == category {
			result := RouteResult{
				Category:   category,
				Model:      route.Model,
				Provider:   route.Provider,
				Confidence: 0.95,
				Reason:     fmt.Sprintf("Matched route: %s → %s", route.Category, route.Model),
			}
			r.recordStat(category)
			return result
		}
	}

	// Default route
	result := RouteResult{
		Category:   category,
		Model:      r.config.DefaultModel,
		Confidence: 0.5,
		Reason:     "No matching route, using default model",
	}
	r.recordStat("default")
	return result
}

// RouteForQuery analyzes a query and routes to the best model.
// This is the automatic classification path (like OMO's category routing).
func (r *Router) RouteForQuery(ctx context.Context, query string, classifyFunc func(string) (string, float64)) RouteResult {
	if !r.config.AutoClassify || classifyFunc == nil {
		return r.Route("general")
	}

	category, confidence := classifyFunc(query)
	result := r.Route(category)
	result.Confidence = confidence
	result.Reason = fmt.Sprintf("Auto-classified as '%s' (confidence: %.2f)", category, confidence)
	return result
}

// CheapModel returns the configured cheap/fast model for simple tasks.
func (r *Router) CheapModel() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.config.CheapModel != "" {
		return r.config.CheapModel
	}
	return r.config.DefaultModel
}

// GetConfig returns the current router configuration.
func (r *Router) GetConfig() Config {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.config
}

// UpdateConfig updates router configuration at runtime.
func (r *Router) UpdateConfig(cfg Config) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.config = cfg
}

// GetStats returns routing statistics.
func (r *Router) GetStats() map[string]int {
	r.stats.mu.RLock()
	defer r.stats.mu.RUnlock()
	cp := make(map[string]int)
	for k, v := range r.stats.Counts {
		cp[k] = v
	}
	return cp
}

func (r *Router) recordStat(category string) {
	r.stats.mu.Lock()
	defer r.stats.mu.Unlock()
	r.stats.Counts[category]++
	r.stats.Total++
}

// defaultRoutes returns sensible default routing rules.
func defaultRoutes() []Route {
	return []Route{
		{
			Category:  "deep",
			Model:     "auto",
			MaxTokens: 4096,
			Desc:      "Autonomous deep work — research, implementation, complex tasks",
		},
		{
			Category:  "quick",
			Model:     "auto",
			MaxTokens: 1024,
			Desc:      "Quick fixes — typos, single-file changes, simple queries",
		},
		{
			Category:  "ultrabrain",
			Model:     "auto",
			MaxTokens: 8192,
			Desc:      "Hard logic — architecture, complex reasoning, system design",
		},
		{
			Category:  "visual",
			Model:     "auto",
			MaxTokens: 4096,
			Desc:      "Visual engineering — frontend, UI/UX, design",
		},
		{
			Category:  "general",
			Model:     "auto",
			MaxTokens: 2048,
			Desc:      "General tasks — default routing",
		},
	}
}

// CategoryKeywords maps keywords to categories for simple classification.
var CategoryKeywords = map[string][]string{
	"deep": {
		"implement", "build", "create", "write", "code", "develop",
		"refactor", "migrate", "integrate", "deploy",
		"buat", "implementasi", "kembangkan",
	},
	"quick": {
		"fix", "typo", "error", "bug", "rename", "update",
		"change", "remove", "delete", "small",
		"perbaiki", "ganti", "hapus",
	},
	"ultrabrain": {
		"architect", "design", "decide", "evaluate", "compare",
		"analyze", "strategy", "plan", "complex",
		"arsitektur", "desain", "analisis", "strategi",
	},
	"visual": {
		"ui", "ux", "frontend", "css", "style", "layout",
		"component", "button", "page", "design",
		"tampilan", "desain",
	},
}

// SimpleClassify uses keyword matching for fast classification.
func SimpleClassify(query string) (string, float64) {
	queryLower := strings.ToLower(query)
	scores := make(map[string]int)

	for category, keywords := range CategoryKeywords {
		for _, kw := range keywords {
			if strings.Contains(queryLower, kw) {
				scores[category]++
			}
		}
	}

	bestCategory := "general"
	bestScore := 0
	for cat, score := range scores {
		if score > bestScore {
			bestScore = score
			bestCategory = cat
		}
	}

	confidence := float64(bestScore) / 3.0
	if confidence > 1.0 {
		confidence = 1.0
	}
	if bestScore == 0 {
		confidence = 0.3
	}

	return bestCategory, confidence
}
