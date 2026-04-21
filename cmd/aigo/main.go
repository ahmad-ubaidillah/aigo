// Aigo — AI Agent in Go
// The Hermes v2 runtime, built for speed, simplicity, and power.
package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/hermes-v2/aigo/internal/agent"
	"github.com/hermes-v2/aigo/internal/autonomy"
	"github.com/hermes-v2/aigo/internal/browsertools"
	"github.com/hermes-v2/aigo/internal/autonomytools"
	"github.com/hermes-v2/aigo/internal/channels/discord"
	"github.com/hermes-v2/aigo/internal/channels/slack"
	"github.com/hermes-v2/aigo/internal/channels/telegram"
	"github.com/hermes-v2/aigo/internal/channels/websocket"
	"github.com/hermes-v2/aigo/internal/channels/whatsapp"
	"github.com/hermes-v2/aigo/internal/actionlog"
	"github.com/hermes-v2/aigo/internal/codex"
	"github.com/hermes-v2/aigo/internal/config"
	"github.com/hermes-v2/aigo/internal/diffsandbox"
	"github.com/hermes-v2/aigo/internal/cron"
	"github.com/hermes-v2/aigo/internal/crontools"
	"github.com/hermes-v2/aigo/internal/diary"
	"github.com/hermes-v2/aigo/internal/engramtools"
	"github.com/hermes-v2/aigo/internal/evolution"
	"github.com/hermes-v2/aigo/internal/evolutiontools"
	"github.com/hermes-v2/aigo/internal/gateway"
	"github.com/hermes-v2/aigo/internal/git"
	"github.com/hermes-v2/aigo/internal/learntools"
	"github.com/hermes-v2/aigo/internal/memory"
	"github.com/hermes-v2/aigo/internal/memory/engram"
	"github.com/hermes-v2/aigo/internal/memory/fts5pkg"
	"github.com/hermes-v2/aigo/internal/memory/pyramid"
	"github.com/hermes-v2/aigo/internal/memory/project"
	"github.com/hermes-v2/aigo/internal/multiagenttools"
	"github.com/hermes-v2/aigo/internal/persona"
	"github.com/hermes-v2/aigo/internal/personatools"
	"github.com/hermes-v2/aigo/internal/plan"
	"github.com/hermes-v2/aigo/internal/planning"
	"github.com/hermes-v2/aigo/internal/providers"
	"github.com/hermes-v2/aigo/internal/memory/vector"
	"github.com/hermes-v2/aigo/internal/mcp"
	"github.com/hermes-v2/aigo/internal/pyramidtools"
	"github.com/hermes-v2/aigo/internal/router"
	"github.com/hermes-v2/aigo/internal/routertools"
	"github.com/hermes-v2/aigo/internal/session"
	"github.com/hermes-v2/aigo/internal/skillhub"
	"github.com/hermes-v2/aigo/internal/skillhubtools"
	"github.com/hermes-v2/aigo/internal/subagent"
	"github.com/hermes-v2/aigo/internal/subagenttools"
	"github.com/hermes-v2/aigo/internal/tools"
	"github.com/hermes-v2/aigo/internal/vectortools"
	"github.com/hermes-v2/aigo/internal/vision"
	"github.com/hermes-v2/aigo/internal/webtools"
	"github.com/hermes-v2/aigo/internal/webui"
)

const version = "0.3.0"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(0)
	}

	cmd := os.Args[1]
	switch cmd {
	case "chat":
		cmdChat()
	case "start":
		cmdStart()
	case "skills":
		cmdSkills(os.Args[2:])
	case "version":
		fmt.Printf("aigo %s\n", version)
	case "help", "--help", "-h":
		printUsage()
	default:
		// Treat as a one-shot query
		cmdQuery(strings.Join(os.Args[1:], " "))
	}
}

func printUsage() {
	fmt.Println(`Aigo — AI Agent in Go (Hermes v2)

Usage:
  aigo chat                    Start interactive chat (CLI)
  aigo start                   Start gateway server (all channels)
  aigo <message>               One-shot query
  aigo version                 Show version
  aigo help                    Show this help

Environment:
  AIGO_API_KEY       API key for the default provider
  AIGO_BASE_URL      Base URL for the API
  AIGO_MODEL         Model name (default: gpt-4o-mini)
  AIGO_PROVIDER      Provider name (default: openai)
  OPENAI_API_KEY     Fallback API key
  ANTHROPIC_API_KEY  Anthropic API key

Config: ~/.aigo/config.yaml`)
}

