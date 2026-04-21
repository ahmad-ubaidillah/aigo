// Package diary implements an agent diary system.
// The agent maintains daily diary entries recording its "thoughts",
// learnings, and observations.
package diary

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Diary manages agent diary entries.
type Diary struct {
	baseDir string
}

// Entry represents a single diary entry.
type Entry struct {
	Date    string `json:"date"`
	Time    string `json:"time"`
	Title   string `json:"title"`
	Content string `json:"content"`
	Mood    string `json:"mood,omitempty"` // happy, thoughtful, frustrated, excited
	Tags    []string `json:"tags,omitempty"`
}

// New creates a new Diary manager.
func New(baseDir string) *Diary {
	os.MkdirAll(baseDir, 0755)
	return &Diary{baseDir: baseDir}
}

// Write adds a new diary entry.
func (d *Diary) Write(entry Entry) error {
	date := entry.Date
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	filename := fmt.Sprintf("%s.md", date)
	path := filepath.Join(d.baseDir, filename)

	// Read existing entries for this day
	existing, _ := os.ReadFile(path)
	
	var timeStr string
	if entry.Time != "" {
		timeStr = entry.Time
	} else {
		timeStr = time.Now().Format("15:04")
	}

	moodEmoji := moodToEmoji(entry.Mood)
	
	var sb strings.Builder
	if len(existing) == 0 {
		sb.WriteString(fmt.Sprintf("# Diary: %s\n\n", date))
	}
	sb.Write(existing)
	sb.WriteString(fmt.Sprintf("\n## %s %s [%s]\n%s\n", timeStr, entry.Title, moodEmoji, entry.Content))
	
	if len(entry.Tags) > 0 {
		sb.WriteString(fmt.Sprintf("\n*Tags: %s*\n", strings.Join(entry.Tags, ", ")))
	}

	return os.WriteFile(path, []byte(sb.String()), 0644)
}

// Read reads all entries for a specific date.
func (d *Diary) Read(date string) (string, error) {
	filename := fmt.Sprintf("%s.md", date)
	path := filepath.Join(d.baseDir, filename)
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Recent reads the last N days of diary entries.
func (d *Diary) Recent(days int) (string, error) {
	entries, err := os.ReadDir(d.baseDir)
	if err != nil {
		return "", err
	}

	var parts []string
	limit := len(entries)
	if limit > days {
		limit = days
	}

	for i := len(entries) - limit; i < len(entries); i++ {
		if entries[i].IsDir() || !strings.HasSuffix(entries[i].Name(), ".md") {
			continue
		}
		data, _ := os.ReadFile(filepath.Join(d.baseDir, entries[i].Name()))
		if len(data) > 0 {
			parts = append(parts, string(data))
		}
	}

	return strings.Join(parts, "\n\n---\n\n"), nil
}

// WriteAuto is called by the autonomous agent to auto-generate a diary entry.
func (d *Diary) WriteAuto(content string, mood string) error {
	return d.Write(Entry{
		Title:   "Daily Reflection",
		Content: content,
		Mood:    mood,
		Tags:    []string{"auto-generated"},
	})
}

func moodToEmoji(mood string) string {
	switch strings.ToLower(mood) {
	case "happy":
		return "😊"
	case "thoughtful":
		return "🤔"
	case "frustrated":
		return "😤"
	case "excited":
		return "🎉"
	case "calm":
		return "😌"
	case "curious":
		return "🧐"
	default:
		return "📝"
	}
}
