package memory

import (
	"math"
	"strings"
)

type Embedding struct {
	dim int
}

func NewEmbedding() *Embedding {
	return &Embedding{
		dim: 128,
	}
}

func (e *Embedding) Encode(text string) []float64 {
	words := strings.Fields(strings.ToLower(text))
	vec := make([]float64, e.dim)

	for i, word := range words {
		hash := simpleHash(word)
		vec[i%e.dim] += float64(hash)
	}

	for i := range vec {
		vec[i] = vec[i] / float64(len(words)+1)
	}

	return vec
}

func (e *Embedding) Normalize(vec []float64) []float64 {
	if len(vec) == 0 {
		return vec
	}

	var sum float64
	for _, v := range vec {
		sum += v * v
	}
	magnitude := math.Sqrt(sum)

	if magnitude == 0 {
		return vec
	}

	result := make([]float64, len(vec))
	for i, v := range vec {
		result[i] = v / magnitude
	}

	return result
}

func simpleHash(s string) int {
	hash := 0
	for i, c := range s {
		hash += int(c) * (i + 1)
	}
	return hash
}