func cmdChat() {
	cfg := loadConfig()
	pm := buildProviders(cfg)
	reg := buildTools(cfg)
	scheduler := buildCronScheduler(reg)
	_ = scheduler
	mem := buildMemory(cfg)
	pyr := buildPyramid(cfg)

	// Persona
	baseDir := filepath.Join(os.Getenv("HOME"), ".aigo")
	personaMgr := persona.New(filepath.Join(baseDir, "persona"))
	personatools.RegisterPersonaTools(reg, personaMgr)
	log.Printf("👤 Persona manager initialized")

	// Engram — structured memory (works alongside pyramid)
	var engBackend *engram.Backend
	engMem := buildEngram(cfg)
	if engMem != nil {
		engBackend = engMem.Backend()
		engramtools.RegisterEngramTools(reg, engBackend)
		log.Printf("🧠 Engram structured memory active")
		// Start a session for this chat
		if err := engBackend.StartSession(""); err != nil {
			log.Printf("Engram session start error: %v", err)
		}
	}

	// Inject memory context into system prompt
	systemPrompt := agent.DefaultSystemPrompt()
	if engMem != nil {
		// Engram context takes priority (structured observations)
		if ctx := engMem.Context(); ctx != "" {
			systemPrompt = ctx + "\n\n" + systemPrompt
		}
	} else if mem != nil {
		if ctx := mem.Context(); ctx != "" {
			systemPrompt = ctx + "\n\n" + systemPrompt
		}
	}

	// Inject persona context into system prompt
	activeProfile := personaMgr.GetActive()
	if activeProfile.Role != "" || activeProfile.Tone != "" {
		personaPrompt := personaMgr.BuildSystemPrompt()
		systemPrompt = personaPrompt + "\n\n" + systemPrompt
	}

	a := agent.New(pm, reg, cfg.Agent.MaxIterations, cfg.Agent.MaxTokens, systemPrompt)

	// Wire planning system (Prometheus/Metis/Momus) into agent
	planner := planning.NewPlanner()
	planner.SetLLMProvider(planning.NewPlanningLLMProvider(pm, cfg.Provider.Default))
	a.SetPlanner(planner)
	metis := planning.NewMetis()
	a.SetMetis(metis)
	momus := planning.NewMomus()
	a.SetMomus(momus)
	log.Printf("📋 Planning system wired (LLM-enabled: true)")

	// Wire pyramid memory into agent
	if pyr != nil {
		a.SetPyramid(pyr)
		pyramidtools.RegisterPyramidTools(reg, pyr)
	}

	// Evolution — self-improvement system
	projectDir := getProjectDir()
	evolMgr := evolution.New(projectDir)
	evolutiontools.RegisterEvolutionTools(reg, evolMgr)
	log.Printf("🔧 Evolution system initialized: %s", projectDir)

	// Multi-agent roundtable
	roundtableBrain := func(prompt string) (string, error) {
		result, err := a.Run(context.Background(), prompt)
		if err != nil {
			return "", err
		}
		return result.Response, nil
	}
	multiagenttools.RegisterMultiAgentTools(reg, roundtableBrain)
	log.Printf("🎭 Multi-agent roundtable tools registered")

	// Phase 10: Vector memory (sqlite-vec)
	vecStore, err := vector.New(filepath.Join(baseDir, "memory", "vector"))
	if err != nil {
		log.Printf("⚠️ Vector memory init error: %v", err)
	} else {
		vectortools.RegisterVectorTools(reg, vecStore)
		log.Printf("🧬 Vector memory active (256-dim SimHash)")
	}

	// Phase 8: Semantic router
	semanticRouter := router.New(router.Config{
		Enabled:      true,
		DefaultModel: cfg.Provider.Model,
		AutoClassify: true,
	})
	routertools.RegisterRouterTools(reg, semanticRouter)
	log.Printf("🔀 Semantic router active")

	// Phase 9: Smart sub-agents (OMO-inspired)
	brainFunc := func(ctx context.Context, systemPrompt, userPrompt string, maxTokens int) (string, error) {
		result, err := a.Run(ctx, userPrompt)
		if err != nil {
			return "", err
		}
		return result.Response, nil
	}
	orch := subagent.NewOrchestrator(brainFunc, func(msg string) { fmt.Println(msg) })
	subagenttools.RegisterSubAgentTools(reg, orch, brainFunc)
	log.Printf("🤖 Sub-agent system active (Sisyphus/Hephaestus/Oracle/Explore)")

	// Phase 7: Skill hub integration (OnlineHub with SQLite FTS5 + Smithery + GitHub)
	skillHub, err := skillhub.NewOnlineHub("")
	if err != nil {
		log.Printf("Skill hub warning: %v", err)
	} else {
		skillhubtools.RegisterSkillHubTools(reg, skillHub)
		stats := skillHub.Stats()
		log.Printf("📦 Skill hub: %v skills indexed (%d sources)", stats["total_indexed"], len(skillHub.ListSources()))
	}

	projectDir = getProjectDir()
	var err2 error
	if err2 = git.RegisterGitTools(reg, projectDir); err2 != nil {
		log.Printf("⚠️ Git tools warning: %v", err2)
	} else {
		log.Printf("📝 Git tools active")
	}

	planBasePath := filepath.Join(baseDir, "memory", "plans")
	if err2 = plan.RegisterPlanTools(reg, planBasePath); err2 != nil {
		log.Printf("⚠️ Plan tools warning: %v", err2)
	} else {
		log.Printf("📋 Plan tools active")
	}

	projBasePath := filepath.Join(baseDir, "memory", "project")
	if err2 = project.RegisterProjectMemoryTools(reg, projBasePath); err2 != nil {
		log.Printf("⚠️ Project memory warning: %v", err2)
	} else {
		log.Printf("🗂️ Project memory active")
	}

	codexBasePath := filepath.Join(baseDir, "codex")
	if err2 = codex.RegisterCodexTools(reg, codexBasePath); err2 != nil {
		log.Printf("⚠️ Codex warning: %v", err2)
	} else {
		log.Printf("🔍 Codex active")
	}

	actionlogBasePath := filepath.Join(baseDir, "memory", "actions")
	if err2 = actionlog.RegisterActionLogTools(reg, actionlogBasePath); err2 != nil {
		log.Printf("⚠️ Action log warning: %v", err2)
	} else {
		log.Printf("📜 Action log active")
	}

	sandboxBasePath := filepath.Join(baseDir, "memory", "sandbox")
	if err2 = diffsandbox.RegisterSandboxTools(reg, sandboxBasePath); err2 != nil {
		log.Printf("⚠️ Diff sandbox warning: %v", err2)
	} else {
		log.Printf("🏖️ Diff sandbox active")
	}

	vision.RegisterVisionTools(reg)
	log.Printf("👁️ Vision pipeline active")

	fmt.Println("🦞 Aigo Chat — type 'exit' to quit, 'clear' to clear screen")
	fmt.Printf("   Provider: %s | Model: %s | Tools: %d\n\n",
		cfg.Provider.Default, cfg.Provider.Model, reg.Count())

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("you> ")
		if !scanner.Scan() {
			break
		}
		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}
		if input == "exit" || input == "quit" {
			fmt.Println("Goodbye! 👋")
			break
		}
		if input == "clear" {
			fmt.Print("\033[2J\033[H")
			continue
		}

		ctx := context.Background()
		result, err := a.Run(ctx, input)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		fmt.Printf("\naigo> %s\n", result.Response)
		fmt.Printf("      [%d steps, %d tokens, %s]\n\n",
			result.Steps, result.Usage.TotalTokens, result.Duration.Round(time.Millisecond))

		// Save to daily memory
		if mem != nil && result.Usage.TotalTokens > 0 {
			mem.SaveDaily(fmt.Sprintf("Q: %s\nA: %s", truncate(input, 100), truncate(result.Response, 200)))
		}

		// Save to engram as structured observation
		if engBackend != nil && result.Usage.TotalTokens > 0 {
			engBackend.SaveObservation("conversation", truncate(input, 120), result.Response, "")
		}
	}

	// End engram session
	if engBackend != nil {
		engBackend.EndSession("Chat session ended")
	}
}

