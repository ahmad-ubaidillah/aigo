// Package onlinehub provides online skill marketplace connectivity.
// It fetches skill indexes from remote sources and enables online skill installation.
package skillhub

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// RemoteSource represents an online skill index source.
type RemoteSource struct {
	Name     string `json:"name"`
	URL      string `json:"url"`
	Type     string `json:"type"` // "hermes-index", "lobehub", "raw-github"
	Enabled  bool   `json:"enabled"`
	LastSync string `json:"last_sync"`
}

// DefaultRemoteSources returns the default online sources.
func DefaultRemoteSources() []RemoteSource {
	return []RemoteSource{
		{
			Name:    "smithery-mcp",
			URL:     "https://registry.smithery.ai/servers",
			Type:    "smithery",
			Enabled: true,
		},
		{
			Name:    "anthropic-skills",
			URL:     "https://api.github.com/repos/anthropics/skills/contents/skills",
			Type:    "github-api",
			Enabled: true,
		},
		{
			Name:    "clawhub-registry",
			URL:     "https://raw.githubusercontent.com/hermes-v2/aigo-registry/main/index.json",
			Type:    "aigo-registry",
			Enabled: true,
		},
	}
}

// OnlineHub extends SkillHub with online capabilities.
type OnlineHub struct {
	*SkillHub
	sources    []RemoteSource
	sourcesPath string
	httpClient *http.Client
}

// NewOnlineHub creates a SkillHub with online connectivity.
func NewOnlineHub(baseDir string) (*OnlineHub, error) {
	hub, err := New(baseDir)
	if err != nil {
		return nil, err
	}

	sourcesPath := filepath.Join(baseDir, "sources.json")
	oh := &OnlineHub{
		SkillHub:    hub,
		sourcesPath: sourcesPath,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}

	// Load or create remote sources
	if data, err := os.ReadFile(sourcesPath); err == nil {
		json.Unmarshal(data, &oh.sources)
	}
	if len(oh.sources) == 0 {
		oh.sources = DefaultRemoteSources()
		oh.saveSources()
	}

	return oh, nil
}

func (oh *OnlineHub) saveSources() {
	data, _ := json.MarshalIndent(oh.sources, "", "  ")
	os.WriteFile(oh.sourcesPath, data, 0644)
}

// SyncIndex fetches the latest skill index from all remote sources.
func (oh *OnlineHub) SyncIndex() (*SyncResult, error) {
	result := &SyncResult{
		StartedAt: time.Now().Format(time.RFC3339),
	}

	for i := range oh.sources {
		src := &oh.sources[i]
		if !src.Enabled {
			continue
		}

		count, err := oh.syncSource(src)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("%s: %v", src.Name, err))
			continue
		}

		src.LastSync = time.Now().Format(time.RFC3339)
		result.Synced = append(result.Synced, SyncEntry{
			Source: src.Name,
			Count:  count,
		})
		result.TotalNew += count
	}

	oh.saveSources()
	result.CompletedAt = time.Now().Format(time.RFC3339)
	return result, nil
}

func (oh *OnlineHub) syncSource(src *RemoteSource) (int, error) {
	switch src.Type {
	case "smithery":
		return oh.syncSmithery(src)
	case "aigo-registry":
		return oh.syncAigoRegistry(src)
	case "github-api":
		return oh.syncGitHubAPI(src)
	case "lobehub":
		return oh.syncLobeHub(src)
	default:
		return 0, fmt.Errorf("unknown source type: %s", src.Type)
	}
}

