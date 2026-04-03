package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ahmad-ubaidillah/aigo/pkg/types"
)

// WebFetchTool fetches content from a URL and strips HTML.
type WebFetchTool struct{}

func (WebFetchTool) Name() string { return "webfetch" }

func (WebFetchTool) Description() string { return "Fetch content from a URL" }

func (WebFetchTool) Schema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"url": map[string]interface{}{
				"type":        "string",
				"description": "URL to fetch",
			},
		},
		"required": []string{"url"},
	}
}

func (WebFetchTool) Execute(ctx context.Context, params map[string]interface{}) (*types.ToolResult, error) {
	urlVal, ok := params["url"]
	if !ok {
		return &types.ToolResult{Success: false, Error: "missing required param: url"}, nil
	}
	url, ok := urlVal.(string)
	if !ok || strings.TrimSpace(url) == "" {
		return &types.ToolResult{Success: false, Error: "invalid param: url"}, nil
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return &types.ToolResult{Success: false, Error: err.Error()}, nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &types.ToolResult{Success: false, Error: err.Error()}, nil
	}
	content := string(body)
	// Strip HTML tags
	re := regexp.MustCompile("<[^>]*>")
	content = re.ReplaceAllString(content, "")
	// Truncate to 50KB by mutating a ToolResult's Output field
	tr := &types.ToolResult{Output: content}
	OutputTruncate(tr, 50*1024)
	content = tr.Output

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &types.ToolResult{Success: false, Output: content, Error: fmt.Sprintf("HTTP status %d", resp.StatusCode)}, nil
	}
	return &types.ToolResult{Success: true, Output: content}, nil
}

// WebSearchTool stub for web search capability.
type WebSearchTool struct{}

func (WebSearchTool) Name() string { return "websearch" }

func (WebSearchTool) Description() string { return "Search the web for information" }

func (WebSearchTool) Schema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"query": map[string]interface{}{
				"type":        "string",
				"description": "Search query",
			},
			"num_results": map[string]interface{}{
				"type":        "integer",
				"description": "Number of results (default 5)",
			},
		},
		"required": []string{"query"},
	}
}

func (WebSearchTool) Execute(ctx context.Context, params map[string]interface{}) (*types.ToolResult, error) {
	qVal, ok := params["query"]
	if !ok {
		return &types.ToolResult{Success: false, Error: "missing required param: query"}, nil
	}
	query, ok := qVal.(string)
	if !ok || strings.TrimSpace(query) == "" {
		return &types.ToolResult{Success: false, Error: "invalid param: query"}, nil
	}

	// Optional: number of results, default 5
	n := 5
	if v, ok := params["num_results"]; ok {
		switch t := v.(type) {
		case int:
			n = t
		case int64:
			n = int(t)
		case float64:
			n = int(t)
		case float32:
			n = int(t)
		case string:
			if iv, err := strconv.Atoi(t); err == nil {
				n = iv
			}
		}
	}
	if n <= 0 {
		n = 5
	}
	if n > 20 {
		n = 20
	}

	// Build DuckDuckGo HTML search URL
	searchURL := "https://html.duckduckgo.com/html/?q=" + url.QueryEscape(query)

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return &types.ToolResult{Success: false, Error: fmt.Sprintf("failed to create request: %v", err)}, nil
	}

	// Set headers to mimic a browser (DuckDuckGo may block requests without user agent)
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")

	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		return &types.ToolResult{Success: false, Error: fmt.Sprintf("search request failed: %v", err)}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &types.ToolResult{Success: false, Error: fmt.Sprintf("search failed with HTTP status %d", resp.StatusCode)}, nil
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &types.ToolResult{Success: false, Error: fmt.Sprintf("failed to read response: %v", err)}, nil
	}

	// Parse search results from HTML
	results := parseDuckDuckGoHTML(string(body), n)

	if len(results) == 0 {
		return &types.ToolResult{Success: true, Output: "No search results found."}, nil
	}

	// Format results
	var sb strings.Builder
	for i, r := range results {
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, r.Title))
		sb.WriteString(fmt.Sprintf("   URL: %s\n", r.URL))
		sb.WriteString(fmt.Sprintf("   Snippet: %s\n", r.Snippet))
		if i < len(results)-1 {
			sb.WriteString("\n")
		}
	}

	return &types.ToolResult{Success: true, Output: sb.String()}, nil
}

// searchResult represents a single search result
type searchResult struct {
	Title   string
	URL     string
	Snippet string
}

