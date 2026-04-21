package recipes

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

type Recipe struct {
	Name         string            `json:"name"`
	Description  string            `json:"description"`
	Requires     []string          `json:"requires"`
	Credentials  map[string]string `json:"credentials"`
	CronSchedule string            `json:"cron_schedule"`
	Enabled      bool              `json:"enabled"`
	LastRun      time.Time         `json:"last_run"`
	Status       string            `json:"status"`
}

type RecipeStore struct {
	db           *sql.DB
	dbPath       string
	mu           sync.RWMutex
	basePath     string
	recipeConfig map[string]Recipe
}

func New(basePath string) (*RecipeStore, error) {
	if basePath == "" {
		home, _ := os.UserHomeDir()
		basePath = filepath.Join(home, ".aigo", "recipes")
	}

	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("recipes mkdir: %w", err)
	}

	dbPath := filepath.Join(basePath, "recipes.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("recipes open: %w", err)
	}

	r := &RecipeStore{
		db:           db,
		dbPath:       dbPath,
		basePath:     basePath,
		recipeConfig: make(map[string]Recipe),
	}

	if err := r.initSchema(); err != nil {
		db.Close()
		return nil, err
	}

	r.loadDefaultRecipes()

	return r, nil
}

func (r *RecipeStore) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS recipes (
		name TEXT PRIMARY KEY,
		description TEXT,
		requires TEXT,
		credentials TEXT,
		cron_schedule TEXT,
		enabled INTEGER DEFAULT 0,
		last_run DATETIME,
		status TEXT DEFAULT 'idle'
	);

	CREATE TABLE IF NOT EXISTS ingested_data (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		recipe_name TEXT NOT NULL,
		source_id TEXT NOT NULL,
		content TEXT,
		entities TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(recipe_name, source_id)
	);
	`

	_, err := r.db.Exec(schema)
	return err
}

func (r *RecipeStore) loadDefaultRecipes() {
	defaults := []Recipe{
		{
			Name:         "ngrok-tunnel",
			Description:  "Fixed URL for MCP + voice (ngrok)",
			Requires:     []string{},
			CronSchedule: "",
			Enabled:      false,
		},
		{
			Name:        "credential-gateway",
			Description: "Gmail + Calendar access",
			Requires:    []string{},
			Credentials: map[string]string{
				"gmail": "oauth",
			},
			Enabled: false,
		},
		{
			Name:         "email-to-brain",
			Description:  "Gmail to entity pages",
			Requires:     []string{"credential-gateway"},
			Enabled:      false,
			CronSchedule: "*/15 * * * *",
		},
		{
			Name:         "calendar-to-brain",
			Description:  "Google Calendar to daily pages",
			Requires:     []string{"credential-gateway"},
			Enabled:      false,
			CronSchedule: "*/30 * * * *",
		},
		{
			Name:         "x-to-brain",
			Description:  "Twitter timeline + mentions",
			Requires:     []string{},
			Enabled:      false,
			CronSchedule: "*/60 * * * *",
		},
		{
			Name:         "meeting-sync",
			Description:  "Circleback transcripts to brain pages",
			Requires:     []string{},
			Enabled:      false,
			CronSchedule: "*/60 * * * *",
		},
		{
			Name:        "voice-to-brain",
			Description: "Phone calls to brain pages (Twilio + OpenAI)",
			Requires:    []string{"ngrok-tunnel"},
			Credentials: map[string]string{
				"twilio": "api_key",
				"openai": "api_key",
			},
			Enabled:      false,
			CronSchedule: "",
		},
		{
			Name:         "data-research",
			Description:  "Extract structured data from email",
			Requires:     []string{"credential-gateway"},
			Enabled:      false,
			CronSchedule: "0 8 * * *",
		},
	}

	for _, recipe := range defaults {
		r.recipeConfig[recipe.Name] = recipe

		requiresJSON, _ := json.Marshal(recipe.Requires)
		credJSON, _ := json.Marshal(recipe.Credentials)

		r.db.Exec(`
			INSERT OR IGNORE INTO recipes (name, description, requires, credentials, cron_schedule, enabled, status)
			VALUES (?, ?, ?, ?, ?, ?, 'idle')
		`, recipe.Name, recipe.Description, string(requiresJSON), string(credJSON), recipe.CronSchedule, 0)
	}
}

func (r *RecipeStore) List() ([]Recipe, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	rows, err := r.db.Query(`
		SELECT name, description, requires, credentials, cron_schedule, enabled, last_run, status
		FROM recipes
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var recipes []Recipe
	for rows.Next() {
		var recipe Recipe
		var requiresJSON, credJSON string
		var lastRun sql.NullTime
		var enabled int

		rows.Scan(&recipe.Name, &recipe.Description, &requiresJSON, &credJSON,
			&recipe.CronSchedule, &enabled, &lastRun, &recipe.Status)

		json.Unmarshal([]byte(requiresJSON), &recipe.Requires)
		json.Unmarshal([]byte(credJSON), &recipe.Credentials)
		recipe.Enabled = enabled == 1
		if lastRun.Valid {
			recipe.LastRun = lastRun.Time
		}

		recipes = append(recipes, recipe)
	}

	return recipes, nil
}

func (r *RecipeStore) Get(name string) (*Recipe, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var recipe Recipe
	var requiresJSON, credJSON string
	var lastRun sql.NullTime
	var enabled int

	err := r.db.QueryRow(`
		SELECT name, description, requires, credentials, cron_schedule, enabled, last_run, status
		FROM recipes WHERE name = ?
	`, name).Scan(&recipe.Name, &recipe.Description, &requiresJSON, &credJSON,
		&recipe.CronSchedule, &enabled, &lastRun, &recipe.Status)

	if err != nil {
		return nil, err
	}

	json.Unmarshal([]byte(requiresJSON), &recipe.Requires)
	json.Unmarshal([]byte(credJSON), &recipe.Credentials)
	recipe.Enabled = enabled == 1
	if lastRun.Valid {
		recipe.LastRun = lastRun.Time
	}

	return &recipe, nil
}

