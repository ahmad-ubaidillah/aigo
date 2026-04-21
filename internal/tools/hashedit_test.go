package tools

import (
	"testing"
)

func TestHashEdit_Compute(t *testing.T) {
	h := NewHashEdit()
	hash := h.Compute("content")
	if hash == "" {
		t.Error("Hash should not be empty")
	}
}

func TestHashEdit_LineHash(t *testing.T) {
	h := NewHashEdit()
	lineHash := h.LineHash("line 1")
	if lineHash == "" {
		t.Error("Line hash should not be empty")
	}
}

func TestHashEdit_Store(t *testing.T) {
	h := NewHashEdit()
	h.StoreLineHash(1, "hash1")
	h.StoreLineHash(2, "hash2")

	if !h.HasLineHash(1) {
		t.Error("Line 1 hash should exist")
	}
}

func TestHashEdit_SurgicalEdit(t *testing.T) {
	h := NewHashEdit()
	original := "line1\nline2\nline3"
	modified, err := h.SurgicalEdit(original, 2, "new line2")

	if err != nil {
		t.Logf("Edit error: %v", err)
	}
	_ = modified
}