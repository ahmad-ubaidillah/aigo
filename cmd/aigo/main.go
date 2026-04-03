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

	// Check and run setup if no config exists
	if !cli.ConfigExists() {
		fmt.Println("Welcome! Let's set up Aigo first...")
		wizard := setup.NewSetupWizard()
		if err := wizard.Run(); err != nil {
			fmt.Printf("Setup cancelled or failed: %v\n", err)
			fmt.Println("You can run 'aigo setup' later to configure manually.")
		} else if wizard.IsComplete() {
			cfg := wizard.GetConfig()
			configPath = cli.GetDefaultConfigPath()
			if err := cli.SaveConfig(*cfg, configPath); err == nil {
				fmt.Printf("✓ Config saved to: %s\n", configPath)
			}
		}
	}

	rootCmd := &cobra.Command{
		Use:   "aigo",
		Short: "Aigo — your buddy aigo",
		Long: `Aigo - your buddy aigo
Execute with Zen - AI coding partner that understands your project context.

Quick Start:
  aigo                  Start interactive mode (default)
  aigo tui              Start TUI mode
  aigo run "fix bug"    Execute a task
  aigo setup            Re-run setup wizard
  aigo providers        List configured LLM providers
  aigo budget           Show token usage and alerts
  aigo doctor           Diagnose issues`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Default: start interactive TUI mode
			return tui.Run()
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
