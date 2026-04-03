package vectordb

import (
	"fmt"
	"math"
	"sort"
	"sync"
)

type Embedding interface {
	Embed(text string) ([]float64, error)
}

type VectorDocument struct {
	ID       string
	Content  string
	Vector   []float64
	Metadata map[string]string
	Score    float64
}

type ChromaClient struct {
	collection string
	documents  []VectorDocument
	mu         sync.RWMutex
	embedding  Embedding
}

func NewChromaClient(collection string, embedding Embedding) *ChromaClient {
	return &ChromaClient{
		collection: collection,
		documents:  make([]VectorDocument, 0),
		embedding:  embedding,
	}
}

func (c *ChromaClient) AddDocument(id, content string, metadata map[string]string) error {
	if c.embedding == nil {
		return fmt.Errorf("no embedding provider configured")
	}
	vec, err := c.embedding.Embed(content)
	if err != nil {
		return fmt.Errorf("embed content: %w", err)
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.documents = append(c.documents, VectorDocument{
		ID:       id,
		Content:  content,
		Vector:   vec,
		Metadata: metadata,
	})
	return nil
}

func (c *ChromaClient) Search(query string, topK int, threshold float64) []VectorDocument {
	if c.embedding == nil || len(c.documents) == 0 {
		return nil
	}
	qvec, err := c.embedding.Embed(query)
	if err != nil {
		return nil
	}
	c.mu.RLock()
	defer c.mu.RUnlock()

	results := make([]VectorDocument, 0, len(c.documents))
	for _, doc := range c.documents {
		sim := cosineSimilarity(qvec, doc.Vector)
		if sim >= threshold {
			d := doc
			d.Score = sim
			results = append(results, d)
		}
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})
	if topK > 0 && len(results) > topK {
		results = results[:topK]
	}
	return results
}

func (c *ChromaClient) DeleteDocument(id string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	for i, doc := range c.documents {
		if doc.ID == id {
			c.documents = append(c.documents[:i], c.documents[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("document %s not found", id)
}

func (c *ChromaClient) Count() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.documents)
}

func cosineSimilarity(a, b []float64) float64 {
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
