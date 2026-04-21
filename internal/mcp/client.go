// Package mcp implements MCP (Model Context Protocol) client for Aigo.
// Connects to MCP servers and exposes their tools to the agent.
package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sync"
)

// Server represents a connected MCP server.
type Server struct {
	Name    string
	Command string
	Args    []string
	Tools   []ToolDef
	cmd     *exec.Cmd
}

// ToolDef is an MCP tool definition.
type ToolDef struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema interface{} `json:"inputSchema"`
}

// Client manages MCP server connections.
type Client struct {
	servers map[string]*Server
	mu      sync.RWMutex
}

// New creates a new MCP client.
func New() *Client {
	return &Client{servers: make(map[string]*Server)}
}

// LoadConfig loads MCP servers from ~/.aigo/mcp_servers.json
// Format compatible with Claude Desktop.
func (c *Client) LoadConfig() error {
	home, _ := os.UserHomeDir()
	configPath := home + "/.aigo/mcp_servers.json"

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No config = no MCP servers
		}
		return err
	}

	var config struct {
		MCPServers map[string]struct {
			Command string   `json:"command"`
			Args    []string `json:"args"`
			Env     []string `json:"env"`
		} `json:"mcpServers"`
	}

	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("parse mcp config: %w", err)
	}

	for name, srv := range config.MCPServers {
		c.mu.Lock()
		c.servers[name] = &Server{
			Name:    name,
			Command: srv.Command,
			Args:    srv.Args,
		}
		c.mu.Unlock()
		log.Printf("MCP: registered server %s", name)
	}

	return nil
}

// Connect starts all configured MCP servers.
func (c *Client) Connect(ctx context.Context) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for name, server := range c.servers {
		cmd := exec.CommandContext(ctx, server.Command, server.Args...)
		cmd.Env = os.Environ()

		if err := cmd.Start(); err != nil {
			log.Printf("MCP: failed to start %s: %v", name, err)
			continue
		}

		server.cmd = cmd
		log.Printf("MCP: started %s (PID %d)", name, cmd.Process.Pid)
	}
	return nil
}

// GetTools returns all tools from all connected MCP servers.
func (c *Client) GetTools() []ToolDef {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var tools []ToolDef
	for _, server := range c.servers {
		tools = append(tools, server.Tools...)
	}
	return tools
}

// CallTool calls an MCP tool by namespaced name (server__tool).
func (c *Client) CallTool(ctx context.Context, namespacedName string, args map[string]interface{}) (string, error) {
	// For now, return a placeholder. Full MCP protocol implementation comes later.
	return fmt.Sprintf("MCP tool '%s' called (full protocol not yet implemented)", namespacedName), nil
}

// Stop stops all MCP servers.
func (c *Client) Stop() {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for name, server := range c.servers {
		if server.cmd != nil && server.cmd.Process != nil {
			server.cmd.Process.Kill()
			log.Printf("MCP: stopped %s", name)
		}
	}
}
