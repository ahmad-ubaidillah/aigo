//go:build !darwin && !linux

package systray

func RunDummy() {}

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/getlantern/systray"
	"github.com/ahmad-ubaidillah/aigo/pkg/types"
)

type Manager struct {
	mu           sync.RWMutex
	iconPath     string
	isRunning    bool
	cfg          types.Config
	onShow       func()
	onHide       func()
	onQuit       func()
	menuItems    map[string]*systray.MenuItem
}

type ManagerOption func(*Manager)

func WithIconPath(path string) ManagerOption {
	return func(m *Manager) {
		m.iconPath = path
	}
}

func WithConfig(cfg types.Config) ManagerOption {
	return func(m *Manager) {
		m.cfg = cfg
	}
}

func WithOnShow(fn func()) ManagerOption {
	return func(m *Manager) {
		m.onShow = fn
	}
}

func WithOnHide(fn func()) ManagerOption {
	return func(m *Manager) {
		m.onHide = fn
	}
}

func WithOnQuit(fn func()) ManagerOption {
	return func(m *Manager) {
		m.onQuit = fn
	}
}

func NewManager(opts ...ManagerOption) *Manager {
	m := &Manager{
		iconPath:  "icon.png",
		menuItems: make(map[string]*systray.MenuItem),
	}

	for _, opt := range opts {
		opt(m)
	}

	return m
}

func (m *Manager) Run() error {
	m.mu.Lock()
	if m.isRunning {
		m.mu.Unlock()
		return fmt.Errorf("systray manager already running")
	}
	m.isRunning = true
	m.mu.Unlock()

	systray.SetIcon(m.getIconData())
	systray.SetTitle("Aigo")
	systray.SetTooltip("Aigo - AI Agent Platform")

	m.buildMenu()

	go m.handleSignals()

	return nil
}

func (m *Manager) getIconData() []byte {
	data, err := os.ReadFile(m.iconPath)
	if err != nil {
		log.Printf("systray: icon not found, using default: %v", err)
		return defaultIcon
	}
	return data
}

func (m *Manager) buildMenu() {
	show := systray.AddMenuItem("Show", "Show Aigo window")
	hide := systray.AddMenuItem("Hide", "Hide Aigo window")
	systray.AddSeparator()

	status := systray.AddMenuItem("Status: Running", "Current status")
	status.Disable()

	systray.AddSeparator()

	newSession := systray.AddMenuItem("New Session", "Create a new session")
	quickTask := systray.AddMenuItem("Quick Task", "Run a quick task")
	systray.AddSeparator()

	settings := systray.AddMenuItem("Settings", "Open settings")
	docs := systray.AddMenuItem("Documentation", "View documentation")
	systray.AddSeparator()

	quit := systray.AddMenuItem("Quit", "Quit Aigo")

	m.menuItems["show"] = show
	m.menuItems["hide"] = hide
	m.menuItems["status"] = status
	m.menuItems["new_session"] = newSession
	m.menuItems["quick_task"] = quickTask
	m.menuItems["settings"] = settings
	m.menuItems["docs"] = docs
	m.menuItems["quit"] = quit

	go func() {
		for {
			select {
			case <-show.ClickedCh:
				if m.onShow != nil {
					m.onShow()
				}
			case <-hide.ClickedCh:
				if m.onHide != nil {
					m.onHide()
				}
			case <-newSession.ClickedCh:
				m.notify("New Session", "Creating new session...")
			case <-quickTask.ClickedCh:
				m.notify("Quick Task", "Opening task input...")
			case <-settings.ClickedCh:
				m.notify("Settings", "Opening settings...")
			case <-docs.ClickedCh:
				m.notify("Docs", "View documentation at docs.aigo.ai")
			case <-quit.ClickedCh:
				if m.onQuit != nil {
					m.onQuit()
				}
				systray.Quit()
				return
			}
		}
	}()
}

func (m *Manager) handleSignals() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigCh
	log.Printf("systray: received signal %v, shutting down", sig)
	systray.Quit()
}

func (m *Manager) notify(title, msg string) {
	log.Printf("systray: %s - %s", title, msg)
}

func (m *Manager) SetStatus(status string) {
	m.mu.RLock()
	item, ok := m.menuItems["status"]
	m.mu.RUnlock()

	if ok {
		item.SetTitle("Status: " + status)
	}
}

func (m *Manager) Stop() {
	m.mu.Lock()
	m.isRunning = false
	m.mu.Unlock()
	systray.Quit()
}

func (m *Manager) UpdateMenu(itemID string, title string) {
	m.mu.RLock()
	item, ok := m.menuItems[itemID]
	m.mu.RUnlock()

	if ok {
		item.SetTitle(title)
	}
}

type SessionManager struct {
	db     *SessionDB
	mu     sync.RWMutex
	active string
}

type SessionDB interface {
	CreateSession(name string) (string, error)
	GetSession(id string) (*types.Session, error)
	ListSessions() ([]types.Session, error)
	DeleteSession(id string) error
}

func NewSessionManager(db SessionDB) *SessionManager {
	return &SessionManager{db: db}
}

func (sm *SessionManager) CreateSession(name string) (string, error) {
	if name == "" {
		name = fmt.Sprintf("session_%d", time.Now().Unix())
	}
	return sm.db.CreateSession(name)
}

func (sm *SessionManager) SetActiveSession(id string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.active = id
}

func (sm *SessionManager) GetActiveSession() string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.active
}

type QuickTask struct {
	Input     string    `json:"input"`
	SessionID string    `json:"session_id"`
	CreatedAt time.Time `json:"created_at"`
	Timeout   time.Duration `json:"timeout"`
}

func (qt *QuickTask) Execute() error {
	if qt.Timeout == 0 {
		qt.Timeout = 5 * time.Minute
	}
	return nil
}

var defaultIcon = []byte{
	0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, 0x00, 0x00, 0x00, 0x0d,
	0x49, 0x48, 0x44, 0x52, 0x00, 0x00, 0x00, 0x10, 0x00, 0x00, 0x00, 0x10,
	0x08, 0x06, 0x00, 0x00, 0x00, 0x1f, 0xf3, 0xff, 0x61, 0x00, 0x00, 0x00,
	0x01, 0x73, 0x52, 0x47, 0x42, 0x00, 0xae, 0xce, 0x1c, 0xe9, 0x00, 0x00,
	0x00, 0x3d, 0x49, 0x44, 0x41, 0x54, 0x38, 0x8d, 0x63, 0x60, 0x20, 0x02,
	0xbc, 0xfe, 0x4f, 0x81, 0x31, 0x03, 0x03, 0x13, 0x23, 0x23, 0xc3, 0x78,
	0x80, 0x31, 0x03, 0x03, 0x13, 0x23, 0x23, 0xc3, 0x78, 0x80, 0x31, 0x03,
	0x03, 0x13, 0x23, 0x23, 0xc3, 0x78, 0x80, 0x31, 0x03, 0x03, 0x13, 0x23,
	0x23, 0x83, 0x00, 0x25, 0x00, 0x02, 0x7c, 0x02, 0x2c, 0x68, 0x2f, 0x8c,
	0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4e, 0x44, 0xae, 0x42, 0x60, 0x82,
}

type Status int

const (
	StatusIdle Status = iota
	StatusRunning
	StatusThinking
	StatusWaiting
	StatusError
)

func (s Status) String() string {
	switch s {
	case StatusIdle:
		return "Idle"
	case StatusRunning:
		return "Running"
	case StatusThinking:
		return "Thinking"
	case StatusWaiting:
		return "Waiting"
	case StatusError:
		return "Error"
	default:
		return "Unknown"
	}
}
