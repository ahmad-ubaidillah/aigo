package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	view        string
	width       int
	height      int
	sessionID   string
	agentStatus string
	activeAgent string
	currentTask string
	action      string
	location    string
	progress    int
	focus       string

	messages    []string
	input       string
	agentFleet  map[string]string
	taskQueue   []string
	activeTools []string
	memory      string
	logs        []string
	kanban      map[string][]string
}

func NewModel() Model {
	return Model{
		view:        "main",
		width:       120,
		height:      30,
		sessionID:   "default",
		agentStatus: "idle",
		activeAgent: "MANAGER",
		currentTask: "Waiting for task...",
		action:      "None",
		location:    "-",
		progress:    0,
		focus:       "Ready for input",

		messages: []string{},
		input:    "",
		agentFleet: map[string]string{
			"MANAGER":  "Plan",
			"CODER":    "Work",
			"TESTER":   "Idle",
			"REVIEWER": "Wait",
		},
		taskQueue: []string{},
		activeTools: []string{
			"RipGrep",
			"Shell-Exec",
			"File-Edit",
		},
		memory: "Long-Term",
		logs:   []string{},
		kanban: map[string][]string{
			"Backlog":     {},
			"In Progress": {},
			"Review":      {},
			"Done":        {},
		},
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) View() string {
	switch m.view {
	case "main":
		return m.renderMainView()
	case "kanban":
		return m.renderKanbanView()
	case "logs":
		return m.renderLogsView()
	case "tools":
		return m.renderToolsView()
	default:
		return m.renderMainView()
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit

	case "enter":
		if m.input != "" {
			m.messages = append(m.messages, "> "+m.input)
			m.focus = "Processing: " + m.input
			m.input = ""
		}
		return m, nil

	case "esc":
		if m.view != "main" {
			m.view = "main"
		}
		m.input = ""
		return m, nil

	case "backspace":
		if len(m.input) > 0 {
			m.input = m.input[:len(m.input)-1]
		}
		return m, nil

	case "f1":
		return m, nil

	case "f2":
		m.view = "kanban"
		return m, nil

	case "f3":
		m.view = "logs"
		return m, nil

	case "f4":
		m.view = "tools"
		return m, nil

	default:
		if len(msg.String()) == 1 {
			m.input += msg.String()
		}
		return m, nil
	}
}

var (
	cyan   = lipgloss.Color("#00FFFF")
	yellow = lipgloss.Color("#FFD700")
	green  = lipgloss.Color("#00FF00")
	blue   = lipgloss.Color("#00BFFF")
	red    = lipgloss.Color("#FF6B6B")
	purple = lipgloss.Color("#9B59B6")
	gray   = lipgloss.Color("#666666")
	darkBG = lipgloss.Color("#1a1a2e")
	border = lipgloss.Color("#444")

	panelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(border).
			Padding(0, 1)

	headerStyle = lipgloss.NewStyle().
			Foreground(cyan).Bold(true).Padding(0, 1)

	agentManagerStyle  = lipgloss.NewStyle().Foreground(yellow)
	agentCoderStyle    = lipgloss.NewStyle().Foreground(blue)
	agentTesterStyle   = lipgloss.NewStyle().Foreground(green)
	agentReviewerStyle = lipgloss.NewStyle().Foreground(gray)

	inputStyle = lipgloss.NewStyle().
			Foreground(green).Background(darkBG).
			Padding(0, 1)
)

func (m Model) renderMainView() string {
	header := m.renderHeader()

	leftPanel := m.renderLeftPanel()
	rightPanel := m.renderRightPanel()

	bodyWidth := m.width - 4
	if bodyWidth < 80 {
		bodyWidth = 80
	}

	leftWidth := 30
	rightWidth := bodyWidth - leftWidth - 1

	body := lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.NewStyle().Width(leftWidth).Render(leftPanel),
		lipgloss.NewStyle().Width(rightWidth).Render(rightPanel),
	)

	focusBar := lipgloss.NewStyle().
		Foreground(cyan).Background(darkBG).
		Padding(0, 2).Bold(true).
		Width(m.width - 2).
		Render(" ⚡ FOCUS: [ " + m.focus + " ] ")

	inputBar := inputStyle.
		Width(m.width - 2).
		Render(" ⌨️ INPUT: [ " + m.input + "_ ] ")

	statusBar := lipgloss.NewStyle().
		Foreground(gray).Padding(0, 1).
		Width(m.width - 2).
		Render(" [F1] Help [F2] Kanban [F3] Logs [F4] Tools [Esc] Back [C-c] Quit ")

	return lipgloss.JoinVertical(lipgloss.Top,
		lipgloss.NewStyle().Width(m.width-2).Render(header),
		lipgloss.NewStyle().Width(m.width-2).Render(body),
		focusBar,
		inputBar,
		statusBar,
	)
}