func cmdStart() {
	cfg := loadConfig()
	pm := buildProviders(cfg)
	reg := buildTools(cfg)
	scheduler := buildCronScheduler(reg)
	mem := buildMemory(cfg)
	pyr := buildPyramid(cfg)

	// Diary
	baseDir := filepath.Join(os.Getenv("HOME"), ".aigo")
	d := diary.New(filepath.Join(baseDir, "diary"))
	log.Printf("📝 Diary initialized")

	// Persona
	personaMgr := persona.New(filepath.Join(baseDir, "persona"))
	personatools.RegisterPersonaTools(reg, personaMgr)
	log.Printf("👤 Persona manager initialized")

	// Engram — structured memory
	var engBackend *engram.Backend
	engMem := buildEngram(cfg)
	if engMem != nil {
		engBackend = engMem.Backend()
		engramtools.RegisterEngramTools(reg, engBackend)
		log.Printf("🧠 Engram structured memory active")
	}

	systemPrompt := agent.DefaultSystemPrompt()
	if engMem != nil {
		if ctx := engMem.Context(); ctx != "" {
			systemPrompt = ctx + "\n\n" + systemPrompt
		}
	} else if mem != nil {
		if ctx := mem.Context(); ctx != "" {
			systemPrompt = ctx + "\n\n" + systemPrompt
		}
	}

	// Inject persona context into system prompt
	activeProfile := personaMgr.GetActive()
	if activeProfile.Role != "" || activeProfile.Tone != "" {
		personaPrompt := personaMgr.BuildSystemPrompt()
		systemPrompt = personaPrompt + "\n\n" + systemPrompt
	}

	a := agent.New(pm, reg, cfg.Agent.MaxIterations, cfg.Agent.MaxTokens, systemPrompt)

	// Wire planning system (Prometheus/Metis/Momus) into agent
	planner := planning.NewPlanner()
	planner.SetLLMProvider(planning.NewPlanningLLMProvider(pm, cfg.Provider.Default))
	a.SetPlanner(planner)
	metis := planning.NewMetis()
	a.SetMetis(metis)
	momus := planning.NewMomus()
	a.SetMomus(momus)
	log.Printf("📋 Planning system wired (LLM-enabled: true)")

	// Wire pyramid memory into agent
	if pyr != nil {
		a.SetPyramid(pyr)
		pyramidtools.RegisterPyramidTools(reg, pyr)
	}

	// Evolution — self-improvement system
	projectDir := getProjectDir()
	evolMgr := evolution.New(projectDir)
	evolutiontools.RegisterEvolutionTools(reg, evolMgr)
	log.Printf("🔧 Evolution system initialized: %s", projectDir)

	// Multi-agent roundtable
	roundtableBrain := func(prompt string) (string, error) {
		result, err := a.Run(context.Background(), prompt)
		if err != nil {
			return "", err
		}
		return result.Response, nil
	}
	multiagenttools.RegisterMultiAgentTools(reg, roundtableBrain)
	log.Printf("🎭 Multi-agent roundtable tools registered")

	// Phase 10: Vector memory (sqlite-vec)
	vecStore, err := vector.New(filepath.Join(baseDir, "memory", "vector"))
	if err != nil {
		log.Printf("⚠️ Vector memory init error: %v", err)
	} else {
		vectortools.RegisterVectorTools(reg, vecStore)
		log.Printf("🧬 Vector memory active (256-dim SimHash)")
	}

	// Phase 8: Semantic router
	semanticRouter := router.New(router.Config{
		Enabled:      true,
		DefaultModel: cfg.Provider.Model,
		AutoClassify: true,
	})
	routertools.RegisterRouterTools(reg, semanticRouter)
	log.Printf("🔀 Semantic router active")

	// Phase 9: Smart sub-agents (OMO-inspired)
	subBrainFunc := func(ctx context.Context, systemPrompt, userPrompt string, maxTokens int) (string, error) {
		result, err := a.Run(ctx, userPrompt)
		if err != nil {
			return "", err
		}
		return result.Response, nil
	}
	orch := subagent.NewOrchestrator(subBrainFunc, func(msg string) { log.Printf("🤖 [SubAgent] %s", msg) })
	subagenttools.RegisterSubAgentTools(reg, orch, subBrainFunc)
	log.Printf("🤖 Sub-agent system active (Sisyphus/Hephaestus/Oracle/Explore)")

	// Phase 7: Skill hub integration (OnlineHub with SQLite FTS5 + Smithery + GitHub)
	skillHub, err := skillhub.NewOnlineHub("")
	if err != nil {
		log.Printf("Skill hub warning: %v", err)
	} else {
		skillhubtools.RegisterSkillHubTools(reg, skillHub)
		stats := skillHub.Stats()
		log.Printf("📦 Skill hub: %v skills indexed (%d sources)", stats["total_indexed"], len(skillHub.ListSources()))
	}

	// Autonomy
	var autoAgent *autonomy.AutonomousAgent
	if cfg.Autonomy.Enabled {
		autoCfg := autonomy.Config{
			AwakeMinMinutes:    cfg.Autonomy.AwakeMinMinutes,
			AwakeMaxMinutes:    cfg.Autonomy.AwakeMaxMinutes,
			SleepStart:         cfg.Autonomy.SleepStart,
			SleepEnd:           cfg.Autonomy.SleepEnd,
			Interests:          cfg.Autonomy.Interests,
			EnableNews:         cfg.Autonomy.EnableNews,
			EnableReflection:   cfg.Autonomy.EnableReflection,
			EnableSpontaneous:  cfg.Autonomy.EnableSpontaneous,
			EnableAutoCompress: false, // pyramid handles compression
		}

		// sendFunc: log messages (for now, channels may not be active)
		sendFunc := func(msg string) error {
			log.Printf("🤖 [Autonomy] Message: %s", msg)
			return nil
		}

		// brainFunc: call the agent's LLM
		brainFunc := func(prompt string) (string, error) {
			result, err := a.Run(context.Background(), prompt)
			if err != nil {
				return "", err
			}
			return result.Response, nil
		}

		autoAgent = autonomy.New(baseDir, autoCfg, sendFunc, brainFunc)
		log.Printf("🤖 Autonomy agent configured")
	}

	// Register autonomy tools
	runningCheck := func() bool {
		return autoAgent != nil
	}
	autonomytools.RegisterAutonomyTools(reg, d, runningCheck)

	gw := gateway.New(a)

	// Register channels
	if cfg.Channels.Telegram.Enabled && cfg.Channels.Telegram.Token != "" {
		gw.Register(telegram.New(cfg.Channels.Telegram.Token))
	}
	if cfg.Channels.Discord.Enabled && cfg.Channels.Discord.Token != "" {
		gw.Register(discord.New(cfg.Channels.Discord.Token))
	}
	if cfg.Channels.Slack.Enabled && cfg.Channels.Slack.AppToken != "" && cfg.Channels.Slack.BotToken != "" {
		gw.Register(slack.New(cfg.Channels.Slack.AppToken, cfg.Channels.Slack.BotToken))
	}
	if cfg.Channels.WebSocket.Enabled {
		gw.Register(websocket.New(cfg.Channels.WebSocket.Port, cfg.Channels.WebSocket.AuthToken))
	}
	if cfg.Channels.WhatsApp.Enabled && cfg.Channels.WhatsApp.AccountSid != "" && cfg.Channels.WhatsApp.AuthToken != "" {
		gw.Register(whatsapp.New(cfg.Channels.WhatsApp.AccountSid, cfg.Channels.WhatsApp.AuthToken, cfg.Channels.WhatsApp.FromNumber))
	}

	fmt.Printf("🦞 Aigo v%s starting...\n", version)
	fmt.Printf("   Provider: %s | Model: %s\n", cfg.Provider.Default, cfg.Provider.Model)
	fmt.Printf("   Tools: %d | Memory: %v\n", reg.Count(), cfg.Memory.Enabled)
	chCount := 0
	if cfg.Channels.Telegram.Enabled && cfg.Channels.Telegram.Token != "" {
		chCount++
	}
	if cfg.Channels.Discord.Enabled && cfg.Channels.Discord.Token != "" {
		chCount++
	}
	if cfg.Channels.Slack.Enabled && cfg.Channels.Slack.AppToken != "" && cfg.Channels.Slack.BotToken != "" {
		chCount++
	}
	if cfg.Channels.WebSocket.Enabled {
		chCount++
	}
	if cfg.Channels.WhatsApp.Enabled && cfg.Channels.WhatsApp.AccountSid != "" {
		chCount++
	}
	fmt.Printf("   Channels: %d registered\n\n", chCount)

	// Start Web UI
	if cfg.WebUI.Enabled {
		ui := webui.New(cfg.WebUI.Port, cfg, nil)

		// Initialize skill hub
		skillHub, err := skillhub.NewOnlineHub("")
		if err != nil {
			log.Printf("Skill hub warning: %v", err)
		} else {
			ui.SetSkillHub(skillHub)
			log.Printf("🛠️ Skill hub initialized (%d indexed)", skillHub.Stats()["total_indexed"])
		}

		// Build auto-memory
		var autoMem *session.AutoMemory
		if cfg.Memory.Enabled && cfg.Memory.UseFTS5 {
			fts5Store, err := fts5pkg.New(cfg.Memory.StoragePath)
			if err == nil {
				autoMem = session.NewAutoMemory(fts5Store)
				log.Printf("🧠 Auto-memory active (FTS5)")
			}
		}

		// Wire chat handler to the agent with auto-memory
		ui.SetChatHandler(func(msg string) (string, error) {
			// 1. Save user turn
			if autoMem != nil {
				autoMem.AddTurn("user", msg)

				// 2. Build session context: recent turns + related past
				ctxParts := []string{}
				if recent := autoMem.GetRecentContext(6); recent != "" {
					ctxParts = append(ctxParts, recent)
				}
				if related := autoMem.SearchRelated(msg, 3); related != "" {
					ctxParts = append(ctxParts, related)
				}
				if len(ctxParts) > 0 {
					a.SetSessionContext(strings.Join(ctxParts, "\n\n"))
				}
			}

			// 3. Run agent
			result, err := a.Run(context.Background(), msg)
			if err != nil {
				return "", err
			}

			// 4. Save assistant turn
			if autoMem != nil {
				autoMem.AddTurn("assistant", result.Response)
			}

			ui.IncrMessages()
			return result.Response, nil
		})

		// Wire streaming chat handler for SSE
		ui.SetChatStreamHandler(func(msg string, onChunk func(string)) (string, error) {
			// 1. Save user turn
			if autoMem != nil {
				autoMem.AddTurn("user", msg)
				ctxParts := []string{}
				if recent := autoMem.GetRecentContext(6); recent != "" {
					ctxParts = append(ctxParts, recent)
				}
				if related := autoMem.SearchRelated(msg, 3); related != "" {
					ctxParts = append(ctxParts, related)
				}
				if len(ctxParts) > 0 {
					a.SetSessionContext(strings.Join(ctxParts, "\n\n"))
				}
			}

			// 2. Run agent with streaming
			result, err := a.RunStream(context.Background(), msg, onChunk)
			if err != nil {
				return "", err
			}

			// 3. Save assistant turn
			if autoMem != nil && result.Response != "" {
				autoMem.AddTurn("assistant", result.Response)
			}

			ui.IncrMessages()
			return result.Response, nil
		})

		// Set live stats
		var channelNames []string
		if cfg.Channels.Telegram.Enabled && cfg.Channels.Telegram.Token != "" {
			channelNames = append(channelNames, "telegram")
		}
		if cfg.Channels.Discord.Enabled && cfg.Channels.Discord.Token != "" {
			channelNames = append(channelNames, "discord")
		}
		if cfg.Channels.Slack.Enabled && cfg.Channels.Slack.AppToken != "" {
			channelNames = append(channelNames, "slack")
		}
		if cfg.Channels.WebSocket.Enabled {
			channelNames = append(channelNames, "websocket")
		}
		ui.SetStats(channelNames, reg.Count(), cfg.Provider.Model, cfg.Provider.Default)

		go func() {
			if err := ui.Start(); err != nil {
				log.Printf("Web UI error: %v", err)
			}
		}()
	}

	// Graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("\nShutting down...")
		cancel()
	}()

	// Start cron scheduler
	go scheduler.Start(ctx)
	log.Printf("⏰ Cron scheduler started")

	// Start MCP server
	mcpServer := mcp.NewMCPServer(a, reg)
	mcpCfg := mcp.ServerConfig{
		Host: "127.0.0.1",
		Port: 3100,
	}
	mcpServer.Configure(mcpCfg)
	if err := mcpServer.Start(ctx); err != nil {
		log.Printf("MCP server warning: %v", err)
	} else {
		log.Printf("🖥️ MCP server started on http://127.0.0.1:3100/mcp")
	}

	// Start autonomous agent
	if autoAgent != nil {
		go autoAgent.Start()
		log.Printf("🤖 Autonomy agent started")
	}

	// Start channels — if no channels but WebUI enabled, just block on context
	if chCount > 0 {
		if err := gw.StartAll(ctx); err != nil {
			log.Printf("Gateway error: %v", err)
		}
	} else if cfg.WebUI.Enabled {
		fmt.Println("   Web UI only mode — waiting for shutdown signal...")
		<-ctx.Done()
	}
}

