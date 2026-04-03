package memory

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ahmad-ubaidillah/aigo/pkg/types"
)

const profileSchema = `
CREATE TABLE IF NOT EXISTS profiles (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL UNIQUE,
	default_model TEXT NOT NULL DEFAULT '',
	coding_model TEXT NOT NULL DEFAULT '',
	api_key TEXT NOT NULL DEFAULT '',
	opencode_binary TEXT NOT NULL DEFAULT '',
	opencode_timeout INTEGER NOT NULL DEFAULT 30,
	opencode_max_turns INTEGER NOT NULL DEFAULT 10,
	platform_prefs TEXT NOT NULL DEFAULT '{}',
	preferences TEXT NOT NULL DEFAULT '{}',
	created_at DATETIME NOT NULL,
	updated_at DATETIME NOT NULL
);
`

func (s *SessionDB) initProfileTable() error {
	_, err := s.db.Exec(profileSchema)
	return err
}

func (s *SessionDB) CreateProfile(id, name string) (*types.Profile, error) {
	now := time.Now()
	prefs, _ := json.Marshal(map[string]string{})
	profile := &types.Profile{
		ID:            id,
		Name:          name,
		PlatformPrefs: make(map[string]string),
		Preferences:   make(map[string]string),
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	_, err := s.db.Exec(
		`INSERT INTO profiles (id, name, default_model, coding_model, api_key,
			opencode_binary, opencode_timeout, opencode_max_turns,
			platform_prefs, preferences, created_at, updated_at)
		 VALUES (?, ?, '', '', '', '', 30, 10, ?, ?, ?, ?)`,
		profile.ID, profile.Name, string(prefs), string(prefs), profile.CreatedAt, profile.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create profile: %w", err)
	}
	return profile, nil
}

func (s *SessionDB) GetProfile(id string) (*types.Profile, error) {
	profile := &types.Profile{}
	var prefsJSON, prefsStr sql.NullString
	var updatedAt time.Time

	err := s.db.QueryRow(
		`SELECT id, name, default_model, coding_model, api_key,
			opencode_binary, opencode_timeout, opencode_max_turns,
			platform_prefs, preferences, created_at, updated_at
		 FROM profiles WHERE id = ?`,
		id,
	).Scan(
		&profile.ID, &profile.Name, &profile.DefaultModel, &profile.CodingModel,
		&profile.APIKey, &profile.OpenCodeBinary, &profile.OpenCodeTimeout,
		&profile.OpenCodeMaxTurns, &prefsJSON, &prefsStr, &profile.CreatedAt, &updatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get profile: %w", err)
	}

	if prefsJSON.Valid {
		_ = json.Unmarshal([]byte(prefsJSON.String), &profile.PlatformPrefs)
	}
	if prefsStr.Valid {
		_ = json.Unmarshal([]byte(prefsStr.String), &profile.Preferences)
	}
	profile.UpdatedAt = updatedAt
	return profile, nil
}

func (s *SessionDB) GetProfileByName(name string) (*types.Profile, error) {
	profile := &types.Profile{}
	var prefsJSON, prefsStr sql.NullString
	var updatedAt time.Time

	err := s.db.QueryRow(
		`SELECT id, name, default_model, coding_model, api_key,
			opencode_binary, opencode_timeout, opencode_max_turns,
			platform_prefs, preferences, created_at, updated_at
		 FROM profiles WHERE name = ?`,
		name,
	).Scan(
		&profile.ID, &profile.Name, &profile.DefaultModel, &profile.CodingModel,
		&profile.APIKey, &profile.OpenCodeBinary, &profile.OpenCodeTimeout,
		&profile.OpenCodeMaxTurns, &prefsJSON, &prefsStr, &profile.CreatedAt, &updatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get profile by name: %w", err)
	}

	if prefsJSON.Valid {
		_ = json.Unmarshal([]byte(prefsJSON.String), &profile.PlatformPrefs)
	}
	if prefsStr.Valid {
		_ = json.Unmarshal([]byte(prefsStr.String), &profile.Preferences)
	}
	profile.UpdatedAt = updatedAt
	return profile, nil
}

func (s *SessionDB) ListProfiles() ([]types.Profile, error) {
	rows, err := s.db.Query(
		`SELECT id, name, default_model, coding_model, api_key,
			opencode_binary, opencode_timeout, opencode_max_turns,
			platform_prefs, preferences, created_at, updated_at
		 FROM profiles ORDER BY updated_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("list profiles: %w", err)
	}
	defer rows.Close()

	var profiles []types.Profile
	for rows.Next() {
		var p types.Profile
		var prefsJSON, prefsStr sql.NullString
		var updatedAt time.Time
		if err := rows.Scan(
			&p.ID, &p.Name, &p.DefaultModel, &p.CodingModel, &p.APIKey,
			&p.OpenCodeBinary, &p.OpenCodeTimeout, &p.OpenCodeMaxTurns,
			&prefsJSON, &prefsStr, &p.CreatedAt, &updatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan profile: %w", err)
		}
		if prefsJSON.Valid {
			_ = json.Unmarshal([]byte(prefsJSON.String), &p.PlatformPrefs)
		}
		if prefsStr.Valid {
			_ = json.Unmarshal([]byte(prefsStr.String), &p.Preferences)
		}
		p.UpdatedAt = updatedAt
		profiles = append(profiles, p)
	}
	return profiles, rows.Err()
}

func (s *SessionDB) UpdateProfile(p *types.Profile) error {
	prefsJSON, _ := json.Marshal(p.PlatformPrefs)
	prefsStr, _ := json.Marshal(p.Preferences)
	p.UpdatedAt = time.Now()

	result, err := s.db.Exec(
		`UPDATE profiles SET name = ?, default_model = ?, coding_model = ?,
			api_key = ?, opencode_binary = ?, opencode_timeout = ?,
			opencode_max_turns = ?, platform_prefs = ?, preferences = ?,
			updated_at = ? WHERE id = ?`,
		p.Name, p.DefaultModel, p.CodingModel, p.APIKey,
		p.OpenCodeBinary, p.OpenCodeTimeout, p.OpenCodeMaxTurns,
		string(prefsJSON), string(prefsStr), p.UpdatedAt, p.ID,
	)
	if err != nil {
		return fmt.Errorf("update profile: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check update result: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("update profile: profile %s not found", p.ID)
	}
	return nil
}

func (s *SessionDB) DeleteProfile(id string) error {
	result, err := s.db.Exec("DELETE FROM profiles WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete profile: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check delete result: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("delete profile: profile %s not found", id)
	}
	return nil
}

func (s *SessionDB) SetProfilePreference(profileID, key, value string) error {
	profile, err := s.GetProfile(profileID)
	if err != nil {
		return err
	}
	if profile.Preferences == nil {
		profile.Preferences = make(map[string]string)
	}
	profile.Preferences[key] = value
	return s.UpdateProfile(profile)
}

func (s *SessionDB) SetPlatformPreference(profileID, platform, token string) error {
	profile, err := s.GetProfile(profileID)
	if err != nil {
		return err
	}
	if profile.PlatformPrefs == nil {
		profile.PlatformPrefs = make(map[string]string)
	}
	profile.PlatformPrefs[platform] = token
	return s.UpdateProfile(profile)
}
