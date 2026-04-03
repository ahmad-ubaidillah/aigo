// Package intent implements two-tier intent classification (rule-based + LLM fallback).
package intent

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"regexp"
	"strings"

	"github.com/ahmad-ubaidillah/aigo/pkg/types"
	"github.com/sashabaranov/go-openai"
)

// Classification holds the result of intent classification.
type Classification struct {
	Intent      string   // one of types.IntentCoding, types.IntentWeb, etc.
	Confidence  float64  // 0.0-1.0
	Workspace   string   // detected workspace path (empty if not detected)
	Description string   // human-readable description of the intent
	Tags        []string // extracted tags/keywords
}

// Classifier holds config and optional OpenAI client for LLM fallback.
type Classifier struct {
	cfg    types.Config
	client *openai.Client
}

var keywordSets = map[string][]string{
	types.IntentCoding: {
		"code", "function", "bug", "fix", "implement", "refactor",
		"test", "build", "compile", "error", "type", "import",
		"package", "struct", "interface", "write file", "create file",
		"edit", "patch", "lsp", "ast",
	},
	types.IntentWeb: {
		"search", "web", "website", "url", "http", "browse",
		"scrape", "extract", "screenshot", "fetch",
	},
	types.IntentFile: {
		"file", "directory", "folder", "copy", "move", "rename",
		"delete file", "list files", "find file", "read file",
	},
	types.IntentGateway: {
		"send message", "telegram", "discord", "slack",
		"whatsapp", "notify", "channel",
	},
	types.IntentMemory: {
		"remember", "memory", "recall", "search memory",
		"save this", "context", "important", "fact",
	},
	types.IntentAutomation: {
		"schedule", "cron", "daily", "weekly", "repeat",
		"automate", "every day",
	},
	types.IntentSkill: {
		"skill", "run skill", "list skills", "add skill",
		"execute skill", "skill marketplace",
	},
	types.IntentResearch: {
		"research", "search code", "lookup", "find docs",
		"web search", "code search", "documentation",
	},
	types.IntentHTTPCall: {
		"http", "api", "request", "endpoint", "rest", "curl",
		"fetch", "call", "post", "get", "put", "delete", "patch",
		"json", "header", "status code", "api call",
	},
	types.IntentBrowser: {
		"browser", "navigate", "click", "fill", "screenshot",
		"scroll", "hover", "select", "dropdown", "checkbox",
		"radio", "button", "link", "form", "automation",
	},
	types.IntentPython: {
		"python", "run python", "execute python", "pip install",
		"import", "numpy", "pandas", "script", "jupyter",
	},
}

var stopWords = map[string]bool{
	"a": true, "an": true, "the": true, "is": true, "are": true,
	"was": true, "were": true, "be": true, "been": true, "being": true,
	"have": true, "has": true, "had": true, "do": true, "does": true,
	"did": true, "will": true, "would": true, "could": true, "should": true,
	"may": true, "might": true, "must": true, "shall": true,
	"to": true, "of": true, "in": true, "for": true, "on": true,
	"with": true, "at": true, "by": true, "from": true, "as": true,
	"into": true, "through": true, "during": true, "before": true,
	"after": true, "above": true, "below": true, "between": true,
	"and": true, "but": true, "or": true, "nor": true, "not": true,
	"so": true, "yet": true, "both": true, "either": true, "neither": true,
	"i": true, "you": true, "he": true, "she": true, "it": true,
	"we": true, "they": true, "what": true, "which": true, "who": true,
	"how": true, "this": true, "that": true, "these": true, "those": true,
	"me": true, "him": true, "her": true, "us": true, "them": true,
	"my": true, "your": true, "his": true, "its": true, "our": true,
	"their": true, "can": true, "please": true, "just": true,
}

// NewClassifier creates a classifier with optional LLM support.
func NewClassifier(cfg types.Config) *Classifier {
	c := &Classifier{cfg: cfg}

	if cfg.Model.Intent == "" {
		return c
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return c
	}

	c.client = openai.NewClient(apiKey)
	return c
}

// Classify classifies user input using rule-based first, then LLM fallback.
func (c *Classifier) Classify(input string) Classification {
	result := c.ruleBasedClassify(input)
	if result.Confidence >= 0.7 {
		return result
	}

	if c.client == nil {
		return result
	}

	llmResult := c.llmClassify(input)
	if llmResult.Confidence > result.Confidence {
		return llmResult
	}

	return result
}

