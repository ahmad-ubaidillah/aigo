package planning

import (
	"regexp"
	"strings"
)

// Decomposer splits tasks into subtasks.
type Decomposer struct {
	maxDepth int
}

// NewDecomposer creates a new Decomposer with default settings.
func NewDecomposer() *Decomposer {
	return &Decomposer{
		maxDepth: 3,
	}
}

// SetMaxDepth sets the maximum decomposition depth.
func (d *Decomposer) SetMaxDepth(depth int) {
	if depth > 0 {
		d.maxDepth = depth
	}
}

// Decompose splits a task into a plan with subtasks.
func (d *Decomposer) Decompose(task string) (*Plan, error) {
	plan := NewPlan(task)
	subtasks := d.splitTask(task, 0)
	
	for i, subtask := range subtasks {
		step := Step{
			ID:          generateStepID(i + 1),
			Description: subtask.description,
			Tool:        subtask.suggestedTool,
			Status:      StatusPending,
			DependsOn:   subtask.dependencies,
		}
		plan.AddStep(step)
	}
	
	// Build dependencies
	plan.Dependencies = d.detectDependencies(subtasks)
	
	return plan, nil
}

// subtask represents a decomposed subtask.
type subtask struct {
	description    string
	suggestedTool  string
	dependencies   []string
	isParallel     bool
}

// splitTask recursively splits a task into subtasks.
func (d *Decomposer) splitTask(task string, depth int) []subtask {
	if depth >= d.maxDepth || d.IsAtomic(task) {
		return []subtask{{
			description:   task,
			suggestedTool: d.inferTool(task),
		}}
	}
	
	var subtasks []subtask
	
	// Try different splitting strategies
	if parts := d.splitByConnectives(task); len(parts) > 1 {
		for i, part := range parts {
			children := d.splitTask(part, depth+1)
			for j := range children {
				if i > 0 {
					children[j].dependencies = append(children[j].dependencies, generateStepID(i))
				}
			}
			subtasks = append(subtasks, children...)
		}
		return subtasks
	}
	
	// Try workflow-based decomposition
	if workflow := d.detectWorkflow(task); workflow != nil {
		return workflow
	}
	
	// Fallback: single task
	return []subtask{{
		description:   task,
		suggestedTool: d.inferTool(task),
	}}
}

// IsAtomic returns true if a task cannot be further decomposed.
func (d *Decomposer) IsAtomic(task string) bool {
	task = strings.ToLower(strings.TrimSpace(task))
	
	// Too short to decompose
	if len(task) < 20 {
		return true
	}
	
	// No connective words
	connectives := []string{" and ", " then ", " after ", " while ", " before ", " also "}
	for _, c := range connectives {
		if strings.Contains(task, c) {
			return false
		}
	}
	
	// Check for workflow patterns
	workflowPatterns := []string{
		"fix", "implement", "create", "build", "refactor",
		"update", "delete", "add", "remove", "test",
	}
	for _, p := range workflowPatterns {
		if strings.Contains(task, p) {
			return false
		}
	}
	
	return true
}

// splitByConnectives splits a task by connecting words.
func (d *Decomposer) splitByConnectives(task string) []string {
	// Match patterns like "do X and Y" or "do X, then Y"
	re := regexp.MustCompile(`(?i)\s+(?:and|then|after that|also)\s+`)
	parts := re.Split(task, -1)
	
	if len(parts) > 1 {
		return parts
	}
	
	// Try comma separation
	if strings.Contains(task, ", ") {
		return strings.Split(task, ", ")
	}
	
	return nil
}

// detectWorkflow returns subtasks for common workflow patterns.
func (d *Decomposer) detectWorkflow(task string) []subtask {
	taskLower := strings.ToLower(task)
	
	// Bug fix workflow
	if strings.Contains(taskLower, "fix") || strings.Contains(taskLower, "bug") {
		return []subtask{
			{description: "Analyze the bug and identify root cause", suggestedTool: "bash"},
			{description: "Implement the fix", suggestedTool: "edit", dependencies: []string{"step-1"}},
			{description: "Test the fix", suggestedTool: "bash", dependencies: []string{"step-2"}},
			{description: "Verify no regressions", suggestedTool: "bash", dependencies: []string{"step-3"}},
		}
	}
	
	// Feature implementation workflow
	if strings.Contains(taskLower, "implement") || strings.Contains(taskLower, "create") || strings.Contains(taskLower, "add") {
		return []subtask{
			{description: "Analyze requirements and plan implementation", suggestedTool: "bash"},
			{description: "Implement the feature", suggestedTool: "write", dependencies: []string{"step-1"}},
			{description: "Write tests for the feature", suggestedTool: "write", dependencies: []string{"step-2"}},
			{description: "Run tests and verify", suggestedTool: "bash", dependencies: []string{"step-3"}},
		}
	}
	
	// Refactor workflow
	if strings.Contains(taskLower, "refactor") || strings.Contains(taskLower, "restructure") {
		return []subtask{
			{description: "Analyze current structure", suggestedTool: "bash"},
			{description: "Plan refactoring approach", suggestedTool: "bash", dependencies: []string{"step-1"}},
			{description: "Apply refactoring changes", suggestedTool: "edit", dependencies: []string{"step-2"}},
			{description: "Run tests to verify", suggestedTool: "bash", dependencies: []string{"step-3"}},
		}
	}
	
	return nil
}

// inferTool suggests a tool based on task description.
func (d *Decomposer) inferTool(task string) string {
	taskLower := strings.ToLower(task)
	
	toolPatterns := map[string]string{
		"read":    "read file|view file|show file|cat|display",
		"write":   "create file|write file|new file|save",
		"edit":    "modify|update|change|edit|fix|patch",
		"bash":    "run|execute|command|shell|build|test|git|npm|go ",
		"glob":    "find files|search files|list files",
		"grep":    "search|find in|grep|look for",
		"webfetch": "fetch|download|get url|http",
	}
	
	for tool, pattern := range toolPatterns {
		matched, _ := regexp.MatchString(pattern, taskLower)
		if matched {
			return tool
		}
	}
	
	return "bash"
}

// detectDependencies builds a dependency map from subtasks.
func (d *Decomposer) detectDependencies(subtasks []subtask) map[string][]string {
	deps := make(map[string][]string)
	for i, st := range subtasks {
		stepID := generateStepID(i + 1)
		if len(st.dependencies) > 0 {
			deps[stepID] = st.dependencies
		}
	}
	return deps
}

func generateStepID(num int) string {
	return strings.TrimPrefix(generateID(), "2") // Simplified ID
}
