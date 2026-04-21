// Package learntools implements learning and memory tools for Aigo.
package learntools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/hermes-v2/aigo/internal/tools"
)

func RegisterLearningTools(reg *tools.Registry, storagePath string) {
	store := &KnowledgeStore{basePath: storagePath}
	store.load()
	reg.Register(&LearnTool{store: store})
	reg.Register(&RecallTool{store: store})
	reg.Register(&KnowledgeListTool{store: store})
}

type KnowledgeStore struct {
	basePath string
	entries  []KnowledgeEntry
	mu       sync.RWMutex
}

type KnowledgeEntry struct {
	ID        string    `json:"id"`
	Category  string    `json:"category"`
	Topic     string    `json:"topic"`
	Content   string    `json:"content"`
	Source    string    `json:"source"`
	CreatedAt time.Time `json:"created_at"`
}

func (ks *KnowledgeStore) filePath() string {
	return filepath.Join(ks.basePath, "knowledge.json")
}

func (ks *KnowledgeStore) load() {
	ks.mu.Lock()
	defer ks.mu.Unlock()
	data, err := os.ReadFile(ks.filePath())
	if err != nil {
		ks.entries = []KnowledgeEntry{}
		return
	}
	json.Unmarshal(data, &ks.entries)
}

func (ks *KnowledgeStore) save() {
	os.MkdirAll(ks.basePath, 0755)
	data, _ := json.MarshalIndent(ks.entries, "", "  ")
	os.WriteFile(ks.filePath(), data, 0644)
}

func (ks *KnowledgeStore) add(entry KnowledgeEntry) {
	ks.mu.Lock()
	defer ks.mu.Unlock()
	entry.ID = fmt.Sprintf("k_%d", time.Now().UnixNano())
	entry.CreatedAt = time.Now()
	ks.entries = append(ks.entries, entry)
	ks.save()
}

func (ks *KnowledgeStore) search(query string, limit int) []KnowledgeEntry {
	ks.mu.RLock()
	defer ks.mu.RUnlock()
	q := strings.ToLower(query)
	var results []KnowledgeEntry
	for i := len(ks.entries) - 1; i >= 0; i-- {
		e := ks.entries[i]
		if strings.Contains(strings.ToLower(e.Content), q) || strings.Contains(strings.ToLower(e.Topic), q) {
			results = append(results, e)
			if len(results) >= limit {
				break
			}
		}
	}
	return results
}

func (ks *KnowledgeStore) list(limit int) []KnowledgeEntry {
	ks.mu.RLock()
	defer ks.mu.RUnlock()
	if limit <= 0 || limit > len(ks.entries) {
		limit = len(ks.entries)
	}
	start := len(ks.entries) - limit
	if start < 0 {
		start = 0
	}
	return ks.entries[start:]
}

// --- learn ---
type LearnTool struct{ store *KnowledgeStore }

func (t *LearnTool) Name() string        { return "learn" }
func (t *LearnTool) Description() string { return "Learn and remember something. Saves knowledge permanently across sessions." }
func (t *LearnTool) Annotations() tools.Annotations {
	return tools.Annotations{Destructive: false, ReadOnly: false}
}
func (t *LearnTool) Schema() tools.ToolSchema {
	return tools.ToolSchema{
		Type: "function",
		Function: tools.ToolFunctionSchema{
			Name:        "learn",
			Description: "Learn and remember something. Saves knowledge permanently across sessions. Use to remember user preferences, corrections, facts, or procedures.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"category": map[string]interface{}{"type": "string", "enum": []string{"preference", "fact", "correction", "skill"}, "description": "Type of knowledge"},
					"topic":    map[string]interface{}{"type": "string", "description": "Short topic label"},
					"content":  map[string]interface{}{"type": "string", "description": "The knowledge to remember"},
				},
				"required": []string{"category", "topic", "content"},
			},
		},
	}
}

func (t *LearnTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	category, _ := args["category"].(string)
	topic, _ := args["topic"].(string)
	content, _ := args["content"].(string)
	if category == "" || topic == "" || content == "" {
		return "", fmt.Errorf("category, topic, and content required")
	}
	t.store.add(KnowledgeEntry{Category: category, Topic: topic, Content: content, Source: "session"})
	return fmt.Sprintf("Learned [%s] %s", category, topic), nil
}

// --- recall ---
type RecallTool struct{ store *KnowledgeStore }

func (t *RecallTool) Name() string        { return "recall" }
func (t *RecallTool) Description() string { return "Search and recall previously learned knowledge." }
func (t *RecallTool) Annotations() tools.Annotations {
	return tools.Annotations{Destructive: false, ReadOnly: true}
}
func (t *RecallTool) Schema() tools.ToolSchema {
	return tools.ToolSchema{
		Type: "function",
		Function: tools.ToolFunctionSchema{
			Name:        "recall",
			Description: "Search previously learned knowledge. Use BEFORE answering about user preferences or past corrections.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{"type": "string", "description": "What to search for"},
					"limit": map[string]interface{}{"type": "integer", "description": "Max results (default 5)"},
				},
				"required": []string{"query"},
			},
		},
	}
}

func (t *RecallTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	query, _ := args["query"].(string)
	limit := 5
	if l, ok := args["limit"].(float64); ok && l > 0 {
		limit = int(l)
	}
	results := t.store.search(query, limit)
	if len(results) == 0 {
		return fmt.Sprintf("No knowledge for: %s", query), nil
	}
	var parts []string
	for _, e := range results {
		parts = append(parts, fmt.Sprintf("[%s] %s: %s", e.Category, e.Topic, e.Content))
	}
	return strings.Join(parts, "\n"), nil
}

// --- knowledge_list ---
type KnowledgeListTool struct{ store *KnowledgeStore }

func (t *KnowledgeListTool) Name() string        { return "knowledge_list" }
func (t *KnowledgeListTool) Description() string { return "List all learned knowledge." }
func (t *KnowledgeListTool) Annotations() tools.Annotations {
	return tools.Annotations{Destructive: false, ReadOnly: true}
}
func (t *KnowledgeListTool) Schema() tools.ToolSchema {
	return tools.ToolSchema{
		Type: "function",
		Function: tools.ToolFunctionSchema{
			Name:        "knowledge_list",
			Description: "List all learned knowledge entries.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"limit": map[string]interface{}{"type": "integer", "description": "Max entries (default 20)"},
				},
			},
		},
	}
}

func (t *KnowledgeListTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	limit := 20
	if l, ok := args["limit"].(float64); ok && l > 0 {
		limit = int(l)
	}
	entries := t.store.list(limit)
	if len(entries) == 0 {
		return "No knowledge yet. Use 'learn' to start.", nil
	}
	var parts []string
	for _, e := range entries {
		parts = append(parts, fmt.Sprintf("[%s] %s: %s", e.Category, e.Topic, e.Content))
	}
	return fmt.Sprintf("Knowledge (%d):\n%s", len(parts), strings.Join(parts, "\n")), nil
}
