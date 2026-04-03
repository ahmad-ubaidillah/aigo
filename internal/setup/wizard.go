// Package setup provides a setup wizard for Aigo.
package setup

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/ahmad-ubaidillah/aigo/pkg/types"
	"github.com/charmbracelet/lipgloss"
)

const (
	ModeCLI string = "cli"
	ModeWeb string = "web"
)

type SetupWizard struct {
	mode     string
	complete bool
	cfg      *types.Config
}

func NewSetupWizard() *SetupWizard {
	return &SetupWizard{
		mode:     ModeCLI,
		complete: false,
		cfg:      &types.Config{},
	}
}

func (w *SetupWizard) Run() error {
	ctx := context.Background()

	fmt.Print("\n" + lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FFFF")).Render("Welcome to Aigo V1.5 Setup!"))
	fmt.Println()
	fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("#00FFFF")).Render("Choose your setup mode:"))
	fmt.Println()
	fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD700")).Render("1. CLI (Terminal UI - Advanced)"))
	fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("#00BFFF")).Render("2. Web UI (Browser - Beginner Friendly)"))
	fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("#00FFFF")).Render("3. Exit"))
	fmt.Println()

	var input string
	fmt.Print("Enter choice [1-3]: ")
	fmt.Scanln(&input)

	switch strings.TrimSpace(input) {
	case "1":
		w.mode = ModeCLI
		return w.setupCLI(ctx)
	case "2":
		w.mode = ModeWeb
		return w.setupWeb(ctx)
	case "3":
		return fmt.Errorf("setup cancelled")
	default:
		return fmt.Errorf("invalid choice: %s", input)
	}
}

func (w *SetupWizard) setupCLI(ctx context.Context) error {
	fmt.Println()
	fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00")).Render("Setting up CLI mode..."))

	if err := w.configureProviders(); err != nil {
		return fmt.Errorf("configure providers: %w", err)
	}

	w.configureTokenBudget()

	if err := w.configureWorkspace(); err != nil {
		return fmt.Errorf("configure workspace: %w", err)
	}

	w.configureGateway()

	w.complete = true

	fmt.Println()
	fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00")).Render("CLI setup complete!"))
	return nil
}

func (w *SetupWizard) setupWeb(ctx context.Context) error {
	fmt.Println()
	fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00")).Render("Setting up Web UI mode..."))

	if err := w.configureProviders(); err != nil {
		return fmt.Errorf("configure providers: %w", err)
	}

	w.configureTokenBudget()

	w.complete = true

	fmt.Println()
	fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00")).Render("Web UI setup complete!"))
	fmt.Println("Run 'aigo web' to start the web interface")
	return nil
}

func (w *SetupWizard) configureProviders() error {
	fmt.Println()
	fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD700")).Render("=== LLM Provider Configuration ==="))
	fmt.Println("You can configure multiple providers for automatic fallback.")
	fmt.Println()

	providers := []string{"openai", "anthropic", "openrouter", "glm", "local", "custom"}

	fmt.Println("Available providers:")
	for i, p := range providers {
		fmt.Printf("  %d. %s\n", i+1, p)
	}
	fmt.Println()

	fmt.Print("Add a provider? (y/n): ")
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))

	if input != "y" && input != "yes" {
		return nil
	}

	w.cfg.LLM.Providers = nil
	w.cfg.LLM.Fallback = nil

	priority := 1
	for input == "y" || input == "yes" {
		fmt.Print("Provider name (e.g., openai, anthropic): ")
		providerInput, _ := reader.ReadString('\n')
		providerName := strings.TrimSpace(providerInput)

		validProvider := false
		for _, p := range providers {
			if providerName == p {
				validProvider = true
				break
			}
		}
		if !validProvider {
			fmt.Printf("Invalid provider: %s\n", providerName)
			continue
		}

		apiKey := ""
		if providerName != "local" {
			fmt.Print("API key (or press Enter to skip): ")
			apiKeyInput, _ := reader.ReadString('\n')
			apiKey = strings.TrimSpace(apiKeyInput)
		}

		model := ""
		fmt.Print("Model (optional, press Enter for default): ")
		modelInput, _ := reader.ReadString('\n')
		model = strings.TrimSpace(modelInput)

		baseURL := ""
		if providerName == "custom" {
			fmt.Print("Base URL: ")
			urlInput, _ := reader.ReadString('\n')
			baseURL = strings.TrimSpace(urlInput)
		}

		w.cfg.LLM.Providers = append(w.cfg.LLM.Providers, types.ProviderConfig{
			Name:     providerName,
			APIKey:   apiKey,
			BaseURL:  baseURL,
			Model:    model,
			Enabled:  true,
			Priority: priority,
			Timeout:  30,
		})
		w.cfg.LLM.Fallback = append(w.cfg.LLM.Fallback, providerName)

		priority++

		fmt.Print("Add another provider? (y/n): ")
		input, _ = reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))
	}

	fmt.Println()
	fmt.Printf("Configured %d provider(s)\n", len(w.cfg.LLM.Providers))
	return nil
}

