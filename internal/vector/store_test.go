package vector

import (
	"testing"
)

func TestMemoryVectorStore_UpsertAndGet(t *testing.T) {
	t.Parallel()

	s := NewMemoryVectorStore()
	err := s.Upsert("doc1", []float64{0.1, 0.2}, map[string]any{"key": "val"})
	if err != nil {
		t.Fatal(err)
	}
	r, err := s.Get("doc1")
	if err != nil {
		t.Fatal(err)
	}
	if r.ID != "doc1" {
		t.Errorf("expected doc1, got %s", r.ID)
	}
}

func TestMemoryVectorStore_Query(t *testing.T) {
	t.Parallel()

	s := NewMemoryVectorStore()
	s.Upsert("doc1", []float64{1, 0}, nil)
	s.Upsert("doc2", []float64{0, 1}, nil)

	results, err := s.Query([]float64{1, 0}, 1, 0.0)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
	if results[0].ID != "doc1" {
		t.Errorf("expected doc1, got %s", results[0].ID)
	}
}

func TestMemoryVectorStore_QueryThreshold(t *testing.T) {
	t.Parallel()

	s := NewMemoryVectorStore()
	s.Upsert("doc1", []float64{1, 0}, nil)
	s.Upsert("doc2", []float64{0, 1}, nil)

	results, err := s.Query([]float64{1, 0}, 5, 0.99)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result with high threshold, got %d", len(results))
	}
}

func TestMemoryVectorStore_Delete(t *testing.T) {
	t.Parallel()

	s := NewMemoryVectorStore()
	s.Upsert("doc1", []float64{0.1}, nil)
	err := s.Delete("doc1")
	if err != nil {
		t.Fatal(err)
	}
	if s.Count() != 0 {
		t.Errorf("expected 0 after delete, got %d", s.Count())
	}
}

func TestMemoryVectorStore_DeleteNotFound(t *testing.T) {
	t.Parallel()

	s := NewMemoryVectorStore()
	err := s.Delete("nonexistent")
	if err == nil {
		t.Error("expected error")
	}
}

func TestMemoryVectorStore_GetNotFound(t *testing.T) {
	t.Parallel()

	s := NewMemoryVectorStore()
	_, err := s.Get("nonexistent")
	if err == nil {
		t.Error("expected error")
	}
}

func TestMemoryVectorStore_Count(t *testing.T) {
	t.Parallel()

	s := NewMemoryVectorStore()
	if s.Count() != 0 {
		t.Error("expected 0")
	}
	s.Upsert("d1", []float64{1}, nil)
	s.Upsert("d2", []float64{2}, nil)
	if s.Count() != 2 {
		t.Errorf("expected 2, got %d", s.Count())
	}
}

func TestMemoryVectorStore_UpsertNilMetadata(t *testing.T) {
	t.Parallel()

	s := NewMemoryVectorStore()
	err := s.Upsert("doc1", []float64{0.1, 0.2}, nil)
	if err != nil {
		t.Fatal(err)
	}
	r, err := s.Get("doc1")
	if err != nil {
		t.Fatal(err)
	}
	if r.Metadata == nil {
		t.Error("expected non-nil metadata")
	}
}

func TestMemoryVectorStore_QueryTopK(t *testing.T) {
	t.Parallel()

	s := NewMemoryVectorStore()
	for i := 0; i < 10; i++ {
		s.Upsert(string(rune('a'+i)), []float64{float64(i)}, nil)
	}
	results, err := s.Query([]float64{5}, 3, 0.0)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 3 {
		t.Errorf("expected 3 results, got %d", len(results))
	}
}

func TestCosineSimilarity_Identical(t *testing.T) {
	t.Parallel()

	sim := CosineSimilarity([]float64{1, 2, 3}, []float64{1, 2, 3})
	if sim < 0.999 {
		t.Errorf("expected ~1.0, got %f", sim)
	}
}

func TestCosineSimilarity_Orthogonal(t *testing.T) {
	t.Parallel()

	sim := CosineSimilarity([]float64{1, 0}, []float64{0, 1})
	if sim > 0.001 {
		t.Errorf("expected ~0.0, got %f", sim)
	}
}

func TestCosineSimilarity_Empty(t *testing.T) {
	t.Parallel()

	sim := CosineSimilarity([]float64{}, []float64{1, 2})
	if sim != 0 {
		t.Errorf("expected 0, got %f", sim)
	}
}

func TestCosineSimilarity_ZeroVector(t *testing.T) {
	t.Parallel()

	sim := CosineSimilarity([]float64{0, 0}, []float64{1, 2})
	if sim != 0 {
		t.Errorf("expected 0, got %f", sim)
	}
}

func TestCosineSimilarity_DifferentLengths(t *testing.T) {
	t.Parallel()

	sim := CosineSimilarity([]float64{1, 2}, []float64{1, 2, 3})
	if sim <= 0 {
		t.Errorf("expected positive similarity, got %f", sim)
	}
}
