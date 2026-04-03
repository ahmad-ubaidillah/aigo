package vectordb

import (
	"fmt"
	"math"
	"testing"
)

type mockEmbedding struct{}

func (m *mockEmbedding) Embed(text string) ([]float64, error) {
	vec := make([]float64, 4)
	for i, c := range text {
		if i < 4 {
			vec[i] = float64(c)
		}
	}
	return vec, nil
}

func TestChromaClient_AddDocument(t *testing.T) {
	t.Parallel()

	c := NewChromaClient("test", &mockEmbedding{})
	err := c.AddDocument("doc1", "hello world", map[string]string{"key": "val"})
	if err != nil {
		t.Fatal(err)
	}
	if c.Count() != 1 {
		t.Errorf("expected 1, got %d", c.Count())
	}
}

func TestChromaClient_AddDocument_NoEmbedding(t *testing.T) {
	t.Parallel()

	c := NewChromaClient("test", nil)
	err := c.AddDocument("doc1", "hello", nil)
	if err == nil {
		t.Error("expected error")
	}
}

func TestChromaClient_DeleteDocument(t *testing.T) {
	t.Parallel()

	c := NewChromaClient("test", &mockEmbedding{})
	c.AddDocument("doc1", "hello", nil)
	err := c.DeleteDocument("doc1")
	if err != nil {
		t.Fatal(err)
	}
	if c.Count() != 0 {
		t.Errorf("expected 0, got %d", c.Count())
	}
}

func TestChromaClient_DeleteDocument_NotFound(t *testing.T) {
	t.Parallel()

	c := NewChromaClient("test", &mockEmbedding{})
	err := c.DeleteDocument("nonexistent")
	if err == nil {
		t.Error("expected error")
	}
}

func TestChromaClient_Search(t *testing.T) {
	t.Parallel()

	c := NewChromaClient("test", &mockEmbedding{})
	c.AddDocument("doc1", "hello world", nil)
	c.AddDocument("doc2", "goodbye world", nil)

	results := c.Search("hello", 5, 0.0)
	if len(results) == 0 {
		t.Error("expected results")
	}
}

func TestChromaClient_Search_NoEmbedding(t *testing.T) {
	t.Parallel()

	c := NewChromaClient("test", nil)
	results := c.Search("hello", 5, 0.0)
	if results != nil {
		t.Error("expected nil results")
	}
}

func TestChromaClient_Search_Threshold(t *testing.T) {
	t.Parallel()

	c := NewChromaClient("test", &mockEmbedding{})
	c.AddDocument("doc1", "\x00\x00\x00\x00", nil)

	results := c.Search("\xff\xff\xff\xff", 5, 0.99)
	if len(results) != 0 {
		t.Errorf("expected 0 results with high threshold, got %d", len(results))
	}
}

func TestCosineSimilarity_Identical(t *testing.T) {
	t.Parallel()

	a := []float64{1, 2, 3}
	b := []float64{1, 2, 3}
	sim := cosineSimilarity(a, b)
	if math.Abs(sim-1.0) > 0.001 {
		t.Errorf("expected 1.0, got %f", sim)
	}
}

func TestCosineSimilarity_Orthogonal(t *testing.T) {
	t.Parallel()

	a := []float64{1, 0}
	b := []float64{0, 1}
	sim := cosineSimilarity(a, b)
	if math.Abs(sim) > 0.001 {
		t.Errorf("expected 0.0, got %f", sim)
	}
}

func TestCosineSimilarity_Empty(t *testing.T) {
	t.Parallel()

	sim := cosineSimilarity([]float64{}, []float64{1, 2})
	if sim != 0 {
		t.Errorf("expected 0, got %f", sim)
	}
}

func TestCosineSimilarity_ZeroVector(t *testing.T) {
	t.Parallel()

	a := []float64{0, 0}
	b := []float64{1, 2}
	sim := cosineSimilarity(a, b)
	if sim != 0 {
		t.Errorf("expected 0, got %f", sim)
	}
}

func TestChromaClient_AddDocument_EmbeddingError(t *testing.T) {
	t.Parallel()

	emb := &mockEmbeddingError{}
	c := NewChromaClient("test", emb)
	err := c.AddDocument("doc1", "hello", nil)
	if err == nil {
		t.Error("expected embedding error")
	}
}

func TestChromaClient_Search_EmptyDocs(t *testing.T) {
	t.Parallel()

	c := NewChromaClient("test", &mockEmbedding{})
	results := c.Search("query", 5, 0.0)
	if results != nil {
		t.Error("expected nil results for empty docs")
	}
}

func TestChromaClient_Search_TopK(t *testing.T) {
	t.Parallel()

	c := NewChromaClient("test", &mockEmbedding{})
	c.AddDocument("doc1", "hello", nil)
	c.AddDocument("doc2", "world", nil)
	c.AddDocument("doc3", "test", nil)

	results := c.Search("query", 1, 0.0)
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
}

type mockEmbeddingError struct{}

func (m *mockEmbeddingError) Embed(text string) ([]float64, error) {
	return nil, fmt.Errorf("embedding failed")
}
