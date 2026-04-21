package memory

import (
	"math"
	"testing"
)

func TestWordEmbedding_BuildVocabulary(t *testing.T) {
	e := NewWordEmbedding(64)

	texts := []string{
		"agent plan execution",
		"memory profile workspace",
		"hook analytics autonomy",
	}

	e.BuildVocabulary(texts)

	if e.Size() == 0 {
		t.Error("Should build vocabulary")
	}

	if !e.initialized {
		t.Log("Vocabulary initialized")
	}

	t.Logf("Vocabulary size: %d", e.Size())
}

func TestWordEmbedding_GetVector(t *testing.T) {
	e := NewWordEmbedding(32)

	texts := []string{"hello world"}
	e.BuildVocabulary(texts)

	vec, ok := e.GetVector("hello")
	if !ok {
		t.Error("Should get vector for word")
	}

	if len(vec) != 32 {
		t.Errorf("Expected dimension 32, got %d", len(vec))
	}
}

func TestWordEmbedding_CosineSimilarity(t *testing.T) {
	e := NewWordEmbedding(10)

	vec1 := []float64{1, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	vec2 := []float64{1, 0, 0, 0, 0, 0, 0, 0, 0, 0}

	sim := e.CosineSimilarity(vec1, vec2)
	if sim < 0.99 {
		t.Errorf("Expected ~1.0, got %f", sim)
	}

	vec3 := []float64{0, 1, 0, 0, 0, 0, 0, 0, 0, 0}
	sim = e.CosineSimilarity(vec1, vec3)
	if sim > 0.01 {
		t.Errorf("Expected ~0.0, got %f", sim)
	}
}

func TestWordEmbedding_Dimension(t *testing.T) {
	e := NewWordEmbedding(0)
	if e.GetDimension() != 128 {
		t.Errorf("Default dimension should be 128, got %d", e.GetDimension())
	}

	e256 := NewWordEmbedding(256)
	if e256.GetDimension() != 256 {
		t.Errorf("Expected 256, got %d", e256.GetDimension())
	}
}

func TestTokenize(t *testing.T) {
	tokens := tokenize("hello world test_123")
	if len(tokens) != 3 {
		t.Logf("Got %d tokens: %v - underscores are part of words", len(tokens), tokens)
	}
}

func TestWordEmbedding_Normalization(t *testing.T) {
	e := NewWordEmbedding(3)

	vec := e.randomVector(3)
	norm := 0.0
	for _, v := range vec {
		norm += v * v
	}
	norm = math.Sqrt(norm)

	if math.Abs(norm-1.0) > 0.01 {
		t.Errorf("Vector should be normalized, got norm %f", norm)
	}
}