package tui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	primary   = lipgloss.Color("#7C3AED")
	secondary = lipgloss.Color("#10B981")
	accent    = lipgloss.Color("#F59E0B")
	warn      = lipgloss.Color("#EF4444")
	dim       = lipgloss.Color("#6B7280")
	text      = lipgloss.Color("#F9FAFB")
	muted     = lipgloss.Color("#9CA3AF")

	setupPanelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primary).
			Padding(0, 1)

	titleStyle = lipgloss.NewStyle().
			Foreground(primary).Bold(true)

	dimStyle = lipgloss.NewStyle().Foreground(dim)
)

type SetupModel struct {
	step      int
	view      string
	input     string
	inputMode bool

	provider  string
	apiKey    string
	model     string
	workspace string
	gateway   bool
	platforms []string
	opencode  bool

	selected int
	width    int
	height   int
	complete bool
}

var setupSteps = []string{
	"Welcome",
	"Model",
	"API Key",
	"Workspace",
	"OpenCode",
	"Gateway",
	"Complete",
}

func NewSetupModel() *SetupModel {
	return &SetupModel{
		step:      0,
		view:      "setup",
		selected:  0,
		provider:  "opencode",
		workspace: "",
		gateway:   false,
		opencode:  true,
		platforms: []string{},
	}
}

func (m *SetupModel) Init() tea.Cmd {
	return nil
}

func (m *SetupModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m *SetupModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit

	case "enter":
		if m.inputMode {
			m.inputMode = false
			return m, nil
		}
		return m.handleEnter()

	case "esc":
		if m.inputMode {
			m.inputMode = false
			m.input = ""
			return m, nil
		}
		if m.step > 0 {
			m.step--
			return m, nil
		}

	case "up", "k":
		if !m.inputMode && m.step == 1 {
			m.selected = (m.selected - 1 + 5) % 5
		}
		return m, nil

	case "down", "j":
		if !m.inputMode && m.step == 1 {
			m.selected = (m.selected + 1) % 5
		}
		return m, nil

	case "backspace":
		if m.inputMode && len(m.input) > 0 {
			m.input = m.input[:len(m.input)-1]
		}
		return m, nil

	case " ":
		if m.step == 4 {
			m.opencode = !m.opencode
		}
		if m.step == 5 && !m.inputMode {
			m.gateway = !m.gateway
		}
		return m, nil

	default:
		if m.inputMode && len(msg.String()) == 1 {
			m.input += msg.String()
		} else if !m.inputMode {
			switch msg.String() {
			case "1":
				if m.step == 1 {
					m.provider = "opencode"
					m.model = "qwen3.6-plus-free"
				}
			case "2":
				if m.step == 1 {
					m.provider = "openai"
					m.model = "gpt-4o"
				}
			case "3":
				if m.step == 1 {
					m.provider = "anthropic"
					m.model = "claude-sonnet-4-20250514"
				}
			case "4":
				if m.step == 1 {
					m.provider = "openrouter"
					m.model = "openai/gpt-4o"
				}
			case "5":
				if m.step == 1 {
					m.provider = "local"
					m.model = "qwen2.5-coder"
				}
			}
		}
		return m, nil
	}
	return m, nil
}

func (m *SetupModel) handleEnter() (tea.Model, tea.Cmd) {
	switch m.step {
	case 0:
		m.step++

	case 1:
		providers := []string{"opencode", "openai", "anthropic", "openrouter", "local"}
		m.provider = providers[m.selected]
		switch m.provider {
		case "opencode":
			m.model = "qwen3.6-plus-free"
		case "openai":
			m.model = "gpt-4o"
		case "anthropic":
			m.model = "claude-sonnet-4-20250514"
		case "openrouter":
			m.model = "openai/gpt-4o"
		case "local":
			m.model = "qwen2.5-coder"
		}
		m.step++

	case 2:
		m.apiKey = m.input
		m.input = ""
		m.step++

	case 3:
		m.workspace = m.input
		if m.workspace == "" {
			wd, _ := os.Getwd()
			m.workspace = wd
		}
		m.input = ""
		m.step++

	case 4:
		m.step++

	case 5:
		m.step++
		m.complete = true
		m.saveConfig()

	case 6:
		return m, tea.Quit
	}
	return m, nil
}

