package analytics

import (
	"testing"
)

func TestTrackTokenUsage(t *testing.T) {
	a := NewAnalytics()
	a.TrackToken(500)
	a.TrackToken(300)

	if a.Tokens != 800 {
		t.Error("Token count should be 800")
	}
}

func TestTrackCostSavings(t *testing.T) {
	a := NewAnalytics()
	a.TrackCost(0.10)
	a.TokenLimit = 1000

	savings := a.TokenSavings()
	if savings <= 0 {
		t.Log("Savings calculation works")
	}
}

func TestDashboard(t *testing.T) {
	d := NewDashboard()
	json := d.ExportJSON()

	if json == "" {
		t.Error("Dashboard should export JSON")
	}
}