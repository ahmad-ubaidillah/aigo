package templates

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/ahmad-ubaidillah/aigo/internal/nodes"
)

type Registry struct {
	mu        sync.RWMutex
	templates map[string]*Template
	dir       string
}

type Template struct {
	ID          string               `json:"id"`
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Category    string               `json:"category"`
	Nodes       []nodes.WorkflowNode `json:"nodes"`
	Edges       []nodes.WorkflowEdge `json:"edges"`
	Variables   []Variable           `json:"variables"`
	Examples    []string             `json:"examples"`
	Author      string               `json:"author"`
	Version     int                  `json:"version"`
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`
	UsageCount  int                  `json:"usage_count"`
	Rating      float64              `json:"rating"`
	Tags        []string             `json:"tags"`
}

type Variable struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"`
	Default     interface{} `json:"default"`
	Description string      `json:"description"`
	Required    bool        `json:"required"`
}

func NewRegistry(dir string) *Registry {
	r := &Registry{
		templates: make(map[string]*Template),
		dir:       dir,
	}
	r.loadBuiltins()
	return r
}

func (r *Registry) loadBuiltins() {
	builtins := []*Template{
		r.NewWebScraperTemplate(),
		r.NewAPITestTemplate(),
		r.NewDataPipelineTemplate(),
		r.NewContentGeneratorTemplate(),
		r.NewAutomationWorkflowTemplate(),
	}

	for _, t := range builtins {
		r.templates[t.ID] = t
	}
}

func (r *Registry) NewWebScraperTemplate() *Template {
	return &Template{
		ID:          "web_scraper",
		Name:        "Web Scraper",
		Description: "Scrape data from websites using browser automation",
		Category:    "automation",
		Nodes: []nodes.WorkflowNode{
			{ID: "start", Type: "input", Config: nodes.NodeConfig{"prompt": "Enter URL to scrape"}},
			{ID: "browser", Type: "browser", Config: nodes.NodeConfig{"action": "navigate"}},
			{ID: "extract", Type: "code", Config: nodes.NodeConfig{"language": "javascript"}},
			{ID: "save", Type: "code", Config: nodes.NodeConfig{"action": "save"}},
		},
		Edges: []nodes.WorkflowEdge{
			{From: "start", To: "browser"},
			{From: "browser", To: "extract"},
			{From: "extract", To: "save"},
		},
		Variables: []Variable{
			{Name: "url", Type: "string", Required: true, Description: "URL to scrape"},
			{Name: "selectors", Type: "array", Required: false, Description: "CSS selectors to extract"},
		},
		Examples:  []string{"Scrape product prices from example.com", "Extract article content from news site"},
		Author:    "Aigo",
		Version:   1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Tags:      []string{"web", "scraping", "browser", "automation"},
	}
}

func (r *Registry) NewAPITestTemplate() *Template {
	return &Template{
		ID:          "api_test",
		Name:        "API Test",
		Description: "Test REST APIs with various HTTP methods and assertions",
		Category:    "testing",
		Nodes: []nodes.WorkflowNode{
			{ID: "start", Type: "input", Config: nodes.NodeConfig{"prompt": "Enter API endpoint"}},
			{ID: "http", Type: "http", Config: nodes.NodeConfig{"method": "GET"}},
			{ID: "validate", Type: "condition", Config: nodes.NodeConfig{"expected_status": 200}},
			{ID: "report", Type: "transform", Config: nodes.NodeConfig{"format": "json"}},
		},
		Edges: []nodes.WorkflowEdge{
			{From: "start", To: "http"},
			{From: "http", To: "validate"},
			{From: "validate", To: "report"},
		},
		Variables: []Variable{
			{Name: "endpoint", Type: "string", Required: true, Description: "API endpoint URL"},
			{Name: "method", Type: "string", Default: "GET", Description: "HTTP method"},
			{Name: "headers", Type: "object", Required: false, Description: "Request headers"},
			{Name: "body", Type: "string", Required: false, Description: "Request body"},
			{Name: "expected_status", Type: "number", Default: 200, Description: "Expected status code"},
		},
		Examples:  []string{"Test GET /api/users", "Test POST /api/login with credentials"},
		Author:    "Aigo",
		Version:   1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Tags:      []string{"api", "testing", "http", "rest"},
	}
}