// parseDuckDuckGoHTML parses DuckDuckGo HTML search results
func parseDuckDuckGoHTML(html string, maxResults int) []searchResult {
	var results []searchResult

	// DuckDuckGo HTML structure:
	// Results are in <div class="result"> or <div class="result results_links results_links_deep">
	// Title: <a class="result__a" href="...">Title</a>
	// Snippet: <a class="result__snippet">Snippet</a>
	// URL is embedded in the redirect URL: //duckduckgo.com/l/?uddg=ENCODED_URL&...

	// Find each result by the result__a class (contains title and URL)
	resultRE := regexp.MustCompile(`<a[^>]*class="result__a"[^>]*href="([^"]+)"[^>]*>([^<]+)</a>`)

	// Find snippet - can be in <a class="result__snippet"> or sometimes other elements
	snippetRE := regexp.MustCompile(`<a[^>]*class="result__snippet"[^>]*>(?s:.*?)</a>`)

	// Find all result entries
	matches := resultRE.FindAllStringSubmatch(html, -1)

	for i, match := range matches {
		if i >= maxResults {
			break
		}

		if len(match) < 3 {
			continue
		}

		rawURL := match[1]
		title := cleanHTML(match[2])

		// Decode the DuckDuckGo redirect URL
		// Format: //duckduckgo.com/l/?uddg=ENCODED_URL&rut=...
		// or: /l/?uddg=ENCODED_URL&...
		actualURL := extractURL(rawURL)
		if actualURL == "" {
			actualURL = rawURL
		}

		// Find snippet near this result (look in surrounding HTML)
		snippet := extractSnippet(html, match[0], snippetRE)

		if title != "" && actualURL != "" {
			results = append(results, searchResult{
				Title:   title,
				URL:     actualURL,
				Snippet: snippet,
			})
		}
	}

	// If the regex approach didn't work well, try alternative parsing
	if len(results) == 0 {
		results = parseDuckDuckGoHTMLAlternative(html, maxResults)
	}

	return results
}

// parseDuckDuckGoHTMLAlternative provides an alternative parsing approach
func parseDuckDuckGoHTMLAlternative(html string, maxResults int) []searchResult {
	var results []searchResult

	// Look for patterns like:
	// <a rel="nofollow" class="result__a" href="...">Title</a>
	titleLinkRE := regexp.MustCompile(`<a[^>]*class="[^"]*result__a[^"]*"[^>]*href="([^"]+)"[^>]*>([^<]+)</a>`)

	matches := titleLinkRE.FindAllStringSubmatch(html, -1)

	for i, match := range matches {
		if i >= maxResults {
			break
		}

		if len(match) < 3 {
			continue
		}

		rawURL := match[1]
		title := cleanHTML(match[2])

		actualURL := extractURL(rawURL)
		if actualURL == "" {
			actualURL = rawURL
		}

		results = append(results, searchResult{
			Title:   title,
			URL:     actualURL,
			Snippet: "",
		})
	}

	return results
}

// extractURL decodes the DuckDuckGo redirect URL to get the actual target URL
func extractURL(redirectURL string) string {
	// Handle protocol-relative URLs
	if strings.HasPrefix(redirectURL, "//") {
		redirectURL = "https:" + redirectURL
	}

	// Parse the URL
	u, err := url.Parse(redirectURL)
	if err != nil {
		return redirectURL
	}

	// Check if it's a DuckDuckGo redirect URL
	if strings.Contains(u.Host, "duckduckgo.com") && u.Path == "/l/" {
		// Extract the uddg parameter which contains the actual URL
		uddg := u.Query().Get("uddg")
		if uddg != "" {
			// URL decode if needed
			decoded, err := url.QueryUnescape(uddg)
			if err == nil {
				return decoded
			}
			return uddg
		}
	}

	return redirectURL
}

// extractSnippet finds the snippet text near a result
func extractSnippet(fullHTML, resultMarker string, snippetRE *regexp.Regexp) string {
	// Find the position of this result in the HTML
	idx := strings.Index(fullHTML, resultMarker)
	if idx == -1 {
		return ""
	}

	// Look for snippet in a window around the result (500 chars after should be enough)
	endIdx := idx + len(resultMarker) + 500
	if endIdx > len(fullHTML) {
		endIdx = len(fullHTML)
	}

	searchArea := fullHTML[idx:endIdx]

	// Try to find snippet
	match := snippetRE.FindString(searchArea)
	if match == "" {
		return ""
	}

	// Extract text content from the snippet element
	// Remove HTML tags
	snippet := cleanHTML(match)

	// Truncate if too long
	if len(snippet) > 300 {
		snippet = snippet[:297] + "..."
	}

	return snippet
}

