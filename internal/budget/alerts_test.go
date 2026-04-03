package budget

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type mockChannel struct {
	name      string
	sendErr   error
	sendCount int32
	lastEvent AlertEvent
	mu        sync.Mutex
}

func (m *mockChannel) Name() string {
	return m.name
}

func (m *mockChannel) Send(event AlertEvent) error {
	m.mu.Lock()
	m.lastEvent = event
	m.mu.Unlock()
	atomic.AddInt32(&m.sendCount, 1)
	return m.sendErr
}

func (m *mockChannel) SendCount() int {
	return int(atomic.LoadInt32(&m.sendCount))
}

func (m *mockChannel) LastEvent() AlertEvent {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.lastEvent
}

func newTestEvent(level AlertLevel) AlertEvent {
	return AlertEvent{
		Level:     level,
		Usage:     5000,
		Budget:    10000,
		Percent:   0.5,
		Provider:  "test-provider",
		Timestamp: time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
		Message:   "Test alert message",
	}
}

func TestNewDispatcher(t *testing.T) {
	t.Parallel()

	d := NewDispatcher()

	if d == nil {
		t.Fatal("NewDispatcher() returned nil")
	}

	if d.channels == nil {
		t.Error("channels slice should not be nil")
	}

	if len(d.channels) != 0 {
		t.Errorf("expected empty channels slice, got %d channels", len(d.channels))
	}
}

func TestRegister(t *testing.T) {
	t.Parallel()

	d := NewDispatcher()
	ch := &mockChannel{name: "test-channel"}

	d.Register(ch)

	if len(d.channels) != 1 {
		t.Errorf("expected 1 channel, got %d", len(d.channels))
	}

	if d.channels[0].Name() != "test-channel" {
		t.Errorf("expected channel name 'test-channel', got %s", d.channels[0].Name())
	}
}

func TestRegisterMultiple(t *testing.T) {
	t.Parallel()

	d := NewDispatcher()
	channels := []AlertChannel{
		&mockChannel{name: "ch1"},
		&mockChannel{name: "ch2"},
		&mockChannel{name: "ch3"},
	}

	for _, ch := range channels {
		d.Register(ch)
	}

	if len(d.channels) != 3 {
		t.Errorf("expected 3 channels, got %d", len(d.channels))
	}
}

func TestDispatchRoutesToAllChannels(t *testing.T) {
	t.Parallel()

	d := NewDispatcher()
	ch1 := &mockChannel{name: "ch1"}
	ch2 := &mockChannel{name: "ch2"}
	ch3 := &mockChannel{name: "ch3"}

	d.Register(ch1)
	d.Register(ch2)
	d.Register(ch3)

	event := newTestEvent(AlertWarning)
	errs := d.Dispatch(event)

	if len(errs) != 0 {
		t.Errorf("expected no errors, got %d", len(errs))
	}

	if ch1.SendCount() != 1 {
		t.Errorf("ch1: expected 1 send, got %d", ch1.SendCount())
	}
	if ch2.SendCount() != 1 {
		t.Errorf("ch2: expected 1 send, got %d", ch2.SendCount())
	}
	if ch3.SendCount() != 1 {
		t.Errorf("ch3: expected 1 send, got %d", ch3.SendCount())
	}
}

func TestDispatchWithNoChannels(t *testing.T) {
	t.Parallel()

	d := NewDispatcher()
	event := newTestEvent(AlertWarning)

	errs := d.Dispatch(event)

	if len(errs) != 0 {
		t.Errorf("expected empty errors slice, got %d errors", len(errs))
	}
}

func TestDispatchWithOneChannelFailing(t *testing.T) {
	t.Parallel()

	d := NewDispatcher()
	ch1 := &mockChannel{name: "ch1"}
	ch2 := &mockChannel{name: "ch2", sendErr: errors.New("send failed")}
	ch3 := &mockChannel{name: "ch3"}

	d.Register(ch1)
	d.Register(ch2)
	d.Register(ch3)

	event := newTestEvent(AlertWarning)
	errs := d.Dispatch(event)

	if len(errs) != 1 {
		t.Errorf("expected 1 error, got %d", len(errs))
	}

	if errs[0].Error() != "channel ch2: send failed" {
		t.Errorf("unexpected error message: %s", errs[0].Error())
	}

	if ch1.SendCount() != 1 {
		t.Errorf("ch1: expected 1 send, got %d", ch1.SendCount())
	}
	if ch2.SendCount() != 1 {
		t.Errorf("ch2: expected 1 send, got %d", ch2.SendCount())
	}
	if ch3.SendCount() != 1 {
		t.Errorf("ch3: expected 1 send, got %d", ch3.SendCount())
	}
}

