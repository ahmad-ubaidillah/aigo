package memory

import (
	"time"
)

type SessionHistory struct {
	SessionID  string    `json:"session_id"`
	LastTask   string    `json:"last_task"`
	Summary    string    `json:"summary"`
	Context    string    `json:"context"`
	TurnCount  int       `json:"turn_count"`
	LastActive time.Time `json:"last_active"`
}

func (s *SessionDB) SaveSessionHistory(sessionID, lastTask, summary, context string, turnCount int) error {
	now := time.Now()
	_, err := s.db.Exec(
		`INSERT OR REPLACE INTO session_history 
		(session_id, last_task, summary, context, turn_count, last_active) 
		VALUES (?, ?, ?, ?, ?, ?)`,
		sessionID, lastTask, summary, context, turnCount, now,
	)
	return err
}

func (s *SessionDB) GetSessionHistory(sessionID string) (*SessionHistory, error) {
	sh := &SessionHistory{}
	err := s.db.QueryRow(
		`SELECT session_id, last_task, summary, context, turn_count, last_active 
		FROM session_history WHERE session_id = ?`,
		sessionID,
	).Scan(&sh.SessionID, &sh.LastTask, &sh.Summary, &sh.Context, &sh.TurnCount, &sh.LastActive)
	if err != nil {
		return nil, err
	}
	return sh, nil
}

func (s *SessionDB) ListSessionHistory(limit int) ([]SessionHistory, error) {
	rows, err := s.db.Query(
		`SELECT session_id, last_task, summary, context, turn_count, last_active 
		FROM session_history ORDER BY last_active DESC LIMIT ?`,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []SessionHistory
	for rows.Next() {
		var sh SessionHistory
		if err := rows.Scan(&sh.SessionID, &sh.LastTask, &sh.Summary, &sh.Context, &sh.TurnCount, &sh.LastActive); err != nil {
			return nil, err
		}
		results = append(results, sh)
	}
	return results, nil
}

func (s *SessionDB) DeleteSessionHistory(sessionID string) error {
	_, err := s.db.Exec(`DELETE FROM session_history WHERE session_id = ?`, sessionID)
	return err
}