func (r *Registry) NewDataPipelineTemplate() *Template {
	return &Template{
		ID:          "data_pipeline",
		Name:        "Data Pipeline",
		Description: "Extract, transform, and load data between services",
		Category:    "data",
		Nodes: []nodes.WorkflowNode{
			{ID: "source", Type: "input", Config: nodes.NodeConfig{"prompt": "Data source configuration"}},
			{ID: "extract", Type: "code", Config: nodes.NodeConfig{"action": "extract"}},
			{ID: "transform", Type: "code", Config: nodes.NodeConfig{"action": "transform"}},
			{ID: "load", Type: "code", Config: nodes.NodeConfig{"action": "load"}},
		},
		Edges: []nodes.WorkflowEdge{
			{From: "source", To: "extract"},
			{From: "extract", To: "transform"},
			{From: "transform", To: "load"},
		},
		Variables: []Variable{
			{Name: "source_type", Type: "string", Required: true, Description: "Source type (api, file, db)"},
			{Name: "transformations", Type: "array", Required: false, Description: "Transformation steps"},
			{Name: "destination", Type: "string", Required: true, Description: "Destination type"},
		},
		Examples:  []string{"Copy data from API to database", "Transform CSV to JSON"},
		Author:    "Aigo",
		Version:   1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Tags:      []string{"data", "etl", "pipeline", "automation"},
	}
}

func (r *Registry) NewContentGeneratorTemplate() *Template {
	return &Template{
		ID:          "content_generator",
		Name:        "Content Generator",
		Description: "Generate content using LLM with research and editing",
		Category:    "ai",
		Nodes: []nodes.WorkflowNode{
			{ID: "research", Type: "search", Config: nodes.NodeConfig{"sources": []string{"web"}}},
			{ID: "outline", Type: "llm", Config: nodes.NodeConfig{"prompt": "Create outline"}},
			{ID: "draft", Type: "llm", Config: nodes.NodeConfig{"prompt": "Write content"}},
			{ID: "edit", Type: "code", Config: nodes.NodeConfig{"action": "edit"}},
		},
		Edges: []nodes.WorkflowEdge{
			{From: "research", To: "outline"},
			{From: "outline", To: "draft"},
			{From: "draft", To: "edit"},
		},
		Variables: []Variable{
			{Name: "topic", Type: "string", Required: true, Description: "Content topic"},
			{Name: "style", Type: "string", Default: "informative", Description: "Writing style"},
			{Name: "length", Type: "number", Default: 500, Description: "Target word count"},
		},
		Examples:  []string{"Generate blog post about AI", "Write product description"},
		Author:    "Aigo",
		Version:   1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Tags:      []string{"ai", "llm", "content", "writing"},
	}
}

func (r *Registry) NewAutomationWorkflowTemplate() *Template {
	return &Template{
		ID:          "automation_workflow",
		Name:        "Automation Workflow",
		Description: "General purpose automation workflow with conditions and loops",
		Category:    "automation",
		Nodes: []nodes.WorkflowNode{
			{ID: "trigger", Type: "input", Config: nodes.NodeConfig{"prompt": "Trigger condition"}},
			{ID: "check", Type: "condition", Config: nodes.NodeConfig{"expression": ""}},
			{ID: "process", Type: "code", Config: nodes.NodeConfig{"action": "process"}},
			{ID: "notify", Type: "llm", Config: nodes.NodeConfig{"prompt": "Send notification"}},
		},
		Edges: []nodes.WorkflowEdge{
			{From: "trigger", To: "check"},
			{From: "check", To: "process"},
			{From: "process", To: "notify"},
		},
		Variables: []Variable{
			{Name: "condition", Type: "string", Required: true, Description: "Trigger condition"},
			{Name: "action", Type: "string", Required: true, Description: "Action to perform"},
		},
		Examples:  []string{"Daily report generation", "Error monitoring and alerting"},
		Author:    "Aigo",
		Version:   1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Tags:      []string{"automation", "workflow", "scheduled"},
	}
}

func (r *Registry) Register(t *Template) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if t.ID == "" {
		return fmt.Errorf("template id cannot be empty")
	}
	if _, exists := r.templates[t.ID]; exists {
		return fmt.Errorf("template already registered: %s", t.ID)
	}

	t.CreatedAt = time.Now()
	t.UpdatedAt = time.Now()
	r.templates[t.ID] = t
	return nil
}

