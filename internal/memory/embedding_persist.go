package memory

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type EmbeddingStore struct {
	version string
	dim    int
	path   string
}

type EmbeddingData struct {
	Version  string               `json:"version"`
	Dim     int                  `json:"dimension"`
	Vocab   map[string][]float64   `json:"vocabulary"`
}

func NewEmbeddingStore(path string, dim int) *EmbeddingStore {
	return &EmbeddingStore{
		version: "1.0",
		dim:    dim,
		path:   path,
	}
}

func (s *EmbeddingStore) Save(e *WordEmbedding) error {
	data := EmbeddingData{
		Version: s.version,
		Dim:     e.dimension,
		Vocab:   e.vocab,
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}

	err = os.WriteFile(s.path, jsonData, 0644)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func (s *EmbeddingStore) Load() (*WordEmbedding, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var embeddingData EmbeddingData
	err = json.Unmarshal(data, &embeddingData)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal: %w", err)
	}

	e := NewWordEmbedding(embeddingData.Dim)
	e.vocab = embeddingData.Vocab
	e.initialized = len(e.vocab) > 0

	return e, nil
}

func (s *EmbeddingStore) Exists() bool {
	_, err := os.Stat(s.path)
	return err == nil
}

func (s *EmbeddingStore) Delete() error {
	return os.Remove(s.path)
}

type EmbeddingIndex struct {
	dimension int
	vectors  [][]float64
	metas    []string
}

func NewEmbeddingIndex(dim int) *EmbeddingIndex {
	return &EmbeddingIndex{
		dimension: dim,
		vectors:  make([][]float64, 0),
		metas:    make([]string, 0),
	}
}

func (idx *EmbeddingIndex) Add(vec []float64, meta string) {
	idx.vectors = append(idx.vectors, vec)
	idx.metas = append(idx.metas, meta)
}

func (idx *EmbeddingIndex) Search(query []float64, topK int) []EmbeddingSearchResult {
	results := make([]EmbeddingSearchResult, 0)

	for i, vec := range idx.vectors {
		sim := cosineSimilarity(query, vec)
		results = append(results, EmbeddingSearchResult{
			Index:     i,
			Score:    sim,
			Metadata: idx.metas[i],
		})
	}

	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].Score > results[i].Score {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	if topK > 0 && topK < len(results) {
		results = results[:topK]
	}

	return results
}

type EmbeddingSearchResult struct {
	Index     int
	Score    float64
	Metadata string
}

func cosineSimilarity(vec1, vec2 []float64) float64 {
	if len(vec1) != len(vec2) {
		return 0
	}

	dot := 0.0
	norm1 := 0.0
	norm2 := 0.0

	for i := 0; i < len(vec1); i++ {
		dot += vec1[i] * vec2[i]
		norm1 += vec1[i] * vec1[i]
		norm2 += vec2[i] * vec2[i]
	}

	if norm1 == 0 || norm2 == 0 {
		return 0
	}

	return dot / (sqrt(norm1) * sqrt(norm2))
}

func sqrt(x float64) float64 {
	if x <= 0 {
		return 0
	}
	return _sqrt(x)
}

func _sqrt(x float64) float64 {
	r := x
	for i := 0; i < 20; i++ {
		r = (r + x/r) / 2
	}
	return r
}

func ParseVector(s string) ([]float64, error) {
	s = strings.TrimSpace(s)
	s = strings.Trim(s, "[]")

	parts := strings.Split(s, ",")
	vec := make([]float64, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		v, err := strconv.ParseFloat(part, 64)
		if err != nil {
			return nil, err
		}
		vec = append(vec, v)
	}

	return vec, nil
}