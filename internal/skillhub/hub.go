// Package skillhub provides Aigo's skill marketplace — search, install, manage skills.
// Supports both Hermes SKILL.md format and Aigo native format.
package skillhub

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

// Skill represents a skill from any source.
type Skill struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Source      string   `json:"source"`      // "hermes", "lobehub", "aigo"
	Identifier  string   `json:"identifier"`  // e.g., "official/devops/docker-management"
	Repo        string   `json:"repo"`        // GitHub repo (for download)
	Path        string   `json:"path"`        // Path within repo
	Tags        []string `json:"tags"`
	TrustLevel  string   `json:"trust_level"` // "builtin", "trusted", "community"
	Installs    int      `json:"installs"`    // Popularity metric
	DetailURL   string   `json:"detail_url"`
}

// SkillContent is the loaded skill content.
type SkillContent struct {
	Skill
	Body        string            `json:"body"`        // Markdown content
	Frontmatter map[string]string `json:"frontmatter"` // YAML frontmatter
	Files       map[string]string `json:"files"`       // Additional files (scripts, templates)
	InstalledAt string            `json:"installed_at"`
}

// SkillHub manages the skill marketplace.
type SkillHub struct {
	baseDir  string // ~/.aigo/skills/
	dbPath   string // ~/.aigo/skills/skills.db
	db       *sql.DB
	index    []Skill
	lobehub  []Skill
}

// New creates a new SkillHub.
func New(baseDir string) (*SkillHub, error) {
	if baseDir == "" {
		home, _ := os.UserHomeDir()
		baseDir = filepath.Join(home, ".aigo", "skills")
	}

	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("create skills dir: %w", err)
	}

	dbPath := filepath.Join(baseDir, "skills.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	hub := &SkillHub{
		baseDir: baseDir,
		dbPath:  dbPath,
		db:      db,
	}

	if err := hub.initDB(); err != nil {
		return nil, fmt.Errorf("init db: %w", err)
	}

	// Load indexes
	if err := hub.loadIndexes(); err != nil {
		fmt.Printf("Warning: could not load skill indexes: %v\n", err)
	}

	return hub, nil
}