func (c *Classifier) ruleBasedClassify(input string) Classification {
	lower := strings.ToLower(input)
	scores := make(map[string]int)

	for intent, keywords := range keywordSets {
		for _, kw := range keywords {
			if strings.Contains(lower, kw) {
				scores[intent]++
			}
		}
	}

	bestIntent := types.IntentGeneral
	bestScore := 0

	for intent, score := range scores {
		if score > bestScore {
			bestIntent = intent
			bestScore = score
		}
	}

	confidence := math.Min(0.9, float64(bestScore)*0.3)
	if confidence < 0.1 {
		confidence = 0.1
	}

	workspace := extractWorkspace(input)
	tags := extractTags(input)
	desc := buildDescription(bestIntent, bestScore)

	return Classification{
		Intent:      bestIntent,
		Confidence:  confidence,
		Workspace:   workspace,
		Description: desc,
		Tags:        tags,
	}
}

func (c *Classifier) llmClassify(input string) Classification {
	if c.client == nil {
		return Classification{
			Intent:      types.IntentGeneral,
			Confidence:  0.5,
			Description: "LLM fallback unavailable, defaulting to general",
		}
	}

	prompt := buildLLMPrompt(input)
	resp, err := c.client.CreateChatCompletion(context.Background(), openai.ChatCompletionRequest{
		Model: c.cfg.Model.Intent,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: prompt,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: input,
			},
		},
		Temperature: 0,
	})
	if err != nil {
		return Classification{
			Intent:      types.IntentGeneral,
			Confidence:  0.5,
			Description: fmt.Sprintf("LLM classification failed: %v", err),
		}
	}

	if len(resp.Choices) == 0 {
		return Classification{
			Intent:      types.IntentGeneral,
			Confidence:  0.5,
			Description: "LLM returned no choices",
		}
	}

	return parseLLMResponse(resp.Choices[0].Message.Content)
}

// buildLLMPrompt creates the system prompt for LLM classification.
func buildLLMPrompt(input string) string {
	return fmt.Sprintf(
		"You are an intent classifier. Classify the user input into one of these categories: %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s. Return JSON with fields: intent, confidence (0.0-1.0), description, tags (array of strings).",
		types.IntentCoding, types.IntentWeb, types.IntentFile,
		types.IntentGateway, types.IntentMemory, types.IntentAutomation,
		types.IntentSkill, types.IntentResearch, types.IntentHTTPCall,
		types.IntentBrowser, types.IntentPython, types.IntentGeneral,
	)
}

// parseLLMResponse parses JSON from the LLM into a Classification.
func parseLLMResponse(content string) Classification {
	var result struct {
		Intent      string   `json:"intent"`
		Confidence  float64  `json:"confidence"`
		Description string   `json:"description"`
		Tags        []string `json:"tags"`
	}

	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return Classification{
			Intent:      types.IntentGeneral,
			Confidence:  0.5,
			Description: fmt.Sprintf("Failed to parse LLM response: %v", err),
		}
	}

	if result.Confidence < 0 || result.Confidence > 1 {
		result.Confidence = 0.5
	}

	if result.Intent == "" {
		result.Intent = types.IntentGeneral
	}

	return Classification{
		Intent:      result.Intent,
		Confidence:  result.Confidence,
		Description: result.Description,
		Tags:        result.Tags,
	}
}

// extractWorkspace finds workspace paths from input patterns.
func extractWorkspace(input string) string {
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`--workspace\s+(\S+)`),
		regexp.MustCompile(`in\s+(/[^\s,;]+)`),
		regexp.MustCompile(`at\s+(/[^\s,;]+)`),
		regexp.MustCompile(`\./([^\s,;]+)`),
	}

	for _, re := range patterns {
		matches := re.FindStringSubmatch(input)
		if len(matches) >= 2 {
			return matches[1]
		}
	}

	return ""
}

// extractTags pulls meaningful keywords from input.
func extractTags(input string) []string {
	words := strings.Fields(input)
	tags := make([]string, 0, len(words))

	for _, w := range words {
		clean := strings.ToLower(strings.Trim(w, ".,!?;:\"'()[]{}"))
		if len(clean) < 3 {
			continue
		}
		if stopWords[clean] {
			continue
		}
		tags = append(tags, clean)
	}

	return tags
}

// buildDescription creates a human-readable description for the classification.
func buildDescription(intent string, score int) string {
	if score == 0 {
		return "No specific intent detected"
	}

	return fmt.Sprintf("Detected %s intent with %d keyword matches", intent, score)
}
