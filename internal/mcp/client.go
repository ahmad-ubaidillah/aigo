package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/ahmad-ubaidillah/aigo/pkg/types"
)

type Client struct {
	conns      map[string]*types.MCPConnection
	httpClient *http.Client
	apiKeys    map[string]string
}

func NewClient() *Client {
	return &Client{
		conns:   make(map[string]*types.MCPConnection),
		apiKeys: make(map[string]string),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) LoadAPIKeys() {
	c.apiKeys["context7"] = os.Getenv("CONTEXT7_API_KEY")
	c.apiKeys["exa"] = os.Getenv("EXA_API_KEY")
	c.apiKeys["grepapp"] = os.Getenv("GREP_APP_API_KEY")
}

func (c *Client) AddConnection(conn *types.MCPConnection) error {
	if conn == nil {
		return fmt.Errorf("connection is nil")
	}
	c.conns[conn.ID] = conn
	return nil
}

func (c *Client) RemoveConnection(id string) error {
	if _, ok := c.conns[id]; !ok {
		return fmt.Errorf("connection not found: %s", id)
	}
	delete(c.conns, id)
	return nil
}

func (c *Client) ListConnections() []*types.MCPConnection {
	result := make([]*types.MCPConnection, 0, len(c.conns))
	for _, conn := range c.conns {
		result = append(result, conn)
	}
	return result
}

func (c *Client) GetConnection(id string) (*types.MCPConnection, error) {
	conn, ok := c.conns[id]
	if !ok {
		return nil, fmt.Errorf("connection not found: %s", id)
	}
	return conn, nil
}

type MCPRequest struct {
	Method string          `json:"method"`
	Params json.RawMessage `json:"params,omitempty"`
	ID     string          `json:"id"`
}

type MCPResponse struct {
	Result json.RawMessage `json:"result,omitempty"`
	Error  *MCPError       `json:"error,omitempty"`
	ID     string          `json:"id"`
}

type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (c *Client) CallTool(ctx context.Context, connID, toolName string, params map[string]interface{}) (*types.MCPToolResult, error) {
	conn, err := c.GetConnection(connID)
	if err != nil {
		return nil, err
	}

	if !conn.Enabled {
		return &types.MCPToolResult{
			Error: fmt.Sprintf("connection %s is disabled", connID),
		}, nil
	}

	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("marshal params: %w", err)
	}

	req := MCPRequest{
		Method: toolName,
		Params: paramsJSON,
		ID:     generateRequestID(),
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", conn.URL, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if apiKey := c.apiKeys[connID]; apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+apiKey)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("call tool: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var mcpResp MCPResponse
	if err := json.Unmarshal(body, &mcpResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if mcpResp.Error != nil {
		return &types.MCPToolResult{
			Error: mcpResp.Error.Message,
		}, nil
	}

	return &types.MCPToolResult{
		Success: true,
		Output:  string(mcpResp.Result),
	}, nil
}

func (c *Client) CallContext7(query string) (*types.MCPToolResult, error) {
	apiKey := c.apiKeys["context7"]
	if apiKey == "" {
		return c.fallbackContext7(query)
	}

	url := "https://api.context7.com/v1/query"
	body, _ := json.Marshal(map[string]string{"query": query})

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return c.fallbackContext7(query)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	return &types.MCPToolResult{
		Success: resp.StatusCode == 200,
		Output:  string(respBody),
	}, nil
}

func (c *Client) fallbackContext7(query string) (*types.MCPToolResult, error) {
	resp, err := c.httpClient.Get("https://context7.com/api/v1/search?q=" + query)
	if err != nil {
		return &types.MCPToolResult{
			Success: false,
			Error:   "Context7 not available",
		}, nil
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	return &types.MCPToolResult{
		Success: true,
		Output:  string(body),
	}, nil
}

func (c *Client) CallGrepApp(pattern, language string) (*types.MCPToolResult, error) {
	apiKey := c.apiKeys["grepapp"]
	if apiKey == "" {
		return c.fallbackGrepApp(pattern, language)
	}

	url := "https://grep.app/api/search"
	body, _ := json.Marshal(map[string]string{
		"pattern":  pattern,
		"language": language,
	})

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return c.fallbackGrepApp(pattern, language)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	return &types.MCPToolResult{
		Success: resp.StatusCode == 200,
		Output:  string(respBody),
	}, nil
}

func (c *Client) fallbackGrepApp(pattern, language string) (*types.MCPToolResult, error) {
	resp, err := c.httpClient.Get("https://grep.app/search?q=" + pattern + "&lang=" + language)
	if err != nil {
		return &types.MCPToolResult{
			Success: false,
			Error:   "grep.app not available",
		}, nil
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	return &types.MCPToolResult{
		Success: true,
		Output:  string(body),
	}, nil
}

func (c *Client) CallExaSearch(query string, numResults int) (*types.MCPToolResult, error) {
	apiKey := c.apiKeys["exa"]
	if apiKey == "" {
		return c.fallbackWebSearch(query, numResults)
	}

	url := "https://api.exa.ai/search"
	body, _ := json.Marshal(map[string]interface{}{
		"query":       query,
		"num_results": numResults,
	})

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return c.fallbackWebSearch(query, numResults)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	return &types.MCPToolResult{
		Success: resp.StatusCode == 200,
		Output:  string(respBody),
	}, nil
}

func (c *Client) CallWebSearch(query string, numResults int) (*types.MCPToolResult, error) {
	return c.CallExaSearch(query, numResults)
}

func (c *Client) fallbackWebSearch(query string, numResults int) (*types.MCPToolResult, error) {
	resp, err := c.httpClient.Get(fmt.Sprintf("https://api.duckduckgo.com/?q=%s&format=json&no_html=1", query))
	if err != nil {
		return &types.MCPToolResult{
			Success: false,
			Error:   "Web search not available",
		}, nil
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	return &types.MCPToolResult{
		Success: true,
		Output:  string(body),
	}, nil
}

func generateRequestID() string {
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}
