package minions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

type Job struct {
	ID          int64                  `json:"id"`
	Name        string                 `json:"name"`
	Params      map[string]interface{} `json:"params"`
	Status      string                 `json:"status"`
	ParentID    int64                  `json:"parent_id"`
	Depth       int                    `json:"depth"`
	Result      string                 `json:"result"`
	Error       string                 `json:"error"`
	CreatedAt   time.Time              `json:"created_at"`
	StartedAt   *time.Time             `json:"started_at"`
	CompletedAt *time.Time             `json:"completed_at"`
	TimeoutMs   int64                  `json:"timeout_ms"`
	MaxChildren int                    `json:"max_children"`
}

type Worker struct {
	jobQueue   *JobQueue
	handlers   map[string]JobHandler
	concurrency int
	running    bool
	mu         sync.RWMutex
}

type JobHandler func(ctx context.Context, job *Job) (string, error)

type JobQueue struct {
	db       *sql.DB
	dbPath   string
	mu       sync.RWMutex
	basePath string
}

const (
	StatusPending   = "pending"
	StatusRunning   = "running"
	StatusCompleted = "completed"
	StatusFailed    = "failed"
	StatusCancelled = "cancelled"
)

func New(basePath string) (*JobQueue, error) {
	if basePath == "" {
		home, _ := os.UserHomeDir()
		basePath = filepath.Join(home, ".aigo", "memory", "minions")
	}

	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("minions mkdir: %w", err)
	}

	dbPath := filepath.Join(basePath, "minions.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("minions open: %w", err)
	}

	q := &JobQueue{
		db:       db,
		dbPath:   dbPath,
		basePath: basePath,
	}

	if err := q.initSchema(); err != nil {
		db.Close()
		return nil, err
	}

	return q, nil
}

func (q *JobQueue) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS jobs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		params TEXT,
		status TEXT DEFAULT 'pending',
		parent_id INTEGER DEFAULT 0,
		depth INTEGER DEFAULT 0,
		result TEXT,
		error TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		started_at DATETIME,
		completed_at DATETIME,
		timeout_ms INTEGER DEFAULT 60000,
		max_children INTEGER DEFAULT 0,
		timeout_at DATETIME
	);

	CREATE INDEX IF NOT EXISTS idx_jobs_status ON jobs(status);
	CREATE INDEX IF NOT EXISTS idx_jobs_parent ON jobs(parent_id);
	CREATE INDEX IF NOT EXISTS idx_jobs_name ON jobs(name);
	`

	_, err := q.db.Exec(schema)
	return err
}

func (q *JobQueue) Submit(name string, params map[string]interface{}, parentID int64) (int64, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	paramsJSON, _ := json.Marshal(params)

	var depth int
	if parentID > 0 {
		var parentDepth int
		q.db.QueryRow("SELECT depth FROM jobs WHERE id = ?", parentID).Scan(&parentDepth)
		depth = parentDepth + 1
	}

	result, err := q.db.Exec(`
		INSERT INTO jobs (name, params, parent_id, depth, status)
		VALUES (?, ?, ?, ?, 'pending')
	`, name, string(paramsJSON), parentID, depth)

	if err != nil {
		return 0, err
	}

	return result.LastInsertId(), nil
}

func (q *JobQueue) SubmitBatch(name string, items []map[string]interface{}, parentID int64) ([]int64, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	var ids []int64

	for i, item := range items {
		if parentID > 0 && i == 0 {
			continue
		}

		paramsJSON, _ := json.Marshal(item)

		var depth int
		if parentID > 0 {
			var parentDepth int
			q.db.QueryRow("SELECT depth FROM jobs WHERE id = ?", parentID).Scan(&parentDepth)
			depth = parentDepth + 1
		}

		result, err := q.db.Exec(`
			INSERT INTO jobs (name, params, parent_id, depth, status)
			VALUES (?, ?, ?, ?, 'pending')
		`, name, string(paramsJSON), parentID, depth)

		if err != nil {
			continue
		}

		id, _ := result.LastInsertId()
		ids = append(ids, id)
	}

	return ids, nil
}

func (q *JobQueue) Get(id int64) (*Job, error) {
	q.mu.RLock()
	defer q.mu.RUnlock()

	var job Job
	var paramsJSON string
	var startedAt, completedAt sql.NullTime

	err := q.db.QueryRow(`
		SELECT id, name, params, status, parent_id, depth, result, error, 
		       created_at, started_at, completed_at, timeout_ms, max_children
		FROM jobs WHERE id = ?
	`, id).Scan(
		&job.ID, &job.Name, &paramsJSON, &job.Status, &job.ParentID, &job.Depth,
		&job.Result, &job.Error, &job.CreatedAt, &startedAt, &completedAt,
		&job.TimeoutMs, &job.MaxChildren,
	)

	if err != nil {
		return nil, err
	}

	json.Unmarshal([]byte(paramsJSON), &job.Params)

	if startedAt.Valid {
		job.StartedAt = &startedAt.Time
	}
	if completedAt.Valid {
		job.CompletedAt = &completedAt.Time
	}

	return &job, nil
}

func (q *JobQueue) GetPending(limit int) ([]Job, error) {
	q.mu.RLock()
	defer q.mu.RUnlock()

	if limit <= 0 {
		limit = 10
	}

	rows, err := q.db.Query(`
		SELECT id, name, params, status, parent_id, depth, result, error,
		       created_at, started_at, completed_at, timeout_ms, max_children
		FROM jobs 
		WHERE status = 'pending' AND (parent_id = 0 OR 
		       EXISTS (SELECT 1 FROM jobs j2 WHERE j2.id = jobs.parent_id AND j2.status = 'completed'))
		ORDER BY depth, created_at
		LIMIT ?
	`, limit)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []Job
	for rows.Next() {
		var job Job
		var paramsJSON string
		var startedAt, completedAt sql.NullTime

		err := rows.Scan(
			&job.ID, &job.Name, &paramsJSON, &job.Status, &job.ParentID, &job.Depth,
			&job.Result, &job.Error, &job.CreatedAt, &startedAt, &completedAt,
			&job.TimeoutMs, &job.MaxChildren,
		)
		if err != nil {
			continue
		}

		json.Unmarshal([]byte(paramsJSON), &job.Params)
		if startedAt.Valid {
			job.StartedAt = &startedAt.Time
		}
		if completedAt.Valid {
			job.CompletedAt = &completedAt.Time
		}

		jobs = append(jobs, job)
	}

	return jobs, nil
}

func (q *JobQueue) Start(id int64) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	_, err := q.db.Exec(`
		UPDATE jobs SET status = 'running', started_at = CURRENT_TIMESTAMP
		WHERE id = ? AND status = 'pending'
	`, id)

	return err
}

func (q *JobQueue) Complete(id int64, result string) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	_, err := q.db.Exec(`
		UPDATE jobs SET status = 'completed', result = ?, completed_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, result, id)

	return err
}

