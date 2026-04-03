package embedding

import (
	"testing"
)

func TestLocalEmbedder_Embed(t *testing.T) {
	t.Parallel()

	e := NewLocalEmbedder(128)
	vec, err := e.Embed("hello world")
	if err != nil {
		t.Fatal(err)
	}
	if len(vec) != 128 {
		t.Errorf("expected 128 dims, got %d", len(vec))
	}
}

func TestLocalEmbedder_EmbedBatch(t *testing.T) {
	t.Parallel()

	e := NewLocalEmbedder(64)
	vecs, err := e.EmbedBatch([]string{"hello", "world"})
	if err != nil {
		t.Fatal(err)
	}
	if len(vecs) != 2 {
		t.Errorf("expected 2 vectors, got %d", len(vecs))
	}
}

func TestLocalEmbedder_DimZero(t *testing.T) {
	t.Parallel()

	e := NewLocalEmbedder(0)
	vec, err := e.Embed("test")
	if err != nil {
		t.Fatal(err)
	}
	if len(vec) != 128 {
		t.Errorf("expected default 128 dims, got %d", len(vec))
	}
}

func TestLocalEmbedder_DimNegative(t *testing.T) {
	t.Parallel()

	e := NewLocalEmbedder(-5)
	vec, err := e.Embed("test")
	if err != nil {
		t.Fatal(err)
	}
	if len(vec) != 128 {
		t.Errorf("expected default 128 dims, got %d", len(vec))
	}
}

func TestEmbeddingCache(t *testing.T) {
	t.Parallel()

	c := NewEmbeddingCache()
	vec := []float64{0.1, 0.2, 0.3}
	c.Set("hello", vec)

	got, ok := c.Get("hello")
	if !ok {
		t.Fatal("expected cache hit")
	}
	if len(got) != 3 {
		t.Errorf("expected 3 elements, got %d", len(got))
	}

	_, ok = c.Get("missing")
	if ok {
		t.Error("expected cache miss")
	}
}

func TestCachedEmbedder(t *testing.T) {
	t.Parallel()

	e := NewLocalEmbedder(128)
	ce := NewCachedEmbedder(e)

	vec1, err := ce.Embed("test")
	if err != nil {
		t.Fatal(err)
	}
	vec2, err := ce.Embed("test")
	if err != nil {
		t.Fatal(err)
	}
	if len(vec1) != len(vec2) {
		t.Error("expected same length vectors")
	}
}

func TestCachedEmbedder_Batch(t *testing.T) {
	t.Parallel()

	e := NewLocalEmbedder(128)
	ce := NewCachedEmbedder(e)

	vecs, err := ce.EmbedBatch([]string{"a", "b", "c"})
	if err != nil {
		t.Fatal(err)
	}
	if len(vecs) != 3 {
		t.Errorf("expected 3 vectors, got %d", len(vecs))
	}
}

func TestOpenAIEmbedder_EmptyModel(t *testing.T) {
	t.Parallel()

	e := NewOpenAIEmbedder("test-key", "")
	if e.model != "text-embedding-ada-002" {
		t.Errorf("expected default model, got %s", e.model)
	}
}

func TestOpenAIEmbedder_CustomModel(t *testing.T) {
	t.Parallel()

	e := NewOpenAIEmbedder("test-key", "custom-model")
	if e.model != "custom-model" {
		t.Errorf("expected custom-model, got %s", e.model)
	}
}

func TestOpenAIEmbedder_CacheSize(t *testing.T) {
	t.Parallel()

	e := NewOpenAIEmbedder("test-key", "")
	if e.CacheSize() != 0 {
		t.Errorf("expected 0 cache size, got %d", e.CacheSize())
	}
}

func TestOpenAIEmbedder_ClearCache(t *testing.T) {
	t.Parallel()

	e := NewOpenAIEmbedder("test-key", "")
	e.ClearCache()
	if e.CacheSize() != 0 {
		t.Error("expected 0 after clear")
	}
}