func TestDispatchMultipleChannelsFailing(t *testing.T) {
	t.Parallel()

	d := NewDispatcher()
	ch1 := &mockChannel{name: "ch1", sendErr: errors.New("error 1")}
	ch2 := &mockChannel{name: "ch2"}
	ch3 := &mockChannel{name: "ch3", sendErr: errors.New("error 3")}

	d.Register(ch1)
	d.Register(ch2)
	d.Register(ch3)

	event := newTestEvent(AlertCritical)
	errs := d.Dispatch(event)

	if len(errs) != 2 {
		t.Errorf("expected 2 errors, got %d", len(errs))
	}

	if ch1.SendCount() != 1 || ch2.SendCount() != 1 || ch3.SendCount() != 1 {
		t.Error("not all channels received the event")
	}
}

func TestLogChannelSend(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		level AlertLevel
	}{
		{"warning level", AlertWarning},
		{"critical level", AlertCritical},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ch := LogChannel{}
			event := newTestEvent(tt.level)

			err := ch.Send(event)

			if err != nil {
				t.Errorf("expected no error, got: %v", err)
			}
		})
	}
}

func TestLogChannelName(t *testing.T) {
	t.Parallel()

	ch := LogChannel{}
	if ch.Name() != "log" {
		t.Errorf("expected name 'log', got %s", ch.Name())
	}
}

func TestTUIChannelSendCallsCallback(t *testing.T) {
	t.Parallel()

	var receivedEvent AlertEvent
	var callbackCalled bool

	ch := TUIChannel{
		onUpdate: func(event AlertEvent) {
			receivedEvent = event
			callbackCalled = true
		},
	}

	event := newTestEvent(AlertWarning)
	err := ch.Send(event)

	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	if !callbackCalled {
		t.Error("callback was not called")
	}

	if receivedEvent.Message != event.Message {
		t.Errorf("expected message %s, got %s", event.Message, receivedEvent.Message)
	}
	if receivedEvent.Level != event.Level {
		t.Errorf("expected level %s, got %s", event.Level, receivedEvent.Level)
	}
}

