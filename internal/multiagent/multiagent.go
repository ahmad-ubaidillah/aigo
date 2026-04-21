// Package multiagent implements interactive multi-agent collaboration.
// Inspired by Golem's InteractiveMultiAgent but using Go concurrency
// for parallel agent execution instead of sequential.
//
// Features:
// - Preset team configurations (Tech, Debate, Creative, Business)
// - Interactive rounds with user interruption
// - Shared memory between agents
// - Consensus detection
// - Summary generation
package multiagent

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// AgentConfig defines a single agent's personality and expertise.
type AgentConfig struct {
	Name        string   `json:"name"`
	Role        string   `json:"role"`
	Personality string   `json:"personality"`
	Expertise   []string `json:"expertise"`
}

// Conversation tracks an active multi-agent discussion.
type Conversation struct {
	ID            string
	Task          string
	Agents        []AgentConfig
	AgentMap      map[string]AgentConfig
	Context       string
	Round         int
	MaxRounds     int
	Messages      []Message
	SharedMemory  []MemoryEntry
	Status        string // active, completed, interrupted
	WaitingForUser bool
	mu            sync.RWMutex
}

// Message is a single agent or user message.
type Message struct {
	Round     int       `json:"round"`
	Speaker   string    `json:"speaker"`
	Role      string    `json:"role"`
	Type      string    `json:"type"` // agent, user, summary
	Content   string    `json:"content"`
	HadMemory bool      `json:"had_memory"`
	HadAction bool      `json:"had_action"`
	Timestamp time.Time `json:"timestamp"`
}

// MemoryEntry is shared information between agents.
type MemoryEntry struct {
	Agent   string    `json:"agent"`
	Content string    `json:"content"`
	Round   int       `json:"round"`
	Source  string    `json:"source"`
}

// Preset team configurations.
var (
	TechTeam = []AgentConfig{
		{Name: "Alex", Role: "Frontend Engineer", Personality: "UX-focused, aesthetic-driven", Expertise: []string{"React", "Next.js", "UI/UX", "CSS"}},
		{Name: "Bob", Role: "Backend Engineer", Personality: "Cautious, security-minded", Expertise: []string{"Go", "Database", "API", "Architecture"}},
		{Name: "Carol", Role: "Product Manager", Personality: "User-focused, business-minded", Expertise: []string{"Requirements", "Product Planning", "Strategy"}},
	}

	DebateTeam = []AgentConfig{
		{Name: "Devil", Role: "Devil's Advocate", Personality: "Critical thinker, challenges assumptions", Expertise: []string{"Risk Analysis", "Logic"}},
		{Name: "Angel", Role: "Optimist", Personality: "Positive thinker, opportunity-finder", Expertise: []string{"Vision Planning", "Opportunity Mining"}},
		{Name: "Judge", Role: "Neutral Arbiter", Personality: "Rational, balanced perspective", Expertise: []string{"Decision Analysis", "Synthesis"}},
	}

	CreativeTeam = []AgentConfig{
		{Name: "Writer", Role: "Copywriter", Personality: "Imaginative, storytelling", Expertise: []string{"Storytelling", "Copy", "Content Strategy"}},
		{Name: "Designer", Role: "Visual Designer", Personality: "Artistic, aesthetic", Expertise: []string{"Visual Design", "Branding"}},
		{Name: "Strategist", Role: "Strategy Consultant", Personality: "Logical, structured", Expertise: []string{"Market Analysis", "Strategy"}},
	}

	BusinessTeam = []AgentConfig{
		{Name: "Finance", Role: "Financial Advisor", Personality: "Numbers-driven", Expertise: []string{"Financial Planning", "Cost Analysis"}},
		{Name: "Marketing", Role: "Marketing Expert", Personality: "Creative, data-driven", Expertise: []string{"Brand Strategy", "Growth"}},
		{Name: "Operations", Role: "Operations Expert", Personality: "Execution-focused", Expertise: []string{"Process Design", "Efficiency"}},
	}

	allPresets = map[string][]AgentConfig{
		"tech":     TechTeam,
		"debate":   DebateTeam,
		"creative": CreativeTeam,
		"business": BusinessTeam,
	}
)

// BrainFunc is the signature for LLM calls.
type BrainFunc func(prompt string) (string, error)

// SendFunc is the signature for sending messages to user.
type SendFunc func(msg string) error

