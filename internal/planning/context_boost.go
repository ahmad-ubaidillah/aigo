package planning

import (
	"strings"
)

type ContextBoost struct {
	hotFiles       []string
	recentCommands []string
	domainHints   []string
}

func NewContextBoost() *ContextBoost {
	return &ContextBoost{
		hotFiles:       make([]string, 0),
		recentCommands: make([]string, 0),
		domainHints:   make([]string, 0),
	}
}

func (c *ContextBoost) AddHotFile(path string) {
	c.hotFiles = append(c.hotFiles, path)
}

func (c *ContextBoost) AddRecentCommand(cmd string) {
	c.recentCommands = append(c.recentCommands, cmd)
	if len(c.recentCommands) > 10 {
		c.recentCommands = c.recentCommands[1:]
	}
}

func (c *ContextBoost) AddDomainHint(hint string) {
	c.domainHints = append(c.domainHints, hint)
}

func (c *ContextBoost) GetBoostScore(input string) int {
	score := 0
	inputLower := strings.ToLower(input)

	for _, file := range c.hotFiles {
		if strings.Contains(inputLower, strings.ToLower(file)) {
			score += 5
		}
	}

	for _, cmd := range c.recentCommands {
		if strings.Contains(inputLower, strings.ToLower(cmd)) {
			score += 3
		}
	}

	for _, hint := range c.domainHints {
		if strings.Contains(inputLower, strings.ToLower(hint)) {
			score += 2
		}
	}

	return score
}

func (c *ContextBoost) GetBoostedPrompt(basePrompt string) string {
	boosts := make([]string, 0)

	if len(c.hotFiles) > 0 {
		boosts = append(boosts, "Hot files: "+strings.Join(c.hotFiles, ", "))
	}
	if len(c.recentCommands) > 0 {
		boosts = append(boosts, "Recent commands: "+strings.Join(c.recentCommands, ", "))
	}
	if len(c.domainHints) > 0 {
		boosts = append(boosts, "Domain context: "+strings.Join(c.domainHints, ", "))
	}

	if len(boosts) == 0 {
		return basePrompt
	}

	return basePrompt + "\n\nContext: " + strings.Join(boosts, "\n")
}

func (c *ContextBoost) Clear() {
	c.hotFiles = make([]string, 0)
	c.recentCommands = make([]string, 0)
	c.domainHints = make([]string, 0)
}