func cmdQuery(query string) {
	cfg := loadConfig()
	pm := buildProviders(cfg)
	reg := buildTools(cfg)

	a := agent.New(pm, reg, cfg.Agent.MaxIterations, cfg.Agent.MaxTokens, agent.DefaultSystemPrompt())

	ctx := context.Background()
	result, err := a.Run(ctx, query)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(result.Response)
}

func cmdSkills(args []string) {
	if len(args) == 0 {
		fmt.Println(`Aigo Skills — Skill Marketplace

Usage:
  aigo skills search <query>         Search for skills
  aigo skills info <identifier>       Show skill details
  aigo skills install <identifier>   Install a skill
  aigo skills list                   List installed skills
  aigo skills sync                   Sync online index (Smithery, GitHub)
  aigo skills sources                List online sources
  aigo skills popular                Show popular skills
  aigo skills remove <name>          Remove a skill
  aigo skills stats                  Show hub statistics`)
		return
	}

	hub, err := skillhub.NewOnlineHub("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer hub.Close()

	subcmd := args[0]
	switch subcmd {
	case "search":
		if len(args) < 2 {
			fmt.Println("Usage: aigo skills search <query>")
			return
		}
		query := strings.Join(args[1:], " ")
		results, err := hub.Search(query, 10)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Search error: %v\n", err)
			return
		}
		if len(results) == 0 {
			fmt.Printf("No skills found for: %s\n", query)
			return
		}
		fmt.Printf("Found %d skills:\n\n", len(results))
		for i, s := range results {
			fmt.Printf("%d. [%s] %s\n   %s\n   ID: %s\n\n",
				i+1, s.Source, s.Name, truncate(s.Description, 100), s.Identifier)
		}

	case "info":
		if len(args) < 2 {
			fmt.Println("Usage: aigo skills info <identifier>")
			return
		}
		identifier := strings.Join(args[1:], " ")
		skill, err := hub.FindByIdentifier(identifier)
		if err != nil {
			skill, err = hub.FindByName(identifier)
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "Skill not found: %s\n", identifier)
			return
		}
		fmt.Printf("📋 Skill Details:\n\n")
		fmt.Printf("  Name:        %s\n", skill.Name)
		fmt.Printf("  Description: %s\n", skill.Description)
		fmt.Printf("  Source:      %s\n", skill.Source)
		fmt.Printf("  Identifier:  %s\n", skill.Identifier)
		fmt.Printf("  Trust Level: %s\n", skill.TrustLevel)
		if skill.Repo != "" {
			fmt.Printf("  Repo:        %s\n", skill.Repo)
		}
		if skill.Path != "" {
			fmt.Printf("  Path:        %s\n", skill.Path)
		}
		if skill.Installs > 0 {
			fmt.Printf("  Installs:    %d\n", skill.Installs)
		}
		if skill.DetailURL != "" {
			fmt.Printf("  URL:         %s\n", skill.DetailURL)
		}
		if len(skill.Tags) > 0 {
			fmt.Printf("  Tags:        %s\n", strings.Join(skill.Tags, ", "))
		}

	case "install":
		if len(args) < 2 {
			fmt.Println("Usage: aigo skills install <identifier>")
			return
		}
		identifier := args[1]
		if err := hub.Install(identifier); err != nil {
			fmt.Fprintf(os.Stderr, "Install failed: %v\n", err)
			return
		}
		fmt.Printf("✅ Installed: %s\n", identifier)

	case "list":
		skills, err := hub.ListInstalled()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return
		}
		if len(skills) == 0 {
			fmt.Println("No skills installed. Use 'aigo skills search <query>' to find skills.")
			return
		}
		fmt.Printf("Installed skills (%d):\n\n", len(skills))
		for _, s := range skills {
			fmt.Printf("• %s (%s)\n", s.Name, s.Source)
		}

	case "sync":
		fmt.Println("🔄 Syncing online sources...")
		result, err := hub.SyncIndex()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Sync error: %v\n", err)
			return
		}
		fmt.Print(result.String())

	case "sources":
		sources := hub.ListSources()
		fmt.Println("📡 Online Skill Sources:")
		for _, s := range sources {
			status := "❌"
			if s.Enabled {
				status = "✅"
			}
			lastSync := "never"
			if s.LastSync != "" {
				lastSync = s.LastSync[:16]
			}
			fmt.Printf("%s %s\n   Type: %s | Last sync: %s\n\n",
				status, s.Name, s.Type, lastSync)
		}

	case "popular":
		skills, err := hub.PopularSkills(10)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return
		}
		if len(skills) == 0 {
			fmt.Println("No data yet. Run 'aigo skills sync' first.")
			return
		}
		fmt.Printf("🔥 Popular Skills:\n\n")
		for i, s := range skills {
			installs := ""
			if s.Installs > 0 {
				installs = fmt.Sprintf(" (%d installs)", s.Installs)
			}
			fmt.Printf("%d. %s%s\n   %s\n   Source: %s\n\n",
				i+1, s.Name, installs, truncate(s.Description, 80), s.Source)
		}

	case "remove":
		if len(args) < 2 {
			fmt.Println("Usage: aigo skills remove <name>")
			return
		}
		if err := hub.Remove(args[1]); err != nil {
			fmt.Fprintf(os.Stderr, "Remove failed: %v\n", err)
			return
		}
		fmt.Printf("✅ Removed: %s\n", args[1])

	case "stats":
		stats := hub.Stats()
		fmt.Printf("📊 Skill Hub Stats:\n")
		fmt.Printf("  Total indexed: %d\n", stats["total_indexed"])
		fmt.Printf("  Installed: %d\n", stats["installed"])
		fmt.Printf("  Categories: %d\n", stats["categories"])
		fmt.Printf("  Hermes index: %d\n", stats["hermes_index"])
		fmt.Printf("  DB: %s\n", stats["db_path"])
		fmt.Printf("  Skills dir: %s\n", stats["skills_dir"])

	default:
		fmt.Printf("Unknown command: %s\nUse 'aigo skills help' for usage.\n", subcmd)
	}
}

