package autonomy

import (
	"testing"
)

func TestErrorAnalysis(t *testing.T) {
	ea := NewErrorAnalyzer()
	pattern := ea.AnalyzePattern("null pointer")

	if pattern == "" {
		t.Error("Pattern should be detected")
	}
}

func TestAutoRetry(t *testing.T) {
	ar := NewAutoRetry()
	ar.SetMaxAttempts(3)

	for i := 1; i <= 3; i++ {
		if !ar.ShouldRetry(i) {
			t.Error("Should retry up to max attempts")
		}
	}
}

func TestSkillSelection(t *testing.T) {
	ss := NewSkillSelector()
	skill := ss.Select("fix bug")

	if skill == nil {
		t.Fatal("Skill should be selected")
	}
}

func TestGoalDecomposition(t *testing.T) {
	gd := NewGoalDecomposer()
	subtasks := gd.Decompose("implement login")

	if len(subtasks) == 0 {
		t.Error("Should decompose into subtasks")
	}
}

func TestContextPrioritization(t *testing.T) {
	cp := NewContextPrioritizer()
	priority := cp.Prioritize("important context")

	if priority <= 0 {
		t.Error("Context should have priority")
	}
}