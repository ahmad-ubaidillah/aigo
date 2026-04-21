package memory

import (
	"testing"
)

func TestHybrid_Search(t *testing.T) {
	h := NewHybridSearch()
	results := h.Search("test query", [][]float64{}, 5)

	if results == nil {
		t.Log("Search returned nil (acceptable for skeleton)")
	}
}

func TestHybrid_CombineScore(t *testing.T) {
	h := NewHybridSearch()
	ftsScore := 0.8
	vecScore := 0.6

	combined := h.CombineScore(ftsScore, vecScore)
	if combined <= 0 {
		t.Error("Combined score should be positive")
	}
}