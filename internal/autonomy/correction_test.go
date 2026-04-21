package autonomy

import (
	"testing"
)

func TestSelfCorrection_Detect(t *testing.T) {
	sc := NewSelfCorrector()
	if sc == nil {
		t.Fatal("SelfCorrector is nil")
	}
}

func TestCorrection_Analyze(t *testing.T) {
	sc := NewSelfCorrector()
	fix := sc.AnalyzeError("undefined error")

	if fix == "" {
		t.Error("Analyze should return fix")
	}
}

func TestCorrection_Retry(t *testing.T) {
	sc := NewSelfCorrector()
	shouldRetry := sc.ShouldRetry(1)

	if !shouldRetry {
		t.Error("Should retry on first attempt")
	}
}

func TestCorrection_AddPattern(t *testing.T) {
	sc := NewSelfCorrector()
	sc.AddPattern("custom error", "custom fix")

	fix := sc.AnalyzeError("custom error")
	if fix != "custom fix" {
		t.Errorf("Expected 'custom fix', got '%s'", fix)
	}
}

func TestCorrection_GetPatterns(t *testing.T) {
	sc := NewSelfCorrector()
	patterns := sc.GetPatterns()

	if len(patterns) == 0 {
		t.Error("Should have default patterns")
	}
}

func TestCorrection_RetryWithFix(t *testing.T) {
	sc := NewSelfCorrector()
	result := sc.RetryWithFix("syntax error")

	if result == "" {
		t.Error("RetryWithFix should return result")
	}
}