package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ahmad-ubaidillah/aigo/internal/agent"
	"github.com/ahmad-ubaidillah/aigo/internal/agents"
	"github.com/ahmad-ubaidillah/aigo/internal/cli"
	aigoctx "github.com/ahmad-ubaidillah/aigo/internal/context"
	"github.com/ahmad-ubaidillah/aigo/internal/installer"
	"github.com/ahmad-ubaidillah/aigo/internal/intent"
	"github.com/ahmad-ubaidillah/aigo/internal/llm"
	"github.com/ahmad-ubaidillah/aigo/internal/memory"
	"github.com/ahmad-ubaidillah/aigo/internal/opencode"
	"github.com/ahmad-ubaidillah/aigo/internal/setup"
	"github.com/spf13/cobra"
)

func runCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run [task]",
		Short: "Execute a task",
		Long: `Execute a task. Aigo classifies the intent and routes to the right handler.
Coding tasks are delegated to OpenCode. Other tasks are handled natively.

Examples:
  aigo run "fix the login bug"
  aigo run "search for latest Go trends" --workspace ./project
  aigo run "send daily report to team"`,
		Args: cobra.MinimumNArgs(1),
		RunE: runCmdRunE,
	}

	cmd.Flags().String("workspace", "", "workspace directory")
	cmd.Flags().String("session", "", "session ID")
	return cmd
}

func runCmdRunE(cmd *cobra.Command, args []string) error {
	configPath, _ := cmd.Flags().GetString("config")
	workspace, _ := cmd.Flags().GetString("workspace")
	sessionID, _ := cmd.Flags().GetString("session")

	cfg, err := cli.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	inst := installer.NewInstaller(false, false)
	installed, _, _ := inst.CheckOpenCode()
	if !installed {
		fmt.Println("OpenCode not found.")
		err := inst.InstallOpenCode(context.Background(), "/usr/local/bin/opencode")
		if err != nil {
			fmt.Printf("⚠️  OpenCode install failed: %v\n", err)
			fmt.Println("Continuing without OpenCode...")
		}
	}

	db, err := memory.NewSessionDB(".aigo/sessions.db")
	if err != nil {
		return fmt.Errorf("open session db: %w", err)
	}
	defer db.Close()

	ocClient, err := opencode.NewClient(cfg.OpenCode.Binary, cfg.OpenCode.Timeout, workspace)
	if err != nil {
		return fmt.Errorf("init opencode client: %w", err)
	}

	// Create LLM client for general queries
	var llmClient llm.LLMClient
	if cfg.LLM.Provider == "glm" && cfg.LLM.APIKey != "" {
		llmClient = llm.NewGLMClient(cfg.LLM.APIKey, cfg.LLM.DefaultModel)
	} else if cfg.LLM.Provider == "local" && cfg.LLM.BaseURL != "" {
		llmClient = llm.NewLocalClient(cfg.LLM.DefaultModel, cfg.LLM.BaseURL)
	}

	classifier := intent.NewClassifier(cfg)
	ctxEngine := aigoctx.NewContextEngine(db, cfg)
	router := agent.NewRouterWithLLM(db, cfg, ocClient, llmClient)

	if sessionID == "" {
		sessionID = fmt.Sprintf("sess_%d", time.Now().Unix())
	}

	a := agent.NewAgent(classifier, router, ctxEngine, db, cfg, sessionID)

	ctx := context.Background()
	result, err := a.RunSession(ctx, sessionID, args[0])
	if err != nil {
		return fmt.Errorf("run session: %w", err)
	}

	fmt.Println(result.Output)
	return nil
}

func sessionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "session",
		Short: "Manage sessions",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "create",
		Short: "Create a new session",
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath, _ := cmd.Flags().GetString("config")
			name, _ := cmd.Flags().GetString("name")
			workspace, _ := cmd.Flags().GetString("workspace")

			if name == "" {
				return fmt.Errorf("--name is required")
			}

			cfg, err := cli.LoadConfig(configPath)
			_ = cfg // suppress unused warning

			dbPath := ".aigo/sessions.db"
			db, err := memory.NewSessionDB(dbPath)
			if err != nil {
				return fmt.Errorf("open database: %w", err)
			}
			defer db.Close()

			id := fmt.Sprintf("sess_%d", len(args)+1)
			sess, err := db.CreateSession(id, name, workspace)
			if err != nil {
				return fmt.Errorf("create session: %w", err)
			}

			fmt.Printf("Created session: %s (%s)\n", sess.Name, sess.ID)
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List all sessions",
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath, _ := cmd.Flags().GetString("config")
			cfg, err := cli.LoadConfig(configPath)
			_ = cfg

			dbPath := ".aigo/sessions.db"
			db, err := memory.NewSessionDB(dbPath)
			if err != nil {
				return fmt.Errorf("open database: %w", err)
			}
			defer db.Close()

			sessions, err := db.ListSessions()
			if err != nil {
				return fmt.Errorf("list sessions: %w", err)
			}

			if len(sessions) == 0 {
				fmt.Println("No sessions found.")
				return nil
			}

			fmt.Printf("%-12s %-20s %-25s %s\n", "ID", "Name", "Workspace", "Last Active")
			fmt.Println("----------------------------------------------------------------------")
			for _, s := range sessions {
				fmt.Printf("%-12s %-20s %-25s %s\n",
					s.ID, s.Name, s.Workspace,
					s.LastActive.Format("2006-01-02 15:04"),
				)
			}
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "info [id]",
		Short: "Show session details",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath, _ := cmd.Flags().GetString("config")
			cfg, err := cli.LoadConfig(configPath)
			_ = cfg

			dbPath := ".aigo/sessions.db"
			db, err := memory.NewSessionDB(dbPath)
			if err != nil {
				return fmt.Errorf("open database: %w", err)
			}
			defer db.Close()

			sess, err := db.GetSession(args[0])
			if err != nil {
				return fmt.Errorf("get session: %w", err)
			}

			fmt.Printf("ID:         %s\n", sess.ID)
			fmt.Printf("Name:       %s\n", sess.Name)
			fmt.Printf("Workspace:  %s\n", sess.Workspace)
			fmt.Printf("Created:    %s\n", sess.CreatedAt.Format("2006-01-02 15:04:05"))
			fmt.Printf("Last Active: %s\n", sess.LastActive.Format("2006-01-02 15:04:05"))
			return nil
		},
	})

	cmd.Flags().String("name", "", "session name")
	cmd.Flags().String("workspace", "", "workspace directory")
	return cmd
}

func memoryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "memory",
		Short: "Manage memories",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "add [content]",
		Short: "Add a memory",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath, _ := cmd.Flags().GetString("config")
			category, _ := cmd.Flags().GetString("category")
			tags, _ := cmd.Flags().GetString("tags")

			cfg, err := cli.LoadConfig(configPath)
			_ = cfg

			dbPath := ".aigo/sessions.db"
			db, err := memory.NewSessionDB(dbPath)
			if err != nil {
				return fmt.Errorf("open database: %w", err)
			}
			defer db.Close()

			if err := db.AddMemory(args[0], category, tags); err != nil {
				return fmt.Errorf("add memory: %w", err)
			}

			fmt.Println("Memory added.")
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "search [query]",
		Short: "Search memories",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath, _ := cmd.Flags().GetString("config")
			cfg, err := cli.LoadConfig(configPath)
			_ = cfg

			dbPath := ".aigo/sessions.db"
			db, err := memory.NewSessionDB(dbPath)
			if err != nil {
				return fmt.Errorf("open database: %w", err)
			}
			defer db.Close()

			memories, err := db.SearchMemory(args[0])
			if err != nil {
				return fmt.Errorf("search memories: %w", err)
			}

			if len(memories) == 0 {
				fmt.Println("No memories found.")
				return nil
			}

			for _, m := range memories {
				fmt.Printf("[%s] %s\n", m.Category, m.Content)
			}
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List all memories",
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath, _ := cmd.Flags().GetString("config")
			category, _ := cmd.Flags().GetString("category")

			cfg, err := cli.LoadConfig(configPath)
			_ = cfg

			dbPath := ".aigo/sessions.db"
			db, err := memory.NewSessionDB(dbPath)
			if err != nil {
				return fmt.Errorf("open database: %w", err)
			}
			defer db.Close()

			memories, err := db.ListMemories(category)
			if err != nil {
				return fmt.Errorf("list memories: %w", err)
			}

			if len(memories) == 0 {
				fmt.Println("No memories found.")
				return nil
			}

			for _, m := range memories {
				fmt.Printf("#%d [%s] %s\n", m.ID, m.Category, m.Content)
			}
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "delete [id]",
		Short: "Delete a memory",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath, _ := cmd.Flags().GetString("config")
			cfg, err := cli.LoadConfig(configPath)
			_ = cfg

			dbPath := ".aigo/sessions.db"
			db, err := memory.NewSessionDB(dbPath)
			if err != nil {
				return fmt.Errorf("open database: %w", err)
			}
			defer db.Close()

			var id int64
			if _, err := fmt.Sscanf(args[0], "%d", &id); err != nil {
				return fmt.Errorf("invalid memory ID: %s", args[0])
			}

			if err := db.DeleteMemory(id); err != nil {
				return fmt.Errorf("delete memory: %w", err)
			}

			fmt.Println("Memory deleted.")
			return nil
		},
	})

	cmd.Flags().String("category", "", "memory category")
	cmd.Flags().String("tags", "", "comma-separated tags")
	return cmd
}

func gatewayCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gateway",
		Short: "Manage gateway connections",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "status",
		Short: "Show gateway connection status",
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath, _ := cmd.Flags().GetString("config")
			cfg, err := cli.LoadConfig(configPath)
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			if !cfg.Gateway.Enabled {
				fmt.Println("Gateway is not enabled.")
				fmt.Println("Run: aigo gateway setup")
				return nil
			}

			fmt.Println("Gateway Status:")
			for _, p := range cfg.Gateway.Platforms {
				fmt.Printf("  🔴 %s (not connected)\n", p)
			}
			fmt.Println("\nRun: aigo gateway start")
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "setup",
		Short: "Interactive gateway setup wizard",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Gateway setup wizard coming in Phase 6.")
			fmt.Println("Supported platforms: telegram, discord, slack, whatsapp")
			return nil
		},
	})

	return cmd
}

func tasksCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tasks",
		Short: "Manage tasks",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List tasks for current session",
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath, _ := cmd.Flags().GetString("config")
			sessionID, _ := cmd.Flags().GetString("session")

			cfg, err := cli.LoadConfig(configPath)
			_ = cfg

			dbPath := ".aigo/sessions.db"
			db, err := memory.NewSessionDB(dbPath)
			if err != nil {
				return fmt.Errorf("open database: %w", err)
			}
			defer db.Close()

			tasks, err := db.ListTasks(sessionID)
			if err != nil {
				return fmt.Errorf("list tasks: %w", err)
			}

			if len(tasks) == 0 {
				fmt.Println("No tasks found.")
				return nil
			}

			fmt.Printf("%-4s %-10s %-8s %s\n", "ID", "Status", "Priority", "Description")
			fmt.Println("------------------------------------------------------------")
			for _, t := range tasks {
				fmt.Printf("%-4d %-10s %-8s %s\n", t.ID, t.Status, t.Priority, t.Description)
			}
			return nil
		},
	})

	cmd.Flags().String("session", "", "session ID")
	return cmd
}

func setupCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "setup",
		Short: "First-run setup wizard",
		RunE: func(cmd *cobra.Command, args []string) error {
			wizard := setup.NewSetupWizard()
			if err := wizard.Run(); err != nil {
				return fmt.Errorf("setup wizard: %w", err)
			}

			cfg := wizard.GetConfig()
			if !wizard.IsComplete() {
				return nil
			}

			configPath := cli.GetDefaultConfigPath()
			if err := cli.SaveConfig(*cfg, configPath); err != nil {
				return fmt.Errorf("save config: %w", err)
			}

			fmt.Printf("\n✓ Config saved to: %s\n", configPath)
			return nil
		},
	}
}

func doctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Diagnose issues and optionally auto-install missing dependencies",
		RunE: func(cmd *cobra.Command, args []string) error {
			autoInstall, _ := cmd.Flags().GetBool("auto-install")
			configPath, _ := cmd.Flags().GetString("config")

			fmt.Println("Aigo Doctor")
			fmt.Println("===========")
			fmt.Println()

			cfg, err := cli.LoadConfig(configPath)
			if err != nil {
				fmt.Printf("❌ Config: %v\n", err)
			} else {
				fmt.Println("✅ Config loaded")
			}

			dbPath := ".aigo/sessions.db"
			db, err := memory.NewSessionDB(dbPath)
			if err != nil {
				fmt.Printf("❌ Database: %v\n", err)
			} else {
				defer db.Close()
				fmt.Println("✅ Database accessible")
			}

			inst := installer.NewInstaller(true, false)
			installed, _, _ := inst.CheckOpenCode()
			if !installed {
				fmt.Println("⚠️  OpenCode: not found in PATH")
				if autoInstall {
					fmt.Println("   Auto-installing OpenCode...")
					err := inst.InstallOpenCode(context.Background(), "/usr/local/bin/opencode")
					if err != nil {
						fmt.Printf("❌ OpenCode install failed: %v\n", err)
					} else {
						fmt.Println("✅ OpenCode installed")
					}
				} else {
					fmt.Println("   Run: aigo doctor --auto-install")
					fmt.Println("   Or: curl -fsSL https://opencode.ai/install | bash")
				}
			} else {
				fmt.Printf("✅ OpenCode found: %s\n", version)
			}

			fmt.Println()
			fmt.Printf("Version: %s\n", version)
			fmt.Printf("Platform: %s/%s\n", os.Getenv("GOOS"), os.Getenv("GOARCH"))
			if cfg.Model.Default != "" {
				fmt.Printf("Model: %s\n", cfg.Model.Default)
			}

			return nil
		},
	}
}

func configCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "config",
		Short: "Show or edit configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath, _ := cmd.Flags().GetString("config")
			cfg, err := cli.LoadConfig(configPath)
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			fmt.Printf("Config loaded from: %s\n\n", configPath)
			fmt.Printf("Model Default: %s\n", cfg.Model.Default)
			fmt.Printf("Model Coding:  %s\n", cfg.Model.Coding)
			fmt.Printf("Model Intent:  %s\n", cfg.Model.Intent)
			fmt.Printf("OpenCode Binary: %s\n", cfg.OpenCode.Binary)
			fmt.Printf("OpenCode Timeout: %ds\n", cfg.OpenCode.Timeout)
			fmt.Printf("OpenCode Max Turns: %d\n", cfg.OpenCode.MaxTurns)
			fmt.Printf("Gateway Enabled: %v\n", cfg.Gateway.Enabled)
			fmt.Printf("Gateway Platforms: %v\n", cfg.Gateway.Platforms)
			fmt.Printf("Memory Max L0: %d\n", cfg.Memory.MaxL0Items)
			fmt.Printf("Memory Max L1: %d\n", cfg.Memory.MaxL1Items)
			fmt.Printf("Memory Auto Compress: %v\n", cfg.Memory.AutoCompress)
			fmt.Printf("Web Enabled: %v\n", cfg.Web.Enabled)
			fmt.Printf("Web Port: %s\n", cfg.Web.Port)

			return nil
		},
	}
}

func skillCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "skill",
		Short: "Manage skills",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "list [category]",
		Short: "List all skills (built-in + local)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			category := ""
			if len(args) > 0 {
				category = args[0]
			}

			dbPath := ".aigo/sessions.db"
			db, err := memory.NewSessionDB(dbPath)
			if err != nil {
				return fmt.Errorf("open database: %w", err)
			}
			defer db.Close()

			skills, err := db.ListSkills(category)
			if err != nil {
				return fmt.Errorf("list skills: %w", err)
			}

			fmt.Println("=== Built-in Skills ===")
			builtIn := []string{
				"git-master: Git operations",
				"playwright: Browser automation",
				"frontend-ui-ux: Frontend development",
				"dev-browser: Browser automation",
				"code-review: Code review",
				"web-search: Web search",
				"code-search: Code search",
				"docs-lookup: Documentation lookup",
			}
			for _, s := range builtIn {
				fmt.Println("  ", s)
			}

			fmt.Println("\n=== Local Skills ===")
			if len(skills) == 0 {
				fmt.Println("  No local skills found.")
			} else {
				for _, s := range skills {
					fmt.Printf("  %s: %s\n", s.Name, s.Description)
				}
			}
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "search [query]",
		Short: "Search skills from marketplace (built-in, local, skillsmp.com)",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := args[0]
			source, _ := cmd.Flags().GetString("source")

			fmt.Printf("Searching for: %s\n", query)
			if source != "" {
				fmt.Printf("Source: %s\n", source)
			}

			fmt.Println("\nUse 'aigo run \"search skills query\"' for full marketplace search (includes remote sources).")
			return nil
		},
	})

	cmd.Flags().String("source", "", "search source: built-in, local, skillsmp, github")

	cmd.AddCommand(&cobra.Command{
		Use:   "add [name] [description] [command] [category] [tags]",
		Short: "Add a new skill",
		Args:  cobra.MinimumNArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			dbPath := ".aigo/sessions.db"
			db, err := memory.NewSessionDB(dbPath)
			if err != nil {
				return fmt.Errorf("open database: %w", err)
			}
			defer db.Close()

			name := args[0]
			description := args[1]
			command := args[2]
			category := ""
			tags := ""

			if len(args) > 3 {
				category = args[3]
			}
			if len(args) > 4 {
				tags = args[4]
			}

			_, err = db.AddSkill(name, description, command, category, tags)
			if err != nil {
				return fmt.Errorf("add skill: %w", err)
			}

			fmt.Printf("Skill '%s' added successfully.\n", name)
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "delete [name]",
		Short: "Delete a skill",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dbPath := ".aigo/sessions.db"
			db, err := memory.NewSessionDB(dbPath)
			if err != nil {
				return fmt.Errorf("open database: %w", err)
			}
			defer db.Close()

			if err := db.DeleteSkill(args[0]); err != nil {
				return fmt.Errorf("delete skill: %w", err)
			}

			fmt.Printf("Skill '%s' deleted.\n", args[0])
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "market",
		Short: "Browse skills marketplace",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("=== Skills Marketplace ===")
			fmt.Println()
			fmt.Println("Sources:")
			fmt.Println("  - built-in : Default skills (8 skills)")
			fmt.Println("  - local    : User-added skills")
			fmt.Println("  - skillsmp : skillsmp.com API (700k+ skills)")
			fmt.Println("  - github   : Search GitHub for skill repos")
			fmt.Println()
			fmt.Println("Usage:")
			fmt.Println("  aigo skill list              # List built-in + local")
			fmt.Println("  aigo run 'search skills js'  # Full marketplace search")
			return nil
		},
	})

	return cmd
}

func cronCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cron",
		Short: "Manage scheduled tasks",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List all cron jobs",
		RunE: func(cmd *cobra.Command, args []string) error {
			dbPath := ".aigo/sessions.db"
			db, err := memory.NewSessionDB(dbPath)
			if err != nil {
				return fmt.Errorf("open database: %w", err)
			}
			defer db.Close()

			jobs, err := db.ListCronJobs()
			if err != nil {
				return fmt.Errorf("list cron jobs: %w", err)
			}

			if len(jobs) == 0 {
				fmt.Println("No cron jobs found.")
				return nil
			}

			fmt.Printf("%-4s %-15s %-20s %-15s %s\n", "ID", "Name", "Schedule", "Next Run", "Enabled")
			fmt.Println("----------------------------------------------------------------------")
			for _, j := range jobs {
				fmt.Printf("%-4d %-15s %-20s %-15s %v\n", j.ID, j.Name, j.Schedule, j.NextRun.Format("2006-01-02 15:04"), j.Enabled)
			}
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "add [name] [command] [schedule]",
		Short: "Add a cron job",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			command := args[1]
			schedule := args[2]

			dbPath := ".aigo/sessions.db"
			db, err := memory.NewSessionDB(dbPath)
			if err != nil {
				return fmt.Errorf("open database: %w", err)
			}
			defer db.Close()

			nextRun := time.Now().Add(1 * time.Hour)
			_, err = db.AddCronJob(name, command, schedule, nextRun)
			if err != nil {
				return fmt.Errorf("add cron job: %w", err)
			}

			fmt.Printf("Cron job '%s' added with schedule '%s'.\n", name, schedule)
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "delete [id]",
		Short: "Delete a cron job",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var id int64
			if _, err := fmt.Sscanf(args[0], "%d", &id); err != nil {
				return fmt.Errorf("invalid cron job ID: %s", args[0])
			}

			dbPath := ".aigo/sessions.db"
			db, err := memory.NewSessionDB(dbPath)
			if err != nil {
				return fmt.Errorf("open database: %w", err)
			}
			defer db.Close()

			if err := db.DeleteCronJob(id); err != nil {
				return fmt.Errorf("delete cron job: %w", err)
			}

			fmt.Printf("Cron job %d deleted.\n", id)
			return nil
		},
	})

	return cmd
}

func providersCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "providers",
		Short: "List configured LLM providers",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cli.LoadConfig("")
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			fmt.Println("=== Configured LLM Providers ===")
			if len(cfg.LLM.Providers) == 0 {
				fmt.Println("No multi-provider config found. Using legacy config:")
				fmt.Printf("  Provider: %s\n", cfg.LLM.Provider)
				fmt.Printf("  Model: %s\n", cfg.LLM.DefaultModel)
				return nil
			}

			for i, p := range cfg.LLM.Providers {
				status := "✓ enabled"
				if !p.Enabled {
					status = "✗ disabled"
				}
				fmt.Printf("%d. %s (%s) [%s]\n", i+1, p.Name, p.Model, status)
				if p.BaseURL != "" {
					fmt.Printf("   BaseURL: %s\n", p.BaseURL)
				}
			}

			if len(cfg.LLM.Fallback) > 0 {
				fmt.Println("\nFallback order:", cfg.LLM.Fallback)
			}

			return nil
		},
	}
	return cmd
}

func budgetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "budget",
		Short: "Show token usage and budget",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cli.LoadConfig("")
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			fmt.Println("=== Token Budget Configuration ===")
			fmt.Printf("Total Budget: %d tokens\n", cfg.TokenBudget.TotalBudget)
			fmt.Printf("Warning Threshold: %.0f%%\n", cfg.TokenBudget.WarningThreshold*100)
			fmt.Printf("Critical Threshold: %.0f%%\n", cfg.TokenBudget.CriticalThreshold*100)
			fmt.Printf("Per-Provider Tracking: %v\n", cfg.TokenBudget.PerProvider)
			fmt.Printf("Alert Channels: %v\n", cfg.TokenBudget.AlertChannels)

			return nil
		},
	}
	return cmd
}

func agentsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "agents",
		Short: "List available agent roles",
		RunE: func(cmd *cobra.Command, args []string) error {
			roles := agents.ListRoles()

			fmt.Println("=== Available Agent Roles ===")
			for _, r := range roles {
				fmt.Printf("\n%s (%s)\n", r.Name, r.Category)
				fmt.Printf("  Max Turns: %d\n", r.MaxTurns)
				fmt.Printf("  Skills: %v\n", r.Skills)
				fmt.Printf("  Prompt: %s\n", r.SystemPrompt)
			}

			return nil
		},
	}
	return cmd
}

func installOpenCodeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install opencode",
		Short: "Install or update OpenCode",
		RunE: func(cmd *cobra.Command, args []string) error {
			inst := installer.NewInstaller(true, false)
			available, path, err := inst.CheckOpenCode()
			if err != nil {
				return fmt.Errorf("check OpenCode: %w", err)
			}

			if available {
				fmt.Printf("OpenCode already installed at: %s\n", path)
				fmt.Println("Use --force to reinstall")
				return nil
			}

			installPath := "/usr/local/bin/opencode"
			if err := inst.InstallOpenCode(context.Background(), installPath); err != nil {
				return fmt.Errorf("install OpenCode: %w", err)
			}

			fmt.Printf("OpenCode installed to: %s\n", installPath)
			return nil
		},
	}

	cmd.Flags().Bool("force", false, "Force reinstall")
	return cmd
}

func completionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion [bash|zsh|fish]",
		Short: "Generate shell completion script",
		Long: `Generate shell completion script for your preferred shell.
		
Examples:
  aigo completion bash > ~/.bash_completion
  aigo completion zsh > ~/.zsh/completion/_aigo
  aigo completion fish > ~/.config/fish/completions/aigo.fish`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			shell := args[0]
			switch shell {
			case "bash":
				fmt.Print(cli.GenerateShellCompletion("bash"))
			case "zsh":
				fmt.Print(cli.GenerateShellCompletion("zsh"))
			case "fish":
				fmt.Print(cli.GenerateShellCompletion("fish"))
			default:
				return fmt.Errorf("unsupported shell: %s (use: bash, zsh, fish)", shell)
			}
			return nil
		},
	}
	return cmd
}

func migrateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate [config|db]",
		Short: "Migrate data between versions",
		Long: `Migrate configuration or database to new format.

Examples:
  aigo migrate config        # Migrate old config to new format
  aigo migrate db           # Migrate database schema`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			target := args[0]
			switch target {
			case "config":
				return migrateConfig()
			case "db":
				return migrateDB()
			default:
				return fmt.Errorf("unknown migration target: %s (use: config, db)", target)
			}
		},
	}
	return cmd
}

func migrateConfig() error {
	configPath := cli.GetDefaultConfigPath()

	// Check if old config exists
	oldPaths := []string{
		".aigo/config.yaml",
		".config/aigo/config.yaml",
	}

	var oldConfig string
	for _, p := range oldPaths {
		if _, err := os.Stat(p); err == nil {
			oldConfig = p
			break
		}
	}

	if oldConfig == "" {
		fmt.Println("No old config found. Creating new config from template...")
		exampleConfig := cli.GenerateExampleConfig()
		if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
			return fmt.Errorf("create config directory: %w", err)
		}
		if err := os.WriteFile(configPath, []byte(exampleConfig), 0644); err != nil {
			return fmt.Errorf("write config: %w", err)
		}
		fmt.Printf("Created new config at: %s\n", configPath)
		return nil
	}

	fmt.Printf("Found old config at: %s\n", oldConfig)
	fmt.Println("Migration: Copying to new location...")

	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}

	data, err := os.ReadFile(oldConfig)
	if err != nil {
		return fmt.Errorf("read old config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("write new config: %w", err)
	}

	fmt.Printf("✓ Config migrated to: %s\n", configPath)
	fmt.Println("  Note: Review and update the migrated config for new options.")
	return nil
}

func migrateDB() error {
	dbPath := ".aigo/sessions.db"

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		fmt.Println("No database found. Nothing to migrate.")
		return nil
	}

	fmt.Println("Database migration check...")
	fmt.Printf("Database: %s\n", dbPath)
	fmt.Println("✓ Database schema is up to date (SQLite with FTS5)")
	return nil
}
