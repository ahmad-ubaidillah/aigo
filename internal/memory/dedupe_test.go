package memory

import (
	"testing"
)

func TestDedupe_Hash(t *testing.T) {
	d := NewDedupe()
	hash1 := d.Hash("test content")
	hash2 := d.Hash("test content")

	if hash1 != hash2 {
		t.Error("Same content should produce same hash")
	}
}

func TestDedupe_IsDuplicate(t *testing.T) {
	d := NewDedupe()
	d.Add("content1", "hash1")

	if !d.IsDuplicate("content1") {
		t.Error("Known content should be duplicate")
	}
}

func TestDedupe_RemoveDuplicate(t *testing.T) {
	d := NewDedupe()
	d.Add("content1", "hash1")
	d.RemoveDuplicate("content1")

	if d.IsDuplicate("content1") {
		t.Error("Content should be removed")
	}
}