func (w *SetupWizard) configureTokenBudget() {
	fmt.Println()
	fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD700")).Render("=== Token Budget Configuration ==="))
	fmt.Println("Configure budget alerts to avoid running out of tokens mid-task.")
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Monthly token budget (e.g., 100000): ")
	budgetInput, _ := reader.ReadString('\n')
	budgetInput = strings.TrimSpace(budgetInput)

	w.cfg.TokenBudget = types.TokenBudgetConfig{
		WarningThreshold:  0.7,
		CriticalThreshold: 0.9,
		AlertChannels:     []string{"log", "tui"},
		PerProvider:       true,
	}

	if budgetInput != "" {
		fmt.Sscanf(budgetInput, "%d", &w.cfg.TokenBudget.TotalBudget)
	}

	fmt.Println()
	fmt.Println("Token budget configured:")
	fmt.Printf("  Budget: %d tokens\n", w.cfg.TokenBudget.TotalBudget)
	fmt.Printf("  Warning at: %.0f%%\n", w.cfg.TokenBudget.WarningThreshold*100)
	fmt.Printf("  Critical at: %.0f%%\n", w.cfg.TokenBudget.CriticalThreshold*100)
}

func (w *SetupWizard) configureWorkspace() error {
	fmt.Println()
	fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD700")).Render("=== Workspace Configuration ==="))
	fmt.Println("Set your default workspace directory.")
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Workspace directory (press Enter for current directory): ")
	workspaceInput, _ := reader.ReadString('\n')
	workspaceInput = strings.TrimSpace(workspaceInput)

	if workspaceInput == "" {
		wd, _ := os.Getwd()
		w.cfg.Workspace = wd
	} else {
		if _, err := os.Stat(workspaceInput); os.IsNotExist(err) {
			fmt.Printf("Directory does not exist. Create it? (y/n): ")
			createInput, _ := reader.ReadString('\n')
			createInput = strings.TrimSpace(strings.ToLower(createInput))
			if createInput == "y" || createInput == "yes" {
				if err := os.MkdirAll(workspaceInput, 0755); err != nil {
					return fmt.Errorf("create directory: %w", err)
				}
				w.cfg.Workspace = workspaceInput
			}
		} else {
			w.cfg.Workspace = workspaceInput
		}
	}

	fmt.Printf("  Workspace: %s\n", w.cfg.Workspace)
	return nil
}

func (w *SetupWizard) configureGateway() {
	fmt.Println()
	fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD700")).Render("=== Gateway Configuration (Optional) ==="))
	fmt.Println("Connect messaging platforms: Telegram, Discord, Slack, WhatsApp.")
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enable gateway? (y/n): ")
	gatewayInput, _ := reader.ReadString('\n')
	gatewayInput = strings.TrimSpace(strings.ToLower(gatewayInput))

	if gatewayInput != "y" && gatewayInput != "yes" {
		w.cfg.Gateway.Enabled = false
		return
	}

	w.cfg.Gateway.Enabled = true

	platforms := []string{"telegram", "discord", "slack", "whatsapp"}
	fmt.Println("Available platforms:")
	for i, p := range platforms {
		fmt.Printf("  %d. %s\n", i+1, p)
	}
	fmt.Println()

	fmt.Print("Select platforms (comma-separated, e.g. 1,2): ")
	platformInput, _ := reader.ReadString('\n')
	platformInput = strings.TrimSpace(platformInput)

	w.cfg.Gateway.Platforms = nil

	for _, p := range strings.Split(platformInput, ",") {
		p = strings.TrimSpace(p)
		idx := 0
		if _, err := fmt.Sscanf(p, "%d", &idx); err == nil && idx > 0 && idx <= len(platforms) {
			w.cfg.Gateway.Platforms = append(w.cfg.Gateway.Platforms, platforms[idx-1])
		}
	}

	if len(w.cfg.Gateway.Platforms) == 0 {
		fmt.Println("  No platforms selected.")
		return
	}

	for _, plat := range w.cfg.Gateway.Platforms {
		fmt.Printf("\nConfiguring %s:\n", plat)
		switch plat {
		case "telegram":
			fmt.Print("  Bot Token: ")
			tokenInput, _ := reader.ReadString('\n')
			w.cfg.Gateway.Telegram = types.TelegramConfig{
				BotToken: strings.TrimSpace(tokenInput),
			}
		case "discord":
			fmt.Print("  Bot Token: ")
			tokenInput, _ := reader.ReadString('\n')
			w.cfg.Gateway.Discord = types.DiscordConfig{
				BotToken: strings.TrimSpace(tokenInput),
			}
		case "slack":
			fmt.Print("  Bot Token: ")
			tokenInput, _ := reader.ReadString('\n')
			fmt.Print("  Signing Secret: ")
			secretInput, _ := reader.ReadString('\n')
			w.cfg.Gateway.Slack = types.SlackConfig{
				BotToken:      strings.TrimSpace(tokenInput),
				SigningSecret: strings.TrimSpace(secretInput),
			}
		case "whatsapp":
			fmt.Print("  Phone Number ID: ")
			phoneInput, _ := reader.ReadString('\n')
			fmt.Print("  Access Token: ")
			tokenInput, _ := reader.ReadString('\n')
			w.cfg.Gateway.WhatsApp = types.WhatsAppConfig{
				PhoneNumberID: strings.TrimSpace(phoneInput),
				AccessToken:   strings.TrimSpace(tokenInput),
			}
		}
	}

	fmt.Println()
	fmt.Printf("Gateway enabled with: %s\n", strings.Join(w.cfg.Gateway.Platforms, ", "))
}

func (w *SetupWizard) IsComplete() bool {
	return w.complete
}

func (w *SetupWizard) GetMode() string {
	return w.mode
}

func (w *SetupWizard) GetConfig() *types.Config {
	return w.cfg
}