func (m *SetupModel) saveConfig() {
	home, _ := os.UserHomeDir()
	cfgDir := filepath.Join(home, ".aigo")
	os.MkdirAll(cfgDir, 0755)

	cfgPath := filepath.Join(cfgDir, "config.yaml")

	config := fmt.Sprintf(`llm:
  provider: %s
  api_key: %s
  default_model: %s

model:
  default: %s
  coding: auto
  intent: gpt-4o-mini

workspace: %s

gateway:
  enabled: %v
  platforms: %v

opencode:
  binary: ""
  timeout: 300
  max_turns: 50

memory:
  max_l0_items: 20
  max_l1_items: 50
  auto_compress: true

web:
  enabled: false
  port: :8080

token_budget:
  total_budget: 100000
  warning_threshold: 0.7
  critical_threshold: 0.9
  alert_channels:
    - log
    - tui
  per_provider: false
`, m.provider, m.apiKey, m.model, m.model, m.workspace, m.gateway, m.platforms)

	os.WriteFile(cfgPath, []byte(config), 0644)

	if m.opencode {
		installOpenCode()
	}
}

func installOpenCode() {
	fmt.Println("\n📦 Installing OpenCode...")

	checkCmd := exec.Command("which", "opencode")
	if checkCmd.Run() == nil {
		fmt.Println("   ✓ OpenCode already installed")
		return
	}

	installCmd := exec.Command("bash", "-c", "curl -fsSL https://opencode.ai/install | bash")
	installCmd.Stdout = os.Stdout
	installCmd.Stderr = os.Stderr
	if err := installCmd.Run(); err != nil {
		fmt.Printf("   ✗ Failed to install OpenCode: %v\n", err)
		return
	}
	fmt.Println("   ✓ OpenCode installed successfully")
}

func (m *SetupModel) View() string {
	return m.renderSetupWizard()
}

func (m *SetupModel) renderSetupWizard() string {
	if m.complete {
		return m.renderComplete()
	}

	progress := ""
	for i, s := range setupSteps {
		if i == m.step {
			progress += titleStyle.Render("[" + s + "] ")
		} else if i < m.step {
			progress += dimStyle.Render("[" + s + "] ")
		} else {
			progress += lipgloss.NewStyle().Foreground(lipgloss.Color("#374151")).Render("[" + s + "] ")
		}
	}

	var content string
	switch m.step {
	case 0:
		content = m.renderWelcome()
	case 1:
		content = m.renderProvider()
	case 2:
		content = m.renderAPIKey()
	case 3:
		content = m.renderWorkspace()
	case 4:
		content = m.renderOpenCode()
	case 5:
		content = m.renderGateway()
	}

	help := dimStyle.Render("↑↓: select | Enter: confirm | Esc: back | Space: toggle | Ctrl+C: quit")

	return lipgloss.JoinVertical(lipgloss.Top,
		lipgloss.NewStyle().Padding(0, 1).Width(m.width-2).Render(progress),
		setupPanelStyle.Width(m.width-6).Render(content),
		help,
	)
}

func (m *SetupModel) renderWelcome() string {
	return fmt.Sprintf(`%s

%s

Features:
  • Multi-Provider LLM Support (OpenCode, OpenAI, Anthropic, etc.)
  • OpenCode Integration for free models
  • Interactive TUI Mode
  • Gateway Support (Telegram, Discord, Slack, WhatsApp)
  • Token Budget Management

Press [Enter] to continue.`,
		titleStyle.Render("⚡ Aigo Setup Wizard"),
		dimStyle.Render("Execute with Zen"))
}