func (h *SkillHub) initDB() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS skills (
			name TEXT PRIMARY KEY,
			description TEXT,
			source TEXT,
			identifier TEXT UNIQUE,
			repo TEXT,
			path TEXT,
			tags TEXT,
			trust_level TEXT,
			installs INTEGER DEFAULT 0,
			detail_url TEXT,
			installed INTEGER DEFAULT 0,
			installed_at TEXT,
			body TEXT
		)`,
		`CREATE VIRTUAL TABLE IF NOT EXISTS skills_fts USING fts5(
			name, description, tags, body,
			content='skills',
			content_rowid='rowid'
		)`,
		`CREATE TRIGGER IF NOT EXISTS skills_ai AFTER INSERT ON skills BEGIN
			INSERT INTO skills_fts(rowid, name, description, tags, body)
			VALUES (new.rowid, new.name, new.description, new.tags, new.body);
		END`,
		`CREATE TRIGGER IF NOT EXISTS skills_au AFTER UPDATE ON skills BEGIN
			INSERT INTO skills_fts(skills_fts, rowid, name, description, tags, body)
			VALUES ('delete', old.rowid, old.name, old.description, old.tags, old.body);
			INSERT INTO skills_fts(rowid, name, description, tags, body)
			VALUES (new.rowid, new.name, new.description, new.tags, new.body);
		END`,
		`CREATE TRIGGER IF NOT EXISTS skills_ad AFTER DELETE ON skills BEGIN
			INSERT INTO skills_fts(skills_fts, rowid, name, description, tags, body)
			VALUES ('delete', old.rowid, old.name, old.description, old.tags, old.body);
		END`,
	}

	for _, q := range queries {
		if _, err := h.db.Exec(q); err != nil {
			return fmt.Errorf("exec query: %w", err)
		}
	}
	return nil
}

// loadIndexes loads the Hermes + LobeHub indexes from hermes installation.
func (h *SkillHub) loadIndexes() error {
	// Try to find hermes index
	possiblePaths := []string{
		filepath.Join(os.Getenv("HOME"), ".hermes", "skills", ".hub", "index-cache", "hermes-index.json"),
		"/mnt/hermes/skills/.hub/index-cache/hermes-index.json",
	}

	for _, indexPath := range possiblePaths {
		if data, err := os.ReadFile(indexPath); err == nil {
			var index struct {
				Skills []Skill `json:"skills"`
			}
			if err := json.Unmarshal(data, &index); err == nil {
				h.index = index.Skills
				// Insert into DB (skip if already exists)
				h.syncIndexToDB()
				fmt.Printf("Loaded %d skills from Hermes index\n", len(h.index))
				break
			}
		}
	}

	// Try LobeHub index
	lobePaths := []string{
		filepath.Join(os.Getenv("HOME"), ".hermes", "skills", ".hub", "index-cache", "lobehub_index.json"),
		"/mnt/hermes/skills/.hub/index-cache/lobehub_index.json",
	}

	for _, indexPath := range lobePaths {
		if data, err := os.ReadFile(indexPath); err == nil {
			var index struct {
				Agents []struct {
					Identifier string `json:"identifier"`
					Meta       struct {
						Description string `json:"description"`
						Avatar      string `json:"avatar"`
					} `json:"meta"`
					Author string `json:"author"`
				} `json:"agents"`
			}
			if err := json.Unmarshal(data, &index); err == nil {
				for _, a := range index.Agents {
					h.lobehub = append(h.lobehub, Skill{
						Name:        a.Identifier,
						Description: a.Meta.Description,
						Source:      "lobehub",
						Identifier:  "lobehub/" + a.Identifier,
						TrustLevel:  "community",
					})
				}
				fmt.Printf("Loaded %d LobeHub skills\n", len(h.lobehub))
				break
			}
		}
	}

	return nil
}

func (h *SkillHub) syncIndexToDB() {
	for _, s := range h.index {
		tagsJSON, _ := json.Marshal(s.Tags)
		h.db.Exec(`INSERT OR IGNORE INTO skills (name, description, source, identifier, repo, path, tags, trust_level, installs, detail_url)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			s.Name, s.Description, s.Source, s.Identifier, s.Repo, s.Path,
			string(tagsJSON), s.TrustLevel, s.Installs, s.DetailURL)
	}
}

// scanSkill scans a row into a Skill, handling NULL values.
func scanSkill(rows interface{ Scan(dest ...interface{}) error }) (Skill, string, error) {
	var s Skill
	var tagsJSON, repo, path, detailURL sql.NullString
	err := rows.Scan(&s.Name, &s.Description, &s.Source, &s.Identifier,
		&repo, &path, &tagsJSON, &s.TrustLevel, &s.Installs, &detailURL)
	if err != nil {
		return s, "", err
	}
	if repo.Valid {
		s.Repo = repo.String
	}
	if path.Valid {
		s.Path = path.String
	}
	if detailURL.Valid {
		s.DetailURL = detailURL.String
	}
	tagsStr := ""
	if tagsJSON.Valid {
		tagsStr = tagsJSON.String
	}
	return s, tagsStr, nil
}

// Search searches for skills by query.
func (h *SkillHub) Search(query string, limit int) ([]Skill, error) {
	if limit <= 0 {
		limit = 10
	}

	// Use FTS5 search
	rows, err := h.db.Query(`
		SELECT s.name, s.description, s.source, s.identifier, s.repo, s.path, s.tags, s.trust_level, s.installs, s.detail_url
		FROM skills_fts fts
		JOIN skills s ON s.rowid = fts.rowid
		WHERE skills_fts MATCH ?
		ORDER BY rank
		LIMIT ?`, query, limit)
	if err != nil {
		// Fallback to LIKE search
		return h.searchLike(query, limit)
	}
	defer rows.Close()

	var results []Skill
	for rows.Next() {
		s, tagsJSON, err := scanSkill(rows)
		if err != nil {
			continue
		}
		json.Unmarshal([]byte(tagsJSON), &s.Tags)
		results = append(results, s)
	}
	return results, nil
}