func (q *JobQueue) Fail(id int64, errMsg string) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	_, err := q.db.Exec(`
		UPDATE jobs SET status = 'failed', error = ?, completed_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, errMsg, id)

	return err
}

func (q *JobQueue) Cancel(id int64) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	_, err := q.db.Exec(`
		UPDATE jobs SET status = 'cancelled', completed_at = CURRENT_TIMESTAMP
		WHERE id = ? OR parent_id = ?
	`, id, id)

	return err
}

func (q *JobQueue) List(status string, limit int) ([]Job, error) {
	q.mu.RLock()
	defer q.mu.RUnlock()

	if limit <= 0 {
		limit = 20
	}

	query := `SELECT id, name, params, status, parent_id, depth, result, error,
	          created_at, started_at, completed_at, timeout_ms, max_children
	          FROM jobs`

	args := []interface{}{}
	if status != "" {
		query += " WHERE status = ?"
		args = append(args, status)
	}

	query += " ORDER BY created_at DESC LIMIT ?"
	args = append(args, limit)

	rows, err := q.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []Job
	for rows.Next() {
		var job Job
		var paramsJSON string
		var startedAt, completedAt sql.NullTime

		err := rows.Scan(
			&job.ID, &job.Name, &paramsJSON, &job.Status, &job.ParentID, &job.Depth,
			&job.Result, &job.Error, &job.CreatedAt, &startedAt, &completedAt,
			&job.TimeoutMs, &job.MaxChildren,
		)
		if err != nil {
			continue
		}

		json.Unmarshal([]byte(paramsJSON), &job.Params)
		if startedAt.Valid {
			job.StartedAt = &startedAt.Time
		}
		if completedAt.Valid {
			job.CompletedAt = &completedAt.Time
		}

		jobs = append(jobs, job)
	}

	return jobs, nil
}

func (q *JobQueue) Stats() (map[string]int, error) {
	q.mu.RLock()
	defer q.mu.RUnlock()

	stats := make(map[string]int)

	rows, err := q.db.Query("SELECT status, COUNT(*) FROM jobs GROUP BY status")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var status string
		var count int
		rows.Scan(&status, &count)
		stats[status] = count
	}

	q.db.QueryRow("SELECT COUNT(*) FROM jobs").Scan(&stats["total"])

	return stats, nil
}

func (q *JobQueue) CleanOld(maxAge time.Duration) (int64, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	result, err := q.db.Exec(`
		DELETE FROM jobs 
		WHERE status IN ('completed', 'failed', 'cancelled') 
		AND completed_at < ?
	`, time.Now().Add(-maxAge))

	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

func (q *JobQueue) Close() error {
	return q.db.Close()
}

func (w *Worker) RegisterHandler(name string, handler JobHandler) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.handlers == nil {
		w.handlers = make(map[string]JobHandler)
	}
	w.handlers[name] = handler
}

func (w *Worker) Start(ctx context.Context) {
	w.mu.Lock()
	if w.running {
		w.mu.Unlock()
		return
	}
	w.running = true
	w.mu.Unlock()

	for w.running {
		select {
		case <-ctx.Done():
			w.running = false
			return
		default:
		}

		jobs, err := w.jobQueue.GetPending(w.concurrency)
		if err != nil || len(jobs) == 0 {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		for _, job := range jobs {
			go w.runJob(ctx, job)
		}
	}
}

func (w *Worker) runJob(ctx context.Context, job Job) {
	w.mu.RLock()
	handler, ok := w.handlers[job.Name]
	w.mu.RUnlock()

	if !ok {
		w.jobQueue.Fail(job.ID, fmt.Sprintf("no handler for job: %s", job.Name))
		return
	}

	w.jobQueue.Start(job.ID)

	timeout := time.Duration(job.TimeoutMs) * time.Millisecond
	if timeout <= 0 {
		timeout = 60 * time.Second
	}

	jobCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	result, err := handler(jobCtx, &job)
	if err != nil {
		w.jobQueue.Fail(job.ID, err.Error())
	} else {
		w.jobQueue.Complete(job.ID, result)
	}
}

func (w *Worker) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.running = false
}

func (w *Worker) Smoke() error {
	jobs, err := w.jobQueue.GetPending(1)
	if err != nil {
		return err
	}

	if len(jobs) == 0 {
		return nil
	}

	return nil
}