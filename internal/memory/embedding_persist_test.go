package memory

import (
	"os"
	"testing"
)

func TestEmbeddingStore_SaveAndLoad(t *testing.T) {
	tmpFile := "/tmp/test_embeddings.json"
	os.Remove(tmpFile)

	store := NewEmbeddingStore(tmpFile, 32)

	e := NewWordEmbedding(32)
	e.BuildVocabulary([]string{"hello world", "test code"})

	err := store.Save(e)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	if !store.Exists() {
		t.Error("Should exist after save")
	}

	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.Size() == 0 {
		t.Error("Should load vocabulary")
	}

	os.Remove(tmpFile)
}

func TestEmbeddingStore_LoadNonExistent(t *testing.T) {
	store := NewEmbeddingStore("/tmp/nonexistent.json", 32)

	_, err := store.Load()
	if err == nil {
		t.Error("Should fail for non-existent file")
	}
}

func TestEmbeddingIndex_Search(t *testing.T) {
	idx := NewEmbeddingIndex(3)

	idx.Add([]float64{1, 0, 0}, "doc1")
	idx.Add([]float64{0, 1, 0}, "doc2")
	idx.Add([]float64{0, 0, 1}, "doc3")

	results := idx.Search([]float64{1, 0, 0}, 2)
	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	if results[0].Metadata != "doc1" {
		t.Logf("First result: %s", results[0].Metadata)
	}
}

func TestEmbeddingIndex_Empty(t *testing.T) {
	idx := NewEmbeddingIndex(3)

	results := idx.Search([]float64{1, 0, 0}, 10)
	if len(results) != 0 {
		t.Error("Should return empty for empty index")
	}
}

func TestParseVector(t *testing.T) {
	vec, err := ParseVector("[1.0, 2.0, 3.0]")
	if err != nil {
		t.Fatalf("ParseVector failed: %v", err)
	}

	if len(vec) != 3 {
		t.Errorf("Expected 3 elements, got %d", len(vec))
	}

	if vec[0] != 1.0 {
		t.Errorf("Expected 1.0, got %f", vec[0])
	}
}