// cleanHTML removes HTML tags and decodes HTML entities
func cleanHTML(s string) string {
	// Remove HTML tags
	tagRE := regexp.MustCompile(`<[^>]*>`)
	s = tagRE.ReplaceAllString(s, " ")

	// Decode common HTML entities
	s = strings.ReplaceAll(s, "&amp;", "&")
	s = strings.ReplaceAll(s, "&lt;", "<")
	s = strings.ReplaceAll(s, "&gt;", ">")
	s = strings.ReplaceAll(s, "&quot;", "\"")
	s = strings.ReplaceAll(s, "&#39;", "'")
	s = strings.ReplaceAll(s, "&nbsp;", " ")

	// Clean up whitespace
	s = strings.Join(strings.Fields(s), " ")

	return strings.TrimSpace(s)
}

// Todo support types
type todoItem struct {
	Content string
	Status  string
}

// In-memory todo list
var (
	mu    sync.Mutex
	todos []todoItem
)

// TodoTool manages a tiny in-process todo list.
type TodoTool struct{}

func (TodoTool) Name() string { return "todo" }

func (TodoTool) Description() string { return "Manage a todo list" }

func (TodoTool) Schema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"action": map[string]interface{}{
				"type":        "string",
				"description": "add|list|complete|cancel",
			},
			"content": map[string]interface{}{
				"type":        "string",
				"description": "Todo text (required for add)",
			},
			"index": map[string]interface{}{
				"type":        "integer",
				"description": "Todo index (required for complete/cancel)",
			},
		},
		"required": []string{"action"},
	}
}

func (TodoTool) Execute(ctx context.Context, params map[string]interface{}) (*types.ToolResult, error) {
	actionVal, ok := params["action"]
	if !ok {
		return &types.ToolResult{Success: false, Error: "missing required param: action"}, nil
	}
	action, _ := actionVal.(string)
	switch action {
	case "add":
		contentVal, ok := params["content"]
		if !ok {
			return &types.ToolResult{Success: false, Error: "missing required param: content"}, nil
		}
		content, ok := contentVal.(string)
		if !ok {
			return &types.ToolResult{Success: false, Error: "invalid param: content"}, nil
		}
		mu.Lock()
		todos = append(todos, todoItem{Content: content, Status: "pending"})
		mu.Unlock()
		return &types.ToolResult{Success: true, Output: "added"}, nil
	case "list":
		mu.Lock()
		var b strings.Builder
		for i, t := range todos {
			b.WriteString(fmt.Sprintf("%d. %s [%s]\n", i+1, t.Content, t.Status))
		}
		mu.Unlock()
		out := strings.TrimSuffix(b.String(), "\n")
		return &types.ToolResult{Success: true, Output: out}, nil
	case "complete":
		idxVal, ok := params["index"]
		if !ok {
			return &types.ToolResult{Success: false, Error: "missing required param: index"}, nil
		}
		idx, err := toInt(idxVal)
		if err != nil {
			return &types.ToolResult{Success: false, Error: "invalid param: index"}, nil
		}
		mu.Lock()
		if idx < 1 || idx > len(todos) {
			mu.Unlock()
			return &types.ToolResult{Success: false, Error: "index out of range"}, nil
		}
		todos[idx-1].Status = "completed"
		mu.Unlock()
		return &types.ToolResult{Success: true, Output: fmt.Sprintf("completed: %d", idx)}, nil
	case "cancel":
		idxVal, ok := params["index"]
		if !ok {
			return &types.ToolResult{Success: false, Error: "missing required param: index"}, nil
		}
		idx, err := toInt(idxVal)
		if err != nil {
			return &types.ToolResult{Success: false, Error: "invalid param: index"}, nil
		}
		mu.Lock()
		if idx < 1 || idx > len(todos) {
			mu.Unlock()
			return &types.ToolResult{Success: false, Error: "index out of range"}, nil
		}
		todos[idx-1].Status = "cancelled"
		mu.Unlock()
		return &types.ToolResult{Success: true, Output: fmt.Sprintf("cancelled: %d", idx)}, nil
	default:
		return &types.ToolResult{Success: false, Error: "invalid action"}, nil
	}
}

// toInt converts a few common numeric representations to int.
func toInt(v interface{}) (int, error) {
	switch t := v.(type) {
	case int:
		return t, nil
	case int64:
		return int(t), nil
	case int32:
		return int(t), nil
	case float64:
		return int(t), nil
	case float32:
		return int(t), nil
	case string:
		return strconv.Atoi(t)
	default:
		return 0, fmt.Errorf("unsupported type")
	}
}
