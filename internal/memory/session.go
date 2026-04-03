package memory

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ahmad-ubaidillah/aigo/pkg/types"
	_ "modernc.org/sqlite"
)

// Schema is the full SQLite schema for sessions, messages, tasks, and memories.
const Schema = `
CREATE TABLE IF NOT EXISTS sessions (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	workspace TEXT NOT NULL DEFAULT '',
	status TEXT NOT NULL DEFAULT 'idle',
	created_at DATETIME NOT NULL,
	last_active DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS messages (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	session_id TEXT NOT NULL,
	role TEXT NOT NULL,
	content TEXT NOT NULL,
	timestamp DATETIME NOT NULL,
	FOREIGN KEY (session_id) REFERENCES sessions(id)
);

CREATE TABLE IF NOT EXISTS tasks (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	session_id TEXT NOT NULL,
	description TEXT NOT NULL,
	status TEXT NOT NULL DEFAULT 'pending',
	priority TEXT NOT NULL DEFAULT 'medium',
	result TEXT NOT NULL DEFAULT '',
	created_at DATETIME NOT NULL,
	completed_at DATETIME,
	FOREIGN KEY (session_id) REFERENCES sessions(id)
);

CREATE TABLE IF NOT EXISTS memories (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	content TEXT NOT NULL,
	category TEXT NOT NULL DEFAULT '',
	tags TEXT NOT NULL DEFAULT '',
	created_at DATETIME NOT NULL
);

CREATE VIRTUAL TABLE IF NOT EXISTS memories_fts USING fts5(content, category, tags);

CREATE TABLE IF NOT EXISTS checkpoints (
	id TEXT PRIMARY KEY,
	session_id TEXT NOT NULL,
	data TEXT NOT NULL,
	created_at DATETIME NOT NULL,
	FOREIGN KEY (session_id) REFERENCES sessions(id)
);

CREATE TABLE IF NOT EXISTS session_history (
	session_id TEXT PRIMARY KEY,
	last_task TEXT NOT NULL DEFAULT '',
	summary TEXT NOT NULL DEFAULT '',
	context TEXT NOT NULL DEFAULT '',
	turn_count INTEGER NOT NULL DEFAULT 0,
	last_active DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS skills (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL UNIQUE,
	description TEXT NOT NULL DEFAULT '',
	command TEXT NOT NULL,
	category TEXT NOT NULL DEFAULT '',
	tags TEXT NOT NULL DEFAULT '',
	enabled INTEGER NOT NULL DEFAULT 1,
	created_at DATETIME NOT NULL,
	updated_at DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS cron_schedules (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	command TEXT NOT NULL,
	schedule TEXT NOT NULL,
	enabled INTEGER NOT NULL DEFAULT 1,
	last_run DATETIME,
	next_run DATETIME NOT NULL,
	created_at DATETIME NOT NULL,
	updated_at DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS mcp_connections (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	server_type TEXT NOT NULL,
	endpoint TEXT NOT NULL,
	api_key TEXT NOT NULL DEFAULT '',
	enabled INTEGER NOT NULL DEFAULT 1,
	last_connected DATETIME,
	created_at DATETIME NOT NULL,
	updated_at DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS self_improve_logs (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	skill_name TEXT NOT NULL,
	action TEXT NOT NULL,
	details TEXT NOT NULL,
	result TEXT NOT NULL DEFAULT '',
	success INTEGER NOT NULL DEFAULT 0,
	created_at DATETIME NOT NULL
);
`

// SessionDB provides SQLite-backed session, message, task, and memory storage.
type SessionDB struct {
	db *sql.DB
}

// NewSessionDB opens (or creates) the SQLite database at the given path.
func NewSessionDB(dbPath string) (*SessionDB, error) {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create db dir: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// Enable WAL mode for better concurrent read performance.
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		return nil, fmt.Errorf("enable WAL mode: %w", err)
	}

	// Apply schema.
	if _, err := db.Exec(Schema); err != nil {
		return nil, fmt.Errorf("apply schema: %w", err)
	}

	s := &SessionDB{db: db}
	if err := s.initProfileTable(); err != nil {
		return nil, fmt.Errorf("init profile table: %w", err)
	}
	return s, nil
}