// MultiAgent manages multi-agent conversations.
type MultiAgent struct {
	mu              sync.RWMutex
	activeConvo     *Conversation
	brainFunc       BrainFunc
	sendFunc        SendFunc
	pausedConvo     map[string]*Conversation
	inputChan       chan string
}

// New creates a new MultiAgent manager.
func New(brainFunc BrainFunc, sendFunc SendFunc) *MultiAgent {
	return &MultiAgent{
		brainFunc:   brainFunc,
		sendFunc:    sendFunc,
		pausedConvo: make(map[string]*Conversation),
	}
}

// StartConversation begins a multi-agent roundtable.
func (m *MultiAgent) StartConversation(task string, teamName string, maxRounds int) error {
	preset, ok := allPresets[strings.ToLower(teamName)]
	if !ok {
		available := make([]string, 0, len(allPresets))
		for k := range allPresets {
			available = append(available, k)
		}
		return fmt.Errorf("unknown team '%s'. Available: %s", teamName, strings.Join(available, ", "))
	}

	if maxRounds <= 0 {
		maxRounds = 3
	}

	convo := &Conversation{
		ID:        fmt.Sprintf("conv_%d", time.Now().UnixNano()),
		Task:      task,
		Agents:    preset,
		AgentMap:  make(map[string]AgentConfig),
		MaxRounds: maxRounds,
		Status:    "active",
	}

	for _, a := range preset {
		convo.AgentMap[strings.ToLower(a.Name)] = a
	}

	m.mu.Lock()
	m.activeConvo = convo
	m.mu.Unlock()

	// Send intro
	intro := fmt.Sprintf("🎭 **Multi-Agent Roundtable**\n\n📋 **Task**: %s\n\n👥 **Team**:\n", task)
	for i, a := range preset {
		intro += fmt.Sprintf("%d. 🤖 **%s** - %s\n   *%s*\n", i+1, a.Name, a.Role, strings.Join(a.Expertise[:2], ", "))
	}
	intro += "\n━━━━━━━━━━━━━━━━━━\n💡 Inter-round commands: `skip`, `end`, `pause`, `@AgentName text`"
	m.sendFunc(intro)

	// Run interactive loop
	m.interactiveLoop()

	// Generate summary
	if convo.Status != "interrupted" {
		m.generateSummary()
	}

	m.mu.Lock()
	m.activeConvo = nil
	m.mu.Unlock()

	return nil
}

func (m *MultiAgent) interactiveLoop() {
	convo := m.activeConvo
	convo.Context = fmt.Sprintf("【Team Task】%s\n【Members】%s\n\n【Discussion】\n",
		convo.Task, agentNames(convo.Agents))

	for round := 1; round <= convo.MaxRounds; round++ {
		m.mu.RLock()
		if convo.Status != "active" {
			m.mu.RUnlock()
			break
		}
		m.mu.RUnlock()

		convo.mu.Lock()
		convo.Round = round
		convo.mu.Unlock()

		m.sendFunc(fmt.Sprintf("\n**━━━ Round %d / %d ━━━**", round, convo.MaxRounds))

		// Agents speak in parallel for speed
		var wg sync.WaitGroup
		results := make(chan agentResult, len(convo.Agents))

		for _, agent := range convo.Agents {
			wg.Add(1)
			go func(a AgentConfig) {
				defer wg.Done()
				response := m.agentSpeak(a, round)
				results <- response
			}(agent)
		}

		go func() {
			wg.Wait()
			close(results)
		}()

		for r := range results {
			if r.err != nil {
				m.sendFunc(fmt.Sprintf("⚠️ %s couldn't speak: %v", r.agentName, r.err))
				continue
			}
			msg := Message{
				Round:     round,
				Speaker:   r.agentName,
				Role:      r.role,
				Type:      "agent",
				Content:   r.reply,
				HadMemory: r.hadMemory,
				Timestamp: time.Now(),
			}
			convo.mu.Lock()
			convo.Messages = append(convo.Messages, msg)
			convo.Context += fmt.Sprintf("[Round %d] %s: %s\n", round, r.agentName, r.reply)
			convo.mu.Unlock()

			badges := ""
			if r.hadMemory {
				badges = " 🧠"
			}
			m.sendFunc(fmt.Sprintf("🤖 **%s** _(%s)_%s\n%s", r.agentName, r.role, badges, r.reply))
		}

		// Check early consensus
		convo.mu.RLock()
		if m.checkConsensus(convo.Messages) {
			convo.mu.RUnlock()
			m.sendFunc("\n✅ Team reached consensus early!")
			convo.mu.Lock()
			convo.Status = "completed"
			convo.mu.Unlock()
			break
		}
		convo.mu.RUnlock()

		// User turn (skip if last round)
		if round < convo.MaxRounds {
			action := m.userTurn(round)
			if action == "END" {
				convo.mu.Lock()
				convo.Status = "completed"
				convo.mu.Unlock()
				m.sendFunc("✅ Discussion ended early.")
				break
			} else if action == "INTERRUPT" {
				convo.mu.Lock()
				convo.Status = "interrupted"
				convo.mu.Unlock()
				m.sendFunc("⏸️ Discussion paused. Use `resume` to continue.")
				return
			}
		}
	}

	if convo.Status == "active" {
		convo.mu.Lock()
		convo.Status = "completed"
		convo.mu.Unlock()
	}
}

