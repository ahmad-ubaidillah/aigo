package wisdom

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

type Learning struct {
	ID        string
	Task      string
	Lesson    string
	Pattern   string
	Timestamp time.Time
	Relevance float64
}

type WisdomStore struct {
	learnings []Learning
	mu        sync.RWMutex
	nextID    int
}

func NewWisdomStore() *WisdomStore {
	return &WisdomStore{learnings: make([]Learning, 0)}
}

func (w *WisdomStore) AddLearning(task, lesson, pattern string) string {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.nextID++
	id := fmt.Sprintf("learn-%d", w.nextID)
	w.learnings = append(w.learnings, Learning{
		ID:        id,
		Task:      task,
		Lesson:    lesson,
		Pattern:   pattern,
		Timestamp: time.Now(),
		Relevance: 1.0,
	})
	return id
}

func (w *WisdomStore) GetLearnings(task string) []Learning {
	w.mu.RLock()
	defer w.mu.RUnlock()
	var result []Learning
	for _, l := range w.learnings {
		if strings.Contains(strings.ToLower(l.Task), strings.ToLower(task)) {
			result = append(result, l)
		}
	}
	return result
}

func (w *WisdomStore) GetPatterns() []string {
	w.mu.RLock()
	defer w.mu.RUnlock()
	seen := make(map[string]bool)
	var patterns []string
	for _, l := range w.learnings {
		if l.Pattern != "" && !seen[l.Pattern] {
			seen[l.Pattern] = true
			patterns = append(patterns, l.Pattern)
		}
	}
	return patterns
}

func (w *WisdomStore) FindRelevant(query string, topK int) []Learning {
	w.mu.RLock()
	defer w.mu.RUnlock()
	q := strings.ToLower(query)
	scored := make([]Learning, len(w.learnings))
	copy(scored, w.learnings)
	for i := range scored {
		scored[i].Relevance = 0
		if strings.Contains(strings.ToLower(scored[i].Task), q) {
			scored[i].Relevance += 0.5
		}
		if strings.Contains(strings.ToLower(scored[i].Lesson), q) {
			scored[i].Relevance += 0.3
		}
		if strings.Contains(strings.ToLower(scored[i].Pattern), q) {
			scored[i].Relevance += 0.2
		}
	}
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].Relevance > scored[j].Relevance
	})
	if topK > 0 && len(scored) > topK {
		scored = scored[:topK]
	}
	return scored
}

func (w *WisdomStore) InjectWisdom(task string) string {
	relevant := w.FindRelevant(task, 3)
	if len(relevant) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("## Relevant Wisdom\n")
	for _, l := range relevant {
		b.WriteString(fmt.Sprintf("- [%s] %s\n", l.Task, l.Lesson))
	}
	return b.String()
}

func (w *WisdomStore) Clear() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.learnings = make([]Learning, 0)
	w.nextID = 0
}

func (w *WisdomStore) Count() int {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return len(w.learnings)
}
