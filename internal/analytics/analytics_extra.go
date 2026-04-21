package analytics

import (
	"encoding/json"
)

type Dashboard struct {
	metrics   *Metrics
	exports   []string
}

func NewDashboard() *Dashboard {
	return &Dashboard{
		metrics: &Metrics{},
		exports:  make([]string, 0),
	}
}

func (d *Dashboard) ExportJSON() string {
	data, _ := json.Marshal(d.metrics)
	return string(data)
}

func (d *Dashboard) GetSummary() string {
	return "Analytics Summary"
}

func (a *Analytics) TokenSavings() float64 {
	if a.TokenLimit == 0 {
		return 0
	}
	return float64(a.TokenLimit-a.Tokens) / float64(a.TokenLimit)
}

type RetentionMetrics struct {
	TotalSessions  int
	ActiveUsers  int
	ChurnRate   float64
}

func (a *Analytics) GetRetention() *RetentionMetrics {
	return &RetentionMetrics{
		TotalSessions: a.planningTotal,
		ActiveUsers:  a.memoryHits,
		ChurnRate:   0.1,
	}
}

type CostBreakdown struct {
	Provider  string
	Model     string
	Tokens    int
	Cost      float64
}

func (a *Analytics) GetCostBreakdown() []*CostBreakdown {
	return []*CostBreakdown{
		{Provider: "openai", Model: "gpt-4o", Tokens: a.Tokens, Cost: a.Cost},
	}
}

type PlanningAccuracyMetrics struct {
	TotalPlans      int
	ApprovedPlans  int
	RejectedPlans  int
	Accuracy       float64
}

func (a *Analytics) GetPlanningMetrics() *PlanningAccuracyMetrics {
	return &PlanningAccuracyMetrics{
		TotalPlans:     a.planningTotal,
		ApprovedPlans: a.memoryHits,
		Accuracy:       a.PlanningAccuracy(),
	}
}

type MemoryMetrics struct {
	TotalStored   int
	Retrievals   int
	Hits         int
	Accuracy     float64
}

func (a *Analytics) GetMemoryMetrics() *MemoryMetrics {
	return &MemoryMetrics{
		TotalStored: a.memoryRetrievals,
		Retrievals:  a.memoryRetrievals,
		Hits:        a.memoryHits,
		Accuracy:    a.MemoryAccuracy(),
	}
}