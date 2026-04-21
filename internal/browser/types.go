// Package browser implements YAML workflow automation via Lightpanda CDP.
package browser

// Workflow defines a YAML-based browser automation workflow.
type Workflow struct {
	Name    string            `yaml:"name"`
	Env     map[string]string `yaml:"env,omitempty"`
	Actions map[string]Action `yaml:"actions"`
}

// Action is a named sequence of browser steps.
type Action struct {
	URL   string `yaml:"url,omitempty"`
	Steps []Step `yaml:"steps"`
}

// Step is a single browser operation.
type Step struct {
	Navigate   string     `yaml:"navigate,omitempty"`
	Click      *ClickStep `yaml:"click,omitempty"`
	Fill       *FillStep  `yaml:"fill,omitempty"`
	Wait       *WaitStep  `yaml:"wait,omitempty"`
	Eval       string     `yaml:"eval,omitempty"`
	Screenshot string     `yaml:"screenshot,omitempty"`
}

// ClickStep clicks an element matching the selector.
type ClickStep struct {
	Selector string `yaml:"selector"`
}

// FillStep fills an input field.
type FillStep struct {
	Selector string `yaml:"selector"`
	Value    string `yaml:"value"`
}

// WaitStep waits for an element or a duration.
type WaitStep struct {
	Selector string `yaml:"selector,omitempty"`
	Timeout  string `yaml:"timeout,omitempty"`
}

// InspectResult holds the discovery output from inspecting a page.
type InspectResult struct {
	URL       string            `json:"url"`
	Title     string            `json:"title"`
	Elements  []ElementInfo     `json:"elements"`
}

// ElementInfo describes a single interactive element found on a page.
type ElementInfo struct {
	Selector    string `json:"selector"`
	Tag         string `json:"tag"`
	Type        string `json:"type,omitempty"`
	Name        string `json:"name,omitempty"`
	Placeholder string `json:"placeholder,omitempty"`
	Text        string `json:"text,omitempty"`
	Href        string `json:"href,omitempty"`
	Role        string `json:"role,omitempty"`
	ID          string `json:"id,omitempty"`
}

// ActionResult is returned after executing a workflow action.
type ActionResult struct {
	Action      string   `json:"action"`
	Success     bool     `json:"success"`
	Error       string   `json:"error,omitempty"`
	StepsRun    int      `json:"steps_run"`
	Screenshots []string `json:"screenshots,omitempty"`
}
