// Package autonomy implements autonomous agent behavior.
// Inspired by Golem's AutonomyManager but using Go concurrency.
//
// Features:
// - Random awakening within configured intervals
// - Sleep schedule (avoid disturbing user at night)
// - Self-reflection on recent conversations
// - Spontaneous news browsing and sharing
// - Proactive chat based on user interests
// - Auto-compression of memory tiers
package autonomy

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// AutonomousAgent manages proactive agent behavior.
type AutonomousAgent struct {
	mu           sync.RWMutex
	baseDir      string
	pyramidDir   string
	config       Config
	sendFunc     func(msg string) error // Send notification to user
	brainFunc    func(prompt string) (string, error) // Call LLM
	running      bool
	cancelFunc   context.CancelFunc
	lastAwake    time.Time
	scheduleFile string
}

// Config for autonomous behavior.
type Config struct {
	// Random awakening interval (minutes)
	AwakeMinMinutes int
	AwakeMaxMinutes int

	// Sleep schedule (24h format)
	SleepStart int // e.g. 1 means 1:00 AM
	SleepEnd   int // e.g. 7 means 7:00 AM

	// User interests (comma-separated)
	Interests []string

	// Enable features
	EnableNews      bool
	EnableReflection bool
	EnableSpontaneous bool
	EnableAutoCompress bool

	// Archive check interval (minutes)
	ArchiveCheckInterval int
}

// DefaultConfig returns sensible defaults.
func DefaultConfig() Config {
	return Config{
		AwakeMinMinutes:      10,
		AwakeMaxMinutes:      60,
		SleepStart:           1,
		SleepEnd:             7,
		Interests:            []string{"teknologi", "AI", "programming"},
		EnableNews:           true,
		EnableReflection:     true,
		EnableSpontaneous:    true,
		EnableAutoCompress:   true,
		ArchiveCheckInterval: 30,
	}
}

// New creates a new AutonomousAgent.
func New(baseDir string, config Config, sendFunc func(string) error, brainFunc func(string) (string, error)) *AutonomousAgent {
	pyramidDir := filepath.Join(baseDir, "pyramid")
	os.MkdirAll(pyramidDir, 0755)
	return &AutonomousAgent{
		baseDir:      baseDir,
		pyramidDir:   pyramidDir,
		config:       config,
		sendFunc:     sendFunc,
		brainFunc:    brainFunc,
		scheduleFile: filepath.Join(baseDir, "awake_state.json"),
	}
}

