package tools

import (
	"testing"
)

func TestHashEdit_Verify(t *testing.T) {
	h := NewHashEdit()
	content := "line1\nline2\nline3"
	h.StoreLineHash(1, h.LineHash("line1"))
	h.StoreLineHash(2, h.LineHash("line2"))
	h.StoreLineHash(3, h.LineHash("line3"))

	if !h.Verify(content) {
		t.Error("Verification should pass")
	}
}

func TestHashEdit_Rollback(t *testing.T) {
	h := NewHashEdit()
	original := "old content"
	hash := h.Compute(original)
	h.StoreLineHash(1, hash)

	restored := h.Rollback(1)
	if restored == "" {
		t.Error("Rollback should return content")
	}
}