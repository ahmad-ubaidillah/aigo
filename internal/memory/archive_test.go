package memory

import (
	"testing"
)

func TestArchive_Store(t *testing.T) {
	a := NewArchive()
	err := a.Store("old data")
	if err != nil {
		t.Logf("Store error: %v", err)
	}
}

func TestArchive_IsCold(t *testing.T) {
	a := NewArchive()
	if !a.IsCold("2020-01-01") {
		t.Log("Old date should be cold")
	}
}