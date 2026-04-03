package budget

import (
	"sync"
	"testing"
)

func TestNewManager(t *testing.T) {
	t.Parallel()

	m := NewManager(10000, Thresholds{Warning: 0.7, Critical: 0.9})
	used, budget, percent := m.Usage()
	if budget != 10000 {
		t.Errorf("expected budget 10000, got %d", budget)
	}
	if used != 0 {
		t.Errorf("expected used 0, got %d", used)
	}
	if percent != 0 {
		t.Errorf("expected percent 0, got %f", percent)
	}
}

func TestManager_Add_IncrementsUsage(t *testing.T) {
	t.Parallel()

	m := NewManager(10000, Thresholds{Warning: 0.7, Critical: 0.9})
	m.Add(1000, "openai")
	m.Add(2000, "anthropic")

	used, _, _ := m.Usage()
	if used != 3000 {
		t.Errorf("expected used 3000, got %d", used)
	}

	perProvider := m.PerProvider()
	if perProvider["openai"] != 1000 {
		t.Errorf("expected openai 1000, got %d", perProvider["openai"])
	}
	if perProvider["anthropic"] != 2000 {
		t.Errorf("expected anthropic 2000, got %d", perProvider["anthropic"])
	}
}

func TestManager_Add_WarningAlert(t *testing.T) {
	t.Parallel()

	m := NewManager(10000, Thresholds{Warning: 0.7, Critical: 0.9})
	alerts := m.Add(7000, "openai")

	if len(alerts) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(alerts))
	}
	if alerts[0].Level != AlertWarning {
		t.Errorf("expected warning level, got %s", alerts[0].Level)
	}
	if alerts[0].Usage != 7000 {
		t.Errorf("expected usage 7000, got %d", alerts[0].Usage)
	}
	if alerts[0].Percent != 0.7 {
		t.Errorf("expected percent 0.7, got %f", alerts[0].Percent)
	}
}

func TestManager_Add_CriticalAlert(t *testing.T) {
	t.Parallel()

	m := NewManager(10000, Thresholds{Warning: 0.7, Critical: 0.9})
	alerts := m.Add(9000, "openai")

	if len(alerts) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(alerts))
	}
	if alerts[0].Level != AlertCritical {
		t.Errorf("expected critical level, got %s", alerts[0].Level)
	}
}

func TestManager_Add_NoAlertBelowThreshold(t *testing.T) {
	t.Parallel()

	m := NewManager(10000, Thresholds{Warning: 0.7, Critical: 0.9})
	alerts := m.Add(5000, "openai")

	if len(alerts) != 0 {
		t.Errorf("expected 0 alerts, got %d", len(alerts))
	}
}

func TestManager_OnAlert_FiresHandler(t *testing.T) {
	t.Parallel()

	m := NewManager(10000, Thresholds{Warning: 0.7, Critical: 0.9})
	var received AlertEvent
	var mu sync.Mutex

	m.OnAlert(func(event AlertEvent) {
		mu.Lock()
		defer mu.Unlock()
		received = event
	})

	m.Add(7000, "openai")

	mu.Lock()
	defer mu.Unlock()
	if received.Level != AlertWarning {
		t.Errorf("expected warning level, got %s", received.Level)
	}
	if received.Provider != "openai" {
		t.Errorf("expected provider openai, got %s", received.Provider)
	}
}

func TestManager_Usage(t *testing.T) {
	t.Parallel()

	m := NewManager(10000, Thresholds{Warning: 0.7, Critical: 0.9})
	m.Add(3000, "openai")

	used, budget, percent := m.Usage()
	if used != 3000 {
		t.Errorf("expected used 3000, got %d", used)
	}
	if budget != 10000 {
		t.Errorf("expected budget 10000, got %d", budget)
	}
	if percent != 0.3 {
		t.Errorf("expected percent 0.3, got %f", percent)
	}
}

func TestManager_PerProvider(t *testing.T) {
	t.Parallel()

	m := NewManager(10000, Thresholds{Warning: 0.7, Critical: 0.9})
	m.Add(1000, "openai")
	m.Add(2000, "anthropic")
	m.Add(500, "openai")

	perProvider := m.PerProvider()
	if perProvider["openai"] != 1500 {
		t.Errorf("expected openai 1500, got %d", perProvider["openai"])
	}
	if perProvider["anthropic"] != 2000 {
		t.Errorf("expected anthropic 2000, got %d", perProvider["anthropic"])
	}
}

func TestManager_History(t *testing.T) {
	t.Parallel()

	m := NewManager(10000, Thresholds{Warning: 0.7, Critical: 0.9})
	m.Add(1000, "openai")
	m.Add(2000, "anthropic")

	history := m.History()
	if len(history) != 2 {
		t.Fatalf("expected 2 history entries, got %d", len(history))
	}
	if history[0].Used != 1000 {
		t.Errorf("expected first used 1000, got %d", history[0].Used)
	}
	if history[1].Used != 3000 {
		t.Errorf("expected second used 3000, got %d", history[1].Used)
	}
}

func TestManager_History_ReturnsCopy(t *testing.T) {
	t.Parallel()

	m := NewManager(10000, Thresholds{Warning: 0.7, Critical: 0.9})
	m.Add(1000, "openai")

	h1 := m.History()
	h1[0].Used = 99999

	h2 := m.History()
	if h2[0].Used == 99999 {
		t.Error("expected History() to return a copy")
	}
}

func TestManager_Reset(t *testing.T) {
	t.Parallel()

	m := NewManager(10000, Thresholds{Warning: 0.7, Critical: 0.9})
	m.Add(5000, "openai")
	m.Reset()

	used, _, percent := m.Usage()
	if used != 0 {
		t.Errorf("expected used 0 after reset, got %d", used)
	}
	if percent != 0 {
		t.Errorf("expected percent 0 after reset, got %f", percent)
	}

	perProvider := m.PerProvider()
	if len(perProvider) != 0 {
		t.Errorf("expected empty perProvider after reset, got %d entries", len(perProvider))
	}
}

func TestManager_WithDispatcher(t *testing.T) {
	t.Parallel()

	m := NewManager(10000, Thresholds{Warning: 0.7, Critical: 0.9})
	dispatcher := NewDispatcher()
	var received []AlertEvent
	var mu sync.Mutex

	dispatcher.Register(&testChannel{onSend: func(event AlertEvent) {
		mu.Lock()
		defer mu.Unlock()
		received = append(received, event)
	}})

	m.WithDispatcher(dispatcher)
	m.Add(7000, "openai")

	mu.Lock()
	defer mu.Unlock()
	if len(received) != 1 {
		t.Fatalf("expected 1 alert received, got %d", len(received))
	}
	if received[0].Level != AlertWarning {
		t.Errorf("expected warning, got %s", received[0].Level)
	}
}

func TestManager_ConcurrentAdd(t *testing.T) {
	t.Parallel()

	m := NewManager(1000000, Thresholds{Warning: 0.7, Critical: 0.9})
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				m.Add(100, "openai")
			}
		}()
	}

	wg.Wait()
	used, _, _ := m.Usage()
	if used != 100000 {
		t.Errorf("expected used 100000, got %d", used)
	}
}

func TestManager_ZeroBudget(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for zero budget")
		}
	}()

	m := NewManager(0, Thresholds{Warning: 0.7, Critical: 0.9})
	m.Add(100, "openai")
}

type testChannel struct {
	onSend func(event AlertEvent)
}

func (c *testChannel) Name() string { return "test" }
func (c *testChannel) Send(event AlertEvent) error {
	if c.onSend != nil {
		c.onSend(event)
	}
	return nil
}
