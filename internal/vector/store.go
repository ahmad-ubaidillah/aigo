package vector

import (
	"fmt"
	"math"
	"sort"
	"sync"
)

type VectorStore interface {
	Upsert(id string, embedding []float64, metadata map[string]any) error
	Query(embedding []float64, n int, threshold float64) ([]Result, error)
	Delete(id string) error
	Get(id string) (*Result, error)
	Count() int
}

type Result struct {
	ID        string
	Score     float64
	Embedding []float64
	Metadata  map[string]any
}

type MemoryVectorStore struct {
	vectors map[string]*Result
	mu      sync.RWMutex
}

func NewMemoryVectorStore() *MemoryVectorStore {
	return &MemoryVectorStore{
		vectors: make(map[string]*Result),
	}
}

func (s *MemoryVectorStore) Upsert(id string, embedding []float64, metadata map[string]any) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if metadata == nil {
		metadata = make(map[string]any)
	}
	s.vectors[id] = &Result{
		ID:        id,
		Embedding: embedding,
		Metadata:  metadata,
	}
	return nil
}

func (s *MemoryVectorStore) Query(embedding []float64, n int, threshold float64) ([]Result, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []Result
	for _, v := range s.vectors {
		sim := cosineSim(embedding, v.Embedding)
		if sim >= threshold {
			r := *v
			r.Score = sim
			results = append(results, r)
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if n > 0 && len(results) > n {
		results = results[:n]
	}
	return results, nil
}

func (s *MemoryVectorStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.vectors[id]; !ok {
		return fmt.Errorf("vector %s not found", id)
	}
	delete(s.vectors, id)
	return nil
}

func (s *MemoryVectorStore) Get(id string) (*Result, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.vectors[id]
	if !ok {
		return nil, fmt.Errorf("vector %s not found", id)
	}
	r := *v
	return &r, nil
}

func (s *MemoryVectorStore) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.vectors)
}

func cosineSim(a, b []float64) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	minLen := len(a)
	if len(b) < minLen {
		minLen = len(b)
	}
	var dot, normA, normB float64
	for i := 0; i < minLen; i++ {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}

func CosineSimilarity(a, b []float64) float64 {
	return cosineSim(a, b)
}
