// Package webtools implements web search and URL fetch tools for Aigo.
package webtools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/hermes-v2/aigo/internal/distiller"
	"github.com/hermes-v2/aigo/internal/tools"
)

const lightpandaWS = "ws://localhost:9222"

func RegisterWebTools(reg *tools.Registry) {
	reg.Register(&WebSearchTool{})
	reg.Register(&WebFetchTool{})
	reg.Register(&WebBrowseTool{})
	reg.Register(&WebScreenshotTool{})
}

// --- web_search ---

type WebSearchTool struct{}

func (t *WebSearchTool) Name() string        { return "web_search" }
func (t *WebSearchTool) Description() string { return "Search the web using DuckDuckGo. Returns instant answers and related topics." }
func (t *WebSearchTool) Annotations() tools.Annotations {
	return tools.Annotations{Destructive: false, ReadOnly: true}
}
func (t *WebSearchTool) Schema() tools.ToolSchema {
	return tools.ToolSchema{
		Type: "function",
		Function: tools.ToolFunctionSchema{
			Name:        "web_search",
			Description: "Search the web using DuckDuckGo. Returns instant answers and related topics. Use this when you need current information or don't know something.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "The search query",
					},
				},
				"required": []string{"query"},
			},
		},
	}
}

func (t *WebSearchTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	query, _ := args["query"].(string)
	if query == "" {
		return "", fmt.Errorf("query is required")
	}
	result := searchDuckDuckGo(query)
	if result != "" {
		return result, nil
	}
	return fmt.Sprintf("No instant answer found for: %s", query), nil
}

func searchDuckDuckGo(query string) string {
	client := &http.Client{Timeout: 10 * time.Second}
	apiURL := fmt.Sprintf("https://api.duckduckgo.com/?q=%s&format=json&no_html=1&skip_disambig=1",
		url.QueryEscape(query))

	resp, err := client.Get(apiURL)
	if err != nil {
		return fmt.Sprintf("Search error: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Sprintf("Read error: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Sprintf("Parse error: %v", err)
	}

	var parts []string

	if abstract, _ := result["Abstract"].(string); abstract != "" {
		source, _ := result["AbstractSource"].(string)
		u, _ := result["AbstractURL"].(string)
		parts = append(parts, fmt.Sprintf("%s\nSource: %s (%s)", abstract, source, u))
	}

	if def, _ := result["Definition"].(string); def != "" {
		parts = append(parts, fmt.Sprintf("Definition: %s", def))
	}

	if answer, _ := result["Answer"].(string); answer != "" {
		parts = append(parts, fmt.Sprintf("Answer: %s", answer))
	}

	if topics, ok := result["RelatedTopics"].([]interface{}); ok {
		var topicStrs []string
		for i, t := range topics {
			if i >= 5 {
				break
			}
			if topic, _ := t.(map[string]interface{}); topic != nil {
				if text, _ := topic["Text"].(string); text != "" && len(text) > 10 {
					topicStrs = append(topicStrs, fmt.Sprintf("• %s", trunc(text, 120)))
				}
			}
		}
		if len(topicStrs) > 0 {
			parts = append(parts, "Related:\n"+strings.Join(topicStrs, "\n"))
		}
	}

	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, "\n\n")
}

// --- web_fetch ---

type WebFetchTool struct{}

func (t *WebFetchTool) Name() string        { return "web_fetch" }
func (t *WebFetchTool) Description() string { return "Fetch a URL and extract readable content. Converts HTML to clean text." }
func (t *WebFetchTool) Annotations() tools.Annotations {
	return tools.Annotations{Destructive: false, ReadOnly: true}
}
func (t *WebFetchTool) Schema() tools.ToolSchema {
	return tools.ToolSchema{
		Type: "function",
		Function: tools.ToolFunctionSchema{
			Name:        "web_fetch",
			Description: "Fetch a URL and extract readable content. Converts HTML to clean text/Markdown. Use this to read web pages, articles, documentation. Supports 3 distillation modes: research (full), smart (query-filtered), compact (aggressive).",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"url": map[string]interface{}{
						"type":        "string",
						"description": "The URL to fetch",
					},
					"max_chars": map[string]interface{}{
						"type":        "integer",
						"description": "Maximum characters to return (default 5000)",
					},
					"mode": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"research", "smart", "compact"},
						"description": "Distillation mode: 'research' (full content), 'smart' (query-filtered, default), 'compact' (aggressive filtering)",
					},
					"query": map[string]interface{}{
						"type":        "string",
						"description": "Keywords for distillation filtering. Sentences not containing these keywords will be filtered out. Only used in smart/compact mode.",
					},
				},
				"required": []string{"url"},
			},
		},
	}
}

