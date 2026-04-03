package memory

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type WisdomEntry struct {
	ID        string    `json:"id"`
	Pattern   string    `json:"pattern"`
	Lesson    string    `json:"lesson"`
	Frequency int       `json:"frequency"`
	FirstSeen time.Time `json:"first_seen"`
	LastSeen  time.Time `json:"last_seen"`
}

type WisdomStore struct {
	entries  map[string]*WisdomEntry
	patterns map[string]int
	mu       sync.RWMutex
	filePath string
}

func NewWisdomStore(filePath string) (*WisdomStore, error) {
	w := &WisdomStore{
		entries:  make(map[string]*WisdomEntry),
		patterns: make(map[string]int),
		filePath: filePath,
	}
	if filePath != "" {
		if err := w.Load(); err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("load wisdom: %w", err)
		}
	}
	return w, nil
}

func (w *WisdomStore) RecordPattern(pattern string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.patterns[pattern]++
}

func (w *WisdomStore) RecognizePatterns() []string {
	w.mu.RLock()
	defer w.mu.RUnlock()
	var recognized []string
	for pattern, count := range w.patterns {
		if count >= 3 {
			recognized = append(recognized, pattern)
		}
	}
	return recognized
}

func (w *WisdomStore) AddLesson(pattern, lesson string) string {
	w.mu.Lock()
	defer w.mu.Unlock()

	id := fmt.Sprintf("wisdom-%d", time.Now().UnixNano())
	now := time.Now()

	if entry, ok := w.entries[pattern]; ok {
		entry.Frequency++
		entry.LastSeen = now
		entry.Lesson = lesson
		return entry.ID
	}

	w.entries[pattern] = &WisdomEntry{
		ID:        id,
		Pattern:   pattern,
		Lesson:    lesson,
		Frequency: 1,
		FirstSeen: now,
		LastSeen:  now,
	}
	return id
}

func (w *WisdomStore) GetLessons(pattern string) []WisdomEntry {
	w.mu.RLock()
	defer w.mu.RUnlock()
	var results []WisdomEntry
	for _, entry := range w.entries {
		if pattern == "" || entry.Pattern == pattern {
			results = append(results, *entry)
		}
	}
	return results
}

func (w *WisdomStore) GetTopLessons(n int) []WisdomEntry {
	w.mu.RLock()
	defer w.mu.RUnlock()

	all := make([]*WisdomEntry, 0, len(w.entries))
	for _, e := range w.entries {
		all = append(all, e)
	}

	for i := 0; i < len(all) && i < n; i++ {
		maxIdx := i
		for j := i + 1; j < len(all); j++ {
			if all[j].Frequency > all[maxIdx].Frequency {
				maxIdx = j
			}
		}
		all[i], all[maxIdx] = all[maxIdx], all[i]
	}

	results := make([]WisdomEntry, 0, n)
	for i := 0; i < len(all) && i < n; i++ {
		results = append(results, *all[i])
	}
	return results
}

func (w *WisdomStore) Save() error {
	w.mu.RLock()
	defer w.mu.RUnlock()

	if w.filePath == "" {
		return nil
	}

	data := struct {
		Entries  map[string]*WisdomEntry `json:"entries"`
		Patterns map[string]int          `json:"patterns"`
	}{
		Entries:  w.entries,
		Patterns: w.patterns,
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal wisdom: %w", err)
	}

	dir := filepath.Dir(w.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create wisdom dir: %w", err)
	}

	return os.WriteFile(w.filePath, jsonData, 0644)
}

func (w *WisdomStore) Load() error {
	if w.filePath == "" {
		return nil
	}

	data, err := os.ReadFile(w.filePath)
	if err != nil {
		return err
	}

	var stored struct {
		Entries  map[string]*WisdomEntry `json:"entries"`
		Patterns map[string]int          `json:"patterns"`
	}
	if err := json.Unmarshal(data, &stored); err != nil {
		return fmt.Errorf("unmarshal wisdom: %w", err)
	}

	w.mu.Lock()
	defer w.mu.Unlock()
	w.entries = stored.Entries
	w.patterns = stored.Patterns
	if w.entries == nil {
		w.entries = make(map[string]*WisdomEntry)
	}
	if w.patterns == nil {
		w.patterns = make(map[string]int)
	}
	return nil
}

func (w *WisdomStore) Count() int {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return len(w.entries)
}
