package ui

import (
	"sync"
)

type Theme struct {
	Name      string
	Colors    map[string]string
	FontSize int
}

var themes = map[string]*Theme{
	"dark": {
		Name:      "dark",
		Colors:    map[string]string{"bg": "#1a1a1a", "fg": "#ffffff", "accent": "#007acc"},
		FontSize: 14,
	},
	"light": {
		Name:      "light",
		Colors:    map[string]string{"bg": "#ffffff", "fg": "#000000", "accent": "#007acc"},
		FontSize: 14,
	},
}

type ThemeSwitcher struct {
	current string
	mu      sync.RWMutex
}

func GetTheme(name string) *Theme {
	if t, ok := themes[name]; ok {
		return t
	}
	return themes["dark"]
}

func NewThemeSwitcher() *ThemeSwitcher {
	return &ThemeSwitcher{current: "dark"}
}

func (ts *ThemeSwitcher) SetTheme(name string) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	if _, ok := themes[name]; ok {
		ts.current = name
	}
}

func (ts *ThemeSwitcher) Current() string {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	return ts.current
}

type SSE struct {
	clients map[chan string]bool
	mu      sync.RWMutex
}

func NewSSE() *SSE {
	return &SSE{clients: make(map[chan string]bool)}
}

func (s *SSE) Add(ch chan string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clients[ch] = true
}

func (s *SSE) Broadcast(msg string) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for ch := range s.clients {
		ch <- msg
	}
}