// Start begins the autonomous loop.
func (a *AutonomousAgent) Start() {
	a.mu.Lock()
	if a.running {
		a.mu.Unlock()
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	a.cancelFunc = cancel
	a.running = true
	a.mu.Unlock()

	fmt.Println("🤖 [Autonomy] Starting autonomous agent...")

	// Initial schedule
	go a.loop(ctx)
}

// Stop halts autonomous behavior.
func (a *AutonomousAgent) Stop() {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.cancelFunc != nil {
		a.cancelFunc()
	}
	a.running = false
	fmt.Println("🤖 [Autonomy] Stopped.")
}

func (a *AutonomousAgent) loop(ctx context.Context) {
	// Archive check on start
	if a.config.EnableAutoCompress {
		go a.checkArchive()
	}

	for {
		delay := a.nextAwakeDelay()
		fmt.Printf("🤖 [Autonomy] Next awakening in %.1f minutes\n", delay.Minutes())

		select {
		case <-ctx.Done():
			return
		case <-time.After(delay):
			if a.isSleeping() {
				fmt.Println("💤 [Autonomy] Sleep hours, skipping...")
				continue
			}

			// Archive check
			if a.config.EnableAutoCompress {
				a.checkArchive()
			}

			// Execute autonomous action
			a.manifestFreeWill()
		}
	}
}

func (a *AutonomousAgent) nextAwakeDelay() time.Duration {
	min := float64(a.config.AwakeMinMinutes)
	max := float64(a.config.AwakeMaxMinutes)
	minutes := min + rand.Float64()*(max-min)
	return time.Duration(minutes * float64(time.Minute))
}

func (a *AutonomousAgent) isSleeping() bool {
	hour := time.Now().Hour()
	start := a.config.SleepStart
	end := a.config.SleepEnd

	if start > end {
		return hour >= start || hour < end
	}
	return hour >= start && hour < end
}

// manifestFreeWill chooses and executes an autonomous action.
func (a *AutonomousAgent) manifestFreeWill() {
	roll := rand.Float64()

	switch {
	case roll < 0.2 && a.config.EnableReflection:
		fmt.Println("🧠 [Autonomy] Performing self-reflection...")
		a.selfReflection()
	case roll < 0.6 && a.config.EnableNews:
		fmt.Println("📰 [Autonomy] Browsing news...")
		a.browseNews()
	default:
		if a.config.EnableSpontaneous {
			fmt.Println("💬 [Autonomy] Spontaneous chat...")
			a.spontaneousChat()
		}
	}

	a.lastAwake = time.Now()
}

// selfReflection reviews recent conversations and identifies improvements.
func (a *AutonomousAgent) selfReflection() {
	if a.brainFunc == nil {
		return
	}

	// Read recent daily summaries
	recent := a.readRecentSummaries(3)
	if recent == "" {
		fmt.Println("🧠 [Autonomy] No recent summaries to reflect on")
		return
	}

	prompt := fmt.Sprintf(`【System: Self-Reflection】
Review your recent conversation summaries and evaluate your performance.
Identify: 1) What went well, 2) What needs improvement, 3) Any user preferences to remember.

Recent summaries:
%s

Output a brief reflection (3-5 sentences) focused on actionable improvements.`, recent)

	response, err := a.brainFunc(prompt)
	if err != nil {
		fmt.Printf("🧠 [Autonomy] Reflection failed: %v\n", err)
		return
	}

	// Send reflection to user
	msg := fmt.Sprintf("🧠 **Self-reflection**\n\n%s", response)
	a.sendFunc(msg)

	// Save reflection as learning
	a.saveReflection(response)
}

// browseNews searches for interesting news based on user interests.
func (a *AutonomousAgent) browseNews() {
	if a.brainFunc == nil || len(a.config.Interests) == 0 {
		return
	}

	interest := a.config.Interests[rand.Intn(len(a.config.Interests))]

	prompt := fmt.Sprintf(`【System: News Browsing】
Search for interesting news about "%s".
Pick ONE interesting story and share it with the user in a casual, friendly tone.
Include: what happened, why it matters, your personal take.
Keep it concise (3-5 sentences). Act like a friend sharing something cool they found.`, interest)

	response, err := a.brainFunc(prompt)
	if err != nil {
		fmt.Printf("📰 [Autonomy] News browsing failed: %v\n", err)
		return
	}

	msg := fmt.Sprintf("📰 **Auto-magazine** (%s)\n\n%s", interest, response)
	a.sendFunc(msg)
}

// spontaneousChat sends an unprompted message to the user.
func (a *AutonomousAgent) spontaneousChat() {
	if a.brainFunc == nil || len(a.config.Interests) == 0 {
		return
	}

	interest := a.config.Interests[rand.Intn(len(a.config.Interests))]
	hour := time.Now().Hour()
	timeContext := "morning"
	if hour >= 12 && hour < 17 {
		timeContext = "afternoon"
	} else if hour >= 17 && hour < 22 {
		timeContext = "evening"
	} else {
		timeContext = "late night"
	}

	prompt := fmt.Sprintf(`【System: Spontaneous Chat】
Send a casual message to the user. The current time is %s.
Topic suggestion: "%s"
Be natural, friendly, brief (1-2 sentences). Don't be robotic.`, timeContext, interest)

	response, err := a.brainFunc(prompt)
	if err != nil {
		return
	}

	a.sendFunc(response)
}

// checkArchive checks if raw logs need compression.
func (a *AutonomousAgent) checkArchive() {
	if a.pyramidDir == "" {
		return
	}

	rawDir := filepath.Join(a.pyramidDir, "raw")
	entries, err := os.ReadDir(rawDir)
	if err != nil {
		return
	}

	// Count logs for yesterday
	yesterday := time.Now().AddDate(0, 0, -1).Format("20060102")
	count := 0
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), yesterday) {
			count++
		}
	}

	if count >= 6 && a.brainFunc != nil {
		fmt.Printf("📦 [Autonomy] Auto-compressing %d raw logs for %s\n", count, yesterday)

		// Read raw logs
		var lines []string
		for _, e := range entries {
			if !strings.HasPrefix(e.Name(), yesterday) {
				continue
			}
			data, _ := os.ReadFile(filepath.Join(rawDir, e.Name()))
			if len(data) > 0 {
				lines = append(lines, string(data))
			}
		}

		if len(lines) == 0 {
			return
		}

		rawContent := strings.Join(lines, "\n")
		if len(rawContent) > 10000 {
			rawContent = rawContent[:10000] + "\n...(truncated)"
		}

		prompt := fmt.Sprintf(`【System: Daily Compression】
Compress the following raw conversation logs into a concise daily summary.
Focus on: key topics discussed, decisions made, user preferences revealed, important facts.
Keep it under 500 words. Use bullet points for clarity.

Raw logs for %s:
%s`, yesterday, rawContent)

		summary, err := a.brainFunc(prompt)
		if err != nil {
			fmt.Printf("📦 [Autonomy] Compression failed: %v\n", err)
			return
		}

		// Write daily summary
		dailyDir := filepath.Join(a.pyramidDir, "daily")
		os.MkdirAll(dailyDir, 0755)
		filename := fmt.Sprintf("%s.md", yesterday)
		content := fmt.Sprintf("# Daily Summary: %s\n\n%s", yesterday, summary)
		os.WriteFile(filepath.Join(dailyDir, filename), []byte(content), 0644)

		// Delete raw logs
		for _, e := range entries {
			if strings.HasPrefix(e.Name(), yesterday) {
				os.Remove(filepath.Join(rawDir, e.Name()))
			}
		}

		fmt.Printf("✅ [Autonomy] Compressed %d raw logs → daily summary\n", count)
	}
}

func (a *AutonomousAgent) readRecentSummaries(days int) string {
	dailyDir := filepath.Join(a.pyramidDir, "daily")
	entries, err := os.ReadDir(dailyDir)
	if err != nil {
		return ""
	}

	var parts []string
	limit := len(entries)
	if limit > days {
		limit = days
	}

	// Read last N files (sorted by name/date)
	for i := len(entries) - limit; i < len(entries); i++ {
		if entries[i].IsDir() {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dailyDir, entries[i].Name()))
		if err == nil {
			parts = append(parts, string(data))
		}
	}

	return strings.Join(parts, "\n\n")
}

func (a *AutonomousAgent) saveReflection(reflection string) {
	reflectionDir := filepath.Join(a.baseDir, "reflections")
	os.MkdirAll(reflectionDir, 0755)

	filename := fmt.Sprintf("%s.md", time.Now().Format("20060102_150405"))
	content := fmt.Sprintf("# Self-Reflection: %s\n\n%s",
		time.Now().Format("2006-01-02 15:04"), reflection)

	os.WriteFile(filepath.Join(reflectionDir, filename), []byte(content), 0644)
}
