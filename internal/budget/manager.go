package budget

import (
	"fmt"
	"sync"
	"time"
)

type AlertLevel string

const (
	AlertWarning  AlertLevel = "warning"
	AlertCritical AlertLevel = "critical"
)

type AlertEvent struct {
	Level     AlertLevel
	Usage     int
	Budget    int
	Percent   float64
	Provider  string
	Timestamp time.Time
	Message   string
}

type AlertHandler func(event AlertEvent)

type UsageSnapshot struct {
	Used      int
	Budget    int
	Percent   float64
	Provider  string
	Timestamp time.Time
}

type Thresholds struct {
	Warning  float64
	Critical float64
}

type Manager struct {
	mu            sync.RWMutex
	totalBudget   int
	used          int
	perProvider   map[string]int
	thresholds    Thresholds
	alertHandlers []AlertHandler
	history       []UsageSnapshot
}

func NewManager(budget int, thresholds Thresholds) *Manager {
	m := &Manager{
		totalBudget:   budget,
		used:          0,
		perProvider:   make(map[string]int),
		thresholds:    thresholds,
		alertHandlers: make([]AlertHandler, 0),
	}
	return m
}

func (m *Manager) Add(used int, provider string) []AlertEvent {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.used += used
	m.perProvider[provider] += used

	percent := float64(m.used) / float64(m.totalBudget)
	if percent > 1 {
		panic("token budget exceeded 100%")
	}

	snapshot := UsageSnapshot{
		Used:      m.used,
		Budget:    m.totalBudget,
		Percent:   percent,
		Provider:  provider,
		Timestamp: time.Now(),
	}
	m.history = append(m.history, snapshot)

	var alerts []AlertEvent
	if percent >= m.thresholds.Critical {
		event := AlertEvent{
			Level:     AlertCritical,
			Usage:     m.used,
			Budget:    m.totalBudget,
			Percent:   percent,
			Provider:  provider,
			Timestamp: time.Now(),
			Message:   fmt.Sprintf("Token budget CRITICAL: %d/%d (%.0f%%) used via %s", m.used, m.totalBudget, percent*100, provider),
		}
		alerts = append(alerts, event)
		for _, h := range m.alertHandlers {
			h(event)
		}
	} else if percent >= m.thresholds.Warning {
		event := AlertEvent{
			Level:     AlertWarning,
			Usage:     m.used,
			Budget:    m.totalBudget,
			Percent:   percent,
			Provider:  provider,
			Timestamp: time.Now(),
			Message:   fmt.Sprintf("Token budget WARNING: %d/%d (%.0f%%) used via %s", m.used, m.totalBudget, percent*100, provider),
		}
		alerts = append(alerts, event)
		for _, h := range m.alertHandlers {
			h(event)
		}
	}

	return alerts
}

func (m *Manager) OnAlert(h AlertHandler) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.alertHandlers = append(m.alertHandlers, h)
}

func (m *Manager) Usage() (used, budget int, percent float64) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.used, m.totalBudget, float64(m.used) / float64(m.totalBudget)
}

func (m *Manager) PerProvider() map[string]int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make(map[string]int)
	for k, v := range m.perProvider {
		result[k] = v
	}
	return result
}

func (m *Manager) History() []UsageSnapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]UsageSnapshot, len(m.history))
	copy(result, m.history)
	return result
}

func (m *Manager) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.used = 0
	m.perProvider = make(map[string]int)
}

func (m *Manager) WithDispatcher(d *AlertDispatcher) *Manager {
	m.OnAlert(func(event AlertEvent) {
		d.Dispatch(event)
	})
	return m
}
