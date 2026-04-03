package mcp

import (
	"os"
	"testing"

	"github.com/ahmad-ubaidillah/aigo/pkg/types"
)

func TestNewClient(t *testing.T) {
	t.Parallel()
	c := NewClient()
	if c == nil {
		t.Error("expected client")
	}
}

func TestClient_AddConnection(t *testing.T) {
	t.Parallel()
	c := NewClient()
	conn := &types.MCPConnection{ID: "ctx7", Name: "Context7"}
	err := c.AddConnection(conn)
	if err != nil {
		t.Fatal(err)
	}
}

func TestClient_AddNilConnection(t *testing.T) {
	t.Parallel()
	c := NewClient()
	err := c.AddConnection(nil)
	if err == nil {
		t.Error("expected error for nil connection")
	}
}

func TestClient_RemoveConnection(t *testing.T) {
	t.Parallel()
	c := NewClient()
	c.AddConnection(&types.MCPConnection{ID: "ctx7", Name: "Context7"})
	err := c.RemoveConnection("ctx7")
	if err != nil {
		t.Fatal(err)
	}
}

func TestClient_RemoveConnectionNotFound(t *testing.T) {
	t.Parallel()
	c := NewClient()
	err := c.RemoveConnection("nonexistent")
	if err == nil {
		t.Error("expected error")
	}
}

func TestClient_ListConnections(t *testing.T) {
	t.Parallel()
	c := NewClient()
	c.AddConnection(&types.MCPConnection{ID: "ctx7", Name: "Context7"})
	conns := c.ListConnections()
	if len(conns) != 1 {
		t.Errorf("expected 1 connection, got %d", len(conns))
	}
}

func TestClient_GetConnection(t *testing.T) {
	t.Parallel()
	c := NewClient()
	c.AddConnection(&types.MCPConnection{ID: "ctx7", Name: "Context7"})
	conn, err := c.GetConnection("ctx7")
	if err != nil {
		t.Fatal(err)
	}
	if conn.Name != "Context7" {
		t.Errorf("expected Context7, got %s", conn.Name)
	}
}

func TestClient_GetConnectionNotFound(t *testing.T) {
	t.Parallel()
	c := NewClient()
	_, err := c.GetConnection("nonexistent")
	if err == nil {
		t.Error("expected error")
	}
}

func TestClient_LoadAPIKeys(t *testing.T) {
	t.Parallel()
	os.Setenv("CONTEXT7_API_KEY", "test-key")
	c := NewClient()
	c.LoadAPIKeys()
	os.Unsetenv("CONTEXT7_API_KEY")
	if c.apiKeys["context7"] != "test-key" {
		t.Errorf("expected test-key, got %s", c.apiKeys["context7"])
	}
}
