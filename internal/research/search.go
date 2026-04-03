package research

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ahmad-ubaidillah/aigo/pkg/types"
)

type SearchClient struct {
	httpClient *http.Client
}

func NewSearchClient() *SearchClient {
	return &SearchClient{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

type WebSearchResult struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Content string `json:"content"`
}

type WebSearchResponse struct {
	Results []WebSearchResult `json:"results"`
}

func (c *SearchClient) WebSearch(ctx context.Context, query string, numResults int) (*types.ResearchQuery, error) {
	apiURL := fmt.Sprintf("https://api.duckduckgo.com/?q=%s&format=json&no_html=1&skip_disambig=1", query)

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("search: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Results []struct {
			Text     string `json:"Text"`
			FirstURL string `json:"FirstURL"`
		} `json:"RelatedTopics"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}

	output := ""
	for i, r := range result.Results {
		if i >= numResults {
			break
		}
		output += fmt.Sprintf("%s\n%s\n\n", r.Text, r.FirstURL)
	}

	return &types.ResearchQuery{
		ID:        generateQueryID(),
		Query:     query,
		Sources:   []string{"web"},
		Result:    output,
		CreatedAt: time.Now(),
	}, nil
}

type CodeSearchResult struct {
	Repo    string `json:"repo"`
	File    string `json:"file"`
	Content string `json:"content"`
	Line    int    `json:"line"`
}

type CodeSearchResponse struct {
	Results []CodeSearchResult `json:"results"`
}

func (c *SearchClient) CodeSearch(ctx context.Context, pattern, language string) (*types.ResearchQuery, error) {
	apiURL := fmt.Sprintf("https://grep.app/search?q=%s&lang=%s", pattern, language)

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("search: %w", err)
	}
	defer resp.Body.Close()

	var result CodeSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}

	output := ""
	for i, r := range result.Results {
		if i >= 10 {
			break
		}
		output += fmt.Sprintf("%s:%d\n%s\n\n", r.File, r.Line, r.Content)
	}

	return &types.ResearchQuery{
		ID:        generateQueryID(),
		Query:     pattern,
		Sources:   []string{"code"},
		Result:    output,
		CreatedAt: time.Now(),
	}, nil
}

func (c *SearchClient) DocsSearch(ctx context.Context, query, docType string) (*types.ResearchQuery, error) {
	sources := []string{"docs"}

	var output string
	switch docType {
	case "go":
		output = c.searchGoDocs(ctx, query)
	case "python":
		output = c.searchPythonDocs(ctx, query)
	default:
		output = "unsupported doc type"
	}

	return &types.ResearchQuery{
		ID:        generateQueryID(),
		Query:     query,
		Sources:   sources,
		Result:    output,
		CreatedAt: time.Now(),
	}, nil
}

func (c *SearchClient) searchGoDocs(ctx context.Context, query string) string {
	return fmt.Sprintf("Go documentation search for: %s\nhttps://pkg.go.dev/search?q=%s", query, query)
}

func (c *SearchClient) searchPythonDocs(ctx context.Context, query string) string {
	return fmt.Sprintf("Python documentation search for: %s\nhttps://docs.python.org/3/search.html?q=%s", query, query)
}

func generateQueryID() string {
	return fmt.Sprintf("query_%d", time.Now().UnixNano())
}
