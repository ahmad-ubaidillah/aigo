// Package pyramidtools implements tool functions for pyramidal memory.
package pyramidtools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hermes-v2/aigo/internal/memory/pyramid"
	"github.com/hermes-v2/aigo/internal/tools"
)

func RegisterPyramidTools(reg *tools.Registry, p *pyramid.Pyramid) {
	reg.Register(&PyramidStatsTool{pyramid: p})
	reg.Register(&PyramidSearchTool{pyramid: p})
	reg.Register(&PyramidCompressTool{pyramid: p})
}

// --- pyramid_stats ---
type PyramidStatsTool struct{ pyramid *pyramid.Pyramid }

func (t *PyramidStatsTool) Name() string        { return "pyramid_stats" }
func (t *PyramidStatsTool) Description() string  { return "Show pyramidal memory statistics (files and sizes per tier)." }
func (t *PyramidStatsTool) Annotations() tools.Annotations {
	return tools.Annotations{Destructive: false, ReadOnly: true}
}
func (t *PyramidStatsTool) Schema() tools.ToolSchema {
	return tools.ToolSchema{
		Type: "function",
		Function: tools.ToolFunctionSchema{
			Name:        "pyramid_stats",
			Description: "Show pyramidal memory statistics: file counts and sizes per tier (raw, daily, monthly, yearly, epoch).",
			Parameters: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
	}
}
func (t *PyramidStatsTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	stats := t.pyramid.Stats()
	data, err := json.MarshalIndent(stats, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// --- pyramid_search ---
type PyramidSearchTool struct{ pyramid *pyramid.Pyramid }

func (t *PyramidSearchTool) Name() string        { return "pyramid_search" }
func (t *PyramidSearchTool) Description() string  { return "Search across all pyramid memory tiers." }
func (t *PyramidSearchTool) Annotations() tools.Annotations {
	return tools.Annotations{Destructive: false, ReadOnly: true}
}
func (t *PyramidSearchTool) Schema() tools.ToolSchema {
	return tools.ToolSchema{
		Type: "function",
		Function: tools.ToolFunctionSchema{
			Name:        "pyramid_search",
			Description: "Search across all pyramid memory tiers (raw, daily, monthly, yearly, epoch) for a query string. Returns matching entries.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{"type": "string", "description": "Search query"},
					"limit": map[string]interface{}{"type": "integer", "description": "Max results per tier (default 5)"},
				},
				"required": []string{"query"},
			},
		},
	}
}
func (t *PyramidSearchTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	query, _ := args["query"].(string)
	if query == "" {
		return "", fmt.Errorf("query is required")
	}
	limit := 5
	if l, ok := args["limit"].(float64); ok && l > 0 {
		limit = int(l)
	}

	// Search across tiers 1-4 (summaries), plus recent raw logs
	var results []string
	queryLower := strings.ToLower(query)

	// Search raw logs
	raw, _ := t.pyramid.ReadRecent(48)
	if raw != "" {
		for _, line := range strings.Split(raw, "\n") {
			if strings.Contains(strings.ToLower(line), queryLower) {
				results = append(results, "[raw] "+line)
				if len(results) >= limit {
					break
				}
			}
		}
	}

	// Search summary tiers
	tierNames := map[int]string{1: "daily", 2: "monthly", 3: "yearly", 4: "epoch"}
	for tier := 1; tier <= 4; tier++ {
		entries, _ := t.pyramid.ReadTier(tier, 10)
		for _, e := range entries {
			if strings.Contains(strings.ToLower(e.Content), queryLower) {
				results = append(results, fmt.Sprintf("[%s] %s: %s", tierNames[tier], e.Date, truncate(e.Content, 200)))
			}
		}
	}

	if len(results) == 0 {
		return fmt.Sprintf("No results for: %s", query), nil
	}
	return strings.Join(results, "\n"), nil
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

// --- pyramid_compress ---
type PyramidCompressTool struct{ pyramid *pyramid.Pyramid }

func (t *PyramidCompressTool) Name() string        { return "pyramid_compress" }
func (t *PyramidCompressTool) Description() string  { return "Trigger manual compression of pyramid memory." }
func (t *PyramidCompressTool) Annotations() tools.Annotations {
	return tools.Annotations{Destructive: true, ReadOnly: false}
}
func (t *PyramidCompressTool) Schema() tools.ToolSchema {
	return tools.ToolSchema{
		Type: "function",
		Function: tools.ToolFunctionSchema{
			Name:        "pyramid_compress",
			Description: "Check if raw logs need compression and report status. Use with date parameter to compress a specific day's raw logs into a daily summary.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"date":    map[string]interface{}{"type": "string", "description": "Date to compress (YYYYMMDD format). If empty, just checks status."},
					"summary": map[string]interface{}{"type": "string", "description": "Summary content for the compressed day"},
				},
			},
		},
	}
}
func (t *PyramidCompressTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	date, _ := args["date"].(string)
	summary, _ := args["summary"].(string)

	if date == "" {
		needs, targetDate := t.pyramid.NeedsCompression()
		if needs {
			return fmt.Sprintf("🧠 Raw logs for %s are ready for compression. Provide date and summary to compress.", targetDate), nil
		}
		return "🧠 No compression needed yet. Raw logs are below threshold.", nil
	}

	if summary == "" {
		// Just scan the raw logs for that date
		raw, err := t.pyramid.ScanTier0Raw(date)
		if err != nil {
			return "", fmt.Errorf("scan raw logs: %w", err)
		}
		if raw == "" {
			return fmt.Sprintf("No raw logs found for date: %s", date), nil
		}
		return fmt.Sprintf("Raw logs for %s:\n%s\n\nProvide a 'summary' parameter to compress these.", date, raw), nil
	}

	err := t.pyramid.CompressDaily(date, summary)
	if err != nil {
		return "", fmt.Errorf("compress: %w", err)
	}
	return fmt.Sprintf("🧠 Compressed raw logs for %s into daily summary.", date), nil
}
