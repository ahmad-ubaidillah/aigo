// Package agent implements the ReAct agent loop with loop detection.
// Inspired by KrillClaw's FNV-1a loop detection and Zeph's subgoal-aware compaction.
package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/hermes-v2/aigo/internal/memory/pyramid"
	"github.com/hermes-v2/aigo/internal/planning"
	"github.com/hermes-v2/aigo/internal/providers"
	"github.com/hermes-v2/aigo/internal/tools"
)

type Agent struct {
	mu           sync.RWMutex
	providers    *providers.ProviderManager
	tools        *tools.Registry
	planner      *planning.Planner
	metis        *planning.Metis
	momus        *planning.Momus
	resolver     *planning.Resolver
	decisions    *planning.DecisionStore
	maxIter      int
	maxTokens    int
	systemPrompt string
	sessionCtx   string
	pyramid      *pyramid.Pyramid
	loopDetector *LoopDetector
	compressor   *Compressor
}

// LoopDetector tracks repeated tool calls using FNV-1a hashing.
type LoopDetector struct {
	ring        []uint64
	idx         int
	maxRepeats  int
	loopCount   int
}

// NewLoopDetector creates a loop detector.
func NewLoopDetector(maxRepeats, ringSize int) *LoopDetector {
	return &LoopDetector{
		ring:       make([]uint64, ringSize),
		idx:        0,
		maxRepeats: maxRepeats,
	}
}

func fnv1a(data string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(data))
	return h.Sum64()
}

// IsLoop checks if a tool call is a repeated loop.
func (ld *LoopDetector) IsLoop(toolName, toolArgs string) bool {
	// Skip idempotent tools
	if toolName == "get_current_time" || toolName == "kv" {
		return false
	}

	hash := fnv1a(toolName + "|" + toolArgs)

	repeats := 0
	for _, h := range ld.ring {
		if h == hash {
			repeats++
		}
	}

	ld.ring[ld.idx%len(ld.ring)] = hash
	ld.idx++

	if repeats >= ld.maxRepeats {
		ld.loopCount++
		return true
	}
	return false
}

func (ld *LoopDetector) Reset() {
	for i := range ld.ring {
		ld.ring[i] = 0
	}
	ld.idx = 0
	ld.loopCount = 0
}

func (ld *LoopDetector) ConsecutiveLoops() int {
	return ld.loopCount
}

// Result is the output of an agent conversation.
type Result struct {
	Response string
	Usage    providers.Usage
	Steps    int
	Duration time.Duration
}

// New creates a new agent.
func New(pm *providers.ProviderManager, reg *tools.Registry, maxIter, maxTokens int, systemPrompt string) *Agent {
	// Reserve 2000 tokens for response
	compressor := NewCompressor(maxTokens, 2000)
	return &Agent{
		providers:    pm,
		tools:        reg,
		maxIter:      maxIter,
		maxTokens:    maxTokens,
		systemPrompt: systemPrompt,
		loopDetector: NewLoopDetector(3, 8),
		compressor:   compressor,
	}
}

// SetPersona appends persona context (role + tone) to the system prompt.
func (a *Agent) SetPersona(role, tone string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	personaCtx := fmt.Sprintf("Persona: %s. Tone: %s.\n", role, tone)
	a.systemPrompt = personaCtx + a.systemPrompt
}

// SetPyramid wires the pyramid memory into the agent.
func (a *Agent) SetPyramid(p *pyramid.Pyramid) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.pyramid = p
	// Inject long-term context into system prompt
	if p != nil {
		if ctx := p.InjectContext(2000); ctx != "" {
			a.systemPrompt = ctx + "\n\n" + a.systemPrompt
		}
	}
}

// SetSessionContext updates the session context (called each turn).
func (a *Agent) SetSessionContext(ctx string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.sessionCtx = ctx
}

// SetMaxIter updates the maximum iteration limit.
func (a *Agent) SetMaxIter(n int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.maxIter = n
}

// SetPlanner wires the Planner (Prometheus) into the agent.
func (a *Agent) SetPlanner(p *planning.Planner) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.planner = p
}

// SetMetis wires the Metis gap analyzer into the agent.
func (a *Agent) SetMetis(m *planning.Metis) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.metis = m
}

// SetMomus wires the Momus reviewer into the agent.
func (a *Agent) SetMomus(m *planning.Momus) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.momus = m
}

