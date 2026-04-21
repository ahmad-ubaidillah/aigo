package autonomy

type ErrorAnalyzer struct {
	patterns map[string]string
}

func NewErrorAnalyzer() *ErrorAnalyzer {
	return &ErrorAnalyzer{
		patterns: map[string]string{
			"null pointer":     "check for nil before use",
			"undefined":       "check import statements",
			"type mismatch":   "check type assertions",
			"index out of bounds": "check array length",
			"deadlock":        "check goroutine synchronization",
		},
	}
}

func (ea *ErrorAnalyzer) AnalyzePattern(errorMsg string) string {
	for pattern, fix := range ea.patterns {
		if Contains(errorMsg, pattern) {
			return fix
		}
	}
	return "manual review required"
}

func Contains(s, substr string) bool {
	return len(s) >= len(substr) && 
		(s == substr || 
		 (len(s) > len(substr) && 
		  (s[:len(substr)] == substr || 
		   containsAt(s, substr))))
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

type AutoRetry struct {
	maxAttempts int
	backoffMs   int
}

func NewAutoRetry() *AutoRetry {
	return &AutoRetry{
		maxAttempts: 3,
		backoffMs:   100,
	}
}

func (ar *AutoRetry) SetMaxAttempts(n int) {
	ar.maxAttempts = n
}

func (ar *AutoRetry) ShouldRetry(attempt int) bool {
	return attempt <= ar.maxAttempts
}

func (ar *AutoRetry) GetBackoff(attempt int) int {
	return ar.backoffMs * attempt
}

type SkillSelector struct {
	skills map[string][]string
}

func NewSkillSelector() *SkillSelector {
	return &SkillSelector{
		skills: map[string][]string{
			"fix":    {"debug", "code-review"},
			"test":   {"testgen", "debug"},
			"build":  {"docker", "deploy"},
			"docs":   {"readme", "apigen"},
		},
	}
}

func (ss *SkillSelector) Select(task string) []string {
	taskLower := ToLower(task)
	for key, skills := range ss.skills {
		if Contains(taskLower, key) {
			return skills
		}
	}
	return []string{"default"}
}

func ToLower(s string) string {
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

type GoalDecomposer struct{}

func NewGoalDecomposer() *GoalDecomposer {
	return &GoalDecomposer{}
}

func (gd *GoalDecomposer) Decompose(goal string) []string {
	subtasks := make([]string, 0)
	
	subtasks = append(subtasks, "analyze: "+goal)
	subtasks = append(subtasks, "plan: "+goal)
	subtasks = append(subtasks, "execute: "+goal)
	subtasks = append(subtasks, "verify: "+goal)
	
	return subtasks
}

func (gd *GoalDecomposer) EstimateEffort(goal string) int {
	return len(gd.Decompose(goal)) * 10
}

type ContextPrioritizer struct{}

func NewContextPrioritizer() *ContextPrioritizer {
	return &ContextPrioritizer{}
}

func (cp *ContextPrioritizer) Prioritize(context string) int {
	highPriority := []string{"error", "bug", "fix", "important", "critical"}
	mediumPriority := []string{"feature", "add", "new", "update"}
	
	for _, kw := range highPriority {
		if Contains(ToLower(context), kw) {
			return 10
		}
	}
	for _, kw := range mediumPriority {
		if Contains(ToLower(context), kw) {
			return 5
		}
	}
	return 1
}