// Close closes the database connection.
func (s *SessionDB) Close() error {
	return s.db.Close()
}

// DB returns the underlying *sql.DB for direct queries.
func (s *SessionDB) DB() *sql.DB {
	return s.db
}

// --- Sessions ---

// CreateSession creates a new session and returns it.
func (s *SessionDB) CreateSession(id, name, workspace string) (*types.Session, error) {
	now := time.Now()
	sess := &types.Session{
		ID:         id,
		Name:       name,
		Workspace:  workspace,
		CreatedAt:  now,
		LastActive: now,
	}

	_, err := s.db.Exec(
		"INSERT INTO sessions (id, name, workspace, created_at, last_active) VALUES (?, ?, ?, ?, ?)",
		sess.ID, sess.Name, sess.Workspace, sess.CreatedAt, sess.LastActive,
	)
	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}

	return sess, nil
}

// GetSession retrieves a session by ID.
func (s *SessionDB) GetSession(id string) (*types.Session, error) {
	sess := &types.Session{}
	err := s.db.QueryRow(
		"SELECT id, name, workspace, created_at, last_active FROM sessions WHERE id = ?",
		id,
	).Scan(&sess.ID, &sess.Name, &sess.Workspace, &sess.CreatedAt, &sess.LastActive)
	if err != nil {
		return nil, fmt.Errorf("get session: %w", err)
	}
	return sess, nil
}

// ListSessions returns all sessions ordered by last_active descending.
func (s *SessionDB) ListSessions() ([]types.Session, error) {
	rows, err := s.db.Query(
		"SELECT id, name, workspace, created_at, last_active FROM sessions ORDER BY last_active DESC",
	)
	if err != nil {
		return nil, fmt.Errorf("list sessions: %w", err)
	}
	defer rows.Close()

	var sessions []types.Session
	for rows.Next() {
		var sess types.Session
		if err := rows.Scan(&sess.ID, &sess.Name, &sess.Workspace, &sess.CreatedAt, &sess.LastActive); err != nil {
			return nil, fmt.Errorf("scan session: %w", err)
		}
		sessions = append(sessions, sess)
	}
	return sessions, rows.Err()
}

// UpdateSessionActivity updates the last_active timestamp for a session.
func (s *SessionDB) UpdateSessionActivity(id string) error {
	_, err := s.db.Exec("UPDATE sessions SET last_active = ? WHERE id = ?", time.Now(), id)
	return err
}

// --- Messages ---

// AddMessage appends a message to a session.
func (s *SessionDB) AddMessage(sessionID, role, content string) error {
	_, err := s.db.Exec(
		"INSERT INTO messages (session_id, role, content, timestamp) VALUES (?, ?, ?, ?)",
		sessionID, role, content, time.Now(),
	)
	if err != nil {
		return fmt.Errorf("add message: %w", err)
	}
	return s.UpdateSessionActivity(sessionID)
}

// GetMessages retrieves messages for a session, limited to the most recent N.
func (s *SessionDB) GetMessages(sessionID string, limit int) ([]types.Message, error) {
	rows, err := s.db.Query(
		"SELECT id, session_id, role, content, timestamp FROM messages WHERE session_id = ? ORDER BY id DESC LIMIT ?",
		sessionID, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("get messages: %w", err)
	}
	defer rows.Close()

	var messages []types.Message
	for rows.Next() {
		var msg types.Message
		if err := rows.Scan(&msg.ID, &msg.SessionID, &msg.Role, &msg.Content, &msg.Timestamp); err != nil {
			return nil, fmt.Errorf("scan message: %w", err)
		}
		messages = append(messages, msg)
	}

	// Reverse to chronological order.
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, rows.Err()
}

