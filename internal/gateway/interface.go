package gateway

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Platform interface {
	Name() string
	Connect(ctx context.Context) error
	Disconnect(ctx context.Context) error
	SendMessage(ctx context.Context, chatID, text string) error
	Listen(ctx context.Context, handler func(Message)) error
}

type Message struct {
	Platform  string
	ChatID    string
	UserID    string
	UserName  string
	Text      string
	MediaURL  string
	Timestamp time.Time
}

type Manager struct {
	platforms map[string]Platform
	running   map[string]bool
	mu        sync.Mutex
}

func NewManager() *Manager {
	return &Manager{
		platforms: make(map[string]Platform),
		running:   make(map[string]bool),
	}
}

func (m *Manager) Register(p Platform) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.platforms[p.Name()] = p
}

func (m *Manager) Start(ctx context.Context, name string) error {
	m.mu.Lock()
	p, ok := m.platforms[name]
	m.mu.Unlock()

	if !ok {
		return fmt.Errorf("platform %s not found", name)
	}

	if err := p.Connect(ctx); err != nil {
		return fmt.Errorf("connect %s: %w", name, err)
	}

	m.mu.Lock()
	m.running[name] = true
	m.mu.Unlock()

	return nil
}

func (m *Manager) Stop(name string) error {
	m.mu.Lock()
	p, ok := m.platforms[name]
	m.mu.Unlock()

	if !ok {
		return fmt.Errorf("platform %s not found", name)
	}

	if err := p.Disconnect(context.Background()); err != nil {
		return fmt.Errorf("disconnect %s: %w", name, err)
	}

	m.mu.Lock()
	m.running[name] = false
	m.mu.Unlock()

	return nil
}

func (m *Manager) Send(platform, chatID, text string) error {
	m.mu.Lock()
	p, ok := m.platforms[platform]
	m.mu.Unlock()

	if !ok {
		return fmt.Errorf("platform %s not found", platform)
	}

	if err := p.SendMessage(context.Background(), chatID, text); err != nil {
		return fmt.Errorf("send via %s: %w", platform, err)
	}

	return nil
}

func (m *Manager) Status() map[string]bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	status := make(map[string]bool)
	for name := range m.platforms {
		status[name] = m.running[name]
	}

	return status
}