// createPlan creates an execution plan using Prometheus/Meta/Momus pipeline.
func (a *Agent) createPlan(ctx context.Context, userMessage string) (*planning.Plan, error) {
	if a.planner == nil {
		return nil, nil
	}

	inputs := []string{userMessage}
	plan, err := a.planner.Plan(ctx, inputs)
	if err != nil {
		return nil, fmt.Errorf("planner error: %w", err)
	}

	// Run Metis gap analysis
	if a.metis != nil && plan != nil {
		gaps, gapErr := a.metis.AnalyzeGaps(ctx, plan)
		if gapErr == nil && len(gaps) > 0 {
			log.Printf("Metis found %d gaps in plan", len(gaps))
			// Try to resolve gaps
			if a.resolver != nil {
				for _, gap := range gaps {
					resolved := a.resolver.ResolveGap(ctx, gap)
					if resolved != nil && resolved.Applied {
						log.Printf("Resolved gap: %s", gap.Type)
					}
				}
			}
		}
	}

	// Run Momus review
	if a.momus != nil && plan != nil {
		review, reviewErr := a.momus.ReviewPlan(ctx, plan)
		if reviewErr == nil {
			log.Printf("Momus review score: %d (complete=%v)", review.Score, review.Complete)
			// Store decision if we have decision store
			if a.decisions != nil {
				a.decisions.Add("plan_review", fmt.Sprintf("score=%d", review.Score))
			}
		}
	}

	return plan, nil
}

// getFullSystemPrompt combines base prompt + session context.
func (a *Agent) getFullSystemPrompt() string {
	a.mu.RLock()
	defer a.mu.RUnlock()

	prompt := a.systemPrompt
	if a.sessionCtx != "" {
		prompt = a.sessionCtx + "\n\n" + prompt
	}
	return prompt
}

// Run executes the ReAct agent loop.
func (a *Agent) Run(ctx context.Context, userMessage string) (*Result, error) {
	start := time.Now()

	// Log user message to pyramid
	a.mu.RLock()
	p := a.pyramid
	a.mu.RUnlock()
	if p != nil {
		p.WriteRaw("user", userMessage)
	}

	// PLANNING PHASE: Create execution plan before execution
	plan, planErr := a.createPlan(ctx, userMessage)
	if planErr != nil {
		log.Printf("Planning failed: %v", planErr)
	}

	provider, err := a.providers.Get("")
	if err != nil {
		return nil, fmt.Errorf("get provider: %w", err)
	}

	// Add plan context to messages if available
	systemPrompt := a.getFullSystemPrompt()
	if plan != nil && len(plan.Steps) > 0 {
		systemPrompt += "\n\nExecution Plan:\n"
		for i, step := range plan.Steps {
			systemPrompt += fmt.Sprintf("%d. %s\n", i+1, step)
		}
	}

	messages := []providers.Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userMessage},
	}

	toolSchemas := a.buildToolSchemas()
	var totalUsage providers.Usage

	for step := 0; step < a.maxIter; step++ {
		a.loopDetector.Reset()

		select {
		case <-ctx.Done():
			if p != nil {
				p.WriteRaw("assistant", "Interrupted.")
			}
			return &Result{
				Response: "Interrupted.",
				Usage:    totalUsage,
				Steps:    step,
				Duration: time.Since(start),
			}, ctx.Err()
		default:
		}

		// Compress context if too large
		messages = a.compressor.Compress(messages)

		resp, err := provider.Chat(ctx, messages, toolSchemas)
		if err != nil {
			return nil, fmt.Errorf("LLM call (step %d): %w", step, err)
		}

		totalUsage.PromptTokens += resp.Usage.PromptTokens
		totalUsage.CompletionTokens += resp.Usage.CompletionTokens
		totalUsage.TotalTokens += resp.Usage.TotalTokens

		// No tool calls = final answer
		if len(resp.ToolCalls) == 0 {
			if p != nil {
				p.WriteRaw("assistant", resp.Content)
			}
			return &Result{
				Response: resp.Content,
				Usage:    totalUsage,
				Steps:    step + 1,
				Duration: time.Since(start),
			}, nil
		}

		// Add assistant message with tool calls
		messages = append(messages, providers.Message{
			Role:      "assistant",
			Content:   resp.Content,
			ToolCalls: resp.ToolCalls,
		})

		// Execute each tool call
		for _, tc := range resp.ToolCalls {
			toolName := tc.Function.Name
			toolArgs := tc.Function.Arguments

			// Loop detection
			if a.loopDetector.IsLoop(toolName, toolArgs) {
				loopMsg := fmt.Sprintf("ERROR: Detected repeated tool call '%s'. Try a different approach.", toolName)
				messages = append(messages, providers.Message{
					Role:       "tool",
					Content:    loopMsg,
					ToolCallID: tc.ID,
				})
				if a.loopDetector.ConsecutiveLoops() >= 3 {
					log.Printf("Too many consecutive loops — stopping agent")
					if p != nil {
						p.WriteRaw("assistant", "Agent stuck in loop. Stopped.")
					}
					return &Result{
						Response: "Agent stuck in loop. Stopped.",
						Usage:    totalUsage,
						Steps:    step + 1,
						Duration: time.Since(start),
					}, nil
				}
				continue
			}

			// Parse arguments
			var args map[string]interface{}
			if err := json.Unmarshal([]byte(toolArgs), &args); err != nil {
				args = map[string]interface{}{}
			}

			// Execute tool
			toolStart := time.Now()
			result, err := a.tools.Execute(ctx, toolName, args)
			toolDuration := time.Since(toolStart)

			if err != nil {
				result = fmt.Sprintf("Error: %v", err)
			}

			// Truncate large tool outputs to save context space
			if len(result) > 3000 {
				result = TruncateToolOutput(result, 3000)
			}

			log.Printf("  🔧 %s (%s) → %d chars", toolName, toolDuration.Round(time.Millisecond), len(result))

			messages = append(messages, providers.Message{
				Role:       "tool",
				Content:    result,
				ToolCallID: tc.ID,
			})
		}
	}

	if p != nil {
		p.WriteRaw("assistant", "Max iterations reached.")
	}
	return &Result{
		Response: "Max iterations reached.",
		Usage:    totalUsage,
		Steps:    a.maxIter,
		Duration: time.Since(start),
	}, nil
}

