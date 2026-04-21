package analytics

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

type PlanningResult int

const (
	False PlanningResult = iota
	True
	Partial
)

type Analytics struct {
	Tokens       int
	Cost        float64
	TokenLimit  int
	Savings     float64

	mu               sync.RWMutex
	planningResults []PlanningResult
	planningTotal  int
	memoryRetrievals int
	memoryHits     int
}

func NewAnalytics() *Analytics {
	return &Analytics{
		TokenLimit:  100000,
		planningResults: make([]PlanningResult, 0),
	}
}

func (a *Analytics) TrackToken(tokens int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.Tokens += tokens
}

func (a *Analytics) TrackCost(cost float64) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.Cost += cost
}

func (a *Analytics) CalcROI() float64 {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.TokenLimit == 0 {
		return 0
	}
	return float64(a.TokenLimit-a.Tokens) / float64(a.TokenLimit)
}

func (a *Analytics) RecordPlanning(result PlanningResult) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.planningResults = append(a.planningResults, result)
	a.planningTotal++
}

func (a *Analytics) PlanningAccuracy() float64 {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if a.planningTotal == 0 {
		return 0
	}

	correct := 0
	for _, r := range a.planningResults {
		if r == True {
			correct++
		}
	}
	return float64(correct) / float64(a.planningTotal)
}

func (a *Analytics) TrackMemory(retrieved, hit bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.memoryRetrievals++
	if hit {
		a.memoryHits++
	}
}

func (a *Analytics) MemoryAccuracy() float64 {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if a.memoryRetrievals == 0 {
		return 0
	}
	return float64(a.memoryHits) / float64(a.memoryRetrievals)
}

type Metrics struct {
	TotalTokens    int       `json:"total_tokens"`
	TotalCost    float64   `json:"total_cost"`
	TokenSavings float64   `json:"token_savings"`
	PlanningAcc float64   `json:"planning_acc"`
	MemoryAcc   float64   `json:"memory_acc"`
	LastUpdated time.Time `json:"last_updated"`
}

func (a *Analytics) GetMetrics() *Metrics {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return &Metrics{
		TotalTokens: a.Tokens,
		TotalCost: a.Cost,
		TokenSavings: a.Savings,
		PlanningAcc: a.PlanningAccuracy(),
		MemoryAcc:  a.MemoryAccuracy(),
		LastUpdated: time.Now(),
	}
}

type AnalyticsPersistence struct {
	filePath string
}

func NewAnalyticsPersistence(path string) *AnalyticsPersistence {
	return &AnalyticsPersistence{filePath: path}
}

func (p *AnalyticsPersistence) Save(a *Analytics) error {
	metrics := a.GetMetrics()
	data, err := json.MarshalIndent(metrics, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}
	return os.WriteFile(p.filePath, data, 0644)
}

func (p *AnalyticsPersistence) Load() (*Analytics, error) {
	data, err := os.ReadFile(p.filePath)
	if err != nil {
		return nil, err
	}

	var metrics Metrics
	err = json.Unmarshal(data, &metrics)
	if err != nil {
		return nil, err
	}

	a := NewAnalytics()
	a.Tokens = metrics.TotalTokens
	a.Cost = metrics.TotalCost
	a.Savings = metrics.TokenSavings

	return a, nil
}

type AnalyticsDBStore struct {
	db      *sql.DB
	dbPath string
}

func NewAnalyticsDBStore(dbPath string) (*AnalyticsDBStore, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}

	store := &AnalyticsDBStore{
		db:      db,
		dbPath: dbPath,
	}

	if err := store.initSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to init schema: %w", err)
	}

	return store, nil
}

func (s *AnalyticsDBStore) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS metrics (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		total_tokens INTEGER,
		total_cost REAL,
		token_savings REAL,
		planning_acc REAL,
		memory_acc REAL,
		recorded_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_metrics_recorded ON metrics(recorded_at);
	`
	_, err := s.db.Exec(schema)
	return err
}

func (s *AnalyticsDBStore) SaveMetrics(tokens int, cost, savings, planningAcc, memoryAcc float64) error {
	_, err := s.db.Exec(
		`INSERT INTO metrics(total_tokens, total_cost, token_savings, planning_acc, memory_acc) VALUES(?, ?, ?, ?, ?)`,
		tokens, cost, savings, planningAcc, memoryAcc,
	)
	return err
}

func (s *AnalyticsDBStore) GetMetrics(limit int) ([]MetricRecord, error) {
	rows, err := s.db.Query(
		`SELECT id, total_tokens, total_cost, token_savings, planning_acc, memory_acc, recorded_at 
		 FROM metrics ORDER BY recorded_at DESC LIMIT ?`,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	records := make([]MetricRecord, 0)
	for rows.Next() {
		var r MetricRecord
		if err := rows.Scan(&r.ID, &r.TotalTokens, &r.TotalCost, &r.TokenSavings, &r.PlanningAcc, &r.MemoryAcc, &r.RecordedAt); err != nil {
			return nil, err
		}
		records = append(records, r)
	}

	return records, rows.Err()
}

func (s *AnalyticsDBStore) GetDailySummary() (*MetricRecord, error) {
	row := s.db.QueryRow(`
		SELECT COALESCE(SUM(total_tokens), 0), COALESCE(SUM(total_cost), 0), COALESCE(SUM(token_savings), 0),
		       COALESCE(AVG(planning_acc), 0), COALESCE(AVG(memory_acc), 0)
		FROM metrics 
		WHERE date(recorded_at) = date('now')
	`)

	var r MetricRecord
	err := row.Scan(&r.TotalTokens, &r.TotalCost, &r.TokenSavings, &r.PlanningAcc, &r.MemoryAcc)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

func (s *AnalyticsDBStore) Close() error {
	return s.db.Close()
}

type MetricRecord struct {
	ID          int64
	TotalTokens  int
	TotalCost   float64
	TokenSavings float64
	PlanningAcc float64
	MemoryAcc  float64
	RecordedAt time.Time
}