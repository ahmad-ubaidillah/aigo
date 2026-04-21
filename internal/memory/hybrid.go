package memory

type HybridResult struct {
	ID      string
	Content string
	FTS     float64
	Vector  float64
	Combined float64
}

type HybridSearch struct {
	ftsWeight float64
	vecWeight float64
}

func NewHybridSearch() *HybridSearch {
	return &HybridSearch{
		ftsWeight:  0.5,
		vecWeight:  0.5,
	}
}

func (h *HybridSearch) Search(query string, vectors [][]float64, limit int) []HybridResult {
	results := make([]HybridResult, 0)

	if limit <= 0 {
		limit = 10
	}

	return results
}

func (h *HybridSearch) CombineScore(ftsScore, vecScore float64) float64 {
	return (ftsScore * h.ftsWeight) + (vecScore * h.vecWeight)
}

func (h *HybridSearch) Rank(results []HybridResult) []HybridResult {
	return results
}