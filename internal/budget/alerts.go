package budget

import (
	"fmt"
	"log"
	"sync"
)

type AlertChannel interface {
	Name() string
	Send(event AlertEvent) error
}

type LogChannel struct{}

func (LogChannel) Name() string { return "log" }
func (LogChannel) Send(event AlertEvent) error {
	switch event.Level {
	case AlertCritical:
		log.Printf("🔴 CRITICAL: %s", event.Message)
	case AlertWarning:
		log.Printf("🟡 WARNING: %s", event.Message)
	}
	return nil
}

type TUIChannel struct {
	onUpdate func(event AlertEvent)
}

func (TUIChannel) Name() string { return "tui" }
func (c TUIChannel) Send(event AlertEvent) error {
	if c.onUpdate != nil {
		c.onUpdate(event)
	}
	return nil
}

type WebChannel struct {
	broadcast func(event AlertEvent)
}

func (WebChannel) Name() string { return "web" }
func (c WebChannel) Send(event AlertEvent) error {
	if c.broadcast != nil {
		c.broadcast(event)
	}
	return nil
}

type GatewayChannel struct {
	sendToGateway func(platform, message string) error
	platforms     []string
}

func (c GatewayChannel) Name() string { return "gateway" }
func (c GatewayChannel) Send(event AlertEvent) error {
	msg := fmt.Sprintf("[%s] %s", event.Level, event.Message)
	for _, platform := range c.platforms {
		if err := c.sendToGateway(platform, msg); err != nil {
			return fmt.Errorf("send to %s: %w", platform, err)
		}
	}
	return nil
}

type AlertDispatcher struct {
	channels []AlertChannel
	mu       sync.RWMutex
}

func NewDispatcher() *AlertDispatcher {
	return &AlertDispatcher{channels: make([]AlertChannel, 0)}
}

func (d *AlertDispatcher) Register(ch AlertChannel) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.channels = append(d.channels, ch)
}

func (d *AlertDispatcher) Dispatch(event AlertEvent) []error {
	d.mu.RLock()
	defer d.mu.RUnlock()

	var errs []error
	for _, ch := range d.channels {
		if err := ch.Send(event); err != nil {
			errs = append(errs, fmt.Errorf("channel %s: %w", ch.Name(), err))
		}
	}
	return errs
}

func (d *AlertDispatcher) WithLogChannel() *AlertDispatcher {
	d.Register(LogChannel{})
	return d
}

func (d *AlertDispatcher) WithTUIChannel(onUpdate func(event AlertEvent)) *AlertDispatcher {
	d.Register(TUIChannel{onUpdate: onUpdate})
	return d
}

func (d *AlertDispatcher) WithWebChannel(broadcast func(event AlertEvent)) *AlertDispatcher {
	d.Register(WebChannel{broadcast: broadcast})
	return d
}

func (d *AlertDispatcher) WithGatewayChannel(sendFn func(platform, message string) error, platforms []string) *AlertDispatcher {
	d.Register(GatewayChannel{sendToGateway: sendFn, platforms: platforms})
	return d
}
