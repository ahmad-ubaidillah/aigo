package tools

import (
	"testing"
)

func TestCodeReview_Analyze(t *testing.T) {
	cr := NewCodeReviewer()
	issues := cr.Analyze("func test() { /* code */ }")

	if issues == nil {
		t.Fatal("Issues should be returned")
	}
}

func TestTestGen_Generate(t *testing.T) {
	tg := NewTestGenerator()
	tests := tg.Generate("func Add(a, b int) int { return a + b }")

	if tests == "" {
		t.Error("Tests should be generated")
	}
}

func TestCodeExplain_Explain(t *testing.T) {
	ce := NewCodeExplainer()
	explanation := ce.Explain("fmt.Println")

	if explanation == "" {
		t.Error("Explanation should be provided")
	}
}

func TestDiagnostics_Run(t *testing.T) {
	d := NewDiagnostics()
	results := d.Run("main.go")

	if results == nil {
		t.Fatal("Diagnostics should run")
	}
}