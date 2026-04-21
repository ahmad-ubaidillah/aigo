package memory

import (
	"testing"
)

func TestSimilarity_Cosine(t *testing.T) {
	s := NewSimilarity()
	vec1 := []float64{1.0, 0.0}
	vec2 := []float64{1.0, 0.0}

	sim := s.Cosine(vec1, vec2)
	if sim != 1.0 {
		t.Errorf("Expected 1.0, got %f", sim)
	}
}

func TestSimilarity_TopK(t *testing.T) {
	s := NewSimilarity()
	queries := [][]float64{
		{1.0, 0.0},
		{0.0, 1.0},
	}
	target := []float64{1.0, 0.0}

	top := s.TopK(queries, target, 1)
	if len(top) != 1 {
		t.Error("Expected 1 result")
	}
}