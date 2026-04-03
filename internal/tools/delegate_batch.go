package tools

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ahmad-ubaidillah/aigo/pkg/types"
)

// BatchResult holds results from parallel delegate execution.
type BatchResult struct {
	ChildID string
	Result  *types.ToolResult
	Error   error
}

// BatchExecute runs multiple child agents in parallel.
func (d *DelegateTool) BatchExecute(ctx context.Context, tasks []map[string]any) []BatchResult {
	results := make([]BatchResult, len(tasks))
	var wg sync.WaitGroup

	for i, t := range tasks {
		wg.Add(1)
		go func(idx int, task map[string]any) {
			defer wg.Done()
			desc, _ := task["description"].(string)
			category, _ := task["category"].(string)
			parentID, _ := task["session_id"].(string)

			id := d.SpawnChild(parentID, desc, category, 1)
			results[idx].ChildID = id

			result, err := d.Execute(ctx, task)
			results[idx].Result = result
			results[idx].Error = err
		}(i, t)
	}

	wg.Wait()
	return results
}

// ProgressCallback registers a callback for child session updates.
func (d *DelegateTool) ProgressCallback(sessionID string, callback func(status, result string)) {
	d.mu.RLock()
	child, ok := d.sessions[sessionID]
	d.mu.RUnlock()

	if !ok {
		return
	}

	for child.Status == "pending" || child.Status == "initialized" {
		time.Sleep(100 * time.Millisecond)
		d.mu.RLock()
		status := child.Status
		res := child.Result
		d.mu.RUnlock()

		if callback != nil {
			callback(status, res)
		}
	}
}

// GetChildProgress returns the current status of a child session.
func (d *DelegateTool) GetChildProgress(sessionID string) (string, string, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	child, ok := d.sessions[sessionID]
	if !ok {
		return "", "", fmt.Errorf("child session %s not found", sessionID)
	}
	return child.Status, child.Result, nil
}

// ListAllChildren returns all child sessions with their status.
func (d *DelegateTool) ListAllChildren() []ChildSession {
	d.mu.RLock()
	defer d.mu.RUnlock()

	result := make([]ChildSession, 0, len(d.sessions))
	for _, s := range d.sessions {
		result = append(result, *s)
	}
	return result
}

// CleanupOldSessions removes sessions older than the given duration.
func (d *DelegateTool) CleanupOldSessions(maxAge time.Duration) int {
	d.mu.Lock()
	defer d.mu.Unlock()

	count := 0
	cutoff := time.Now().Add(-maxAge)
	for id, s := range d.sessions {
		if s.CreatedAt.Before(cutoff) {
			delete(d.sessions, id)
			count++
		}
	}
	return count
}