func (r *RecipeStore) Enable(name string, enabled bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, err := r.db.Exec(`
		UPDATE recipes SET enabled = ? WHERE name = ?
	`, enabled, name)

	return err
}

func (r *RecipeStore) SetCredential(name, key, value string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	var credJSON string
	r.db.QueryRow("SELECT credentials FROM recipes WHERE name = ?", name).Scan(&credJSON)

	credentials := make(map[string]string)
	json.Unmarshal([]byte(credJSON), &credentials)
	credentials[key] = value

	credJSONBytes, _ := json.Marshal(credentials)
	credJSON = string(credJSONBytes)

	_, err := r.db.Exec(`
		UPDATE recipes SET credentials = ? WHERE name = ?
	`, credJSON, name)

	return err
}

func (r *RecipeStore) Ingest(recipeName, sourceID, content, entities string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, err := r.db.Exec(`
		INSERT OR REPLACE INTO ingested_data (recipe_name, source_id, content, entities, created_at)
		VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)
	`, recipeName, sourceID, content, entities)

	if err == nil {
		r.db.Exec(`
			UPDATE recipes SET last_run = CURRENT_TIMESTAMP, status = 'idle' WHERE name = ?
		`, recipeName)
	}

	return err
}

func (r *RecipeStore) GetIngestedData(recipeName string, limit int) ([]struct {
	ID        int64
	SourceID  string
	Content   string
	Entities  string
	CreatedAt time.Time
}, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if limit <= 0 {
		limit = 20
	}

	rows, err := r.db.Query(`
		SELECT id, source_id, content, entities, created_at
		FROM ingested_data
		WHERE recipe_name = ?
		ORDER BY created_at DESC
		LIMIT ?
	`, recipeName, limit)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []struct {
		ID        int64
		SourceID  string
		Content   string
		Entities  string
		CreatedAt time.Time
	}

	for rows.Next() {
		var item struct {
			ID        int64
			SourceID  string
			Content   string
			Entities  string
			CreatedAt time.Time
		}
		rows.Scan(&item.ID, &item.SourceID, &item.Content, &item.Entities, &item.CreatedAt)
		results = append(results, item)
	}

	return results, nil
}

func (r *RecipeStore) RunRecipe(ctx context.Context, name string, handler func(context.Context, *Recipe) error) error {
	recipe, err := r.Get(name)
	if err != nil {
		return err
	}

	if !recipe.Enabled {
		return fmt.Errorf("recipe %s is not enabled", name)
	}

	r.mu.Lock()
	r.db.Exec("UPDATE recipes SET status = 'running' WHERE name = ?", name)
	r.mu.Unlock()

	err = handler(ctx, recipe)

	r.mu.Lock()
	if err != nil {
		r.db.Exec("UPDATE recipes SET status = 'failed' WHERE name = ?", name)
	} else {
		r.db.Exec("UPDATE recipes SET status = 'idle', last_run = CURRENT_TIMESTAMP WHERE name = ?", name)
	}
	r.mu.Unlock()

	return err
}

func (r *RecipeStore) CheckDependencies(name string) (bool, []string) {
	recipe, err := r.Get(name)
	if err != nil {
		return false, []string{err.Error()}
	}

	var missing []string
	for _, req := range recipe.Requires {
		reqRecipe, err := r.Get(req)
		if err != nil || !reqRecipe.Enabled {
			missing = append(missing, req)
		}
	}

	return len(missing) == 0, missing
}

func (r *RecipeStore) Stats() (map[string]int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	stats := make(map[string]int)

	var ingestedCount int

	rows, _ := r.db.Query("SELECT name, enabled FROM recipes")
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var name string
			var enabled int
			rows.Scan(&name, &enabled)
			if enabled == 1 {
				stats["enabled"]++
			}
			stats["total"]++
		}
	}

	r.db.QueryRow("SELECT COUNT(*) FROM ingested_data").Scan(&ingestedCount)
	stats["ingested"] = ingestedCount

	return stats, nil
}

func (r *RecipeStore) Close() error {
	return r.db.Close()
}

func (r *RecipeStore) SetupInstructions(recipeName string) string {
	recipe, err := r.Get(recipeName)
	if err != nil {
		return "Recipe not found"
	}

	var lines []string
	lines = append(lines, fmt.Sprintf("# Setup: %s", recipe.Name))
	lines = append(lines, "")
	lines = append(lines, recipe.Description)
	lines = append(lines, "")

	if len(recipe.Requires) > 0 {
		lines = append(lines, "## Dependencies")
		for _, req := range recipe.Requires {
			lines = append(lines, fmt.Sprintf("- Enable `%s` first", req))
		}
		lines = append(lines, "")
	}

	if len(recipe.Credentials) > 0 {
		lines = append(lines, "## Required Credentials")
		for cred, typ := range recipe.Credentials {
			lines = append(lines, fmt.Sprintf("- `%s`: %s", cred, typ))
		}
		lines = append(lines, "")
	}

	if recipe.CronSchedule != "" {
		lines = append(lines, fmt.Sprintf("## Schedule"))
		lines = append(lines, recipe.CronSchedule)
		lines = append(lines, "")
	}

	lines = append(lines, "## Enable")
	lines = append(lines, fmt.Sprintf("Use the dashboard or CLI to enable this recipe."))

	return strings.Join(lines, "\n")
}