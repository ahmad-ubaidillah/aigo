// Package tui implements a rich terminal UI for Aigo chat using Bubble Tea.
package tui

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// AgentRunner is the function signature for running the agent.
type AgentRunner func(ctx context.Context, prompt string) (AgentResult, error)

// AgentResult is the result from the agent.
type AgentResult struct {
	Response string
	Steps    int
	Usage    struct {
		TotalTokens int
	}
	Duration time.Duration
}

// Message represents a chat message.
type Message struct {
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	Meta      string    `json:"meta,omitempty"`
}

// Model is the Bubble Tea model for the chat TUI.
type Model struct {
	viewport   viewport.Model
	textInput  textinput.Model
	messages   []Message
	runner     AgentRunner
	width      int
	height     int
	loading    bool
	spinnerIdx int
	err        error
	quit       bool
	showHelp   bool
	historyPath string
}

// spinnerFrames is a simple ASCII spinner.
var spinnerFrames = []string{"⣾", "⣽", "⣻", "⢿", "⡿", "⣟", "⣯", "⣷"}

// Styles.
var (
	senderStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#58A6FF")).
			Bold(true)

	botStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#3FB950")).
			Bold(true)

	contentStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#C9D1D9"))

	metaStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#8B949E")).
			Italic(true)

	headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F0F6FC")).
			Background(lipgloss.Color("#161B22")).
			Bold(true).
			Padding(0, 1)

	inputStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#C9D1D9"))

	borderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#30363D"))

	helpBoxStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#C9D1D9")).
			Background(lipgloss.Color("#161B22")).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#58A6FF")).
			Padding(1, 2)
)

type responseMsg struct {
	result AgentResult
	err    error
}

type tickMsg struct{}

func historyFile() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".aigo", "tui_history.json")
}

func (m *Model) loadHistory() {
	if m.historyPath == "" {
		return
	}
	data, err := os.ReadFile(m.historyPath)
	if err != nil {
		return
	}
	var msgs []Message
	if json.Unmarshal(data, &msgs) == nil {
		m.messages = msgs
	}
}

func (m *Model) saveHistory() {
	if m.historyPath == "" {
		return
	}
	_ = os.MkdirAll(filepath.Dir(m.historyPath), 0755)
	data, _ := json.MarshalIndent(m.messages, "", "  ")
	_ = os.WriteFile(m.historyPath, data, 0644)
}

// New creates a new TUI chat model.
func New(runner AgentRunner) *Model {
	ti := textinput.New()
	ti.Placeholder = "Type a message..."
	ti.Focus()
	ti.CharLimit = 2000
	ti.Width = 50
	ti.Prompt = "➜ "
	ti.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#58A6FF"))
	ti.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#C9D1D9"))

	vp := viewport.New(80, 20)
	vp.SetContent("")

	m := &Model{
		viewport:    vp,
		textInput:   ti,
		runner:      runner,
		messages:    []Message{},
		historyPath: historyFile(),
	}
	m.loadHistory()
	return m
}

// Init implements tea.Model.
func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
		m.tick(),
	)
}

func (m *Model) tick() tea.Cmd {
	return tea.Tick(time.Millisecond*80, func(t time.Time) tea.Msg {
		return tickMsg{}
	})
}