func (m Model) renderHeader() string {
	version := "v1.5.0"
	statusDot := "●"
	statusColor := green

	if m.agentStatus != "running" {
		statusColor = gray
	}

	status := lipgloss.NewStyle().Foreground(statusColor).Bold(true).Render(statusDot)

	return lipgloss.NewStyle().
		Width(m.width-4).
		Foreground(cyan).
		Bold(true).
		Padding(0, 1).
		Render(" 🤖 AIGO " + version + "  [" + status + "] MULTI-AGENT MODE     [Session: " + m.sessionID + "] ")
}

func (m Model) renderLeftPanel() string {
	agentFleet := m.renderAgentFleet()
	contextTools := m.renderContextTools()
	quickTasks := m.renderQuickTasks()

	return lipgloss.JoinVertical(lipgloss.Top,
		agentFleet,
		contextTools,
		quickTasks,
	)
}

func (m Model) renderAgentFleet() string {
	var fleet []string
	fleet = append(fleet, headerStyle.Render(" 👥 AGENT FLEET "))
	fleet = append(fleet, "")

	styles := map[string]lipgloss.Style{
		"MANAGER":  agentManagerStyle,
		"CODER":    agentCoderStyle,
		"TESTER":   agentTesterStyle,
		"REVIEWER": agentReviewerStyle,
	}

	for agent, status := range m.agentFleet {
		style, ok := styles[agent]
		if !ok {
			style = lipgloss.NewStyle()
		}
		current := ""
		if agent == m.activeAgent {
			current = " [ACTIVE]"
		}
		fleet = append(fleet, style.Render(" "+agent+" ["+status+"]"+current))
	}

	return panelStyle.Width(28).Render(strings.Join(fleet, "\n"))
}

func (m Model) renderContextTools() string {
	var ctx []string
	ctx = append(ctx, headerStyle.Render(" 🧠 CONTEXT & TOOLS "))
	ctx = append(ctx, "")
	ctx = append(ctx, " 🧠 Mem: "+m.memory)
	ctx = append(ctx, "")
	ctx = append(ctx, " 🛠️  Active Tools:")

	for i, tool := range m.activeTools {
		connector := "├"
		if i == len(m.activeTools)-1 {
			connector = "└"
		}
		ctx = append(ctx, "   "+connector+"─ "+tool)
	}

	return panelStyle.Width(28).Render(strings.Join(ctx, "\n"))
}

func (m Model) renderQuickTasks() string {
	var tasks []string
	tasks = append(tasks, headerStyle.Render(" 📋 QUICK TASKS "))
	tasks = append(tasks, "")

	if len(m.taskQueue) == 0 {
		tasks = append(tasks, "  No tasks in queue")
	} else {
		for _, task := range m.taskQueue {
			tasks = append(tasks, " [ ] "+task)
		}
	}

	return panelStyle.Width(28).Render(strings.Join(tasks, "\n"))
}

func (m Model) renderRightPanel() string {
	activity := m.renderActivity()
	interaction := m.renderInteraction()

	return lipgloss.JoinVertical(lipgloss.Top,
		activity,
		interaction,
	)
}