func TestTUIChannelSendWithNilCallback(t *testing.T) {
	t.Parallel()

	ch := TUIChannel{onUpdate: nil}
	event := newTestEvent(AlertWarning)

	err := ch.Send(event)

	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestTUIChannelName(t *testing.T) {
	t.Parallel()

	ch := TUIChannel{}
	if ch.Name() != "tui" {
		t.Errorf("expected name 'tui', got %s", ch.Name())
	}
}

func TestWebChannelSendCallsCallback(t *testing.T) {
	t.Parallel()

	var receivedEvent AlertEvent
	var callbackCalled bool

	ch := WebChannel{
		broadcast: func(event AlertEvent) {
			receivedEvent = event
			callbackCalled = true
		},
	}

	event := newTestEvent(AlertCritical)
	err := ch.Send(event)

	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	if !callbackCalled {
		t.Error("broadcast callback was not called")
	}

	if receivedEvent.Message != event.Message {
		t.Errorf("expected message %s, got %s", event.Message, receivedEvent.Message)
	}
	if receivedEvent.Level != event.Level {
		t.Errorf("expected level %s, got %s", event.Level, receivedEvent.Level)
	}
}

func TestWebChannelSendWithNilCallback(t *testing.T) {
	t.Parallel()

	ch := WebChannel{broadcast: nil}
	event := newTestEvent(AlertWarning)

	err := ch.Send(event)

	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestWebChannelName(t *testing.T) {
	t.Parallel()

	ch := WebChannel{}
	if ch.Name() != "web" {
		t.Errorf("expected name 'web', got %s", ch.Name())
	}
}

func TestGatewayChannelSendToAllPlatforms(t *testing.T) {
	t.Parallel()

	platforms := []string{"telegram", "discord", "slack"}
	received := make(map[string]string)
	var mu sync.Mutex

	ch := GatewayChannel{
		sendToGateway: func(platform, message string) error {
			mu.Lock()
			received[platform] = message
			mu.Unlock()
			return nil
		},
		platforms: platforms,
	}

	event := newTestEvent(AlertWarning)
	err := ch.Send(event)

	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	if len(received) != 3 {
		t.Errorf("expected 3 platforms to receive, got %d", len(received))
	}

	for _, p := range platforms {
		if _, ok := received[p]; !ok {
			t.Errorf("platform %s did not receive message", p)
		}
	}
}

func TestGatewayChannelSendReturnsError(t *testing.T) {
	t.Parallel()

	callCount := 0
	ch := GatewayChannel{
		sendToGateway: func(platform, message string) error {
			callCount++
			if platform == "failing-platform" {
				return errors.New("connection refused")
			}
			return nil
		},
		platforms: []string{"ok-platform", "failing-platform", "another-ok"},
	}

	event := newTestEvent(AlertCritical)
	err := ch.Send(event)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	expectedErr := "send to failing-platform: connection refused"
	if err.Error() != expectedErr {
		t.Errorf("expected error '%s', got '%s'", expectedErr, err.Error())
	}

	if callCount != 2 {
		t.Errorf("expected 2 calls (stops on error), got %d", callCount)
	}
}

func TestGatewayChannelSendEmptyPlatforms(t *testing.T) {
	t.Parallel()

	ch := GatewayChannel{
		sendToGateway: func(platform, message string) error {
			t.Error("should not be called with empty platforms")
			return nil
		},
		platforms: []string{},
	}

	event := newTestEvent(AlertWarning)
	err := ch.Send(event)

	if err != nil {
		t.Errorf("expected no error with empty platforms, got: %v", err)
	}
}

func TestGatewayChannelName(t *testing.T) {
	t.Parallel()

	ch := GatewayChannel{}
	if ch.Name() != "gateway" {
		t.Errorf("expected name 'gateway', got %s", ch.Name())
	}
}

func TestWithLogChannel(t *testing.T) {
	t.Parallel()

	d := NewDispatcher()
	result := d.WithLogChannel()

	if result != d {
		t.Error("WithLogChannel should return the same dispatcher for chaining")
	}

	if len(d.channels) != 1 {
		t.Errorf("expected 1 channel, got %d", len(d.channels))
	}

	if d.channels[0].Name() != "log" {
		t.Errorf("expected log channel, got %s", d.channels[0].Name())
	}
}

func TestWithTUIChannel(t *testing.T) {
	t.Parallel()

	callback := func(event AlertEvent) {}

	d := NewDispatcher()
	result := d.WithTUIChannel(callback)

	if result != d {
		t.Error("WithTUIChannel should return the same dispatcher for chaining")
	}

	if len(d.channels) != 1 {
		t.Errorf("expected 1 channel, got %d", len(d.channels))
	}

	if d.channels[0].Name() != "tui" {
		t.Errorf("expected tui channel, got %s", d.channels[0].Name())
	}
}

func TestWithWebChannel(t *testing.T) {
	t.Parallel()

	callback := func(event AlertEvent) {}

	d := NewDispatcher()
	result := d.WithWebChannel(callback)

	if result != d {
		t.Error("WithWebChannel should return the same dispatcher for chaining")
	}

	if len(d.channels) != 1 {
		t.Errorf("expected 1 channel, got %d", len(d.channels))
	}

	if d.channels[0].Name() != "web" {
		t.Errorf("expected web channel, got %s", d.channels[0].Name())
	}
}

func TestWithGatewayChannel(t *testing.T) {
	t.Parallel()

	sendFn := func(platform, message string) error { return nil }
	platforms := []string{"telegram", "discord"}

	d := NewDispatcher()
	result := d.WithGatewayChannel(sendFn, platforms)

	if result != d {
		t.Error("WithGatewayChannel should return the same dispatcher for chaining")
	}

	if len(d.channels) != 1 {
		t.Errorf("expected 1 channel, got %d", len(d.channels))
	}

	if d.channels[0].Name() != "gateway" {
		t.Errorf("expected gateway channel, got %s", d.channels[0].Name())
	}
}

func TestFluentBuilderChaining(t *testing.T) {
	t.Parallel()

	d := NewDispatcher().
		WithLogChannel().
		WithTUIChannel(func(event AlertEvent) {}).
		WithWebChannel(func(event AlertEvent) {}).
		WithGatewayChannel(func(platform, message string) error { return nil }, []string{"telegram"})

	if len(d.channels) != 4 {
		t.Errorf("expected 4 channels from chaining, got %d", len(d.channels))
	}

	expectedNames := []string{"log", "tui", "web", "gateway"}
	for i, name := range expectedNames {
		if d.channels[i].Name() != name {
			t.Errorf("channel %d: expected %s, got %s", i, name, d.channels[i].Name())
		}
	}
}

func TestConcurrentDispatch(t *testing.T) {
	t.Parallel()

	d := NewDispatcher()
	ch := &mockChannel{name: "concurrent-test"}
	d.Register(ch)

	const goroutines = 100
	const dispatchesPerGoroutine = 10

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < dispatchesPerGoroutine; j++ {
				event := newTestEvent(AlertWarning)
				event.Message = "concurrent test"
				d.Dispatch(event)
			}
		}(i)
	}

	wg.Wait()

	expectedCount := goroutines * dispatchesPerGoroutine
	if ch.SendCount() != expectedCount {
		t.Errorf("expected %d sends, got %d", expectedCount, ch.SendCount())
	}
}

