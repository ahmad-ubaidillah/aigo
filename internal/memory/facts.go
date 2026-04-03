package memory

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/ahmad-ubaidillah/aigo/internal/llm"
)

const (
	FactAdd    = "ADD"
	FactUpdate = "UPDATE"
	FactDelete = "DELETE"
	FactNone   = "NONE"
)

type MemoryFact struct {
	ID        string
	Content   string
	Source    string
	Action    string
	Timestamp time.Time
	UserID    string
	AgentID   string
}

type FactExtractor struct {
	client llm.LLMClient
	model  string
	facts  []MemoryFact
	mu     sync.RWMutex
	nextID int
}

func NewFactExtractor(client llm.LLMClient, model string) *FactExtractor {
	return &FactExtractor{
		client: client,
		model:  model,
		facts:  make([]MemoryFact, 0),
	}
}

func (e *FactExtractor) ExtractFact(conversation string) MemoryFact {
	return MemoryFact{
		Action:    FactNone,
		Content:   "No facts extracted",
		Timestamp: time.Now(),
	}
}

func (e *FactExtractor) ExtractFactWithLLM(ctx context.Context, conversation string) ([]MemoryFact, error) {
	if e.client == nil {
		return e.extractHeuristic(conversation), nil
	}

	prompt := fmt.Sprintf(`Extract factual information from this conversation.
Return facts as a JSON array of objects with: content, category, source.
Categories: preference, decision, error, fact, context.

Conversation:
%s`, conversation)

	messages := []llm.Message{
		{Role: "system", Content: "Extract only concrete facts. Ignore opinions and speculation."},
		{Role: "user", Content: prompt},
	}

	resp, err := e.client.Chat(ctx, messages)
	if err != nil {
		return e.extractHeuristic(conversation), nil
	}

	return e.parseLLMResponse(resp.Content), nil
}

func (e *FactExtractor) extractHeuristic(conversation string) []MemoryFact {
	var facts []MemoryFact
	lines := strings.Split(conversation, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		lower := strings.ToLower(line)
		if strings.Contains(lower, "prefer") || strings.Contains(lower, "always") {
			facts = append(facts, MemoryFact{
				ID:        e.nextIDString(),
				Content:   line,
				Source:    "heuristic",
				Action:    FactAdd,
				Timestamp: time.Now(),
			})
		}
		if strings.Contains(lower, "error") || strings.Contains(lower, "failed") {
			facts = append(facts, MemoryFact{
				ID:        e.nextIDString(),
				Content:   line,
				Source:    "heuristic",
				Action:    FactAdd,
				Timestamp: time.Now(),
			})
		}
	}
	return facts
}

func (e *FactExtractor) parseLLMResponse(content string) []MemoryFact {
	var facts []MemoryFact
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || line == "[" || line == "]" {
			continue
		}
		facts = append(facts, MemoryFact{
			ID:        e.nextIDString(),
			Content:   line,
			Source:    "llm",
			Action:    FactAdd,
			Timestamp: time.Now(),
		})
	}
	return facts
}

func (e *FactExtractor) nextIDString() string {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.nextID++
	return fmt.Sprintf("fact-%d", e.nextID)
}

func (e *FactExtractor) AddFact(content, source, userID, agentID string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.nextID++
	e.facts = append(e.facts, MemoryFact{
		ID:        fmt.Sprintf("fact-%d", e.nextID),
		Content:   content,
		Source:    source,
		Action:    FactAdd,
		Timestamp: time.Now(),
		UserID:    userID,
		AgentID:   agentID,
	})
}

func (e *FactExtractor) GetFacts(userID string) []MemoryFact {
	e.mu.RLock()
	defer e.mu.RUnlock()
	var result []MemoryFact
	for _, f := range e.facts {
		if f.UserID == userID {
			result = append(result, f)
		}
	}
	return result
}

func (e *FactExtractor) UpdateFact(id, content string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	for i := range e.facts {
		if e.facts[i].ID == id {
			e.facts[i].Content = content
			e.facts[i].Action = FactUpdate
			return nil
		}
	}
	return fmt.Errorf("fact %s not found", id)
}

func (e *FactExtractor) DeleteFact(id string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	for i := range e.facts {
		if e.facts[i].ID == id {
			e.facts[i].Action = FactDelete
			e.facts = append(e.facts[:i], e.facts[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("fact %s not found", id)
}

func (e *FactExtractor) ListFacts() []MemoryFact {
	e.mu.RLock()
	defer e.mu.RUnlock()
	out := make([]MemoryFact, len(e.facts))
	copy(out, e.facts)
	return out
}

func (e *FactExtractor) Count() int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return len(e.facts)
}