func (m Model) renderActivity() string {
	var act []string
	act = append(act, headerStyle.Render(" 🚀 LIVE AGENT ACTIVITY "))
	act = append(act, "")

	agentColor := blue
	if m.activeAgent != "" {
		switch m.activeAgent {
		case "MANAGER":
			agentColor = yellow
		case "CODER":
			agentColor = blue
		case "TESTER":
			agentColor = green
		case "REVIEWER":
			agentColor = purple
		}
	}

	agentName := lipgloss.NewStyle().Foreground(agentColor).Bold(true).Render(m.activeAgent)
	act = append(act, " 1. 🟦 ACTIVE AGENT  : [ "+agentName+" ]")
	act = append(act, " 2. 📋 CURRENT TASK  : [ "+m.currentTask+" ]")
	act = append(act, " 3. 🛠️  ACTION       : [ "+m.action+" ]")
	act = append(act, " 4. 📂 LOCATION     : [ "+m.location+" ]")
	act = append(act, " 5. 📊 STATUS       : [ "+m.renderProgressBar()+" ] "+fmt.Sprintf("%d%%", m.progress))
	act = append(act, "")
	act = append(act, " ───────────────────────────────────────")

	return panelStyle.Width(48).Render(strings.Join(act, "\n"))
}

func (m Model) renderProgressBar() string {
	filled := int(float64(m.progress) / 100.0 * 10)
	bar := strings.Repeat("█", filled) + strings.Repeat("░", 10-filled)
	return lipgloss.NewStyle().Foreground(green).Render(bar)
}

func (m Model) renderInteraction() string {
	var chat []string
	chat = append(chat, headerStyle.Render(" 💬 INTERACTION "))
	chat = append(chat, "")

	if len(m.messages) == 0 {
		chat = append(chat, " Awaiting input...")
	} else {
		for _, msg := range m.messages {
			if len(chat) > 15 {
				break
			}
			chat = append(chat, " "+msg)
		}
	}

	return panelStyle.Width(48).Render(strings.Join(chat, "\n"))
}

func (m Model) renderKanbanView() string {
	var kanban []string
	kanban = append(kanban, headerStyle.Render(" 📋 KANBAN VIEW "))
	kanban = append(kanban, "")

	columns := map[string][]string{
		"Backlog":     m.kanban["Backlog"],
		"In Progress": m.kanban["In Progress"],
		"Review":      m.kanban["Review"],
		"Done":        m.kanban["Done"],
	}

	for col, tasks := range columns {
		colStyle := lipgloss.NewStyle().Bold(true).Foreground(cyan)
		kanban = append(kanban, colStyle.Render(" "+col+" ("+fmt.Sprint(len(tasks))+")"))
		if len(tasks) == 0 {
			kanban = append(kanban, "   (empty)")
		}
		for _, task := range tasks {
			kanban = append(kanban, "   • "+task)
		}
		kanban = append(kanban, "")
	}

	statusBar := lipgloss.NewStyle().Foreground(gray).Render(" [Esc] Back [C-c] Quit ")

	return lipgloss.JoinVertical(lipgloss.Top,
		panelStyle.Width(80).Render(strings.Join(kanban, "\n")),
		statusBar,
	)
}

func (m Model) renderLogsView() string {
	var logLines []string
	logLines = append(logLines, headerStyle.Render(" 📜 AGENT LOGS "))
	logLines = append(logLines, "")

	if len(m.logs) == 0 {
		logLines = append(logLines, " No logs yet")
	} else {
		for _, log := range m.logs {
			logLines = append(logLines, " "+log)
		}
	}

	statusBar := lipgloss.NewStyle().Foreground(gray).Render(" [Esc] Back [C-c] Quit ")

	return lipgloss.JoinVertical(lipgloss.Top,
		panelStyle.Width(80).Render(strings.Join(logLines, "\n")),
		statusBar,
	)
}

func (m Model) renderToolsView() string {
	var toolLines []string
	toolLines = append(toolLines, headerStyle.Render(" 🔧 TOOLS CONFIG "))
	toolLines = append(toolLines, "")

	for _, tool := range m.activeTools {
		toolLines = append(toolLines, " ✓ "+tool)
	}

	statusBar := lipgloss.NewStyle().Foreground(gray).Render(" [Esc] Back [C-c] Quit ")

	return lipgloss.JoinVertical(lipgloss.Top,
		panelStyle.Width(80).Render(strings.Join(toolLines, "\n")),
		statusBar,
	)
}
