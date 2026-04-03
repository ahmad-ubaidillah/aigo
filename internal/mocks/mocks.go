package mocks

import (
	"github.com/ahmad-ubaidillah/aigo/pkg/types"
)

type MockSessionDB struct {
	Sessions map[string]*types.Session
	Messages map[string][]types.Message
	Tasks    map[string][]types.Task
	Memories map[string][]types.Memory
}

func NewMockSessionDB() *MockSessionDB {
	return &MockSessionDB{
		Sessions: make(map[string]*types.Session),
		Messages: make(map[string][]types.Message),
		Tasks:    make(map[string][]types.Task),
		Memories: make(map[string][]types.Memory),
	}
}

func (m *MockSessionDB) CreateSession(name, workspace string) (*types.Session, error) {
	id := "test-session-" + name
	session := &types.Session{
		ID:        id,
		Name:      name,
		Workspace: workspace,
	}
	m.Sessions[id] = session
	return session, nil
}

func (m *MockSessionDB) GetSession(id string) (*types.Session, error) {
	return m.Sessions[id], nil
}

func (m *MockSessionDB) AddMessage(sessionID string, role, content string) error {
	msg := types.Message{
		Role:      role,
		Content:   content,
		SessionID: sessionID,
	}
	m.Messages[sessionID] = append(m.Messages[sessionID], msg)
	return nil
}

func (m *MockSessionDB) GetMessages(sessionID string, limit int) ([]types.Message, error) {
	return m.Messages[sessionID], nil
}

func (m *MockSessionDB) CreateTask(sessionID, description, workspace string) (*types.Task, error) {
	task := &types.Task{
		SessionID:   sessionID,
		Description: description,
		Workspace:   workspace,
		Status:      types.TaskPending,
	}
	m.Tasks[sessionID] = append(m.Tasks[sessionID], *task)
	return task, nil
}

func (m *MockSessionDB) UpdateTask(id string, status string) error {
	return nil
}

func (m *MockSessionDB) GetTasks(sessionID string, limit int) ([]types.Task, error) {
	return m.Tasks[sessionID], nil
}

func (m *MockSessionDB) AddMemory(sessionID, content, category string, tags []string) (*types.Memory, error) {
	mem := &types.Memory{
		Content:  content,
		Category: category,
		Tags:     category,
	}
	m.Memories[sessionID] = append(m.Memories[sessionID], *mem)
	return mem, nil
}

func (m *MockSessionDB) SearchMemory(query string, limit int) ([]types.Memory, error) {
	var results []types.Memory
	for _, mems := range m.Memories {
		for _, mem := range mems {
			if len(results) >= limit {
				break
			}
			results = append(results, mem)
		}
	}
	return results, nil
}
