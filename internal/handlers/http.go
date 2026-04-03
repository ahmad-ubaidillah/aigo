package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/ahmad-ubaidillah/aigo/pkg/types"
)

type HTTPHandler struct {
	client *http.Client
}

func NewHTTPHandler() *HTTPHandler {
	return &HTTPHandler{
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (h *HTTPHandler) CanHandle(intent string) bool {
	return intent == types.IntentHTTPCall
}

type HTTPRequest struct {
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers,omitempty"`
	Body    string            `json:"body,omitempty"`
}

type HTTPResponse struct {
	Status     int               `json:"status"`
	StatusText string            `json:"status_text"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
	Duration   int64             `json:"duration_ms"`
}

func (h *HTTPHandler) Execute(ctx context.Context, task *types.Task, workspace string) (*types.ToolResult, error) {
	var req HTTPRequest
	if err := json.Unmarshal([]byte(task.Description), &req); err != nil {
		return &types.ToolResult{
			Success: false,
			Error:   fmt.Sprintf("parse request: %v", err),
		}, nil
	}

	start := time.Now()
	httpReq, err := http.NewRequestWithContext(ctx, req.Method, req.URL, strings.NewReader(req.Body))
	if err != nil {
		return &types.ToolResult{
			Success: false,
			Error:   fmt.Sprintf("create request: %v", err),
		}, nil
	}

	for k, v := range req.Headers {
		httpReq.Header.Set(k, v)
	}

	if httpReq.Header.Get("Content-Type") == "" {
		httpReq.Header.Set("Content-Type", "application/json")
	}

	resp, err := h.client.Do(httpReq)
	if err != nil {
		return &types.ToolResult{
			Success: false,
			Error:   fmt.Sprintf("request failed: %v", err),
		}, nil
	}
	defer resp.Body.Close()

	duration := time.Since(start).Milliseconds()
	body, _ := io.ReadAll(resp.Body)

	headers := make(map[string]string)
	for k, v := range resp.Header {
		if len(v) > 0 {
			headers[k] = v[0]
		}
	}

	result := HTTPResponse{
		Status:     resp.StatusCode,
		StatusText: resp.Status,
		Headers:    headers,
		Body:       string(body),
		Duration:   duration,
	}

	output, _ := json.Marshal(result)
	return &types.ToolResult{
		Success: true,
		Output:  string(output),
	}, nil
}

type HTTPBatchRequest struct {
	Requests []HTTPRequest `json:"requests"`
	Parallel bool          `json:"parallel,omitempty"`
}

func (h *HTTPHandler) ExecuteBatch(ctx context.Context, task *types.Task, workspace string) (*types.ToolResult, error) {
	var batch HTTPBatchRequest
	if err := json.Unmarshal([]byte(task.Description), &batch); err != nil {
		return &types.ToolResult{
			Success: false,
			Error:   fmt.Sprintf("parse batch request: %v", err),
		}, nil
	}

	results := make([]HTTPResponse, len(batch.Requests))
	for i, req := range batch.Requests {
		start := time.Now()
		httpReq, err := http.NewRequestWithContext(ctx, req.Method, req.URL, strings.NewReader(req.Body))
		if err != nil {
			results[i] = HTTPResponse{Status: 0, Body: err.Error()}
			continue
		}

		for k, v := range req.Headers {
			httpReq.Header.Set(k, v)
		}

		resp, err := h.client.Do(httpReq)
		if err != nil {
			results[i] = HTTPResponse{Status: 0, Body: err.Error()}
			continue
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		results[i] = HTTPResponse{
			Status:     resp.StatusCode,
			StatusText: resp.Status,
			Body:       string(body),
			Duration:   time.Since(start).Milliseconds(),
		}
	}

	output, _ := json.Marshal(results)
	return &types.ToolResult{
		Success: true,
		Output:  string(output),
	}, nil
}

func ParseHTTPRequest(input string) (*HTTPRequest, error) {
	input = strings.TrimSpace(input)
	if strings.HasPrefix(input, "{") {
		var req HTTPRequest
		if err := json.Unmarshal([]byte(input), &req); err != nil {
			return nil, err
		}
		return &req, nil
	}

	parts := strings.Fields(input)
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid format: use METHOD URL or JSON")
	}

	return &HTTPRequest{
		Method: strings.ToUpper(parts[0]),
		URL:    parts[1],
	}, nil
}

func ValidateURL(url string) error {
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return fmt.Errorf("URL must start with http:// or https://")
	}
	return nil
}

func SetDefaultHeaders(req *HTTPRequest) {
	if req.Headers == nil {
		req.Headers = make(map[string]string)
	}
	if req.Headers["Content-Type"] == "" {
		req.Headers["Content-Type"] = "application/json"
	}
	if req.Headers["User-Agent"] == "" {
		req.Headers["User-Agent"] = "Aigo/1.0"
	}
}

type HTTPRequestBuilder struct {
	req HTTPRequest
}

func NewHTTPRequestBuilder() *HTTPRequestBuilder {
	return &HTTPRequestBuilder{
		req: HTTPRequest{
			Headers: make(map[string]string),
		},
	}
}

func (b *HTTPRequestBuilder) Method(m string) *HTTPRequestBuilder {
	b.req.Method = strings.ToUpper(m)
	return b
}

func (b *HTTPRequestBuilder) URL(url string) *HTTPRequestBuilder {
	b.req.URL = url
	return b
}

func (b *HTTPRequestBuilder) Header(key, value string) *HTTPRequestBuilder {
	b.req.Headers[key] = value
	return b
}

func (b *HTTPRequestBuilder) Body(body string) *HTTPRequestBuilder {
	b.req.Body = body
	return b
}

func (b *HTTPRequestBuilder) JSONBody(v interface{}) *HTTPRequestBuilder {
	data, _ := json.Marshal(v)
	b.req.Body = string(data)
	b.req.Headers["Content-Type"] = "application/json"
	return b
}

func (b *HTTPRequestBuilder) Build() *HTTPRequest {
	SetDefaultHeaders(&b.req)
	return &b.req
}

func (b *HTTPRequestBuilder) String() string {
	data, _ := json.Marshal(b.req)
	return string(data)
}

type ResponseValidator struct {
	ExpectedStatus int
	CheckBody      bool
	BodyContains   string
}

func (v *ResponseValidator) Validate(resp HTTPResponse) error {
	if v.ExpectedStatus > 0 && resp.Status != v.ExpectedStatus {
		return fmt.Errorf("status mismatch: expected %d, got %d", v.ExpectedStatus, resp.Status)
	}
	if v.CheckBody && v.BodyContains != "" && !strings.Contains(resp.Body, v.BodyContains) {
		return fmt.Errorf("body does not contain: %s", v.BodyContains)
	}
	return nil
}

func FormatHTTPResponse(resp HTTPResponse) string {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("HTTP %d %s (%dms)\n", resp.Status, resp.StatusText, resp.Duration))
	buf.WriteString("Headers:\n")
	for k, v := range resp.Headers {
		buf.WriteString(fmt.Sprintf("  %s: %s\n", k, v))
	}
	buf.WriteString("Body:\n")
	buf.WriteString(resp.Body)
	return buf.String()
}
