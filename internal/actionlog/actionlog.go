package actionlog

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Action struct {
	ID          string                 `json:"id"`
	Timestamp   time.Time              `json:"timestamp"`
	Type        string                 `json:"type"`
	Project     string                 `json:"project"`
	Description string                 `json:"description"`
	File        string                 `json:"file,omitempty"`
	OldContent  string                 `json:"old_content,omitempty"`
	NewContent  string                 `json:"new_content,omitempty"`
	Reversible  bool                   `json:"reversible"`
	Reversed    bool                   `json:"reversed"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

type ActionLog struct {
	actions   []Action
	byProject map[string][]Action
	basePath  string
	maxItems  int
}

func New(basePath string, maxItems int) (*ActionLog, error) {
	if basePath == "" {
		home, _ := os.UserHomeDir()
		basePath = filepath.Join(home, ".aigo", "memory", "actions")
	}

	if maxItems <= 0 {
		maxItems = 1000
	}

	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, err
	}

	l := &ActionLog{
		actions:   make([]Action, 0, maxItems),
		byProject: make(map[string][]Action),
		basePath:  basePath,
		maxItems:  maxItems,
	}

	l.loadAll()
	return l, nil
}

func (l *ActionLog) Log(action Action) error {
	action.Timestamp = time.Now()

	if action.ID == "" {
		action.ID = generateID(action.Timestamp, action.Type, action.Description)
	}

	l.actions = append(l.actions, action)

	project := action.Project
	if project != "" {
		l.byProject[project] = append(l.byProject[project], action)
	}

	if len(l.actions) > l.maxItems {
		l.prune()
	}

	return l.save(action)
}

func (l *ActionLog) LogEdit(project, file, description, oldContent, newContent string) error {
	return l.Log(Action{
		Type:        "edit",
		Project:     project,
		Description: description,
		File:        file,
		OldContent:  oldContent,
		NewContent:  newContent,
		Reversible:  true,
	})
}

func (l *ActionLog) LogCreate(project, file, description, content string) error {
	return l.Log(Action{
		Type:        "create",
		Project:     project,
		Description: description,
		File:        file,
		NewContent:  content,
		Reversible:  true,
	})
}

func (l *ActionLog) LogDelete(project, file, description, content string) error {
	return l.Log(Action{
		Type:        "delete",
		Project:     project,
		Description: description,
		File:        file,
		OldContent:  content,
		Reversible:  true,
	})
}

func (l *ActionLog) LogCommand(project, description, command string) error {
	return l.Log(Action{
		Type:        "command",
		Project:     project,
		Description: description,
		Metadata:    map[string]interface{}{"command": command},
		Reversible:  false,
	})
}

func (l *ActionLog) LogError(project, description, errorMsg string) error {
	return l.Log(Action{
		Type:        "error",
		Project:     project,
		Description: description,
		Metadata:    map[string]interface{}{"error": errorMsg},
		Reversible:  false,
	})
}

func (l *ActionLog) Undo(actionID string) (Action, error) {
	var targetIdx = -1

	for i := len(l.actions) - 1; i >= 0; i-- {
		if l.actions[i].ID == actionID {
			targetIdx = i
			break
		}
	}

	if targetIdx == -1 {
		return Action{}, fmt.Errorf("action not found: %s", actionID)
	}

	action := &l.actions[targetIdx]

	if !action.Reversible {
		return Action{}, fmt.Errorf("action is not reversible: %s", actionID)
	}

	if action.Reversed {
		return Action{}, fmt.Errorf("action already reversed: %s", actionID)
	}

	action.Reversed = true

	l.saveAll()

	return *action, nil
}

func (l *ActionLog) UndoLast(project string) (Action, error) {
	for i := len(l.actions) - 1; i >= 0; i-- {
		action := &l.actions[i]

		if project != "" && action.Project != project {
			continue
		}

		if action.Reversible && !action.Reversed {
			return l.Undo(action.ID)
		}
	}

	return Action{}, fmt.Errorf("no reversible action found for project: %s", project)
}

func (l *ActionLog) Get(actionID string) *Action {
	for i := range l.actions {
		if l.actions[i].ID == actionID {
			return &l.actions[i]
		}
	}
	return nil
}

func (l *ActionLog) List(project string, limit int) []Action {
	if limit <= 0 {
		limit = 50
	}

	var actions []Action

	if project != "" {
		projActions := l.byProject[project]
		start := 0
		if len(projActions) > limit {
			start = len(projActions) - limit
		}
		actions = projActions[start:]
	} else {
		start := 0
		if len(l.actions) > limit {
			start = len(l.actions) - limit
		}
		actions = l.actions[start:]
	}

	return actions
}

func (l *ActionLog) GetReversible(project string) []Action {
	var results []Action

	actions := l.byProject[project]
	for i := len(actions) - 1; i >= 0; i-- {
		action := actions[i]
		if action.Reversible && !action.Reversed {
			results = append(results, action)
			if len(results) >= 10 {
				break
			}
		}
	}

	return results
}

func (l *ActionLog) Search(project, query string) []Action {
	queryLower := strings.ToLower(query)
	var results []Action

	actions := l.byProject[project]
	for i := len(actions) - 1; i >= 0; i-- {
		action := actions[i]
		if strings.Contains(strings.ToLower(action.Description), queryLower) ||
			strings.Contains(strings.ToLower(action.File), queryLower) ||
			strings.Contains(strings.ToLower(action.Type), queryLower) {
			results = append(results, action)
			if len(results) >= 20 {
				break
			}
		}
	}

	return results
}

func (l *ActionLog) GetDiff(actionID string) (string, error) {
	action := l.Get(actionID)
	if action == nil {
		return "", fmt.Errorf("action not found: %s", actionID)
	}

	var diff string

	switch action.Type {
	case "edit":
		diff = fmt.Sprintf("--- %s\n+++ %s\n--- %s\n+++ %s\n",
			action.File, action.File,
			action.OldContent[:min(100, len(action.OldContent))],
			action.NewContent[:min(100, len(action.NewContent))])
	case "create":
		diff = fmt.Sprintf("+ %s (created)\n%s",
			action.File,
			action.NewContent[:min(200, len(action.NewContent))])
	case "delete":
		diff = fmt.Sprintf("- %s (deleted)\n%s",
			action.File,
			action.OldContent[:min(200, len(action.OldContent))])
	}

	return diff, nil
}

func (l *ActionLog) prune() {
	if len(l.actions) <= l.maxItems {
		return
	}

	toRemove := len(l.actions) - l.maxItems

	l.actions = l.actions[toRemove:]

	l.byProject = make(map[string][]Action)
	for _, action := range l.actions {
		l.byProject[action.Project] = append(l.byProject[action.Project], action)
	}
}

func (l *ActionLog) save(action Action) error {
	projDir := filepath.Join(l.basePath, action.Project)
	if err := os.MkdirAll(projDir, 0755); err != nil {
		return err
	}

	filename := fmt.Sprintf("%s_%s.json", action.Timestamp.Format("20060102_150405"), action.ID[:8])
	data, err := json.MarshalIndent(action, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(projDir, filename), data, 0644)
}

func (l *ActionLog) saveAll() error {
	index := make([]Action, 0, len(l.actions))
	for _, action := range l.actions {
		index = append(index, action)
	}

	data, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(l.basePath, "index.json"), data, 0644)
}

func (l *ActionLog) loadAll() {
	indexFile := filepath.Join(l.basePath, "index.json")
	data, err := os.ReadFile(indexFile)
	if err != nil {
		return
	}

	var actions []Action
	if err := json.Unmarshal(data, &actions); err != nil {
		return
	}

	l.actions = actions

	for _, action := range actions {
		l.byProject[action.Project] = append(l.byProject[action.Project], action)
	}
}

func (l *ActionLog) Stats() map[string]int {
	stats := make(map[string]int)
	stats["total"] = len(l.actions)

	typeCounts := make(map[string]int)
	projectCounts := make(map[string]int)
	reversible := 0
	reversed := 0

	for _, action := range l.actions {
		typeCounts[action.Type]++
		projectCounts[action.Project]++
		if action.Reversible {
			reversible++
		}
		if action.Reversed {
			reversed++
		}
	}

	for t, c := range typeCounts {
		stats["type_"+t] = c
	}

	stats["projects"] = len(projectCounts)
	stats["reversible"] = reversible
	stats["reversed"] = reversed

	return stats
}

func generateID(t time.Time, actionType, desc string) string {
	input := fmt.Sprintf("%s_%s_%s_%d", t.Format("20060102150405"), actionType, desc, t.Nanosecond())
	return fmt.Sprintf("%x", []byte(input))[:16]
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}