type agentResult struct {
	agentName string
	role      string
	reply     string
	hadMemory bool
	err       error
}

func (m *MultiAgent) agentSpeak(agent AgentConfig, round int) agentResult {
	// Build prompt
	prompt := m.buildAgentPrompt(agent, round)

	response, err := m.brainFunc(prompt)
	if err != nil {
		return agentResult{agentName: agent.Name, role: agent.Role, err: err}
	}

	// Parse response (simple: use full text as reply)
	reply := strings.TrimSpace(response)

	// Check for memory indicators
	hadMemory := strings.Contains(reply, "[MEMORY]") || strings.Contains(reply, "📌")

	// Clean response
	reply = cleanReply(reply, agent.Name)

	if len(reply) > 300 {
		reply = reply[:297] + "..."
	}

	return agentResult{
		agentName: agent.Name,
		role:      agent.Role,
		reply:     reply,
		hadMemory: hadMemory,
	}
}

func (m *MultiAgent) buildAgentPrompt(agent AgentConfig, round int) string {
	convo := m.activeConvo

	convo.mu.RLock()
	ctx := convo.Context
	sharedMem := convo.SharedMemory
	convo.mu.RUnlock()

	var memCtx string
	if len(sharedMem) > 0 {
		memCtx = "\n【Shared Team Memory】\n"
		recent := sharedMem
		if len(recent) > 5 {
			recent = recent[len(recent)-5:]
		}
		for _, mem := range recent {
			memCtx += fmt.Sprintf("- [%s] %s\n", mem.Agent, mem.Content)
		}
	}

	isLastRound := round >= convo.MaxRounds

	return fmt.Sprintf(`【System: Multi-Agent Collaboration】
🎭 **Your Role**:
- Name: %s
- Position: %s
- Personality: %s
- Expertise: %s
━━━━━━━━━━━━━━━━━━
【Context】
Task: "%s"
Members: %s + User
Progress: Round %d / %d
【Discussion History】
%s
%s
━━━━━━━━━━━━━━━━━━
Instructions:
- Be brief: 2-3 sentences, under 80 words
- Be specific and actionable
- Reference others' points with @Name
- If recording important info, prefix with 📌
%s
Speak as %s:`,
		agent.Name,
		agent.Role,
		agent.Personality,
		strings.Join(agent.Expertise, ", "),
		convo.Task,
		agentNames(convo.Agents),
		round, convo.MaxRounds,
		ctx,
		memCtx,
		func() string {
			if round == 1 {
				return "- Share your initial professional opinion"
			}
			if isLastRound {
				return "- ⚠️ FINAL ROUND: Give your conclusion!"
			}
			return "- Respond to others' points, build on ideas"
		}(),
		agent.Name,
	)
}

// userTurn waits for user input between rounds.
func (m *MultiAgent) userTurn(round int) string {
	m.sendFunc("\n💬 **Your turn** (30s timeout, or `skip` to continue)")
	m.sendFunc("━━━━━━━━━━━━━━━━━━")

	// Use channel with timeout
	inputCh := make(chan string, 1)
	m.mu.Lock()
	m.inputChan = inputCh
	m.mu.Unlock()

	select {
	case input := <-inputCh:
		m.mu.Lock()
		m.inputChan = nil
		m.mu.Unlock()

		input = strings.TrimSpace(strings.ToLower(input))
		if input == "" || input == "skip" || input == "continue" {
			return "CONTINUE"
		}
		if input == "end" || input == "stop" {
			return "END"
		}
		if input == "pause" || input == "interrupt" {
			return "INTERRUPT"
		}

		// Process user input
		m.processUserInput(input, round)
		return "CONTINUE"

	case <-time.After(30 * time.Second):
		m.mu.Lock()
		m.inputChan = nil
		m.mu.Unlock()
		m.sendFunc("⏱️ Timeout, auto-continuing...")
		return "CONTINUE"
	}
}