func (t *WebFetchTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	rawURL, _ := args["url"].(string)
	if rawURL == "" {
		return "", fmt.Errorf("url is required")
	}

	maxChars := 5000
	if mc, ok := args["max_chars"].(float64); ok && mc > 0 {
		maxChars = int(mc)
	}

	// Parse distillation mode
	mode := distiller.ModeSmart // default
	if m, ok := args["mode"].(string); ok {
		switch strings.ToLower(m) {
		case "research":
			mode = distiller.ModeResearch
		case "compact":
			mode = distiller.ModeCompact
		}
	}

	// Parse query for distillation filtering
	query, _ := args["query"].(string)

	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequestWithContext(ctx, "GET", rawURL, nil)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; Aigo/0.3)")
	req.Header.Set("Accept", "text/html,text/plain")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetch: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1024*1024))
	if err != nil {
		return "", fmt.Errorf("read: %w", err)
	}

	rawHTML := string(body)

	// Use 3-layer distillation pipeline
	result := distiller.FullPipeline(rawHTML, query, mode)

	// Build output
	output := fmt.Sprintf("URL: %s\n", rawURL)
	if result.Title != "" {
		output += fmt.Sprintf("Title: %s\n", result.Title)
	}

	// Show distillation stats in research mode
	if mode == distiller.ModeResearch {
		output += fmt.Sprintf("Mode: %s | %d→%d bytes\n", result.Mode, result.OriginalSize, result.FinalSize)
	} else {
		compression := 0.0
		if result.OriginalSize > 0 {
			compression = (1.0 - float64(result.FinalSize)/float64(result.OriginalSize)) * 100
		}
		output += fmt.Sprintf("Mode: %s | %d→%d bytes (%.0f%% compressed) | %d→%d sentences\n",
			result.Mode, result.OriginalSize, result.FinalSize, compression,
			result.SentencesIn, result.SentencesOut)
	}

	content := result.Content
	if len(content) > maxChars {
		content = content[:maxChars-3] + "..."
	}
	output += fmt.Sprintf("\n%s", content)

	return output, nil
}

func htmlToText(html string) string {
	// Remove script and style blocks separately
	scriptRe := regexp.MustCompile(`(?is)<script[^>]*>.*?</script>`)
	html = scriptRe.ReplaceAllString(html, "")
	styleRe := regexp.MustCompile(`(?is)<style[^>]*>.*?</style>`)
	html = styleRe.ReplaceAllString(html, "")
	// Convert elements
	replacements := []struct{ p, r string }{
		{`(?i)<br\s*/?>`, "\n"},
		{`(?i)</?p[^>]*>`, "\n"},
		{`(?i)</?div[^>]*>`, "\n"},
		{`(?i)</?li[^>]*>`, "\n• "},
		{`(?i)<a[^>]*href="([^"]*)"[^>]*>(.*?)</a>`, "$2 [$1]"},
	}
	for _, r := range replacements {
		re := regexp.MustCompile(r.p)
		html = re.ReplaceAllString(html, r.r)
	}
	// Strip tags
	re := regexp.MustCompile(`<[^>]+>`)
	html = re.ReplaceAllString(html, "")
	// Entities
	for old, new := range map[string]string{"&amp;": "&", "&lt;": "<", "&gt;": ">", "&quot;": `"`, "&nbsp;": " "} {
		html = strings.ReplaceAll(html, old, new)
	}
	// Clean whitespace
	wsRe := regexp.MustCompile(`\n{3,}`)
	html = wsRe.ReplaceAllString(html, "\n\n")
	return strings.TrimSpace(html)
}

func extractTitle(html string) string {
	re := regexp.MustCompile(`(?is)<title[^>]*>(.*?)</title>`)
	m := re.FindStringSubmatch(html)
	if len(m) > 1 {
		return strings.TrimSpace(m[1])
	}
	return ""
}

func trunc(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

// --- web_browse ---

type WebBrowseTool struct{}

func (t *WebBrowseTool) Name() string        { return "web_browse" }
func (t *WebBrowseTool) Description() string { return "Browse a web page using a headless browser. Handles JavaScript-rendered pages." }
func (t *WebBrowseTool) Annotations() tools.Annotations {
	return tools.Annotations{Destructive: false, ReadOnly: true}
}
func (t *WebBrowseTool) Schema() tools.ToolSchema {
	return tools.ToolSchema{
		Type: "function",
		Function: tools.ToolFunctionSchema{
			Name:        "web_browse",
			Description: "Browse a web page using a headless browser (Lightpanda). Use this for pages that require JavaScript rendering, SPAs, or dynamic content that web_fetch cannot handle.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"url": map[string]interface{}{
						"type":        "string",
						"description": "The URL to browse",
					},
					"wait_selector": map[string]interface{}{
						"type":        "string",
						"description": "Optional CSS selector to wait for before extracting content",
					},
					"dump_format": map[string]interface{}{
						"type":        "string",
						"description": "Output format: 'text' (default) or 'markdown'",
					},
					"max_chars": map[string]interface{}{
						"type":        "integer",
						"description": "Maximum characters to return (default 5000)",
					},
				},
				"required": []string{"url"},
			},
		},
	}
}

