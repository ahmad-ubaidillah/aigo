package distill

import (
	"strings"
	"testing"
)

func BenchmarkClassify(b *testing.B) {
	c := NewClassifier()
	text := strings.Repeat("2024-01-01T00:00:00Z [INFO] server started\n", 100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Classify(text)
	}
}

func BenchmarkScore(b *testing.B) {
	s := NewScorer()
	text := "error: compilation failed"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Score(text, ContentBuildOutput)
	}
}

func BenchmarkCollapse(b *testing.B) {
	c := NewCollapse()
	text := strings.Repeat("repeated line\n", 500)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Compress(text)
	}
}

func BenchmarkPipeline(b *testing.B) {
	p := NewPipeline()
	text := strings.Repeat("=== RUN   TestFoo\n--- PASS: TestFoo (0.00s)\n", 200)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.Process(text)
	}
}