// syncAigoRegistry fetches from our custom Aigo registry (when available).
func (oh *OnlineHub) syncAigoRegistry(src *RemoteSource) (int, error) {
	resp, err := oh.httpClient.Get(src.URL)
	if err != nil {
		// Registry not yet created — that's OK, skip silently
		return 0, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return 0, nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	var registry struct {
		Skills []Skill `json:"skills"`
	}
	if err := json.Unmarshal(body, &registry); err != nil {
		return 0, err
	}

	count := 0
	for _, s := range registry.Skills {
		s.Source = "clawhub"
		tagsJSON, _ := json.Marshal(s.Tags)
		oh.db.Exec(`INSERT OR REPLACE INTO skills (name, description, source, identifier, repo, path, tags, trust_level, installs, detail_url)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			s.Name, s.Description, s.Source, s.Identifier, s.Repo, s.Path,
			string(tagsJSON), s.TrustLevel, s.Installs, s.DetailURL)
		count++
	}
	return count, nil
}

// syncGitHubAPI fetches skills from GitHub API (public repos).
func (oh *OnlineHub) syncGitHubAPI(src *RemoteSource) (int, error) {
	resp, err := oh.httpClient.Get(src.URL)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return 0, fmt.Errorf("GitHub API returned %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	var contents []struct {
		Name string `json:"name"`
		Type string `json:"type"`
		Path string `json:"path"`
	}
	if err := json.Unmarshal(body, &contents); err != nil {
		return 0, err
	}

	count := 0
	repoBase := "anthropics/skills"
	for _, c := range contents {
		if c.Type != "dir" {
			continue
		}

		s := Skill{
			Name:        c.Name,
			Description: fmt.Sprintf("Anthropic skill: %s", c.Name),
			Source:      "github",
			Identifier:  fmt.Sprintf("github/anthropic/%s", c.Name),
			Repo:        repoBase,
			Path:        c.Path,
			TrustLevel:  "trusted",
		}

		tagsJSON, _ := json.Marshal(s.Tags)
		oh.db.Exec(`INSERT OR IGNORE INTO skills (name, description, source, identifier, repo, path, tags, trust_level)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			s.Name, s.Description, s.Source, s.Identifier, s.Repo, s.Path,
			string(tagsJSON), s.TrustLevel)
		count++
	}
	return count, nil
}

// syncLobeHub fetches the LobeHub agent index.
func (oh *OnlineHub) syncLobeHub(src *RemoteSource) (int, error) {
	resp, err := oh.httpClient.Get(src.URL)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return 0, fmt.Errorf("LobeHub returned %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	var index struct {
		Agents []struct {
			Identifier string `json:"identifier"`
			Meta       struct {
				Description string `json:"description"`
			} `json:"meta"`
		} `json:"agents"`
	}
	if err := json.Unmarshal(body, &index); err != nil {
		return 0, err
	}

	count := 0
	for _, a := range index.Agents {
		s := Skill{
			Name:        a.Identifier,
			Description: a.Meta.Description,
			Source:      "lobehub",
			Identifier:  "lobehub/" + a.Identifier,
			TrustLevel:  "community",
		}
		tagsJSON, _ := json.Marshal(s.Tags)
		oh.db.Exec(`INSERT OR IGNORE INTO skills (name, description, source, identifier, tags, trust_level)
			VALUES (?, ?, ?, ?, ?, ?)`,
			s.Name, s.Description, s.Source, s.Identifier,
			string(tagsJSON), s.TrustLevel)
		count++
	}
	return count, nil
}

// syncSmithery fetches MCP servers from Smithery registry.
// Smithery has 4,775+ MCP servers with REST API.
func (oh *OnlineHub) syncSmithery(src *RemoteSource) (int, error) {
	totalCount := 0
	pageSize := 10
	maxPages := 50 // Fetch up to 500 skills per sync

	for page := 1; page <= maxPages; page++ {
		url := fmt.Sprintf("%s?page=%d&pageSize=%d", src.URL, page, pageSize)
		resp, err := oh.httpClient.Get(url)
		if err != nil {
			return totalCount, err
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return totalCount, err
		}

		if resp.StatusCode == 429 {
			// Rate limited — wait and retry
			time.Sleep(5 * time.Second)
			page-- // Retry this page
			continue
		}
		if resp.StatusCode != 200 {
			return totalCount, fmt.Errorf("Smithery returned %d", resp.StatusCode)
		}

		var result struct {
			Servers []struct {
				QualifiedName string `json:"qualifiedName"`
				DisplayName   string `json:"displayName"`
				Description   string `json:"description"`
				Homepage      string `json:"homepage"`
				UseCount      int    `json:"useCount"`
				Verified      bool   `json:"verified"`
			} `json:"servers"`
			Pagination struct {
				TotalPages int `json:"totalPages"`
			} `json:"pagination"`
		}

		if err := json.Unmarshal(body, &result); err != nil {
			return totalCount, fmt.Errorf("page %d parse error: %w", page, err)
		}

		if len(result.Servers) == 0 {
			break // No more results
		}

		for _, srv := range result.Servers {
			trustLevel := "community"
			if srv.Verified {
				trustLevel = "trusted"
			}

			s := Skill{
				Name:        srv.DisplayName,
				Description: srv.Description,
				Source:      "smithery",
				Identifier:  "smithery/" + srv.QualifiedName,
				TrustLevel:  trustLevel,
				Installs:    srv.UseCount,
				DetailURL:   srv.Homepage,
			}

			tagsJSON, _ := json.Marshal(s.Tags)
			oh.db.Exec(`INSERT OR REPLACE INTO skills (name, description, source, identifier, tags, trust_level, installs, detail_url)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
				s.Name, s.Description, s.Source, s.Identifier,
				string(tagsJSON), s.TrustLevel, s.Installs, s.DetailURL)
			totalCount++
		}
	}

	return totalCount, nil
}

// AddSource adds a new remote source (like adding a "tap").
func (oh *OnlineHub) AddSource(name, url, sourceType string) error {
	for _, s := range oh.sources {
		if s.Name == name {
			return fmt.Errorf("source already exists: %s", name)
		}
	}

	oh.sources = append(oh.sources, RemoteSource{
		Name:    name,
		URL:     url,
		Type:    sourceType,
		Enabled: true,
	})
	oh.saveSources()
	return nil
}

// RemoveSource removes a remote source.
func (oh *OnlineHub) RemoveSource(name string) error {
	for i, s := range oh.sources {
		if s.Name == name {
			oh.sources = append(oh.sources[:i], oh.sources[i+1:]...)
			oh.saveSources()
			return nil
		}
	}
	return fmt.Errorf("source not found: %s", name)
}

// ListSources returns all configured remote sources.
func (oh *OnlineHub) ListSources() []RemoteSource {
	return oh.sources
}

// FindByIdentifier finds a skill by its identifier (public wrapper).
func (oh *OnlineHub) FindByIdentifier(identifier string) (*Skill, error) {
	return oh.findByIdentifier(identifier)
}

// FindByName finds a skill by its name.
func (oh *OnlineHub) FindByName(name string) (*Skill, error) {
	var s Skill
	var tagsJSON, repo, path sql.NullString
	err := oh.db.QueryRow(`
		SELECT name, description, source, identifier, repo, path, tags, trust_level, installs, detail_url
		FROM skills WHERE name = ?`, name).
		Scan(&s.Name, &s.Description, &s.Source, &s.Identifier,
			&repo, &path, &tagsJSON, &s.TrustLevel, &s.Installs, &s.DetailURL)
	if err != nil {
		return nil, fmt.Errorf("skill not found by name: %s", name)
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

// BrowseOnline fetches and displays skills from a specific source without installing.
func (oh *OnlineHub) BrowseOnline(source string, limit int) ([]Skill, error) {
	if limit <= 0 {
		limit = 20
	}

	// Search in DB for skills from this source
	rows, err := oh.db.Query(`
		SELECT name, description, source, identifier, repo, path, tags, trust_level, installs
		FROM skills WHERE source = ?
		ORDER BY installs DESC
		LIMIT ?`, source, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []Skill
	for rows.Next() {
		var s Skill
		var tagsJSON string
		if err := rows.Scan(&s.Name, &s.Description, &s.Source, &s.Identifier,
			&s.Repo, &s.Path, &tagsJSON, &s.TrustLevel, &s.Installs); err != nil {
			continue
		}
		json.Unmarshal([]byte(tagsJSON), &s.Tags)
		results = append(results, s)
	}
	return results, nil
}

// PopularSkills returns the most popular skills across all sources.
func (oh *OnlineHub) PopularSkills(limit int) ([]Skill, error) {
	if limit <= 0 {
		limit = 10
	}

	rows, err := oh.db.Query(`
		SELECT name, description, source, identifier, repo, path, tags, trust_level, installs
		FROM skills
		WHERE installs > 0
		ORDER BY installs DESC
		LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []Skill
	for rows.Next() {
		var s Skill
		var tagsJSON, repo, path sql.NullString
		if err := rows.Scan(&s.Name, &s.Description, &s.Source, &s.Identifier,
			&repo, &path, &tagsJSON, &s.TrustLevel, &s.Installs); err != nil {
			continue
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
		results = append(results, s)
	}
	return results, nil
}

// SyncResult contains the result of a sync operation.
type SyncResult struct {
	StartedAt   string      `json:"started_at"`
	CompletedAt string      `json:"completed_at"`
	TotalNew    int         `json:"total_new"`
	Synced      []SyncEntry `json:"synced"`
	Errors      []string    `json:"errors,omitempty"`
}

type SyncEntry struct {
	Source string `json:"source"`
	Count  int    `json:"count"`
}

func (r *SyncResult) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("🔄 Sync complete: %d new skills\n", r.TotalNew))
	for _, s := range r.Synced {
		sb.WriteString(fmt.Sprintf("  ✅ %s: %d skills\n", s.Source, s.Count))
	}
	for _, e := range r.Errors {
		sb.WriteString(fmt.Sprintf("  ❌ %s\n", e))
	}
	return sb.String()
}
