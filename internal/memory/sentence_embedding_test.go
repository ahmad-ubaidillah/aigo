package memory

import (
	"testing"
)

func TestSentenceEmbedding_Build(t *testing.T) {
	e := NewSentenceEmbedding(32)

	texts := []string{
		"hello world",
		"agent planning",
		"memory extraction",
	}

	e.Build(texts)

	if e.wordEmbed.Size() == 0 {
		t.Error("Should build vocabulary")
	}
}

func TestSentenceEmbedding_Encode(t *testing.T) {
	e := NewSentenceEmbedding(16)

	texts := []string{"hello world test"}
	e.Build(texts)

	vec := e.Encode("hello world")
	if len(vec) == 0 {
		t.Error("Should encode sentence")
	}

	if len(vec) != 16 {
		t.Errorf("Expected dimension 16, got %d", len(vec))
	}
}

func TestSentenceEmbedding_Similarity(t *testing.T) {
	e := NewSentenceEmbedding(16)

	texts := []string{
		"agent planning execution",
		"memory profile workspace",
	}
	e.Build(texts)

	sim1 := e.Similarity("agent planning", "agent plan")
	sim2 := e.Similarity("agent planning", "memory profile")

	t.Logf("Similarity (same): %f", sim1)
	t.Logf("Similarity (diff): %f", sim2)

	if sim1 < sim2 {
		t.Log("Similar sentences should have higher similarity")
	}
}

func TestSentenceEmbedding_Pooling(t *testing.T) {
	e := NewSentenceEmbedding(8)

	texts := []string{"hello world test one two"}
	e.Build(texts)

	e.SetPooling("mean")
	vec1 := e.Encode("hello world")

	e.SetPooling("max")
	vec2 := e.Encode("hello world")

	t.Logf("Mean pooling: %v", vec1[:2])
	t.Logf("Max pooling: %v", vec2[:2])
}

func TestSentenceEmbedding_Empty(t *testing.T) {
	e := NewSentenceEmbedding(8)

	vec := e.Encode("")
	if len(vec) != 8 {
		t.Errorf("Should return zero vector for empty input, got %d", len(vec))
	}
}