func (t *WebBrowseTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	rawURL, _ := args["url"].(string)
	if rawURL == "" {
		return "", fmt.Errorf("url is required")
	}

	waitSelector, _ := args["wait_selector"].(string)

	maxChars := 5000
	if mc, ok := args["max_chars"].(float64); ok && mc > 0 {
		maxChars = int(mc)
	}

	// Create context with 15s timeout
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	// Connect to Lightpanda CDP
	allocCtx, allocCancel := chromedp.NewRemoteAllocator(ctx, lightpandaWS)
	defer allocCancel()

	browserCtx, browserCancel := chromedp.NewContext(allocCtx)
	defer browserCancel()

	var title string
	var content string

	tasks := chromedp.Tasks{
		chromedp.ActionFunc(func(ctx context.Context) error {
			// Enable network events
			return network.Enable().Do(ctx)
		}),
		chromedp.Navigate(rawURL),
	}

	if waitSelector != "" {
		tasks = append(tasks, chromedp.WaitReady(waitSelector, chromedp.ByQuery))
	} else {
		tasks = append(tasks, chromedp.WaitReady("body"))
	}

	tasks = append(tasks,
		chromedp.Title(&title),
		chromedp.ActionFunc(func(ctx context.Context) error {
			// Get the page text content via JS
			return chromedp.Evaluate(`
				(function() {
					// Remove script and style elements
					var clone = document.body.cloneNode(true);
					var scripts = clone.querySelectorAll('script, style, noscript');
					scripts.forEach(function(el) { el.remove(); });
					return clone.innerText || clone.textContent || '';
				})()
			`, &content).Do(ctx)
		}),
	)

	if err := chromedp.Run(browserCtx, tasks); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("timeout connecting to Lightpanda browser at %s. Make sure Lightpanda is running: lightpanda serve", lightpandaWS)
		}
		return "", fmt.Errorf("browser error (is Lightpanda running at %s?): %w", lightpandaWS, err)
	}

	// Clean up whitespace
	wsRe := regexp.MustCompile(`\n{3,}`)
	content = wsRe.ReplaceAllString(content, "\n\n")
	content = strings.TrimSpace(content)
	content = trunc(content, maxChars)

	result := fmt.Sprintf("URL: %s\n", rawURL)
	if title != "" {
		result += fmt.Sprintf("Title: %s\n", title)
	}
	result += fmt.Sprintf("\n%s", content)

	return result, nil
}

// --- web_screenshot ---

type WebScreenshotTool struct{}

func (t *WebScreenshotTool) Name() string        { return "web_screenshot" }
func (t *WebScreenshotTool) Description() string { return "Capture a screenshot of a web page using a headless browser." }
func (t *WebScreenshotTool) Annotations() tools.Annotations {
	return tools.Annotations{Destructive: false, ReadOnly: true}
}
func (t *WebScreenshotTool) Schema() tools.ToolSchema {
	return tools.ToolSchema{
		Type: "function",
		Function: tools.ToolFunctionSchema{
			Name:        "web_screenshot",
			Description: "Capture a screenshot of a web page using the headless browser. Returns the path to the saved PNG file.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"url": map[string]interface{}{
						"type":        "string",
						"description": "The URL to screenshot",
					},
					"output_path": map[string]interface{}{
						"type":        "string",
						"description": "Path to save the screenshot (default /tmp/aigo_screenshot.png)",
					},
				},
				"required": []string{"url"},
			},
		},
	}
}

func (t *WebScreenshotTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	rawURL, _ := args["url"].(string)
	if rawURL == "" {
		return "", fmt.Errorf("url is required")
	}

	outputPath := "/tmp/aigo_screenshot.png"
	if op, ok := args["output_path"].(string); ok && op != "" {
		outputPath = op
	}

	// Create context with 15s timeout
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	// Connect to Lightpanda CDP
	allocCtx, allocCancel := chromedp.NewRemoteAllocator(ctx, lightpandaWS)
	defer allocCancel()

	browserCtx, browserCancel := chromedp.NewContext(allocCtx)
	defer browserCancel()

	var buf []byte

	tasks := chromedp.Tasks{
		chromedp.Navigate(rawURL),
		chromedp.WaitReady("body"),
		chromedp.FullScreenshot(&buf, 90),
	}

	if err := chromedp.Run(browserCtx, tasks); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("timeout connecting to Lightpanda browser at %s. Make sure Lightpanda is running: lightpanda serve", lightpandaWS)
		}
		return "", fmt.Errorf("browser error (is Lightpanda running at %s?): %w", lightpandaWS, err)
	}

	if err := os.WriteFile(outputPath, buf, 0644); err != nil {
		return "", fmt.Errorf("write screenshot: %w", err)
	}

	return fmt.Sprintf("Screenshot saved to: %s", outputPath), nil
}