// HandleInput feeds user input to the active multi-agent conversation.
func (m *MultiAgent) HandleInput(input string) bool {
	m.mu.RLock()
	ch := m.inputChan
	m.mu.RUnlock()

	if ch != nil {
		ch <- input
		return true
	}
	return false
}

func (m *MultiAgent) processUserInput(input string, round int) {
	convo := m.activeConvo

	m.sendFunc(fmt.Sprintf("👤 **Your input**: %s", input))

	msg := Message{
		Round:     round,
		Speaker:   "User",
		Role:      "User",
		Type:      "user",
		Content:   input,
		Timestamp: time.Now(),
	}

	convo.mu.Lock()
	convo.Messages = append(convo.Messages, msg)
	convo.Context += fmt.Sprintf("[User]: %s\n", input)
	convo.mu.Unlock()

	// Check for @mentions
	if strings.Contains(input, "@") {
		for _, agent := range convo.Agents {
			mention := "@" + strings.ToLower(agent.Name)
			if strings.Contains(strings.ToLower(input), mention) {
				m.sendFunc(fmt.Sprintf("🎤 _%s responding to your question..._", agent.Name))
				result := m.agentSpeak(agent, round)
				if result.err == nil {
					m.sendFunc(fmt.Sprintf("🤖 **%s** _(to you)_\n%s", result.agentName, result.reply))
				}
			}
		}
	}
}

func (m *MultiAgent) generateSummary() {
	convo := m.activeConvo
	m.sendFunc("\n━━━━━━━━━━━━━━━━━━\n🎯 **Generating team summary...**")

	convo.mu.RLock()
	ctx := convo.Context
	memCount := len(convo.SharedMemory)
	msgCount := len(convo.Messages)
	convo.mu.RUnlock()

	prompt := fmt.Sprintf(`【System: Meeting Summary】
Synthesize the following discussion into a professional summary.

Task: %s
Team: %s + User
Messages: %d
Shared Memory: %d entries

Discussion:
%s

Format:
## Core Conclusion
(2-3 sentences)

## Key Decisions
- Decision 1
- Decision 2

## Action Items
- Action 1
- Action 2`, convo.Task, agentNames(convo.Agents), msgCount, memCount, ctx)

	summary, err := m.brainFunc(prompt)
	if err != nil {
		m.sendFunc("⚠️ Summary generation failed")
		return
	}

	m.sendFunc(fmt.Sprintf("🎯 **Team Summary**\n\n%s\n\n━━━━━━━━━━━━━━━━━━\n📊 Stats: %d messages / %d rounds / %d memories",
		summary, msgCount, convo.Round, memCount))
}

func (m *MultiAgent) checkConsensus(messages []Message) bool {
	if len(messages) < 6 {
		return false
	}
	recent := messages[len(messages)-3:]
	keywords := []string{"agree", "setuju", "consensus", "sepakat", "sounds good", "oke", "ok"}
	for _, msg := range recent {
		lower := strings.ToLower(msg.Content)
		for _, kw := range keywords {
			if strings.Contains(lower, kw) {
				return true
			}
		}
	}
	return false
}

// GetStatus returns current conversation status for WebUI.
func (m *MultiAgent) GetStatus() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.activeConvo == nil {
		return map[string]interface{}{"active": false}
	}

	return map[string]interface{}{
		"active":    true,
		"task":      m.activeConvo.Task,
		"round":     m.activeConvo.Round,
		"max_rounds": m.activeConvo.MaxRounds,
		"messages":  len(m.activeConvo.Messages),
		"status":    m.activeConvo.Status,
	}
}

func agentNames(agents []AgentConfig) string {
	names := make([]string, len(agents))
	for i, a := range agents {
		names[i] = a.Name
	}
	return strings.Join(names, ", ")
}

func cleanReply(reply, agentName string) string {
	// Remove agent name prefix if present
	prefixes := []string{
		agentName + ": ",
		agentName + "：",
		"**" + agentName + "**: ",
	}
	for _, p := range prefixes {
		if strings.HasPrefix(reply, p) {
			reply = reply[len(p):]
		}
	}
	return strings.TrimSpace(reply)
}

// GetPresetNames returns available team preset names.
func GetPresetNames() []string {
	names := make([]string, 0, len(allPresets))
	for k := range allPresets {
		names = append(names, k)
	}
	return names
}
