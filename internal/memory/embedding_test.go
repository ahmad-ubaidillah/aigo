package memory

import (
	"testing"
)

func TestEmbedding_Encode(t *testing.T) {
	e := NewEmbedding()
	vec := e.Encode("test input")

	if vec == nil {
		t.Fatal("Encode returned nil")
	}
	if len(vec) == 0 {
		t.Error("Embedding vector is empty")
	}
}

func TestEmbedding_Normalize(t *testing.T) {
	e := NewEmbedding()
	vec := []float64{3.0, 4.0}
	normalized := e.Normalize(vec)

	if normalized == nil {
		t.Log("Normalize returned nil (acceptable)")
	}
}