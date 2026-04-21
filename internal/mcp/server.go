package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/hermes-v2/aigo/internal/agent"
	"github.com/hermes-v2/aigo/internal/tools"
)

type MCPServer struct {
	agent   *agent.Agent
	tools   *tools.Registry
	config  ServerConfig
	server  *http.Server
	running bool
	mu      sync.RWMutex
}

type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
	Stdio bool  `yaml:"stdio"`
}

type MCPTool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
	ID      interface{}     `json:"id,omitempty"`
}

type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	Result  interface{} `json:"result,omitempty"`
	Error   *JSONError  `json:"error,omitempty"`
	ID      interface{} `json:"id,omitempty"`
}

type JSONError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func NewMCPServer(agent *agent.Agent, reg *tools.Registry) *MCPServer {
	return &MCPServer{
		agent: agent,
		tools: reg,
		config: ServerConfig{
			Host: "127.0.0.1",
			Port: 3100,
			Stdio: false,
		},
	}
}

func (s *MCPServer) Configure(cfg ServerConfig) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.config = cfg
}

func (s *MCPServer) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return nil
	}

	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	mux := http.NewServeMux()
	mux.HandleFunc("/mcp", s.handleMCP)
	mux.HandleFunc("/health", s.handleHealth)

	s.server = &http.Server{Addr: addr, Handler: mux}
	s.running = true
	s.mu.Unlock()

	go func() {
		log.Printf("🖥️ MCP server listening on %s", addr)
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("MCP server error: %v", err)
		}
	}()

	return nil
}

func (s *MCPServer) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.running || s.server == nil {
		return nil
	}
	s.running = false
	return s.server.Shutdown(context.Background())
}

func (s *MCPServer) handleMCP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodGet {
		s.handleInitialize(w, r)
		return
	}

	var req JSONRPCRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendError(w, -32700, "Parse error", nil, nil)
		return
	}

	var resp JSONRPCResponse
	switch req.Method {
	case "initialize":
		resp = s.handleInitializeRequest(req)
	case "tools/list":
		resp = s.handleToolsListRequest(req)
	case "tools/call":
		resp = s.handleToolsCallRequest(req)
	case "resources/list":
		resp = s.handleResourcesListRequest(req)
	default:
		resp = s.handleNotImplemented(req)
	}

	json.NewEncoder(w).Encode(resp)
}

func (s *MCPServer) handleInitialize(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		Result: map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{},
			},
			"serverInfo": map[string]string{
				"name":    "aigo",
				"version": "0.3.0",
			},
		},
	}
	json.NewEncoder(w).Encode(resp)
}

func (s *MCPServer) handleInitializeRequest(req JSONRPCRequest) JSONRPCResponse {
	return JSONRPCResponse{
		JSONRPC: "2.0",
		Result: map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{},
			},
			"serverInfo": map[string]string{
				"name":    "aigo",
				"version": "0.3.0",
			},
		},
		ID: req.ID,
	}
}

func (s *MCPServer) handleToolsListRequest(req JSONRPCRequest) JSONRPCResponse {
	toolList := s.tools.ListTools()
	toolsSlice := make([]MCPTool, 0, len(toolList))
	for _, t := range toolList {
		params, _ := t.Schema().Function.Parameters.(map[string]interface{})
		toolsSlice = append(toolsSlice, MCPTool{
			Name:        t.Name(),
			Description: t.Description(),
			InputSchema: params,
		})
	}

	return JSONRPCResponse{
		JSONRPC: "2.0",
		Result: map[string]interface{}{
			"tools": toolsSlice,
		},
		ID: req.ID,
	}
}

func (s *MCPServer) handleToolsCallRequest(req JSONRPCRequest) JSONRPCResponse {
	var params struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments,omitempty"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return s.sendErrorReturn(req.ID, -32602, "Invalid params")
	}

	result, err := s.tools.Execute(context.Background(), params.Name, params.Arguments)
	if err != nil {
		return s.sendErrorReturn(req.ID, -32603, fmt.Sprintf("Tool error: %v", err))
	}

	return JSONRPCResponse{
		JSONRPC: "2.0",
		Result: map[string]interface{}{
			"content": []map[string]string{
				{"type": "text", "text": fmt.Sprintf("%v", result)},
			},
		},
		ID: req.ID,
	}
}

func (s *MCPServer) handleResourcesListRequest(req JSONRPCRequest) JSONRPCResponse {
	return JSONRPCResponse{
		JSONRPC: "2.0",
		Result: map[string]interface{}{
			"resources": []map[string]string{},
		},
		ID: req.ID,
	}
}

func (s *MCPServer) handleNotImplemented(req JSONRPCRequest) JSONRPCResponse {
	return s.sendErrorReturn(req.ID, -32601, "Method not found")
}

func (s *MCPServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *MCPServer) sendError(w http.ResponseWriter, code int, message string, data interface{}, id interface{}) {
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		Error: &JSONError{
			Code:    code,
			Message: message,
			Data:    data,
		},
		ID: id,
	}
	json.NewEncoder(w).Encode(resp)
}

func (s *MCPServer) sendErrorReturn(id interface{}, code int, message string) JSONRPCResponse {
	return JSONRPCResponse{
		JSONRPC: "2.0",
		Error: &JSONError{
			Code:    code,
			Message: message,
		},
		ID: id,
	}
}

func LoadServerConfig(path string) (ServerConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return ServerConfig{}, err
	}

	var cfg struct {
		MCP struct {
			Host string `yaml:"host"`
			Port int    `yaml:"port"`
			Stdio bool  `yaml:"stdio"`
		} `yaml:"mcp"`
	}

	if err := json.Unmarshal(data, &cfg); err != nil {
		return ServerConfig{}, err
	}

	return ServerConfig{
		Host: cfg.MCP.Host,
		Port: cfg.MCP.Port,
		Stdio: cfg.MCP.Stdio,
	}, nil
}