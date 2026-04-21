package api

import (
	"encoding/json"
	"net/http"
	"time"
)

type DashboardServer struct {
	addr      string
	analytics *AnalyticsHandler
	mux      *http.ServeMux
}

type AnalyticsHandler struct {
	tokenCost   float64
	planningOK  int
	planningFail int
	memoryHits int
	memoryMiss int
}

func NewAnalyticsHandler() *AnalyticsHandler {
	return &AnalyticsHandler{
		tokenCost: 0,
	}
}

func NewDashboardServer(addr string) *DashboardServer {
	return &DashboardServer{
		addr:      addr,
		analytics: NewAnalyticsHandler(),
		mux:      http.NewServeMux(),
	}
}

func (s *DashboardServer) Start() error {
	s.setupRoutes()
	return http.ListenAndServe(s.addr, s.mux)
}

func (s *DashboardServer) setupRoutes() {
	s.mux.HandleFunc("/api/analytics/token", s.handleTokenAnalytics)
	s.mux.HandleFunc("/api/analytics/cost", s.handleCostAnalytics)
	s.mux.HandleFunc("/api/analytics/planning", s.handlePlanningAnalytics)
	s.mux.HandleFunc("/api/analytics/memory", s.handleMemoryAnalytics)
	s.mux.HandleFunc("/api/health", s.handleHealth)
}

func (s *DashboardServer) handleTokenAnalytics(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"total_tokens": 1000,
		"token_limit":  100000,
		"usage_pct":    1.0,
		"remaining":    99000,
	}
	sendJSON(w, response)
}

func (s *DashboardServer) handleCostAnalytics(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"total_cost":     0.05,
		"cost_per_token": 0.00005,
		"currency":       "USD",
		"period":         "daily",
	}
	sendJSON(w, response)
}

func (s *DashboardServer) handlePlanningAnalytics(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"total_plans":       10,
		"successful_plans":  8,
		"failed_plans":     2,
		"success_rate":      0.8,
		"avg_plan_duration": "1.2s",
	}
	sendJSON(w, response)
}

func (s *DashboardServer) handleMemoryAnalytics(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"total_memories":    100,
		"daily_memories":    10,
		"longterm_memories": 90,
		"memory_hits":       50,
		"memory_misses":     10,
		"hit_rate":          0.83,
	}
	sendJSON(w, response)
}

func (s *DashboardServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"version":   "1.0.0",
	}
	sendJSON(w, response)
}

func sendJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

type TimeWindowedMetrics struct {
	window  time.Duration
	metrics []MetricPoint
}

type MetricPoint struct {
	Timestamp time.Time
	Value     float64
}

func NewTimeWindowedMetrics(window time.Duration) *TimeWindowedMetrics {
	return &TimeWindowedMetrics{
		window:  window,
		metrics: make([]MetricPoint, 0),
	}
}

func (m *TimeWindowedMetrics) Add(value float64) {
	now := time.Now()
	m.metrics = append(m.metrics, MetricPoint{
		Timestamp: now,
		Value:     value,
	})
	m.prune()
}

func (m *TimeWindowedMetrics) prune() {
	cutoff := time.Now().Add(-m.window)
	filtered := make([]MetricPoint, 0)
	for _, p := range m.metrics {
		if p.Timestamp.After(cutoff) {
			filtered = append(filtered, p)
		}
	}
	m.metrics = filtered
}

func (m *TimeWindowedMetrics) Average() float64 {
	if len(m.metrics) == 0 {
		return 0
	}
	sum := 0.0
	for _, p := range m.metrics {
		sum += p.Value
	}
	return sum / float64(len(m.metrics))
}

func (m *TimeWindowedMetrics) Sum() float64 {
	sum := 0.0
	for _, p := range m.metrics {
		sum += p.Value
	}
	return sum
}

func (m *TimeWindowedMetrics) Max() float64 {
	if len(m.metrics) == 0 {
		return 0
	}
	max := m.metrics[0].Value
	for _, p := range m.metrics {
		if p.Value > max {
			max = p.Value
		}
	}
	return max
}

type DailySummary struct {
	Date        string  `json:"date"`
	TotalTokens int     `json:"total_tokens"`
	TotalCost  float64 `json:"total_cost"`
	SuccessRate float64 `json:"success_rate"`
}

func CalculateTrend(points []float64) string {
	if len(points) < 2 {
		return "stable"
	}

	first := points[0]
	last := points[len(points)-1]
	changePct := (last - first) / first * 100

	if changePct > 10 {
		return "increasing"
	} else if changePct < -10 {
		return "decreasing"
	}
	return "stable"
}