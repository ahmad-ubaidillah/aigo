package main

import (
	"fmt"
	"os"

	"github.com/ahmad-ubaidillah/aigo/internal/cli"
	"github.com/spf13/cobra"
)

const version = "1.5.0"

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
- Agent Roles: Aizen, Atlas, Cody, Nova, Testa

Usage:
  aigo                  Interactive TUI mode
  aigo run "fix bug"    Execute a task
  aigo setup            First-run setup (install OpenCode + inject superpowers)
  aigo providers        List configured LLM providers
  aigo budget           Show token usage and alerts
  aigo agents           List available agent roles
  aigo doctor           Diagnose issues`,
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := cli.LoadConfig(configPath)
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}
			fmt.Println("Aigo v" + version + " - Never-Die Architecture")
			fmt.Println()
			fmt.Println("V1.5 Features:")
			fmt.Println("  aigo providers        # List configured LLM providers")
			fmt.Println("  aigo budget           # Show token usage")
			fmt.Println("  aigo agents           # List agent roles")
			fmt.Println("  aigo install opencode # Install OpenCode")
			fmt.Println()
			fmt.Println("Quick start:")
			fmt.Println("  aigo setup           # Install OpenCode + OMO superpowers")
			fmt.Println("  aigo run \"fix bug\" # Execute a task")
			fmt.Println("  aigo tui           # Interactive mode")
			return nil
		},
	}

	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "", "config file path")
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

	doctorCmd().Flags().Bool("auto-install", false, "Automatically install missing dependencies")
	setupCmd().Flags().Bool("force", false, "Force reinstall even if already installed")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