func (h *SkillHub) searchLike(query string, limit int) ([]Skill, error) {
	pattern := "%" + query + "%"
	rows, err := h.db.Query(`
		SELECT name, description, source, identifier, repo, path, tags, trust_level, installs, detail_url
		FROM skills
		WHERE name LIKE ? OR description LIKE ? OR tags LIKE ?
		ORDER BY installs DESC
		LIMIT ?`, pattern, pattern, pattern, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []Skill
	for rows.Next() {
		s, tagsJSON, err := scanSkill(rows)
		if err != nil {
			continue
		}
		json.Unmarshal([]byte(tagsJSON), &s.Tags)
		results = append(results, s)
	}
	return results, nil
}

// Install downloads and installs a skill by its identifier.
// Supports: GitHub (Hermes/Anthropic), Smithery API, and local generation.
func (h *SkillHub) Install(identifier string) error {
	// Find skill in index
	skill, err := h.findByIdentifier(identifier)
	if err != nil {
		return fmt.Errorf("skill not found: %s", identifier)
	}

	// Check if already installed
	installDir := filepath.Join(h.baseDir, skill.Source, skill.Name)
	skillFile := filepath.Join(installDir, "SKILL.md")
	if _, err := os.Stat(skillFile); err == nil {
		return fmt.Errorf("already installed: %s", skill.Name)
	}

	// Download based on source
	var content string

	switch {
	case skill.Source == "smithery":
		content, err = h.installSmithery(skill)
	case skill.Source == "github" || skill.Source == "official":
		// For official/Hermes skills, check local installation first
		if skill.Source == "official" && skill.Path != "" {
			localPaths := []string{
				filepath.Join(os.Getenv("HOME"), ".hermes", "skills", skill.Path, "SKILL.md"),
				filepath.Join("/mnt/hermes/skills", skill.Path, "SKILL.md"),
			}
			for _, lp := range localPaths {
				if data, err := os.ReadFile(lp); err == nil {
					content = string(data)
					break
				}
			}
		}
		// If not found locally, try GitHub
		if content == "" {
			if skill.Repo == "" {
				// Default to hermes repo
				skill.Repo = "hermes-v2/awesome-hermes-agent"
			}
			content, err = h.downloadFromGitHub(skill.Repo, skill.Path)
		}
	case skill.Source == "lobehub":
		content, err = h.installLobeHub(skill)
	default:
		// Try GitHub as fallback if repo exists
		if skill.Repo != "" {
			content, err = h.downloadFromGitHub(skill.Repo, skill.Path)
			if err != nil {
				// Download failed — generate from metadata
				content, err = h.installGenerated(skill)
			}
		} else {
			content, err = h.installGenerated(skill)
		}
	}
	if err != nil {
		return fmt.Errorf("download: %w", err)
	}

	// Save
	if err := os.MkdirAll(installDir, 0755); err != nil {
		return fmt.Errorf("create dir: %w", err)
	}

	if err := os.WriteFile(skillFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("write skill: %w", err)
	}

	// Update DB
	h.db.Exec(`UPDATE skills SET installed = 1, installed_at = ?, body = ? WHERE identifier = ?`,
		time.Now().Format(time.RFC3339), content[:min(len(content), 10000)], identifier)

	fmt.Printf("Installed: %s → %s\n", skill.Name, installDir)
	return nil
}

// installSmithery fetches a Smithery MCP server's README as SKILL.md.
func (h *SkillHub) installSmithery(skill *Skill) (string, error) {
	// Smithery detail_url format: https://smithery.ai/servers/EthanHenrickson/math-mcp
	// Try to fetch README from the GitHub repo associated with this MCP server
	// Extract repo from qualifiedName: EthanHenrickson/math-mcp → github.com/EthanHenrickson/math-mcp
	parts := strings.SplitN(skill.Identifier, "/", 3) // smithery/EthanHenrickson/math-mcp
	if len(parts) >= 3 {
		owner := parts[1]
		repo := parts[2]
		// Try GitHub raw README
		for _, branch := range []string{"main", "master"} {
			url := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/README.md", owner, repo, branch)
			resp, err := http.Get(url)
			if err == nil && resp.StatusCode == 200 {
				body, _ := io.ReadAll(resp.Body)
				resp.Body.Close()
				skillMd := fmt.Sprintf("---\nname: %s\nsource: smithery\nidentifier: %s\n---\n\n# %s\n\n%s\n\n---\n*Install via Smithery: `npx -y @anthropic-ai/skills install smithery/%s/%s`*",
					skill.Name, skill.Identifier, skill.Name, string(body), owner, repo)
				return skillMd, nil
			}
			if resp != nil {
				resp.Body.Close()
			}
		}
	}

	// Fallback: generate from metadata
	return h.installGenerated(skill)
}

// installLobeHub generates a SKILL.md for LobeHub agents.
func (h *SkillHub) installLobeHub(skill *Skill) (string, error) {
	return fmt.Sprintf(`---
name: %s
source: lobehub
identifier: %s
---

# %s

%s

---
*Source: LobeHub Agent Marketplace*
`, skill.Name, skill.Identifier, skill.Name, skill.Description), nil
}

// installGenerated generates a basic SKILL.md from skill metadata.
func (h *SkillHub) installGenerated(skill *Skill) (string, error) {
	tags := strings.Join(skill.Tags, ", ")
	url := skill.DetailURL
	if url == "" {
		url = "N/A"
	}
	return fmt.Sprintf(`---
name: %s
source: %s
identifier: %s
trust_level: %s
tags: %s
---

# %s

%s

---

**Source:** %s
**Trust Level:** %s
**Installs:** %d
**URL:** %s
`, skill.Name, skill.Source, skill.Identifier, skill.TrustLevel, tags,
		skill.Name, skill.Description, skill.Source, skill.TrustLevel, skill.Installs, url), nil
}

func (h *SkillHub) findByIdentifier(identifier string) (*Skill, error) {
	// Check in-memory index first
	for _, s := range h.index {
		if s.Identifier == identifier {
			return &s, nil
		}
	}

	// Check DB
	var s Skill
	var tagsJSON, repo, path sql.NullString
	err := h.db.QueryRow(`
		SELECT name, description, source, identifier, repo, path, tags, trust_level, installs, detail_url
		FROM skills WHERE identifier = ?`, identifier).
		Scan(&s.Name, &s.Description, &s.Source, &s.Identifier,
			&repo, &path, &tagsJSON, &s.TrustLevel, &s.Installs, &s.DetailURL)
	if err != nil {
		return nil, err
	}
	if repo.Valid {
		s.Repo = repo.String
	}
	if path.Valid {
		s.Path = path.String
	}
	if tagsJSON.Valid {
		json.Unmarshal([]byte(tagsJSON.String), &s.Tags)
	}
	return &s, nil
}

func (h *SkillHub) downloadFromGitHub(repo, path string) (string, error) {
	// Try raw.githubusercontent.com
	url := fmt.Sprintf("https://raw.githubusercontent.com/%s/main/%s/SKILL.md", repo, path)
	resp, err := http.Get(url)
	if err != nil {
		// Try master branch
		url = fmt.Sprintf("https://raw.githubusercontent.com/%s/master/%s/SKILL.md", repo, path)
		resp, err = http.Get(url)
		if err != nil {
			return "", err
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("HTTP %d for %s", resp.StatusCode, url)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

// View loads a skill's content.
func (h *SkillHub) View(name string) (*SkillContent, error) {
	// Try installed path first
	skillDirs := []string{
		filepath.Join(h.baseDir),
	}

	for _, dir := range skillDirs {
		skillFile := filepath.Join(dir, name, "SKILL.md")
		if data, err := os.ReadFile(skillFile); err == nil {
			sc := &SkillContent{
				Skill:       Skill{Name: name},
				Body:        string(data),
				Frontmatter: parseFrontmatter(string(data)),
			}
			return sc, nil
		}
	}

	// Try DB
	var body string
	var identifier string
	err := h.db.QueryRow(`SELECT body, identifier FROM skills WHERE name = ?`, name).
		Scan(&body, &identifier)
	if err == nil && body != "" {
		return &SkillContent{
			Skill:       Skill{Name: name, Identifier: identifier},
			Body:        body,
			Frontmatter: parseFrontmatter(body),
		}, nil
	}

	return nil, fmt.Errorf("skill not found: %s", name)
}

// ListInstalled returns all installed skills.
func (h *SkillHub) ListInstalled() ([]Skill, error) {
	var skills []Skill

	// Walk the skills directory
	filepath.Walk(h.baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.Name() == "SKILL.md" {
			relpath, _ := filepath.Rel(h.baseDir, filepath.Dir(path))
			parts := strings.Split(relpath, string(filepath.Separator))
			if len(parts) >= 2 {
				skills = append(skills, Skill{
					Name:   parts[1],
					Source: parts[0],
					Path:   relpath,
				})
			}
		}
		return nil
	})

	return skills, nil
}

// ListByCategory returns skills in a category.
func (h *SkillHub) ListByCategory(category string, limit int) ([]Skill, error) {
	if limit <= 0 {
		limit = 20
	}

	rows, err := h.db.Query(`
		SELECT name, description, source, identifier, repo, path, tags, trust_level, installs, detail_url
		FROM skills
		WHERE path LIKE ?
		ORDER BY installs DESC
		LIMIT ?`, category+"/%", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []Skill
	for rows.Next() {
		s, tagsJSON, err := scanSkill(rows)
		if err != nil {
			continue
		}
		json.Unmarshal([]byte(tagsJSON), &s.Tags)
		results = append(results, s)
	}
	return results, nil
}

// Stats returns hub statistics.
func (h *SkillHub) Stats() map[string]interface{} {
	var total, installed, categories int
	h.db.QueryRow(`SELECT COUNT(*) FROM skills`).Scan(&total)
	h.db.QueryRow(`SELECT COUNT(*) FROM skills WHERE installed = 1`).Scan(&installed)

	rows, _ := h.db.Query(`SELECT DISTINCT substr(path, 1, instr(path, '/')-1) FROM skills WHERE path LIKE '%/%'`)
	if rows != nil {
		for rows.Next() {
			categories++
		}
		rows.Close()
	}

	return map[string]interface{}{
		"total_indexed": total,
		"installed":     installed,
		"categories":    categories,
		"hermes_index":  len(h.index),
		"lobehub":       len(h.lobehub),
		"db_path":       h.dbPath,
		"skills_dir":    h.baseDir,
	}
}

// Remove uninstalls a skill.
func (h *SkillHub) Remove(name string) error {
	// Find installed path
	dirs, _ := os.ReadDir(h.baseDir)
	for _, d := range dirs {
		if d.IsDir() {
			skillPath := filepath.Join(h.baseDir, d.Name(), name)
			if _, err := os.Stat(filepath.Join(skillPath, "SKILL.md")); err == nil {
				os.RemoveAll(skillPath)
				h.db.Exec(`UPDATE skills SET installed = 0 WHERE name = ?`, name)
				fmt.Printf("Removed: %s\n", name)
				return nil
			}
		}
	}
	return fmt.Errorf("not installed: %s", name)
}

// Close closes the hub.
func (h *SkillHub) Close() {
	if h.db != nil {
		h.db.Close()
	}
}

// DB returns the underlying database connection (for debugging).
func (h *SkillHub) DB() *sql.DB {
	return h.db
}

// parseFrontmatter extracts YAML frontmatter from SKILL.md content.
func parseFrontmatter(content string) map[string]string {
	result := make(map[string]string)
	re := regexp.MustCompile(`(?s)^---\s*\n(.*?)\n---`)
	m := re.FindStringSubmatch(content)
	if len(m) < 2 {
		return result
	}

	fm := m[1]
	for _, line := range strings.Split(fm, "\n") {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			val = strings.Trim(val, `"`)
			result[key] = val
		}
	}
	return result
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
