package selfimprove

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ahmad-ubaidillah/aigo/internal/skills"
	"github.com/ahmad-ubaidillah/aigo/pkg/types"
)

type Learner struct {
	db        LearnerStore
	registry  *skills.Registry
	generator *Generator
}

type LearnerStore interface {
	SaveLog(log *types.SelfImproveLog) error
	ListLogs(sessionID string, limit int) ([]types.SelfImproveLog, error)
	GetStats() (LearnerStats, error)
}

type LearnerStats struct {
	TotalTurns     int
	SuccessRate    float64
	SkillsCreated  int
	FailurePattern int
}

func NewLearner(store LearnerStore, registry *skills.Registry) *Learner {
	return &Learner{
		db:        store,
		registry:  registry,
		generator: NewGenerator(),
	}
}

func (l *Learner) LogTurn(ctx context.Context, sessionID, input, output, outcome string, skillGen bool) error {
	log := &types.SelfImproveLog{
		ID:         generateLogID(),
		SessionID:  sessionID,
		TurnInput:  input,
		TurnOutput: output,
		Outcome:    outcome,
		SkillGen:   skillGen,
		CreatedAt:  time.Now(),
	}

	return l.db.SaveLog(log)
}

func (l *Learner) GetRecentTurns(ctx context.Context, sessionID string, limit int) ([]types.SelfImproveLog, error) {
	return l.db.ListLogs(sessionID, limit)
}

func (l *Learner) AnalyzeForSkillCreation(ctx context.Context, logs []types.SelfImproveLog) ([]SkillProposal, error) {
	var proposals []SkillProposal

	patternCount := make(map[string]int)
	patternOutputs := make(map[string]string)

	for _, log := range logs {
		if log.Outcome == "failure" || log.Outcome == "partial" {
			trigger := extractTrigger(log.TurnInput)
			patternCount[trigger]++
			if patternOutputs[trigger] == "" {
				patternOutputs[trigger] = log.TurnOutput
			}
		}
	}

	for pattern, count := range patternCount {
		if count >= 2 {
			proposals = append(proposals, SkillProposal{
				Trigger:     pattern,
				Suggestion:  fmt.Sprintf("Auto-generated skill for: %s", pattern),
				SampleError: patternOutputs[pattern],
				Confidence:  float64(count) / float64(len(logs)),
			})
		}
	}

	return proposals, nil
}

func extractTrigger(input string) string {
	input = strings.ToLower(input)
	words := strings.Fields(input)
	if len(words) > 0 {
		return strings.Join(words[:min(5, len(words))], " ")
	}
	return input
}

func (l *Learner) GenerateAndRegisterSkill(ctx context.Context, proposal SkillProposal) error {
	skill := &types.Skill{
		ID:          generateSkillID(),
		Name:        sanitizeName(proposal.Trigger),
		Description: proposal.Suggestion,
		Code:        generateSkillCode(proposal.Trigger, proposal.SampleError),
		Category:    "auto-generated",
		Version:     1,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Enabled:     true,
		UsageCount:  0,
		Rating:      0,
	}

	err := l.registry.Register(skill.Name, skill.Description, skill.Code, skill.Category, "")
	if err != nil {
		return fmt.Errorf("register skill: %w", err)
	}

	l.generator.AddProposal(proposal)

	return nil
}

func (l *Learner) GetStats(ctx context.Context) (LearnerStats, error) {
	return l.db.GetStats()
}

func (l *Learner) SuggestSkills(ctx context.Context) ([]SkillProposal, error) {
	logs, err := l.db.ListLogs("", 100)
	if err != nil {
		return nil, err
	}

	return l.AnalyzeForSkillCreation(ctx, logs)
}

type SkillProposal struct {
	Trigger     string
	Suggestion  string
	SampleError string
	Confidence  float64
}

type Generator struct {
	proposals []SkillProposal
}

func NewGenerator() *Generator {
	return &Generator{
		proposals: make([]SkillProposal, 0),
	}
}

func (g *Generator) AddProposal(p SkillProposal) {
	g.proposals = append(g.proposals, p)
}

func (g *Generator) GenerateSkill() (*types.Skill, error) {
	if len(g.proposals) == 0 {
		return nil, fmt.Errorf("no proposals available")
	}

	prop := g.proposals[0]
	g.proposals = g.proposals[1:]

	return &types.Skill{
		ID:          generateSkillID(),
		Name:        sanitizeName(prop.Trigger),
		Description: prop.Suggestion,
		Code:        generateSkillCode(prop.Trigger, prop.SampleError),
		Category:    "auto-generated",
		Version:     1,
		CreatedAt:   time.Now(),
		Enabled:     false,
	}, nil
}

func (g *Generator) GetProposals() []SkillProposal {
	return g.proposals
}

func sanitizeName(trigger string) string {
	name := strings.ToLower(trigger)
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, "-", "_")
	if len(name) > 30 {
		name = name[:30]
	}
	return "auto_" + name
}

func generateLogID() string {
	return fmt.Sprintf("log_%d", time.Now().UnixNano())
}

func generateSkillID() string {
	return fmt.Sprintf("auto_%d", time.Now().UnixNano())
}

func generateSkillCode(trigger, sampleError string) string {
	return fmt.Sprintf(`// Auto-generated skill for: %s
// Generated from failure pattern analysis
// Sample error: %s

package main

import "fmt"

func Execute_%s(input string) (string, error) {
    // TODO: Implement skill logic based on pattern
    // This skill was auto-generated because similar commands failed multiple times
    return fmt.Sprintf("Executed skill for: %%s", input), nil
}

func main() {
    result, err := Execute_%s("")
    if err != nil {
        fmt.Println("Error:", err)
    }
    fmt.Println(result)
}
`, trigger, sampleError, sanitizeName(trigger), sanitizeName(trigger))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
