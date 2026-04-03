package tui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ahmad-ubaidillah/aigo/internal/cli"
	"github.com/ahmad-ubaidillah/aigo/pkg/types"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type SetupModel struct {
	step      int
	steps     []string
	provider  string
	apiKey    string
	model     string
	workspace string
	providers []string
	selected  int
	input     string
	inputMode bool
	width     int
	height    int
	complete  bool
	cfg       *types.Config
}

var setupSteps = []string{
	"Welcome",
	"Select Provider",
	"Configure API Key",
	"Select Model",
	"Set Workspace",
	"Install OpenCode",
	"Complete",
}

func NewSetupModel() *SetupModel {
	return &SetupModel{
		step:      0,
		steps:     setupSteps,
		provider:  "openai",
		providers: []string{"openai", "anthropic", "openrouter", "glm", "local"},
		selected:  0,
		complete:  false,
		cfg:       &types.Config{},
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
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m *SetupModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
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
			m.selected = (m.selected - 1 + len(m.providers)) % len(m.providers)
		}
		return m, nil

	case "down", "j":
		if !m.inputMode && m.step == 1 {
			m.selected = (m.selected + 1) % len(m.providers)
		}
		return m, nil

	case "tab":
		if !m.inputMode && m.step == 1 {
			m.selected = (m.selected + 1) % len(m.providers)
		}
		return m, nil

	case "backspace":
		if m.inputMode && len(m.input) > 0 {
			m.input = m.input[:len(m.input)-1]
		}
		return m, nil

	default:
		if m.inputMode && len(msg.String()) == 1 {
			m.input += msg.String()
		} else if !m.inputMode {
			switch msg.String() {
			case "1":
				if m.step == 2 {
					m.provider = "openai"
				}
			case "2":
				if m.step == 2 {
					m.provider = "anthropic"
				}
			case "3":
				if m.step == 2 {
					m.provider = "openrouter"
				}
			case "4":
				if m.step == 2 {
					m.provider = "glm"
				}
			case "5":
				if m.step == 2 {
					m.provider = "local"
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
		m.provider = m.providers[m.selected]
		m.step++
	case 2:
		if m.input != "" {
			m.apiKey = m.input
			m.input = ""
		}
		m.step++
	case 3:
		if m.input != "" {
			m.model = m.input
			m.input = ""
		}
		m.step++
	case 4:
		if m.input != "" {
			m.workspace = m.input
			m.input = ""
		}
		m.step++
	case 5:
		m.step++
		m.complete = true
	case 6:
		m.saveConfig()
		return m, tea.Quit
	}
	return m, nil
}

func (m *SetupModel) saveConfig() {
	home, _ := os.UserHomeDir()
	cfgDir := filepath.Join(home, ".aigo")
	os.MkdirAll(cfgDir, 0755)

	m.cfg.LLM.Provider = m.provider
	m.cfg.LLM.APIKey = m.apiKey
	if m.model != "" {
		m.cfg.Model.Coding = m.model
	}

	cfgPath := filepath.Join(cfgDir, "config.yaml")
	cli.SaveConfig(*m.cfg, cfgPath)

	if m.workspace != "" {
		workspaceFile := filepath.Join(cfgDir, ".workspace")
		os.WriteFile(workspaceFile, []byte(m.workspace), 0644)
	}
}

func (m *SetupModel) View() string {
	stepStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FFFF")).Bold(true)
	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD700")).Bold(true)
	inputStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00")).Background(lipgloss.Color("#1a1a2e"))
	boxStyle := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#6366F1")).Padding(1, 2)

	var content string

	switch m.step {
	case 0:
		content = boxStyle.Render(fmt.Sprintf(`%s

Welcome to Aigo Setup!

This wizard will help you configure:
- LLM Provider selection
- API Key configuration
- Model selection
- Workspace directory

Press [Enter] to continue.`, stepStyle.Render("Welcome to Aigo V1.5")))

	case 1:
		providerList := ""
		for i, p := range m.providers {
			if i == m.selected {
				providerList += selectedStyle.Render("▶ "+p+" ") + "\n"
			} else {
				providerList += "  " + p + "\n"
			}
		}
		content = boxStyle.Render(fmt.Sprintf(`%s

Select your LLM provider:

%s
Use ↑/↓ or Tab to select, Enter to confirm.`, stepStyle.Render("Provider Selection"), providerList))

	case 2:
		note := ""
		if m.provider == "local" {
			note = "\n(Note: Local provider doesn't need API key)"
		}
		input := m.input
		if m.inputMode {
			input = inputStyle.Render(m.input + "_")
		} else {
			input = inputStyle.Render(strings.Repeat("•", len(m.input)))
		}
		content = boxStyle.Render(fmt.Sprintf(`%s

Provider: %s

Enter your API key:%s

%s

Press Enter to skip.`, stepStyle.Render("API Key"), m.provider, note, input))

	case 3:
		defaultModels := map[string]string{
			"openai":     "gpt-4o",
			"anthropic":  "claude-sonnet-4-20250514",
			"openrouter": "openai/gpt-4o",
			"glm":        "glm-4-plus",
			"local":      "qwen2.5-coder",
		}
		defaultModel := defaultModels[m.provider]
		input := m.input
		if m.inputMode {
			input = inputStyle.Render(m.input + "_")
		} else if m.input == "" {
			input = lipgloss.NewStyle().Foreground(lipgloss.Color("#666")).Render(defaultModel + " (default)")
		} else {
			input = m.input
		}
		content = boxStyle.Render(fmt.Sprintf(`%s

Provider: %s

Enter model name (or press Enter for default):%s

%s

Default: %s`, stepStyle.Render("Model Selection"), m.provider, lipgloss.NewStyle().Foreground(lipgloss.Color("#666")).Render("(optional)"), input, defaultModel))

	case 4:
		input := m.input
		if m.inputMode {
			input = inputStyle.Render(m.input + "_")
		} else if m.input == "" {
			input = lipgloss.NewStyle().Foreground(lipgloss.Color("#666")).Render("current directory (default)")
		} else {
			input = m.input
		}
		content = boxStyle.Render(fmt.Sprintf(`%s

Enter your workspace directory:%s

%s

Press Enter to use current directory.`, stepStyle.Render("Workspace"), lipgloss.NewStyle().Foreground(lipgloss.Color("#666")).Render("(optional)"), input))

	case 5:
		content = boxStyle.Render(fmt.Sprintf(`%s

Summary:
- Provider: %s
- Model: %s
- Workspace: %s

Press [Enter] to install OpenCode and complete setup.`, stepStyle.Render("Ready to Complete"), m.provider, m.model, m.workspace))

	case 6:
		content = boxStyle.Render(fmt.Sprintf(`%s

Setup Complete!

Run 'aigo tui' to start the interactive mode.
Run 'aigo run "your task"' to execute a task.

Enjoy!`, stepStyle.Render("Congratulations!")))
	}

	progress := ""
	for i, s := range m.steps {
		if i == m.step {
			progress += stepStyle.Render("[" + s + "] ")
		} else if i < m.step {
			progress += lipgloss.NewStyle().Foreground(lipgloss.Color("#666")).Render("[" + s + "] ")
		} else {
			progress += lipgloss.NewStyle().Foreground(lipgloss.Color("#333")).Render("[" + s + "] ")
		}
	}

	help := "↑↓/Tab: select | Enter: confirm | Esc: back | Ctrl+C: quit"
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666")).Padding(0, 1)

	return lipgloss.JoinVertical(lipgloss.Top,
		lipgloss.NewStyle().Padding(0, 1).Render(progress),
		boxStyle.Width(60).Render(content),
		helpStyle.Render(help),
	)
}

func RunSetupWizard() error {
	p := tea.NewProgram(NewSetupModel(), tea.WithAltScreen())
	_, err := p.Run()
	return err
}

func installOpenCodeIfNeeded() error {
	cmd := exec.Command("which", "opencode")
	if cmd.Run() == nil {
		return nil
	}

	fmt.Println("OpenCode not found. Installing...")
	installCmd := exec.Command("bash", "-c", "curl -fsSL https://opencode.ai/install | bash")
	installCmd.Stdout = os.Stdout
	installCmd.Stderr = os.Stderr
	return installCmd.Run()
}
