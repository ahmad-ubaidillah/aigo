package research

import (
	"context"
	"strings"
	"testing"
)

func TestNewSearchClient(t *testing.T) {
	t.Parallel()
	c := NewSearchClient()
	if c == nil {
		t.Error("expected client")
	}
	if c.httpClient == nil {
		t.Error("expected http client")
	}
}

func TestSearchClient_DocsSearch_Go(t *testing.T) {
	t.Parallel()
	c := NewSearchClient()
	result, err := c.DocsSearch(context.Background(), "fmt", "go")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result.Result, "Go documentation") {
		t.Errorf("expected Go docs, got %s", result.Result)
	}
}

func TestSearchClient_DocsSearch_Python(t *testing.T) {
	t.Parallel()
	c := NewSearchClient()
	result, err := c.DocsSearch(context.Background(), "os", "python")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result.Result, "Python documentation") {
		t.Errorf("expected Python docs, got %s", result.Result)
	}
}

func TestSearchClient_DocsSearch_Unsupported(t *testing.T) {
	t.Parallel()
	c := NewSearchClient()
	result, err := c.DocsSearch(context.Background(), "test", "ruby")
	if err != nil {
		t.Fatal(err)
	}
	if result.Result != "unsupported doc type" {
		t.Errorf("expected unsupported, got %s", result.Result)
	}
}

func TestSearchClient_SearchGoDocs(t *testing.T) {
	t.Parallel()
	c := NewSearchClient()
	result := c.searchGoDocs(context.Background(), "fmt")
	if !strings.Contains(result, "pkg.go.dev") {
		t.Errorf("expected pkg.go.dev, got %s", result)
	}
}

func TestSearchClient_SearchPythonDocs(t *testing.T) {
	t.Parallel()
	c := NewSearchClient()
	result := c.searchPythonDocs(context.Background(), "os")
	if !strings.Contains(result, "docs.python.org") {
		t.Errorf("expected docs.python.org, got %s", result)
	}
}

func TestGenerateQueryID(t *testing.T) {
	t.Parallel()
	id := generateQueryID()
	if !strings.HasPrefix(id, "query_") {
		t.Errorf("expected query_ prefix, got %s", id)
	}
}

func TestWebSearchResult(t *testing.T) {
	t.Parallel()
	r := WebSearchResult{Title: "Test", URL: "http://test.com", Content: "content"}
	if r.Title != "Test" {
		t.Errorf("expected Test, got %s", r.Title)
	}
}

func TestWebSearchResponse(t *testing.T) {
	t.Parallel()
	resp := WebSearchResponse{Results: []WebSearchResult{{Title: "Test"}}}
	if len(resp.Results) != 1 {
		t.Errorf("expected 1 result, got %d", len(resp.Results))
	}
}

func TestCodeSearchResult(t *testing.T) {
	t.Parallel()
	r := CodeSearchResult{Repo: "test/repo", File: "main.go", Content: "func main()", Line: 1}
	if r.File != "main.go" {
		t.Errorf("expected main.go, got %s", r.File)
	}
}

func TestCodeSearchResponse(t *testing.T) {
	t.Parallel()
	resp := CodeSearchResponse{Results: []CodeSearchResult{{File: "main.go"}}}
	if len(resp.Results) != 1 {
		t.Errorf("expected 1 result, got %d", len(resp.Results))
	}
}
