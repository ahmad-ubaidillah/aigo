package skills

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/ahmad-ubaidillah/aigo/pkg/types"
)

const (
	SkillsMPAPI = "https://skillsmp.com/api/v1/skills"
	GithubAPI   = "https://api.github.com"
)

type Marketplace struct {
	registry   *Registry
	httpClient *http.Client
}

type MarketplaceSkill struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Command     string `json:"command"`
	Category    string `json:"category"`
	Tags        string `json:"tags"`
	Author      string `json:"author"`
	Version     string `json:"version"`
	Stars       int    `json:"stars"`
	Source      string `json:"source"`
}

type SkillsMPResponse struct {
	Skills []MarketplaceSkill `json:"skills"`
	Total  int                `json:"total"`
	Page   int                `json:"page"`
}

func NewMarketplace(reg *Registry) *Marketplace {
	return &Marketplace{
		registry: reg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (m *Marketplace) ListBuiltIn() []MarketplaceSkill {
	return []MarketplaceSkill{
		{
			Name:        "git-master",
			Description: "Git operations - commit, rebase, branch management",
			Command:     "skill:git-master",
			Category:    "development",
			Tags:        "git,version-control,vcs",
			Author:      "aigo",
			Version:     "1.0.0",
			Source:      "built-in",
		},
		{
			Name:        "playwright",
			Description: "Browser automation - testing, scraping, screenshots",
			Command:     "skill:playwright",
			Category:    "automation",
			Tags:        "browser,automation,testing,web",
			Author:      "aigo",
			Version:     "1.0.0",
			Source:      "built-in",
		},
		{
			Name:        "frontend-ui-ux",
			Description: "Frontend development - React, Vue, styling, animations",
			Command:     "skill:frontend-ui-ux",
			Category:    "development",
			Tags:        "frontend,ui,ux,react,vue,css",
			Author:      "aigo",
			Version:     "1.0.0",
			Source:      "built-in",
		},
		{
			Name:        "dev-browser",
			Description: "Browser automation with persistent state",
			Command:     "skill:dev-browser",
			Category:    "automation",
			Tags:        "browser,automation,web-scraping",
			Author:      "aigo",
			Version:     "1.0.0",
			Source:      "built-in",
		},
		{
			Name:        "code-review",
			Description: "Code review with lint, security, and best practices",
			Command:     "skill:code-review",
			Category:    "development",
			Tags:        "code-review,linter,security",
			Author:      "aigo",
			Version:     "1.0.0",
			Source:      "built-in",
		},
		{
			Name:        "web-search",
			Description: "Web search using Exa AI for real-time information",
			Command:     "skill:web-search",
			Category:    "research",
			Tags:        "search,web,research,exa",
			Author:      "aigo",
			Version:     "1.0.0",
			Source:      "built-in",
		},
		{
			Name:        "code-search",
			Description: "Search code patterns across GitHub repositories",
			Command:     "skill:code-search",
			Category:    "research",
			Tags:        "search,code,github,grep",
			Author:      "aigo",
			Version:     "1.0.0",
			Source:      "built-in",
		},
		{
			Name:        "docs-lookup",
			Description: "Search official documentation for libraries and frameworks",
			Command:     "skill:docs-lookup",
			Category:    "research",
			Tags:        "docs,documentation,reference",
			Author:      "aigo",
			Version:     "1.0.0",
			Source:      "built-in",
		},
	}
}

func (m *Marketplace) ListLocal() ([]types.Skill, error) {
	return m.registry.List("")
}

func (m *Marketplace) SearchLocal(query string) ([]MarketplaceSkill, error) {
	builtIn := m.ListBuiltIn()
	var results []MarketplaceSkill
	for _, s := range builtIn {
		if containsIgnoreCase(s.Name, query) ||
			containsIgnoreCase(s.Description, query) ||
			containsIgnoreCase(s.Tags, query) {
			results = append(results, s)
		}
	}

	local, err := m.registry.Search(query)
	if err == nil {
		for _, s := range local {
			results = append(results, MarketplaceSkill{
				Name:        s.Name,
				Description: s.Description,
				Command:     s.Code,
				Category:    s.Category,
				Tags:        s.Tags,
				Author:      "local",
				Version:     "1.0.0",
				Source:      "local",
			})
		}
	}

	return results, nil
}

func (m *Marketplace) Search(query string, sources ...string) ([]MarketplaceSkill, error) {
	var results []MarketplaceSkill
	sourcesEnabled := sources
	if len(sourcesEnabled) == 0 {
		sourcesEnabled = []string{"built-in", "local", "skillsmp"}
	}

	for _, source := range sourcesEnabled {
		switch source {
		case "built-in":
			results = append(results, m.ListBuiltIn()...)
		case "local":
			local, err := m.SearchLocal(query)
			if err == nil {
				results = append(results, local...)
			}
		case "skillsmp":
			remote, err := m.FetchSkillsMP(query)
			if err == nil {
				results = append(results, remote...)
			}
		case "github":
			remote, err := m.FetchGithubSkills(query)
			if err == nil {
				results = append(results, remote...)
			}
		}
	}

	if query != "" {
		filtered := filterByQuery(results, query)
		return filtered, nil
	}

	return results, nil
}

func (m *Marketplace) FetchSkillsMP(query string) ([]MarketplaceSkill, error) {
	url := SkillsMPAPI
	if query != "" {
		url += "?q=" + query
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch from skillsmp: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("skillsmp API returned status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var result SkillsMPResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	var skills []MarketplaceSkill
	for _, s := range result.Skills {
		s.Source = "skillsmp"
		skills = append(skills, s)
	}

	return skills, nil
}

func (m *Marketplace) FetchGithubSkills(query string) ([]MarketplaceSkill, error) {
	url := fmt.Sprintf("%s/search/repositories?q=%s+skills&sort=stars&order=desc",
		GithubAPI, query)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch from github: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github API returned status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var result struct {
		Items []struct {
			Name        string `json:"name"`
			Description string `json:"description"`
			FullName    string `json:"full_name"`
			Stargazers  int    `json:"stargazers_count"`
			HTMLURL     string `json:"html_url"`
		} `json:"items"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	var skills []MarketplaceSkill
	for _, r := range result.Items {
		skills = append(skills, MarketplaceSkill{
			Name:        r.Name,
			Description: r.Description,
			Command:     r.HTMLURL,
			Category:    "github",
			Tags:        "github,claude-skills",
			Author:      r.FullName,
			Version:     "1.0.0",
			Stars:       r.Stargazers,
			Source:      "github",
		})
	}

	return skills, nil
}

func (m *Marketplace) Install(skill MarketplaceSkill) error {
	return m.registry.Register(
		skill.Name,
		skill.Description,
		skill.Command,
		skill.Category,
		skill.Tags,
	)
}

func (m *Marketplace) InstallFromGitHub(repoURL string) error {
	parts := strings.Split(strings.TrimSuffix(repoURL, ".git"), "/")
	if len(parts) < 2 {
		return fmt.Errorf("invalid repo URL: %s", repoURL)
	}

	owner := parts[len(parts)-2]
	repo := parts[len(parts)-1]

	tmpDir := fmt.Sprintf("/tmp/aigo-skills-%d", time.Now().UnixNano())
	defer os.RemoveAll(tmpDir)

	cloneCmd := exec.Command("git", "clone", "--depth", "1", repoURL, tmpDir)
	cloneCmd.Stdout = os.Stderr
	if err := cloneCmd.Run(); err != nil {
		return fmt.Errorf("clone repo: %w", err)
	}

	skillMDPath := tmpDir + "/SKILL.md"
	if _, err := os.Stat(skillMDPath); err != nil {
		return fmt.Errorf("SKILL.md not found in repo")
	}

	content, err := os.ReadFile(skillMDPath)
	if err != nil {
		return fmt.Errorf("read SKILL.md: %w", err)
	}

	name := repo
	description := extractDescription(string(content))
	command := extractCommand(string(content))

	category := "github"
	tags := fmt.Sprintf("github,%s,%s", owner, repo)

	return m.registry.Register(name, description, command, category, tags)
}

func extractDescription(skillMD string) string {
	lines := strings.Split(skillMD, "\n")
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			if i+1 < len(lines) {
				return strings.TrimSpace(lines[i+1])
			}
			return strings.TrimPrefix(line, "# ")
		}
	}
	return "Auto-installed skill from GitHub"
}

func extractCommand(skillMD string) string {
	if strings.Contains(skillMD, "skill:") {
		re := regexp.MustCompile(`skill:(\S+)`)
		matches := re.FindStringSubmatch(skillMD)
		if len(matches) > 1 {
			return matches[1]
		}
	}
	return "shell"
}

func (m *Marketplace) FetchRemote(url string) ([]MarketplaceSkill, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var skills []MarketplaceSkill
	if err := json.Unmarshal(body, &skills); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}

	return skills, nil
}

func filterByQuery(skills []MarketplaceSkill, query string) []MarketplaceSkill {
	query = lower(query)
	var results []MarketplaceSkill
	for _, s := range skills {
		if containsIgnoreCase(s.Name, query) ||
			containsIgnoreCase(s.Description, query) ||
			containsIgnoreCase(s.Tags, query) {
			results = append(results, s)
		}
	}
	return results
}

func containsIgnoreCase(s, substr string) bool {
	s = lower(s)
	substr = lower(substr)
	if substr == "" {
		return true
	}
	return strings.Contains(s, substr)
}

func lower(s string) string {
	return strings.ToLower(s)
}