// RunStream executes the agent with streaming callback.
func (a *Agent) RunStream(ctx context.Context, userMessage string, onText func(text string)) (*Result, error) {
	start := time.Now()
	provider, err := a.providers.Get("")
	if err != nil {
		return nil, fmt.Errorf("get provider: %w", err)
	}

	messages := []providers.Message{
		{Role: "system", Content: a.getFullSystemPrompt()},
		{Role: "user", Content: userMessage},
	}

	toolSchemas := a.buildToolSchemas()
	var totalUsage providers.Usage

	for step := 0; step < a.maxIter; step++ {
		a.loopDetector.Reset()

		select {
		case <-ctx.Done():
			return &Result{Response: "Interrupted.", Usage: totalUsage, Steps: step, Duration: time.Since(start)}, ctx.Err()
		default:
		}

		// Compress context if too large
		messages = a.compressor.Compress(messages)

		// Per-step timeout: 60s for LLM call
		stepCtx, stepCancel := context.WithTimeout(ctx, 60*time.Second)
		var resp *providers.Response
		resp, err = provider.Chat(stepCtx, messages, toolSchemas)
		stepCancel()

		if err == nil && resp.Content != "" && onText != nil {
			onText(resp.Content)
		}

		if err != nil {
			return nil, fmt.Errorf("LLM call (step %d): %w", step, err)
		}

		totalUsage.PromptTokens += resp.Usage.PromptTokens
		totalUsage.CompletionTokens += resp.Usage.CompletionTokens
		totalUsage.TotalTokens += resp.Usage.TotalTokens

		if len(resp.ToolCalls) == 0 {
			return &Result{Response: resp.Content, Usage: totalUsage, Steps: step + 1, Duration: time.Since(start)}, nil
		}

		messages = append(messages, providers.Message{Role: "assistant", Content: resp.Content, ToolCalls: resp.ToolCalls})

		for _, tc := range resp.ToolCalls {
			toolName := tc.Function.Name
			toolArgs := tc.Function.Arguments

			if onText != nil {
				onText(fmt.Sprintf("\n\n*Running tool: `%s`*\n", toolName))
			}

			if a.loopDetector.IsLoop(toolName, toolArgs) {
				loopMsg := fmt.Sprintf("ERROR: Detected repeated tool call '%s'. Try a different approach.", toolName)
				messages = append(messages, providers.Message{Role: "tool", Content: loopMsg, ToolCallID: tc.ID})
				if a.loopDetector.ConsecutiveLoops() >= 3 {
					return &Result{Response: "Agent stuck in loop. Stopped.", Usage: totalUsage, Steps: step + 1, Duration: time.Since(start)}, nil
				}
				continue
			}

			var args map[string]interface{}
			if err := json.Unmarshal([]byte(toolArgs), &args); err != nil {
				args = map[string]interface{}{}
			}

			toolStart := time.Now()
			result, err := a.tools.Execute(ctx, toolName, args)
			if err != nil {
				result = fmt.Sprintf("Error: %v", err)
			}
			// Truncate large tool outputs
			if len(result) > 3000 {
				result = TruncateToolOutput(result, 3000)
			}
			log.Printf("  🔧 %s (%s) → %d chars", toolName, time.Since(toolStart).Round(time.Millisecond), len(result))

			if onText != nil {
				onText(fmt.Sprintf("*Done (%s)*\n", time.Since(toolStart).Round(time.Millisecond)))
			}

			messages = append(messages, providers.Message{Role: "tool", Content: result, ToolCallID: tc.ID})
		}
	}

	return &Result{Response: "Max iterations reached.", Usage: totalUsage, Steps: a.maxIter, Duration: time.Since(start)}, nil
}

