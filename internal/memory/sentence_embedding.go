package memory

import (
	"math"
	"strings"
)

type SentenceEmbedding struct {
	wordEmbed *WordEmbedding
	pooling  string
}

func NewSentenceEmbedding(dim int) *SentenceEmbedding {
	return &SentenceEmbedding{
		wordEmbed: NewWordEmbedding(dim),
		pooling:  "mean",
	}
}

func (e *SentenceEmbedding) Build(texts []string) {
	e.wordEmbed.BuildVocabulary(texts)
}

func (e *SentenceEmbedding) Encode(sentence string) []float64 {
	words := tokenize(strings.ToLower(sentence))
	if len(words) == 0 {
		return make([]float64, e.wordEmbed.GetDimension())
	}

	vectors := make([][]float64, 0)
	for _, word := range words {
		if vec, ok := e.wordEmbed.GetVector(word); ok {
			vectors = append(vectors, vec)
		}
	}

	if len(vectors) == 0 {
		return make([]float64, e.wordEmbed.GetDimension())
	}

	switch e.pooling {
	case "mean":
		return e.meanPooling(vectors)
	case "max":
		return e.maxPooling(vectors)
	default:
		return e.meanPooling(vectors)
	}
}

func (e *SentenceEmbedding) meanPooling(vectors [][]float64) []float64 {
	dim := len(vectors[0])
	result := make([]float64, dim)

	for _, vec := range vectors {
		for i := 0; i < dim; i++ {
			result[i] += vec[i]
		}
	}

	for i := 0; i < dim; i++ {
		result[i] /= float64(len(vectors))
	}

	return e.normalize(result)
}

func (e *SentenceEmbedding) maxPooling(vectors [][]float64) []float64 {
	dim := len(vectors[0])
	result := make([]float64, dim)

	for i := 0; i < dim; i++ {
		maxVal := vectors[0][i]
		for j := 1; j < len(vectors); j++ {
			if vectors[j][i] > maxVal {
				maxVal = vectors[j][i]
			}
		}
		result[i] = maxVal
	}

	return e.normalize(result)
}

func (e *SentenceEmbedding) normalize(vec []float64) []float64 {
	norm := 0.0
	for _, v := range vec {
		norm += v * v
	}
	norm = math.Sqrt(norm)

	if norm == 0 {
		return vec
	}

	result := make([]float64, len(vec))
	for i, v := range vec {
		result[i] = v / norm
	}

	return result
}

func (e *SentenceEmbedding) Similarity(s1, s2 string) float64 {
	vec1 := e.Encode(s1)
	vec2 := e.Encode(s2)

	return e.wordEmbed.CosineSimilarity(vec1, vec2)
}

func (e *SentenceEmbedding) SetPooling(strategy string) {
	if strategy == "mean" || strategy == "max" {
		e.pooling = strategy
	}
}