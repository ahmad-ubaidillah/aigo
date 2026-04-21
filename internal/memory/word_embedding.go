package memory

import (
	"math"
	"math/rand"
)

type WordEmbedding struct {
	dimension int
 vocab    map[string][]float64
	initialized bool
}

func NewWordEmbedding(dim int) *WordEmbedding {
	if dim <= 0 {
		dim = 128
	}
	return &WordEmbedding{
		dimension: dim,
		vocab:    make(map[string][]float64),
		initialized: false,
	}
}

func (e *WordEmbedding) BuildVocabulary(texts []string) {
	e.vocab = make(map[string][]float64)

	for _, text := range texts {
		words := tokenize(text)
		for _, word := range words {
			if _, exists := e.vocab[word]; !exists {
				e.vocab[word] = e.randomVector(e.dimension)
			}
		}
	}

	e.initialized = len(e.vocab) > 0
}

func (e *WordEmbedding) GetVector(word string) ([]float64, bool) {
	vec, ok := e.vocab[word]
	return vec, ok
}

func (e *WordEmbedding) GetDimension() int {
	return e.dimension
}

func (e *WordEmbedding) Size() int {
	return len(e.vocab)
}

func (e *WordEmbedding) randomVector(dim int) []float64 {
	vec := make([]float64, dim)
	norm := 0.0

	for i := 0; i < dim; i++ {
		vec[i] = rand.Float64()*2 - 1
		norm += vec[i] * vec[i]
	}

	norm = math.Sqrt(norm)
	for i := 0; i < dim; i++ {
		vec[i] /= norm
	}

	return vec
}

func (e *WordEmbedding) CosineSimilarity(vec1, vec2 []float64) float64 {
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

	return dot / (math.Sqrt(norm1) * math.Sqrt(norm2))
}

func tokenize(text string) []string {
	var tokens []string
	var word []rune

	for _, r := range text {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			word = append(word, r)
		} else if len(word) > 0 {
			tokens = append(tokens, string(word))
			word = nil
		}
	}

	if len(word) > 0 {
		tokens = append(tokens, string(word))
	}

	return tokens
}