func (m *SetupModel) renderProvider() string {
	providers := []struct {
		name   string
		models string
		cost   string
	}{
		{"OpenCode", "qwen3.6, mimo, minimax", "FREE"},
		{"OpenAI", "gpt-4, gpt-4o", "Paid"},
		{"Anthropic", "claude-sonnet, claude-opus", "Paid"},
		{"OpenRouter", "openai/gpt-4o, anthropic/*", "Paid"},
		{"Local", "qwen2.5, llama3", "Self-hosted"},
	}

	var lines []string
	lines = append(lines, titleStyle.Render("▸ Select Model Provider"))
	lines = append(lines, "")

	for i, p := range providers {
		marker := "  "
		if i == m.selected {
			marker = "▶ "
		}
		costStr := p.cost
		if p.cost == "Paid" {
			costStr = lipgloss.NewStyle().Foreground(accent).Render(p.cost)
		} else {
			costStr = lipgloss.NewStyle().Foreground(secondary).Render(p.cost)
		}
		lines = append(lines, fmt.Sprintf("%s%s  %s  (%s)",
			marker, p.name, p.models, costStr))
	}

	return strings.Join(lines, "\n")
}

func (m *SetupModel) renderAPIKey() string {
	inputDisplay := m.input
	if m.inputMode {
		inputDisplay = m.input + "_"
	} else if m.input != "" {
		inputDisplay = "********"
	} else if m.provider == "opencode" || m.provider == "local" {
		inputDisplay = dimStyle.Render("(not required)")
	}

	note := ""
	if m.provider == "opencode" || m.provider == "local" {
		note = "\n" + dimStyle.Render("(No API key needed for this provider)")
	}

	return fmt.Sprintf(`%s

Provider: %s
Model: %s

Enter your API key:%s

%s`,
		titleStyle.Render("▸ API Key Configuration"),
		m.provider, m.model,
		note,
		inputDisplay)
}

func (m *SetupModel) renderWorkspace() string {
	ws := m.workspace
	if ws == "" {
		ws, _ = os.Getwd()
	}

	inputDisplay := ws
	if m.inputMode {
		inputDisplay = m.input
	}

	return fmt.Sprintf(`%s

Enter your workspace directory.

%s`,
		titleStyle.Render("▸ Workspace Configuration"),
		inputDisplay)
}

func (m *SetupModel) renderOpenCode() string {
	checked := " "
	if m.opencode {
		checked = "✓"
	}

	status := "not installed"
	checkCmd := exec.Command("which", "opencode")
	if checkCmd.Run() == nil {
		status = "already installed"
	}

	return fmt.Sprintf(`%s

[%s] Install OpenCode for free models

Status: %s

OpenCode provides free access to:
  • qwen3.6-plus-free
  • mimo-v2-omni-free
  • minimax-m2.5-free

Press [Space] to toggle, [Enter] to continue.`,
		titleStyle.Render("▸ OpenCode Installation"),
		checked,
		dimStyle.Render(status))
}

func (m *SetupModel) renderGateway() string {
	checked := " "
	if m.gateway {
		checked = "✓"
	}

	return fmt.Sprintf(`[%s] Enable Gateway (Telegram, Discord, Slack, WhatsApp)

If enabled, you can interact with Aigo from messaging platforms.

Press [Space] to toggle, [Enter] to continue.`,
		checked)
}

func (m *SetupModel) renderComplete() string {
	return fmt.Sprintf(`%s

✓ Configuration saved to: ~/.aigo/config.yaml
✓ Workspace: %s
✓ Provider: %s / %s

%s

Run 'aigo tui' to start the interactive mode.
Run 'aigo run "your task"' to execute a task.
`,
		titleStyle.Render("✓ Setup Complete!"),
		m.workspace,
		m.provider,
		m.model,
		dimStyle.Render("Thank you for choosing Aigo!"))
}

func RunSetupWizard() error {
	p := tea.NewProgram(NewSetupModel(), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