func loadConfig() config.Config {
	cfgPath := os.Getenv("AIGO_CONFIG")
	if cfgPath == "" {
		cfgPath = config.ConfigPath()
	}
	cfg, err := config.Load(cfgPath)
	if err != nil {
		log.Printf("Config warning: %v", err)
	}
	return cfg
}

func buildProviders(cfg config.Config) *providers.ProviderManager {
	pm := providers.NewProviderManager()

	// Register default provider with API key from config (if provided)
	if cfg.Provider.APIKey != "" {
		pm.RegisterWithAPIKey(cfg.Provider.Default, cfg.Provider.Model, cfg.Provider.APIKey)
	} else {
		pm.Register(cfg.Provider.Default, cfg.Provider.Model)
	}
	pm.SetDefault(cfg.Provider.Default)

	// Register additional providers
	for name, entry := range cfg.Provider.Providers {
		if name == cfg.Provider.Default {
			continue
		}
		model := entry.Model
		if model == "" {
			model = cfg.Provider.Model
		}
		if entry.APIKey != "" {
			pm.RegisterWithAPIKey(name, model, entry.APIKey)
		} else {
			pm.Register(name, model)
		}
	}

	return pm
}

func buildTools(cfg config.Config) *tools.Registry {
	reg := tools.NewRegistry()

	// Register built-in tools
	reg.Register(&tools.TerminalTool{})
	reg.Register(&tools.ReadFileTool{})
	reg.Register(&tools.WriteFileTool{})
	reg.Register(&tools.SearchFilesTool{})
	reg.Register(&tools.GetCurrentTimeTool{})

	// KV store
	kvPath := filepath.Join(os.Getenv("HOME"), ".aigo", "kv")
	reg.Register(tools.NewKVTool(kvPath))

	// Web tools (search + fetch)
	webtools.RegisterWebTools(reg)

	// Browser workflow tools (inspect, run, validate)
	browsertools.RegisterBrowserTools(reg)

	// Learning tools (learn, recall, knowledge_list)
	learnPath := filepath.Join(os.Getenv("HOME"), ".aigo", "knowledge")
	learntools.RegisterLearningTools(reg, learnPath)

	// Cron tools (if scheduler is available — registered separately in cmdStart)
	// Note: cron tools need scheduler instance, so they're added in cmdStart

	return reg
}

