package memory

import (
	"math"
	"sort"
)

type Similarity struct {
	sentenceEmbed *SentenceEmbedding
	threshold  float64
}

func NewSimilarity() *Similarity {
	return &Similarity{
		threshold: 0.7,
	}
}

func NewSimilarityWithEmbedding(dim int) *Similarity {
	return &Similarity{
		sentenceEmbed: NewSentenceEmbedding(dim),
		threshold:   0.7,
	}
}

type TopResult struct {
	Index   int
	Score  float64
}

func (s *Similarity) Cosine(vec1, vec2 []float64) float64 {
	if len(vec1) != len(vec2) || len(vec1) == 0 {
		return 0
	}

	var dot, norm1, norm2 float64
	for i := range vec1 {
		dot += vec1[i] * vec2[i]
		norm1 += vec1[i] * vec1[i]
		norm2 += vec2[i] * vec2[i]
	}

	if norm1 == 0 || norm2 == 0 {
		return 0
	}

	return dot / (math.Sqrt(norm1) * math.Sqrt(norm2))
}

func (s *Similarity) TopK(queries [][]float64, target []float64, k int) []TopResult {
	results := make([]TopResult, 0, len(queries))

	for i, query := range queries {
		score := s.Cosine(query, target)
		results = append(results, TopResult{Index: i, Score: score})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if k > 0 && k < len(results) {
		results = results[:k]
	}

	return results
}

func (s *Similarity) TextSimilarity(text1, text2 string) float64 {
	if s.sentenceEmbed == nil {
		return 0
	}

	return s.sentenceEmbed.Similarity(text1, text2)
}

func (s *Similarity) TextSearch(query string, texts []string, topK int) []TopResult {
	if s.sentenceEmbed == nil {
		return nil
	}

	results := make([]TopResult, 0, len(texts))

	for i, text := range texts {
		score := s.sentenceEmbed.Similarity(query, text)
		results = append(results, TopResult{Index: i, Score: score})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if topK > 0 && topK < len(results) {
		results = results[:topK]
	}

	return results
}

func (s *Similarity) FilterByThreshold(results []TopResult) []TopResult {
	filtered := make([]TopResult, 0)
	for _, r := range results {
		if r.Score >= s.threshold {
			filtered = append(filtered, r)
		}
	}
	return filtered
}

func (s *Similarity) SetThreshold(t float64) {
	if t >= 0 && t <= 1 {
		s.threshold = t
	}
}