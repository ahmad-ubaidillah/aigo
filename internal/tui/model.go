package tui

import tea "github.com/charmbracelet/bubbletea"

type Model struct {
	activeView  string
	width       int
	height      int
	sessionID   string
	messages    []string
	input       string
	focused     bool
	agentStatus string
	taskQueue   []string
	agents      map[string]string
	activity    []string
	tools       map[string]bool
	kanban      map[string][]string
	logs        []string
}

func NewModel() Model {
	return Model{
		activeView:  "dashboard",
		width:       80,
		height:      24,
		sessionID:   "default",
		messages:    []string{},
		input:       "",
		focused:     true,
		agentStatus: "idle",
		taskQueue:   []string{},
		agents:      map[string]string{},
		activity:    []string{},
		tools:       map[string]bool{},
		kanban: map[string][]string{
			"Backlog":     {},
			"In Progress": {},
			"Review":      {},
			"Done":        {},
		},
		logs: []string{},
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}