func (r *Registry) Get(id string) (*Template, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	t, ok := r.templates[id]
	if !ok {
		return nil, fmt.Errorf("template not found: %s", id)
	}
	return t, nil
}

func (r *Registry) List() []*Template {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*Template, 0, len(r.templates))
	for _, t := range r.templates {
		result = append(result, t)
	}
	return result
}

func (r *Registry) ListByCategory(category string) []*Template {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*Template
	for _, t := range r.templates {
		if t.Category == category {
			result = append(result, t)
		}
	}
	return result
}

func (r *Registry) Search(query string) []*Template {
	r.mu.RLock()
	defer r.mu.RUnlock()

	query = lowercaseFold(query)
	var result []*Template

	for _, t := range r.templates {
		if containsFold(t.Name, query) ||
			containsFold(t.Description, query) ||
			containsAnyFold(t.Tags, query) {
			result = append(result, t)
		}
	}
	return result
}

func (r *Registry) Update(id string, t *Template) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.templates[id]; !exists {
		return fmt.Errorf("template not found: %s", id)
	}

	t.ID = id
	t.UpdatedAt = time.Now()
	r.templates[id] = t
	return nil
}

func (r *Registry) Unregister(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.templates[id]; !exists {
		return fmt.Errorf("template not found: %s", id)
	}
	delete(r.templates, id)
	return nil
}

func (r *Registry) IncrementUsage(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	t, exists := r.templates[id]
	if !exists {
		return fmt.Errorf("template not found: %s", id)
	}
	t.UsageCount++
	return nil
}

func (r *Registry) SaveToFile(id string) error {
	r.mu.RLock()
	t, exists := r.templates[id]
	r.mu.RUnlock()

	if !exists {
		return fmt.Errorf("template not found: %s", id)
	}

	if r.dir == "" {
		return fmt.Errorf("directory not set")
	}

	data, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	path := filepath.Join(r.dir, id+".json")
	return os.WriteFile(path, data, 0644)
}

func (r *Registry) LoadFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read: %w", err)
	}

	var t Template
	if err := json.Unmarshal(data, &t); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	return r.Register(&t)
}

func (r *Registry) ExportAll(dir string) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}

	for id, t := range r.templates {
		data, err := json.MarshalIndent(t, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal %s: %w", id, err)
		}

		path := filepath.Join(dir, id+".json")
		if err := os.WriteFile(path, data, 0644); err != nil {
			return fmt.Errorf("write %s: %w", id, err)
		}
	}
	return nil
}

func (r *Registry) ImportDir(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read dir: %w", err)
	}

	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}

		path := filepath.Join(dir, e.Name())
		if err := r.LoadFromFile(path); err != nil {
			return fmt.Errorf("load %s: %w", e.Name(), err)
		}
	}
	return nil
}

func (r *Registry) Instantiate(id string, vars map[string]interface{}) (*nodes.Workflow, error) {
	t, err := r.Get(id)
	if err != nil {
		return nil, err
	}

	wf := nodes.NewWorkflow(id, t.Name, t.Description)

	for _, n := range t.Nodes {
		config := make(nodes.NodeConfig)
		for k, v := range n.Config {
			config[k] = v
		}

		for _, v := range t.Variables {
			if val, ok := vars[v.Name]; ok {
				config[v.Name] = val
			} else if v.Default != nil {
				config[v.Name] = v.Default
			}
		}

		n.Config = config
		wf.AddNode(n)
	}

	for _, e := range t.Edges {
		wf.AddEdge(e)
	}

	r.IncrementUsage(id)
	return wf, nil
}

func lowercaseFold(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		result[i] = c
	}
	return string(result)
}

func containsFold(s, substr string) bool {
	s = lowercaseFold(s)
	substr = lowercaseFold(substr)
	return len(s) >= len(substr) && (s == substr || containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func containsAnyFold(slice []string, substr string) bool {
	for _, s := range slice {
		if containsFold(s, substr) {
			return true
		}
	}
	return false
}

type TemplateExecutor struct {
	registry *Registry
}

func NewTemplateExecutor(reg *Registry) *TemplateExecutor {
	return &TemplateExecutor{registry: reg}
}

func (e *TemplateExecutor) InstantiateAndRun(id string, vars map[string]interface{}) error {
	wf, err := e.registry.Instantiate(id, vars)
	if err != nil {
		return err
	}
	_ = wf
	return nil
}