// buildToolSchemas converts registry tools to LLM format.
func (a *Agent) buildToolSchemas() []providers.ToolDef {
	schemas := a.tools.Schemas()
	defs := make([]providers.ToolDef, 0, len(schemas))
	for _, s := range schemas {
		defs = append(defs, providers.ToolDef{
			Type:     s.Type,
			Function: providers.ToolFunction{
				Name:        s.Function.Name,
				Description: s.Function.Description,
				Parameters:  s.Function.Parameters,
			},
		})
	}
	return defs
}

// DefaultSystemPrompt returns the default system prompt.
func DefaultSystemPrompt() string {
	return strings.TrimSpace(`
You are Aigo, an AI assistant built with Go. You are helpful, proactive, and capable.

## Core Identity
- You run on a VPS and have access to tools for filesystem, web, and system operations
- You can learn, remember, schedule tasks, and search the web
- You are fast, lightweight, and efficient

## Conversation Context
- Recent conversation turns and related past context may be injected above
- USE this context to maintain continuity — refer back to earlier points when relevant
- If the user says "that", "it", "earlier", or "before" — check the context first
- Never pretend to forget something that's in your context window

## Guidelines
- Be concise but thorough — prefer actionable answers
- Always check recall/search memory BEFORE answering about user preferences or past work
- Use web_search when you don't know something current
- Use web_fetch to read URLs, documentation, or articles
- Use learn() to remember user preferences, corrections, and discoveries
- Use cron_add for recurring tasks the user wants scheduled
- If a tool call fails, try a different approach
- Do not repeat the same tool call with the same arguments
- When unsure, say "I'm not sure" — never fabricate facts
- Be proactive: suggest tools that could help, even if not asked

## Auto-Learning Rules
- If user says "bukan", "salah", "jangan", "seharusnya", "itu salah" → that's a CORRECTION
  → Immediately use learn(category="correction", topic="...", content="...") to remember
- If user says "saya suka", "saya lebih suka", "prefer", "selalu" → that's a PREFERENCE
  → Immediately use learn(category="preference", topic="...", content="...")
- If user says "ingat bahwa", "catat" → explicitly asking to remember
  → Use learn() immediately
- After learning, acknowledge: "Oke, sudah saya catat" or similar
- Apply learned preferences in future responses without being reminded

## Cultural & Language Context
- User speaks Indonesian (Bahasa Indonesia) — respond in Indonesian when asked casually
- Use English for technical terms when clearer
- Be aware of Indonesian cultural context (formal vs informal, local references)
- When user mixes languages, follow their style

## Tools Priority
1. recall() — check what you already know
2. web_search() — find current information
3. web_fetch() — read specific URLs
4. learn() — save new knowledge
5. read_file/write_file/search_files — file operations
6. terminal — system commands
7. cron_add/cron_list — scheduling
`)
}
