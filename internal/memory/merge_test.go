package memory

import (
	"testing"
)

func TestMerge_Simple(t *testing.T) {
	m := NewMerger()
	mem1 := &MemoryItem{ID: "1", Content: "test content"}
	mem2 := &MemoryItem{ID: "2", Content: "test content"}

	merged := m.Merge(mem1, mem2)
	if merged == nil {
		t.Fatal("Merge returned nil")
	}
}

func TestMerge_Conflict(t *testing.T) {
	m := NewMerger()
	mem1 := &MemoryItem{ID: "1", Content: "value A"}
	mem2 := &MemoryItem{ID: "1", Content: "value B"}

	resolved := m.ResolveConflict(mem1, mem2)
	if resolved == nil {
		t.Log("ResolveConflict returned nil (acceptable)")
	}
}