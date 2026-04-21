package plan

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Plan struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	Tasks       []Task    `json:"tasks"`
	CurrentTask int       `json:"current_task"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Branch      string     `json:"branch"`
	ParentBranch string   `json:"parent_branch"`
}

type Task struct {
	ID          string   `json:"id"`
	Index       int      `json:"index"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Status      string   `json:"status"`
	Result      string   `json:"result,omitempty"`
	Changes     []Change `json:"changes,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	Output      string   `json:"output,omitempty"`
}

type Change struct {
	File    string `json:"file"`
	Content string `json:"content"`
	Type    string `json:"type"`
	Patch   string `json:"patch,omitempty"`
}

type Manager struct {
	basePath  string
	plans    map[string]*Plan
	current *Plan
}

func New(basePath string) (*Manager, error) {
	if basePath == "" {
		home, _ := os.UserHomeDir()
		basePath = filepath.Join(home, ".aigo", "memory", "plans")
	}

	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, err
	}

	m := &Manager{
		basePath: basePath,
		plans:   make(map[string]*Plan),
	}

	m.loadAll()
	return m, nil
}

func (m *Manager) Create(name, description string) (*Plan, error) {
	plan := &Plan{
		ID:          generateID(name + time.Now().Format("20060102150405")),
		Name:        name,
		Description: description,
		Status:      "active",
		Tasks:       []Task{},
		CurrentTask: 0,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Branch:      sanitizeBranchName(name),
	}

	m.plans[plan.ID] = plan
	m.current = plan

	if err := m.save(plan); err != nil {
		return nil, err
	}

	return plan, nil
}

func (m *Manager) AddTask(planID, title, description string) (*Task, error) {
	plan, ok := m.plans[planID]
	if !ok {
		return nil, fmt.Errorf("plan not found: %s", planID)
	}

	task := Task{
		ID:          generateID(title + fmt.Sprintf("%d", len(plan.Tasks))),
		Index:       len(plan.Tasks),
		Title:       title,
		Description: description,
		Status:      "pending",
	}

	plan.Tasks = append(plan.Tasks, task)
	plan.UpdatedAt = time.Now()

	m.save(plan)
	return &task, nil
}

func (m *Manager) UpdateTask(planID, taskID string, result string, changes []Change) error {
	plan, ok := m.plans[planID]
	if !ok {
		return fmt.Errorf("plan not found: %s", planID)
	}

	for i := range plan.Tasks {
		if plan.Tasks[i].ID == taskID {
			plan.Tasks[i].Status = "completed"
			plan.Tasks[i].Result = result
			plan.Tasks[i].Changes = changes
			now := time.Now()
			plan.Tasks[i].CompletedAt = &now

			if i+1 < len(plan.Tasks) {
				plan.CurrentTask = i + 1
				plan.Tasks[i+1].Status = "in_progress"
			} else {
				plan.Status = "completed"
			}

			plan.UpdatedAt = time.Now()
			return m.save(plan)
		}
	}

	return fmt.Errorf("task not found: %s", taskID)
}

func (m *Manager) GetTask(planID, taskID string) (*Task, error) {
	plan, ok := m.plans[planID]
	if !ok {
		return nil, fmt.Errorf("plan not found: %s", planID)
	}

	for i := range plan.Tasks {
		if plan.Tasks[i].ID == taskID {
			return &plan.Tasks[i], nil
		}
	}

	return nil, fmt.Errorf("task not found: %s", taskID)
}

func (m *Manager) GetCurrentTask(planID string) (*Task, error) {
	plan, ok := m.plans[planID]
	if !ok {
		return nil, fmt.Errorf("plan not found: %s", planID)
	}

	if plan.CurrentTask < len(plan.Tasks) {
		return &plan.Tasks[plan.CurrentTask], nil
	}

	return nil, nil
}

func (m *Manager) Get(planID string) *Plan {
	return m.plans[planID]
}

func (m *Manager) List() []*Plan {
	var list []*Plan
	for _, p := range m.plans {
		list = append(list, p)
	}
	return list
}

func (m *Manager) Current() *Plan {
	return m.current
}

func (m *Manager) SetCurrent(planID string) error {
	plan, ok := m.plans[planID]
	if !ok {
		return fmt.Errorf("plan not found: %s", planID)
	}

	m.current = plan
	return nil
}

func (m *Manager) CompleteTask(planID, taskID string) error {
	return m.UpdateTask(planID, taskID, "", nil)
}

func (m *Manager) FailTask(planID, taskID, errMsg string) error {
	plan, ok := m.plans[planID]
	if !ok {
		return fmt.Errorf("plan not found: %s", planID)
	}

	for i := range plan.Tasks {
		if plan.Tasks[i].ID == taskID {
			plan.Tasks[i].Status = "failed"
			plan.Tasks[i].Output = errMsg

			plan.UpdatedAt = time.Now()
			return m.save(plan)
		}
	}

	return fmt.Errorf("task not found: %s", taskID)
}

func (m *Manager) RevertTask(planID, taskID string) error {
	plan, ok := m.plans[planID]
	if !ok {
		return fmt.Errorf("plan not found: %s", planID)
	}

	for i := range plan.Tasks {
		if plan.Tasks[i].ID == taskID {
			plan.Tasks[i].Status = "pending"
			plan.Tasks[i].Result = ""
			plan.Tasks[i].Changes = nil
			plan.Tasks[i].CompletedAt = nil

			plan.CurrentTask = i
			plan.Status = "active"

			plan.UpdatedAt = time.Now()
			return m.save(plan)
		}
	}

	return fmt.Errorf("task not found: %s", taskID)
}

func (m *Manager) GetChanges(planID string) []Change {
	plan, ok := m.plans[planID]
	if !ok {
		return nil
	}

	var changes []Change
	for _, task := range plan.Tasks {
		if task.Status == "completed" {
			changes = append(changes, task.Changes...)
		}
	}

	return changes
}

func (m *Manager) GetAllChanges(planID string) map[string]string {
	changes := m.GetChanges(planID)
	result := make(map[string]string)

	for _, c := range changes {
		result[c.File] = c.Content
	}

	return result
}

func (m *Manager) Delete(planID string) error {
	if _, ok := m.plans[planID]; !ok {
		return fmt.Errorf("plan not found: %s", planID)
	}

	delete(m.plans, planID)

	planFile := filepath.Join(m.basePath, planID+".json")
	return os.Remove(planFile)
}

func (m *Manager) save(plan *Plan) error {
	data, err := json.MarshalIndent(plan, "", "  ")
	if err != nil {
		return err
	}

	planFile := filepath.Join(m.basePath, plan.ID+".json")
	return os.WriteFile(planFile, data, 0644)
}

func (m *Manager) loadAll() {
	entries, err := os.ReadDir(m.basePath)
	if err != nil {
		return
	}

	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".json") {
			continue
		}

		data, err := os.ReadFile(filepath.Join(m.basePath, e.Name()))
		if err != nil {
			continue
		}

		var plan Plan
		if err := json.Unmarshal(data, &plan); err != nil {
			continue
		}

		m.plans[plan.ID] = &plan
	}
}

func generateID(input string) string {
	hash := 0
	for _, c := range input {
		hash = hash*31 + int(c)
	}
	return fmt.Sprintf("%x", hash)
}

func sanitizeBranchName(name string) string {
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, " ", "-")
	name = strings.ReplaceAll(name, "_", "-")

	var result []rune
	for _, c := range name {
		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' {
			result = append(result, c)
		}
	}

	if len(result) > 50 {
		result = result[:50]
	}

	return string(result)
}

func (m *Manager) Stats() map[string]int {
	stats := make(map[string]int)
	stats["total_plans"] = len(m.plans)

	active := 0
	completed := 0
	totalTasks := 0

	for _, p := range m.plans {
		if p.Status == "active" {
			active++
		} else if p.Status == "completed" {
			completed++
		}
		totalTasks += len(p.Tasks)
	}

	stats["active"] = active
	stats["completed"] = completed
	stats["total_tasks"] = totalTasks

	return stats
}