func TestConcurrentRegister(t *testing.T) {
	t.Parallel()

	d := NewDispatcher()
	const goroutines = 100

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			ch := &mockChannel{name: "ch"}
			d.Register(ch)
		}(i)
	}

	wg.Wait()

	if len(d.channels) != goroutines {
		t.Errorf("expected %d channels, got %d", goroutines, len(d.channels))
	}
}

func TestConcurrentRegisterAndDispatch(t *testing.T) {
	t.Parallel()

	d := NewDispatcher()
	event := newTestEvent(AlertWarning)

	const operations = 50
	var wg sync.WaitGroup
	wg.Add(operations * 2)

	for i := 0; i < operations; i++ {
		go func() {
			defer wg.Done()
			ch := &mockChannel{name: "dynamic-ch"}
			d.Register(ch)
		}()
	}

	for i := 0; i < operations; i++ {
		go func() {
			defer wg.Done()
			d.Dispatch(event)
		}()
	}

	wg.Wait()
}

func TestDispatchEventPreservation(t *testing.T) {
	t.Parallel()

	d := NewDispatcher()
	ch1 := &mockChannel{name: "ch1"}
	ch2 := &mockChannel{name: "ch2"}

	d.Register(ch1)
	d.Register(ch2)

	event := AlertEvent{
		Level:     AlertCritical,
		Usage:     8000,
		Budget:    10000,
		Percent:   0.8,
		Provider:  "openai",
		Timestamp: time.Date(2025, 6, 15, 10, 30, 0, 0, time.UTC),
		Message:   "Critical threshold reached",
	}

	d.Dispatch(event)

	e1 := ch1.LastEvent()
	e2 := ch2.LastEvent()

	if e1.Message != event.Message || e2.Message != event.Message {
		t.Error("event message not preserved")
	}
	if e1.Level != event.Level || e2.Level != event.Level {
		t.Error("event level not preserved")
	}
	if e1.Usage != event.Usage || e2.Usage != event.Usage {
		t.Error("event usage not preserved")
	}
	if e1.Provider != event.Provider || e2.Provider != event.Provider {
		t.Error("event provider not preserved")
	}
}

func TestGatewayChannelMessageFormat(t *testing.T) {
	t.Parallel()

	var receivedMessage string
	ch := GatewayChannel{
		sendToGateway: func(platform, message string) error {
			receivedMessage = message
			return nil
		},
		platforms: []string{"test"},
	}

	event := newTestEvent(AlertWarning)
	event.Message = "Budget warning!"
	ch.Send(event)

	expectedFormat := "[warning] Budget warning!"
	if receivedMessage != expectedFormat {
		t.Errorf("expected message format '%s', got '%s'", expectedFormat, receivedMessage)
	}
}

func TestGatewayChannelMessageFormatCritical(t *testing.T) {
	t.Parallel()

	var receivedMessage string
	ch := GatewayChannel{
		sendToGateway: func(platform, message string) error {
			receivedMessage = message
			return nil
		},
		platforms: []string{"test"},
	}

	event := newTestEvent(AlertCritical)
	event.Message = "Budget exceeded!"
	ch.Send(event)

	expectedFormat := "[critical] Budget exceeded!"
	if receivedMessage != expectedFormat {
		t.Errorf("expected message format '%s', got '%s'", expectedFormat, receivedMessage)
	}
}