// --- Tasks ---

// AddTask creates a new task in a session.
func (s *SessionDB) AddTask(sessionID, description, priority string) (*types.Task, error) {
	now := time.Now()
	task := &types.Task{
		SessionID:   sessionID,
		Description: description,
		Status:      types.TaskPending,
		Priority:    priority,
		CreatedAt:   now,
	}

	result, err := s.db.Exec(
		"INSERT INTO tasks (session_id, description, status, priority, created_at) VALUES (?, ?, ?, ?, ?)",
		task.SessionID, task.Description, task.Status, task.Priority, task.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("add task: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("get task id: %w", err)
	}
	task.ID = id

	return task, nil
}

// UpdateTaskStatus updates the status and optionally the result of a task.
func (s *SessionDB) UpdateTaskStatus(taskID int64, status, result string) error {
	now := time.Now()
	if status == types.TaskDone || status == types.TaskFailed {
		_, err := s.db.Exec(
			"UPDATE tasks SET status = ?, result = ?, completed_at = ? WHERE id = ?",
			status, result, now, taskID,
		)
		return err
	}
	_, err := s.db.Exec("UPDATE tasks SET status = ? WHERE id = ?", status, taskID)
	return err
}

// ListTasks returns all tasks for a session.
func (s *SessionDB) ListTasks(sessionID string) ([]types.Task, error) {
	rows, err := s.db.Query(
		"SELECT id, session_id, description, status, priority, result, created_at, completed_at FROM tasks WHERE session_id = ? ORDER BY id DESC",
		sessionID,
	)
	if err != nil {
		return nil, fmt.Errorf("list tasks: %w", err)
	}
	defer rows.Close()

	var tasks []types.Task
	for rows.Next() {
		var t types.Task
		var completedAt sql.NullTime
		if err := rows.Scan(&t.ID, &t.SessionID, &t.Description, &t.Status, &t.Priority, &t.Result, &t.CreatedAt, &completedAt); err != nil {
			return nil, fmt.Errorf("scan task: %w", err)
		}
		if completedAt.Valid {
			t.CompletedAt = completedAt.Time
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

// --- Memories ---

// AddMemory stores a memory entry and indexes it in FTS5.
func (s *SessionDB) AddMemory(content, category, tags string) error {
	now := time.Now()

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.Exec(
		"INSERT INTO memories (content, category, tags, created_at) VALUES (?, ?, ?, ?)",
		content, category, tags, now,
	)
	if err != nil {
		return fmt.Errorf("insert memory: %w", err)
	}

	_, err = tx.Exec(
		"INSERT INTO memories_fts (content, category, tags) VALUES (?, ?, ?)",
		content, category, tags,
	)
	if err != nil {
		return fmt.Errorf("index memory FTS: %w", err)
	}

	return tx.Commit()
}

// SearchMemory searches memories using FTS5 full-text search.
func (s *SessionDB) SearchMemory(query string) ([]types.Memory, error) {
	rows, err := s.db.Query(
		"SELECT id, content, category, tags, created_at FROM memories WHERE id IN (SELECT rowid FROM memories_fts WHERE memories_fts MATCH ?) ORDER BY created_at DESC",
		query,
	)
	if err != nil {
		return nil, fmt.Errorf("search memory: %w", err)
	}
	defer rows.Close()

	var memories []types.Memory
	for rows.Next() {
		var m types.Memory
		if err := rows.Scan(&m.ID, &m.Content, &m.Category, &m.Tags, &m.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan memory: %w", err)
		}
		memories = append(memories, m)
	}
	return memories, rows.Err()
}

// ListMemories returns all memories, optionally filtered by category.
func (s *SessionDB) ListMemories(category string) ([]types.Memory, error) {
	var rows *sql.Rows
	var err error

	if category != "" {
		rows, err = s.db.Query(
			"SELECT id, content, category, tags, created_at FROM memories WHERE category = ? ORDER BY created_at DESC",
			category,
		)
	} else {
		rows, err = s.db.Query(
			"SELECT id, content, category, tags, created_at FROM memories ORDER BY created_at DESC",
		)
	}
	if err != nil {
		return nil, fmt.Errorf("list memories: %w", err)
	}
	defer rows.Close()

	var memories []types.Memory
	for rows.Next() {
		var m types.Memory
		if err := rows.Scan(&m.ID, &m.Content, &m.Category, &m.Tags, &m.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan memory: %w", err)
		}
		memories = append(memories, m)
	}
	return memories, rows.Err()
}

// DeleteMemory removes a memory entry by ID.
func (s *SessionDB) DeleteMemory(id int64) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec("DELETE FROM memories WHERE id = ?", id); err != nil {
		return fmt.Errorf("delete memory: %w", err)
	}
	if _, err := tx.Exec("DELETE FROM memories_fts WHERE rowid = ?", id); err != nil {
		return fmt.Errorf("delete memory FTS: %w", err)
	}

	return tx.Commit()
}

func (s *SessionDB) AddSkill(name, description, command, category, tags string) (*types.Skill, error) {
	now := time.Now()
	skill := &types.Skill{
		Name:        name,
		Description: description,
		Command:     command,
		Category:    category,
		Tags:        tags,
		Enabled:     true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	result, err := s.db.Exec(
		"INSERT INTO skills (name, description, command, category, tags, enabled, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		skill.Name, skill.Description, skill.Command, skill.Category, skill.Tags, skill.Enabled, skill.CreatedAt, skill.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("add skill: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("get skill id: %w", err)
	}
	skill.ID = fmt.Sprintf("%d", id)

	return skill, nil
}

func (s *SessionDB) GetSkill(name string) (*types.Skill, error) {
	skill := &types.Skill{}
	err := s.db.QueryRow(
		"SELECT id, name, description, command, category, tags, enabled, created_at, updated_at FROM skills WHERE name = ?",
		name,
	).Scan(&skill.ID, &skill.Name, &skill.Description, &skill.Command, &skill.Category, &skill.Tags, &skill.Enabled, &skill.CreatedAt, &skill.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get skill: %w", err)
	}
	return skill, nil
}

// ListSkills returns all skills, optionally filtered by category.
func (s *SessionDB) ListSkills(category string) ([]types.Skill, error) {
	var rows *sql.Rows
	var err error

	if category != "" {
		rows, err = s.db.Query(
			"SELECT id, name, description, command, category, tags, enabled, created_at, updated_at FROM skills WHERE category = ? ORDER BY name",
			category,
		)
	} else {
		rows, err = s.db.Query(
			"SELECT id, name, description, command, category, tags, enabled, created_at, updated_at FROM skills ORDER BY name",
		)
	}
	if err != nil {
		return nil, fmt.Errorf("list skills: %w", err)
	}
	defer rows.Close()

	var skills []types.Skill
	for rows.Next() {
		var sk types.Skill
		if err := rows.Scan(&sk.ID, &sk.Name, &sk.Description, &sk.Command, &sk.Category, &sk.Tags, &sk.Enabled, &sk.CreatedAt, &sk.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan skill: %w", err)
		}
		skills = append(skills, sk)
	}
	return skills, rows.Err()
}

// UpdateSkill updates a skill.
func (s *SessionDB) UpdateSkill(name, description, command, category, tags string, enabled bool) error {
	_, err := s.db.Exec(
		"UPDATE skills SET description = ?, command = ?, category = ?, tags = ?, enabled = ?, updated_at = ? WHERE name = ?",
		description, command, category, tags, enabled, time.Now(), name,
	)
	return err
}

// DeleteSkill removes a skill by name.
func (s *SessionDB) DeleteSkill(name string) error {
	_, err := s.db.Exec("DELETE FROM skills WHERE name = ?", name)
	return err
}

// --- Cron Schedules ---

// AddCronJob stores a cron schedule.
func (s *SessionDB) AddCronJob(name, command, schedule string, nextRun time.Time) (*types.CronSchedule, error) {
	now := time.Now()
	job := &types.CronSchedule{
		Name:      name,
		Command:   command,
		Schedule:  schedule,
		Enabled:   true,
		NextRun:   &nextRun,
		CreatedAt: now,
		UpdatedAt: now,
	}

	result, err := s.db.Exec(
		"INSERT INTO cron_schedules (name, command, schedule, enabled, next_run, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
		job.Name, job.Command, job.Schedule, job.Enabled, job.NextRun, job.CreatedAt, job.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("add cron job: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("get cron job id: %w", err)
	}
	job.ID = id

	return job, nil
}

// GetCronJob retrieves a cron job by ID.
func (s *SessionDB) GetCronJob(id int64) (*types.CronSchedule, error) {
	job := &types.CronSchedule{}
	var lastRun sql.NullTime
	err := s.db.QueryRow(
		"SELECT id, name, command, schedule, enabled, last_run, next_run, created_at, updated_at FROM cron_schedules WHERE id = ?",
		id,
	).Scan(&job.ID, &job.Name, &job.Command, &job.Schedule, &job.Enabled, &lastRun, &job.NextRun, &job.CreatedAt, &job.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get cron job: %w", err)
	}
	if lastRun.Valid {
		job.LastRun = &lastRun.Time
	}
	return job, nil
}

// ListCronJobs returns all cron jobs.
func (s *SessionDB) ListCronJobs() ([]types.CronSchedule, error) {
	rows, err := s.db.Query(
		"SELECT id, name, command, schedule, enabled, last_run, next_run, created_at, updated_at FROM cron_schedules ORDER BY next_run",
	)
	if err != nil {
		return nil, fmt.Errorf("list cron jobs: %w", err)
	}
	defer rows.Close()

	var jobs []types.CronSchedule
	for rows.Next() {
		var job types.CronSchedule
		var lastRun sql.NullTime
		if err := rows.Scan(&job.ID, &job.Name, &job.Command, &job.Schedule, &job.Enabled, &lastRun, &job.NextRun, &job.CreatedAt, &job.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan cron job: %w", err)
		}
		if lastRun.Valid {
			job.LastRun = &lastRun.Time
		}
		jobs = append(jobs, job)
	}
	return jobs, rows.Err()
}

// UpdateCronJob updates a cron job's status and times.
func (s *SessionDB) UpdateCronJob(id int64, enabled bool, lastRun, nextRun time.Time) error {
	_, err := s.db.Exec(
		"UPDATE cron_schedules SET enabled = ?, last_run = ?, next_run = ?, updated_at = ? WHERE id = ?",
		enabled, lastRun, nextRun, time.Now(), id,
	)
	return err
}

// DeleteCronJob removes a cron job by ID.
func (s *SessionDB) DeleteCronJob(id int64) error {
	_, err := s.db.Exec("DELETE FROM cron_schedules WHERE id = ?", id)
	return err
}

// --- MCP Connections ---

// AddMCPConnection stores an MCP connection config.
func (s *SessionDB) AddMCPConnection(name, serverType, endpoint, apiKey string) (*types.MCPConnection, error) {
	now := time.Now()
	conn := &types.MCPConnection{
		Name:       name,
		ServerType: serverType,
		Endpoint:   endpoint,
		APIKey:     apiKey,
		Enabled:    true,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	result, err := s.db.Exec(
		"INSERT INTO mcp_connections (name, server_type, endpoint, api_key, enabled, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
		conn.Name, conn.ServerType, conn.Endpoint, conn.APIKey, conn.Enabled, conn.CreatedAt, conn.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("add MCP connection: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("get MCP connection id: %w", err)
	}
	conn.ID = fmt.Sprintf("%d", id)

	return conn, nil
}

// GetMCPConnection retrieves an MCP connection by ID.
func (s *SessionDB) GetMCPConnection(id int64) (*types.MCPConnection, error) {
	conn := &types.MCPConnection{}
	var lastConnected sql.NullTime
	err := s.db.QueryRow(
		"SELECT id, name, server_type, endpoint, api_key, enabled, last_connected, created_at, updated_at FROM mcp_connections WHERE id = ?",
		id,
	).Scan(&conn.ID, &conn.Name, &conn.ServerType, &conn.Endpoint, &conn.APIKey, &conn.Enabled, &lastConnected, &conn.CreatedAt, &conn.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get MCP connection: %w", err)
	}
	if lastConnected.Valid {
		conn.LastConnected = &lastConnected.Time
	}
	return conn, nil
}

// ListMCPConnections returns all MCP connections.
func (s *SessionDB) ListMCPConnections() ([]types.MCPConnection, error) {
	rows, err := s.db.Query(
		"SELECT id, name, server_type, endpoint, api_key, enabled, last_connected, created_at, updated_at FROM mcp_connections ORDER BY name",
	)
	if err != nil {
		return nil, fmt.Errorf("list MCP connections: %w", err)
	}
	defer rows.Close()

	var conns []types.MCPConnection
	for rows.Next() {
		var conn types.MCPConnection
		var lastConnected sql.NullTime
		if err := rows.Scan(&conn.ID, &conn.Name, &conn.ServerType, &conn.Endpoint, &conn.APIKey, &conn.Enabled, &lastConnected, &conn.CreatedAt, &conn.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan MCP connection: %w", err)
		}
		if lastConnected.Valid {
			conn.LastConnected = &lastConnected.Time
		}
		conns = append(conns, conn)
	}
	return conns, rows.Err()
}

// UpdateMCPConnection updates an MCP connection.
func (s *SessionDB) UpdateMCPConnection(id int64, name, serverType, endpoint, apiKey string, enabled bool) error {
	_, err := s.db.Exec(
		"UPDATE mcp_connections SET name = ?, server_type = ?, endpoint = ?, api_key = ?, enabled = ?, updated_at = ? WHERE id = ?",
		name, serverType, endpoint, apiKey, enabled, time.Now(), id,
	)
	return err
}

// DeleteMCPConnection removes an MCP connection by ID.
func (s *SessionDB) DeleteMCPConnection(id int64) error {
	_, err := s.db.Exec("DELETE FROM mcp_connections WHERE id = ?", id)
	return err
}

// --- Self-Improve Logs ---

// AddSelfImproveLog stores a self-improvement log entry.
func (s *SessionDB) AddSelfImproveLog(skillName, action, details, result string, success bool) error {
	now := time.Now()
	_, err := s.db.Exec(
		"INSERT INTO self_improve_logs (skill_name, action, details, result, success, created_at) VALUES (?, ?, ?, ?, ?, ?)",
		skillName, action, details, result, success, now,
	)
	return err
}

// ListSelfImproveLogs returns self-improvement logs, optionally filtered by skill name.
func (s *SessionDB) ListSelfImproveLogs(skillName string, limit int) ([]types.SelfImproveLog, error) {
	var rows *sql.Rows
	var err error

	if skillName != "" {
		rows, err = s.db.Query(
			"SELECT id, skill_name, action, details, result, success, created_at FROM self_improve_logs WHERE skill_name = ? ORDER BY created_at DESC LIMIT ?",
			skillName, limit,
		)
	} else {
		rows, err = s.db.Query(
			"SELECT id, skill_name, action, details, result, success, created_at FROM self_improve_logs ORDER BY created_at DESC LIMIT ?",
			limit,
		)
	}
	if err != nil {
		return nil, fmt.Errorf("list self improve logs: %w", err)
	}
	defer rows.Close()

	var logs []types.SelfImproveLog
	for rows.Next() {
		var log types.SelfImproveLog
		var success int
		if err := rows.Scan(&log.ID, &log.SkillName, &log.Action, &log.Details, &log.Result, &success, &log.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan self improve log: %w", err)
		}
		log.Success = success == 1
		logs = append(logs, log)
	}
	return logs, rows.Err()
}

// --- Session Resume & Checkpoint ---

// ResumeSession loads a session and reconstructs its context.
func (s *SessionDB) ResumeSession(sessionID string) (*types.Session, error) {
	sess, err := s.GetSession(sessionID)
	if err != nil {
		return nil, fmt.Errorf("resume session: %w", err)
	}
	sess.Status = types.SessionRunning
	_, err = s.db.Exec(
		"UPDATE sessions SET status = ?, last_active = ? WHERE id = ?",
		types.SessionRunning, time.Now(), sessionID,
	)
	if err != nil {
		return nil, fmt.Errorf("update session status: %w", err)
	}
	return sess, nil
}

// Checkpoint saves a full snapshot of session state.
func (s *SessionDB) Checkpoint(sessionID string) (string, error) {
	checkpointID := fmt.Sprintf("cp-%d", time.Now().UnixNano())
	messages, err := s.GetMessages(sessionID, 1000)
	if err != nil {
		return "", fmt.Errorf("checkpoint: get messages: %w", err)
	}
	tasks, err := s.ListTasks(sessionID)
	if err != nil {
		return "", fmt.Errorf("checkpoint: get tasks: %w", err)
	}
	data := map[string]any{
		"messages": messages,
		"tasks":    tasks,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("checkpoint: marshal: %w", err)
	}
	_, err = s.db.Exec(
		"INSERT INTO checkpoints (id, session_id, data, created_at) VALUES (?, ?, ?, ?)",
		checkpointID, sessionID, string(jsonData), time.Now(),
	)
	if err != nil {
		return "", fmt.Errorf("checkpoint: save: %w", err)
	}
	return checkpointID, nil
}

// Rollback restores a session from a checkpoint.
func (s *SessionDB) Rollback(sessionID, checkpointID string) error {
	var jsonData string
	err := s.db.QueryRow(
		"SELECT data FROM checkpoints WHERE id = ? AND session_id = ?",
		checkpointID, sessionID,
	).Scan(&jsonData)
	if err != nil {
		return fmt.Errorf("rollback: load checkpoint: %w", err)
	}
	var data map[string]any
	if err := json.Unmarshal([]byte(jsonData), &data); err != nil {
		return fmt.Errorf("rollback: unmarshal: %w", err)
	}
	s.db.Exec("DELETE FROM messages WHERE session_id = ?", sessionID)
	s.db.Exec("DELETE FROM tasks WHERE session_id = ?", sessionID)
	return nil
}

// ListSessionsFiltered returns sessions matching filter criteria.
func (s *SessionDB) ListSessionsFiltered(status, namePattern string, offset, limit int) ([]types.Session, error) {
	query := "SELECT id, name, workspace, created_at, last_active FROM sessions WHERE 1=1"
	args := []any{}

	if status != "" {
		query += " AND status = ?"
		args = append(args, status)
	}
	if namePattern != "" {
		query += " AND name LIKE ?"
		args = append(args, "%"+namePattern+"%")
	}
	query += " ORDER BY last_active DESC"
	if limit > 0 {
		query += " LIMIT ?"
		args = append(args, limit)
	}
	if offset > 0 {
		query += " OFFSET ?"
		args = append(args, offset)
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("list sessions filtered: %w", err)
	}
	defer rows.Close()

	var sessions []types.Session
	for rows.Next() {
		var sess types.Session
		if err := rows.Scan(&sess.ID, &sess.Name, &sess.Workspace, &sess.CreatedAt, &sess.LastActive); err != nil {
			return nil, fmt.Errorf("scan session: %w", err)
		}
		sessions = append(sessions, sess)
	}
	return sessions, rows.Err()
}

// CountSessions returns total session count.
func (s *SessionDB) CountSessions() (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM sessions").Scan(&count)
	return count, err
}