func buildCronScheduler(reg *tools.Registry) *cron.Scheduler {
	cronPath := filepath.Join(os.Getenv("HOME"), ".aigo", "cron", "jobs.json")
	scheduler := cron.New(cronPath, func(ctx context.Context, job cron.Job) (string, error) {
		// When a cron job fires, create a fresh agent and run the prompt
		log.Printf("⏰ Cron job '%s' executing: %s", job.Name, truncate(job.Prompt, 60))
		return "Job executed (cron handler not wired to agent)", nil
	})

	// Register cron tools
	crontools.RegisterCronTools(reg, scheduler)

	return scheduler
}

func buildMemory(cfg config.Config) memory.Backend {
	if !cfg.Memory.Enabled {
		return nil
	}
	mem, err := memory.NewBackend(cfg.Memory.StoragePath, cfg.Memory.UseFTS5)
	if err != nil {
		log.Printf("Memory init error: %v", err)
		return nil
	}
	return mem
}

func buildPyramid(cfg config.Config) *pyramid.Pyramid {
	if !cfg.Memory.PyramidEnabled {
		return nil
	}
	pyramidDir := filepath.Join(cfg.Memory.StoragePath, "pyramid")
	log.Printf("🧠 Pyramidal memory enabled: %s", pyramidDir)
	return pyramid.New(pyramidDir)
}

func buildEngram(cfg config.Config) *memory.EngramBackend {
	engramDir := filepath.Join(cfg.Memory.StoragePath, "engram")
	b, err := memory.NewEngramBackend(engramDir, "aigo")
	if err != nil {
		log.Printf("Engram init error: %v", err)
		return nil
	}
	return b
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func getProjectDir() string {
	home := os.Getenv("HOME")
	if home != "" {
		return filepath.Join(home, "aigo")
	}
	return "/mnt/projects/Aigo"
}
