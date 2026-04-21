package analytics

import (
	"testing"
)

func TestAnalytics_TrackToken(t *testing.T) {
	a := NewAnalytics()
	a.TrackToken(100)

	if a.Tokens != 100 {
		t.Error("Token count mismatch")
	}
}

func TestAnalytics_Cost(t *testing.T) {
	a := NewAnalytics()
	a.TrackCost(0.05)

	if a.Cost != 0.05 {
		t.Error("Cost mismatch")
	}
}

func TestAnalytics_ROI(t *testing.T) {
	a := NewAnalytics()
	roi := a.CalcROI()
	t.Logf("ROI: %f", roi)
}

func TestAnalytics_PlanningAccuracy(t *testing.T) {
	a := NewAnalytics()
	a.RecordPlanning(True)

	acc := a.PlanningAccuracy()
	if acc < 0 {
		t.Error("Accuracy should be positive")
	}
}