// Update implements tea.Model.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	if m.showHelp {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			m.showHelp = false
			return m, nil
		case tea.WindowSizeMsg:
			m.width = msg.Width
			m.height = msg.Height
		}
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width - 4
		m.viewport.Height = msg.Height - 6
		m.textInput.Width = msg.Width - 10
		m.refreshViewport()

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.quit = true
			m.saveHistory()
			return m, tea.Quit

		case tea.KeyF1:
			m.showHelp = true
			return m, nil

		case tea.KeyEnter:
			if m.loading {
				return m, nil
			}
			input := strings.TrimSpace(m.textInput.Value())
			if input == "" {
				return m, nil
			}
			if input == "/quit" || input == "/exit" {
				m.quit = true
				m.saveHistory()
				return m, tea.Quit
			}
			if input == "/clear" {
				m.messages = nil
				m.textInput.SetValue("")
				m.refreshViewport()
				return m, nil
			}
			if input == "/help" || input == "?" {
				m.showHelp = true
				m.textInput.SetValue("")
				return m, nil
			}

			// Add user message
			m.messages = append(m.messages, Message{
				Role:      "user",
				Content:   input,
				Timestamp: time.Now(),
			})
			m.textInput.SetValue("")
			m.loading = true
			m.refreshViewport()

			// Run agent in background
			return m, m.runAgent(input)

		case tea.KeyUp:
			m.viewport.LineUp(3)
		case tea.KeyDown:
			m.viewport.LineDown(3)
		case tea.KeyPgUp:
			m.viewport.LineUp(m.viewport.Height - 2)
		case tea.KeyPgDown:
			m.viewport.LineDown(m.viewport.Height - 2)
		}

	case responseMsg:
		m.loading = false
		if msg.err != nil {
			m.messages = append(m.messages, Message{
				Role:      "error",
				Content:   msg.err.Error(),
				Timestamp: time.Now(),
				Meta:      "error",
			})
		} else {
			meta := fmt.Sprintf("%d steps · %d tokens · %s",
				msg.result.Steps,
				msg.result.Usage.TotalTokens,
				msg.result.Duration.Round(time.Millisecond))
			m.messages = append(m.messages, Message{
				Role:      "assistant",
				Content:   msg.result.Response,
				Timestamp: time.Now(),
				Meta:      meta,
			})
		}
		m.saveHistory()
		m.refreshViewport()

	case tickMsg:
		if m.loading {
			m.spinnerIdx = (m.spinnerIdx + 1) % len(spinnerFrames)
			m.refreshViewport()
		}
		return m, m.tick()
	}

	// Update sub-models
	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	cmds = append(cmds, cmd)

	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *Model) runAgent(prompt string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
		defer cancel()
		result, err := m.runner(ctx, prompt)
		return responseMsg{result: result, err: err}
	}
}

func (m *Model) refreshViewport() {
	var b strings.Builder

	if len(m.messages) == 0 && !m.loading {
		b.WriteString(metaStyle.Render("Welcome to Aigo Chat! Press F1 or type /help for keybindings.\n\n"))
	}

	for _, msg := range m.messages {
		switch msg.Role {
		case "user":
			b.WriteString(senderStyle.Render("You"))
			b.WriteString(" ")
			b.WriteString(metaStyle.Render(msg.Timestamp.Format("15:04")))
			b.WriteString("\n")
			b.WriteString(contentStyle.Render(msg.Content))
			b.WriteString("\n\n")

		case "assistant":
			b.WriteString(botStyle.Render("Aigo"))
			b.WriteString(" ")
			b.WriteString(metaStyle.Render(msg.Timestamp.Format("15:04")))
			if msg.Meta != "" {
				b.WriteString(" ")
				b.WriteString(metaStyle.Render("[" + msg.Meta + "]"))
			}
			b.WriteString("\n")
			b.WriteString(contentStyle.Render(msg.Content))
			b.WriteString("\n\n")

		case "error":
			b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#F85149")).Bold(true).Render("Error"))
			b.WriteString("\n")
			b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#F85149")).Render(msg.Content))
			b.WriteString("\n\n")
		}
	}

	if m.loading {
		b.WriteString(botStyle.Render("Aigo"))
		b.WriteString(" ")
		b.WriteString(metaStyle.Render("thinking..."))
		b.WriteString(" ")
		b.WriteString(spinnerFrames[m.spinnerIdx])
		b.WriteString("\n")
	}

	m.viewport.SetContent(b.String())
	m.viewport.GotoBottom()
}

func helpContent() string {
	lines := []string{
		lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#58A6FF")).Render("Aigo Chat Keybindings"),
		"",
		"Enter        Send message",
		"Esc / Ctrl+C Quit",
		"F1 / ?       Toggle this help",
		"↑ / ↓        Scroll messages",
		"PgUp / PgDn  Scroll by page",
		"",
		"Commands:",
		"  /clear     Clear chat history",
		"  /quit      Exit chat",
		"  /help      Show this help",
		"",
		"Press any key to close...",
	}
	return strings.Join(lines, "\n")
}

// View implements tea.Model.
func (m *Model) View() string {
	if m.quit {
		return ""
	}

	if m.showHelp {
		box := helpBoxStyle.Render(helpContent())
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
	}

	header := headerStyle.Render("  Aigo Chat — Execute with Zen  ")
	if m.width > 0 {
		header = headerStyle.Width(m.width - 2).Render("  Aigo Chat — Execute with Zen  ")
	}

	inputLine := inputStyle.Render(m.textInput.View())

	help := metaStyle.Render("enter: send · /clear: clear · /quit: exit · F1: help · ↑↓: scroll")
	if m.width > 0 {
		help = metaStyle.Width(m.width - 2).Render(help)
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		borderStyle.Render(m.viewport.View()),
		inputLine,
		help,
	)
}

// Run starts the TUI chat.
func Run(runner AgentRunner) error {
	m := New(runner)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
