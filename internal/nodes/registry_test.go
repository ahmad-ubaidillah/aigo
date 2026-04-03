package nodes

import (
	"context"
	"testing"
)

func TestNewRegistry(t *testing.T) {
	t.Parallel()
	r := NewRegistry()
	if r == nil {
		t.Error("expected registry")
	}
}

func TestRegistry_List(t *testing.T) {
	t.Parallel()
	r := NewRegistry()
	nodes := r.List()
	if len(nodes) < 1 {
		t.Errorf("expected at least 1 builtin node, got %d", len(nodes))
	}
}

func TestRegistry_ListByType(t *testing.T) {
	t.Parallel()
	r := NewRegistry()
	nodes := r.ListByType("code")
	if len(nodes) < 1 {
		t.Errorf("expected at least 1 code node, got %d", len(nodes))
	}
}

func TestCodeNode(t *testing.T) {
	t.Parallel()
	r := NewRegistry()
	nodes := r.List()
	for _, n := range nodes {
		if n.Type() == "code" {
			if n.ID() == "" {
				t.Error("expected non-empty ID")
			}
			_, err := n.Execute(context.Background(), nil)
			if err != nil {
				t.Fatal(err)
			}
			break
		}
	}
}

func TestSearchNode(t *testing.T) {
	t.Parallel()
	r := NewRegistry()
	nodes := r.List()
	for _, n := range nodes {
		if n.Type() == "search" {
			if n.ID() == "" {
				t.Error("expected non-empty ID")
			}
			break
		}
	}
}

func TestLLMNode(t *testing.T) {
	t.Parallel()
	r := NewRegistry()
	nodes := r.List()
	for _, n := range nodes {
		if n.Type() == "llm" {
			if n.ID() == "" {
				t.Error("expected non-empty ID")
			}
			break
		}
	}
}

func TestHTTPNode(t *testing.T) {
	t.Parallel()
	r := NewRegistry()
	nodes := r.List()
	for _, n := range nodes {
		if n.Type() == "http" {
			if n.ID() == "" {
				t.Error("expected non-empty ID")
			}
			break
		}
	}
}

func TestBrowserNode(t *testing.T) {
	t.Parallel()
	r := NewRegistry()
	nodes := r.List()
	for _, n := range nodes {
		if n.Type() == "browser" {
			if n.ID() == "" {
				t.Error("expected non-empty ID")
			}
			break
		}
	}
}

func TestConditionNode(t *testing.T) {
	t.Parallel()
	r := NewRegistry()
	nodes := r.List()
	for _, n := range nodes {
		if n.Type() == "condition" {
			if n.ID() == "" {
				t.Error("expected non-empty ID")
			}
			break
		}
	}
}

func TestLoopNode(t *testing.T) {
	t.Parallel()
	r := NewRegistry()
	nodes := r.List()
	for _, n := range nodes {
		if n.Type() == "loop" {
			if n.ID() == "" {
				t.Error("expected non-empty ID")
			}
			break
		}
	}
}

func TestTransformNode(t *testing.T) {
	t.Parallel()
	r := NewRegistry()
	nodes := r.List()
	for _, n := range nodes {
		if n.Type() == "transform" {
			if n.ID() == "" {
				t.Error("expected non-empty ID")
			}
			break
		}
	}
}

func TestWaitNode(t *testing.T) {
	t.Parallel()
	r := NewRegistry()
	nodes := r.List()
	for _, n := range nodes {
		if n.Type() == "wait" {
			if n.ID() == "" {
				t.Error("expected non-empty ID")
			}
			break
		}
	}
}

func TestInputNode(t *testing.T) {
	t.Parallel()
	r := NewRegistry()
	nodes := r.List()
	for _, n := range nodes {
		if n.Type() == "input" {
			if n.ID() == "" {
				t.Error("expected non-empty ID")
			}
			break
		}
	}
}
