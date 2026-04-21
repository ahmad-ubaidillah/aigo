package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestDashboardServer_Routes(t *testing.T) {
	s := NewDashboardServer(":0")
	s.setupRoutes()

	tests := []struct {
		path   string
		method string
	}{
		{"/api/analytics/token", "GET"},
		{"/api/analytics/cost", "GET"},
		{"/api/analytics/planning", "GET"},
		{"/api/analytics/memory", "GET"},
		{"/api/health", "GET"},
	}

	for _, tt := range tests {
		req := httptest.NewRequest(tt.method, tt.path, nil)
		w := httptest.NewRecorder()

		s.mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200 for %s, got %d", tt.path, w.Code)
		}
	}
}

func TestDashboardServer_Health(t *testing.T) {
	s := NewDashboardServer(":0")
	s.setupRoutes()

	req := httptest.NewRequest("GET", "/api/health", nil)
	w := httptest.NewRecorder()

	s.mux.ServeHTTP(w, req)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if response["status"] != "healthy" {
		t.Errorf("Expected healthy status, got %v", response["status"])
	}
}

func TestTimeWindowedMetrics_Add(t *testing.T) {
	m := NewTimeWindowedMetrics(60 * time.Second)

	m.Add(10.0)
	m.Add(20.0)
	m.Add(30.0)

	if len(m.metrics) != 3 {
		t.Errorf("Expected 3 metrics, got %d", len(m.metrics))
	}
}

func TestTimeWindowedMetrics_Average(t *testing.T) {
	m := NewTimeWindowedMetrics(60 * time.Second)

	m.Add(10.0)
	m.Add(20.0)
	m.Add(30.0)

	avg := m.Average()
	if avg != 20.0 {
		t.Errorf("Expected 20.0, got %f", avg)
	}
}

func TestTimeWindowedMetrics_Sum(t *testing.T) {
	m := NewTimeWindowedMetrics(60 * time.Second)

	m.Add(10.0)
	m.Add(20.0)
	m.Add(30.0)

	sum := m.Sum()
	if sum != 60.0 {
		t.Errorf("Expected 60.0, got %f", sum)
	}
}

func TestTimeWindowedMetrics_Max(t *testing.T) {
	m := NewTimeWindowedMetrics(60 * time.Second)

	m.Add(10.0)
	m.Add(50.0)
	m.Add(30.0)

	max := m.Max()
	if max != 50.0 {
		t.Errorf("Expected 50.0, got %f", max)
	}
}

func TestCalculateTrend(t *testing.T) {
	points := []float64{10, 15, 20, 25}
	trend := CalculateTrend(points)
	if trend != "increasing" {
		t.Errorf("Expected increasing, got %s", trend)
	}

	points = []float64{25, 20, 15, 10}
	trend = CalculateTrend(points)
	if trend != "decreasing" {
		t.Errorf("Expected decreasing, got %s", trend)
	}

	points = []float64{10, 11, 10, 11}
	trend = CalculateTrend(points)
	if trend != "stable" {
		t.Errorf("Expected stable, got %s", trend)
	}
}