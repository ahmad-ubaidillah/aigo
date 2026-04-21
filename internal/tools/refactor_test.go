package tools

import (
	"testing"
)

func TestRefactor_Analyze(t *testing.T) {
	r := NewRefactorAnalyzer()
	deps := r.AnalyzeDependencies("file1.go", []string{"file2.go", "file3.go"})

	if deps == nil {
		t.Fatal("Dependencies should be analyzed")
	}
}

func TestRefactor_AtomicEdit(t *testing.T) {
	r := NewRefactorAnalyzer()
	results := r.AtomicEdit(map[string]string{
		"file1.go": "content1",
		"file2.go": "content2",
	})

	if len(results) != 2 {
		t.Error("Should have 2 results")
	}
}