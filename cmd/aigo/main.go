package main

import (
	"fmt"
	"os"

	"github.com/ahmad-ubaidillah/aigo/internal/cli"
	"github.com/ahmad-ubaidillah/aigo/internal/setup"
	"github.com/ahmad-ubaidillah/aigo/internal/tui"
	"github.com/spf13/cobra"
)

const version = "1.5.0"

var verbose bool

func main() {
	var configPath string

	rootCmd := &cobra.Command{
		Use:   "aigo",
		Short: "Aigo — Execute with Zen",
		Long: `Aigo is a minimal, fast, token-efficient AI agent platform.
With OMO superpowers. Orchestrates OpenCode for coding, handles everything else natively.

V1.5 Never-Die Architecture:
- Multi-Provider LLM Router with automatic fallback
- Token Budget Manager with cross-channel alerts
- Agent Roles: Aigo, Atlas, Cody, Nova, Testa

Usage:
  aigo                  Interactive TUI mode
  aigo run "fix bug"    Execute a task
  aigo setup            First-run setup (install OpenCode + inject superpowers)
  aigo providers        List configured LLM providers
  aigo budget           Show token usage and alerts
  aigo doctor           Diagnose issues`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if cmd.Name() == "help" || cmd.Name() == "completion" || cmd.Name() == "setup" {
				return nil
			}
			if !cli.ConfigExists() {
				fmt.Println("No config found. Running first-time setup...")
				wizard := setup.NewSetupWizard()
				if err := wizard.Run(); err != nil {
					fmt.Printf("Setup cancelled or failed: %v\n", err)
					fmt.Println("You can run 'aigo setup' later to configure manually.")
				} else if wizard.IsComplete() {
					cfg := wizard.GetConfig()
					configPath := cli.GetDefaultConfigPath()
					if err := cli.SaveConfig(*cfg, configPath); err == nil {
						fmt.Printf("✓ Config saved to: %s\n", configPath)
					}
				}
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := cli.LoadConfig(configPath)
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}
			fmt.Println("Aigo v" + version + " - Never-Die Architecture")
			if verbose {
				fmt.Println("Verbose mode enabled")
			}
			fmt.Println()
			fmt.Println("V1.5 Features:")
			fmt.Println("  aigo tui             # Interactive TUI mode")
			fmt.Println("  aigo providers       # List configured LLM providers")
			fmt.Println("  aigo budget          # Show token usage")
			fmt.Println("  aigo agents          # List agent roles")
			fmt.Println("  aigo install opencode # Install OpenCode")
			fmt.Println("  aigo completion bash # Generate shell completion")
			fmt.Println()
			fmt.Println("Quick start:")
			fmt.Println("  aigo tui             # Start interactive mode")
			fmt.Println("  aigo setup           # Re-run setup wizard")
			fmt.Println("  aigo run \"fix bug\" # Execute a task")
			return nil
		},
	}

	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "", "config file path")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.Version = version

	rootCmd.AddCommand(runCmd())
	rootCmd.AddCommand(sessionCmd())
	rootCmd.AddCommand(memoryCmd())
	rootCmd.AddCommand(gatewayCmd())
	rootCmd.AddCommand(tasksCmd())
	rootCmd.AddCommand(setupCmd())
	rootCmd.AddCommand(doctorCmd())
	rootCmd.AddCommand(configCmd())
	rootCmd.AddCommand(skillCmd())
	rootCmd.AddCommand(cronCmd())
	rootCmd.AddCommand(providersCmd())
	rootCmd.AddCommand(budgetCmd())
	rootCmd.AddCommand(agentsCmd())
	rootCmd.AddCommand(installOpenCodeCmd())
	rootCmd.AddCommand(completionCmd())
	rootCmd.AddCommand(migrateCmd())

	doctorCmd().Flags().Bool("auto-install", false, "Automatically install missing dependencies")
	setupCmd().Flags().Bool("force", false, "Force reinstall even if already installed")

	rootCmd.AddCommand(tuiCmd())

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func tuiCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "tui",
		Short: "Start interactive TUI mode",
		Long:  "Opens the interactive terminal UI for Aigo",
		RunE: func(cmd *cobra.Command, args []string) error {
			return tui.Run()
		},
	}
}
