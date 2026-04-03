package distill

import (
	"strings"
	"testing"
)

func TestClassifier_GitDiff(t *testing.T) {
	t.Parallel()

	c := NewClassifier()
	text := "diff --git a/file.go b/file.go\n--- a/file.go\n+++ b/file.go\n@@ -1 +1 @@"
	if ct := c.Classify(text); ct != ContentGitDiff {
		t.Errorf("expected GitDiff, got %v", ct)
	}
}

func TestClassifier_BuildOutput(t *testing.T) {
	t.Parallel()

	c := NewClassifier()
	text := "go build ./...\nerror: undefined reference to foo"
	if ct := c.Classify(text); ct != ContentBuildOutput {
		t.Errorf("expected BuildOutput, got %v", ct)
	}
}

func TestClassifier_TestOutput(t *testing.T) {
	t.Parallel()

	c := NewClassifier()
	text := "=== RUN   TestFoo\n--- PASS: TestFoo (0.00s)\nok\tgithub.com/example\t0.001s"
	if ct := c.Classify(text); ct != ContentTestOutput {
		t.Errorf("expected TestOutput, got %v", ct)
	}
}

func TestClassifier_LogOutput(t *testing.T) {
	t.Parallel()

	c := NewClassifier()
	text := "2024-01-01T00:00:00Z [INFO] server started\n2024-01-01T00:00:01Z [ERROR] connection lost"
	if ct := c.Classify(text); ct != ContentLogOutput {
		t.Errorf("expected LogOutput, got %v", ct)
	}
}

func TestClassifier_StructuredData(t *testing.T) {
	t.Parallel()

	c := NewClassifier()
	text := `{"key": "value", "count": 42}`
	if ct := c.Classify(text); ct != ContentStructuredData {
		t.Errorf("expected StructuredData, got %v", ct)
	}
}

func TestClassifier_Unknown(t *testing.T) {
	t.Parallel()

	c := NewClassifier()
	text := "hello world, this is random text"
	if ct := c.Classify(text); ct != ContentUnknown {
		t.Errorf("expected Unknown, got %v", ct)
	}
}

func TestScorer_BuildCritical(t *testing.T) {
	t.Parallel()

	s := NewScorer()
	text := "error: compilation failed"
	tier := s.Score(text, ContentBuildOutput)
	if tier != TierCritical {
		t.Errorf("expected Critical, got %v", tier)
	}
}

func TestScorer_BuildWarning(t *testing.T) {
	t.Parallel()

	s := NewScorer()
	text := "warning: unused variable"
	tier := s.Score(text, ContentBuildOutput)
	if tier != TierImportant {
		t.Errorf("expected Important, got %v", tier)
	}
}

func TestScorer_TestFail(t *testing.T) {
	t.Parallel()

	s := NewScorer()
	text := "--- FAIL: TestSomething"
	tier := s.Score(text, ContentTestOutput)
	if tier != TierCritical {
		t.Errorf("expected Critical, got %v", tier)
	}
}

func TestScorer_LogError(t *testing.T) {
	t.Parallel()

	s := NewScorer()
	text := "ERROR: database connection lost"
	tier := s.Score(text, ContentLogOutput)
	if tier != TierCritical {
		t.Errorf("expected Critical, got %v", tier)
	}
}

func TestScorer_LogWarn(t *testing.T) {
	t.Parallel()

	s := NewScorer()
	text := "WARN: disk space low"
	tier := s.Score(text, ContentLogOutput)
	if tier != TierImportant {
		t.Errorf("expected Important, got %v", tier)
	}
}

func TestCollapse_RepeatedLines(t *testing.T) {
	t.Parallel()

	c := NewCollapse()
	text := "line1\nline1\nline1\nline1\nline2"
	result := c.Compress(text)
	if !strings.Contains(result, "omitted") {
		t.Errorf("expected omission, got %s", result)
	}
}

func TestCollapse_BlankLines(t *testing.T) {
	t.Parallel()

	c := NewCollapse()
	text := "a\n\n\n\n\nb"
	result := c.Compress(text)
	lines := strings.Split(result, "\n")
	blankCount := 0
	for _, l := range lines {
		if strings.TrimSpace(l) == "" {
			blankCount++
		}
	}
	if blankCount > 1 {
		t.Errorf("expected at most 1 blank line, got %d", blankCount)
	}
}

func TestCollapse_NoRepetition(t *testing.T) {
	t.Parallel()

	c := NewCollapse()
	text := "unique1\nunique2\nunique3"
	result := c.Compress(text)
	if result != text {
		t.Errorf("expected no change, got %s", result)
	}
}

func TestComposer_Filter(t *testing.T) {
	t.Parallel()

	co := NewComposer()
	co.Threshold = TierContext
	lines := []string{"critical", "important", "context", "noise"}
	tiers := []SignalTier{TierCritical, TierImportant, TierContext, TierNoise}
	result, dropped := co.Compose(lines, tiers)
	if dropped != 1 {
		t.Errorf("expected 1 dropped, got %d", dropped)
	}
	if strings.Contains(result, "noise") {
		t.Errorf("noise should be filtered, got %s", result)
	}
}

func TestComposer_LengthMismatch(t *testing.T) {
	t.Parallel()

	co := NewComposer()
	lines := []string{"a", "b"}
	tiers := []SignalTier{TierCritical}
	result, dropped := co.Compose(lines, tiers)
	if dropped != 0 {
		t.Errorf("expected 0 dropped on length mismatch, got %d", dropped)
	}
	if result != "a\nb" {
		t.Errorf("expected all lines returned, got %s", result)
	}
}

func TestPipeline_Process(t *testing.T) {
	t.Parallel()

	p := NewPipeline()
	text := "=== RUN   TestFoo\n--- PASS: TestFoo\n--- PASS: TestBar\nok\tgithub.com/example"
	output, saved, original := p.Process(text)
	if original <= 0 {
		t.Error("expected positive original tokens")
	}
	if saved < 0 {
		t.Error("expected non-negative saved tokens")
	}
	if output == "" {
		t.Error("expected non-empty output")
	}
}

func TestEstimateTokens(t *testing.T) {
	t.Parallel()

	tokens := estimateTokens("hello world")
	if tokens <= 0 {
		t.Errorf("expected positive tokens, got %d", tokens)
	}
}

func TestClassifier_InfraOutput(t *testing.T) {
	t.Parallel()

	c := NewClassifier()
	text := "kubectl apply -f deployment.yaml"
	if ct := c.Classify(text); ct != ContentInfraOutput {
		t.Errorf("expected InfraOutput, got %v", ct)
	}
}

func TestClassifier_InfraOutput_Docker(t *testing.T) {
	t.Parallel()

	c := NewClassifier()
	text := "docker build -t myapp ."
	if ct := c.Classify(text); ct != ContentInfraOutput {
		t.Errorf("expected InfraOutput, got %v", ct)
	}
}

func TestClassifier_TabularData_Pipes(t *testing.T) {
	t.Parallel()

	c := NewClassifier()
	text := "| Name | Age |\n| Alice | 30 |\n| Bob | 25 |\n| Eve | 28 |"
	if ct := c.Classify(text); ct != ContentTabularData {
		t.Errorf("expected TabularData, got %v", ct)
	}
}

func TestClassifier_TabularData_Tabs(t *testing.T) {
	t.Parallel()

	c := NewClassifier()
	text := "Name\tAge\tCity\nAlice\t30\tNYC\nBob\t25\tLA\nEve\t28\tSF"
	if ct := c.Classify(text); ct != ContentTabularData {
		t.Errorf("expected TabularData, got %v", ct)
	}
}

func TestClassifier_Empty(t *testing.T) {
	t.Parallel()

	c := NewClassifier()
	if ct := c.Classify(""); ct != ContentUnknown {
		t.Errorf("expected Unknown for empty, got %v", ct)
	}
}

func TestScorer_GitDiffCritical(t *testing.T) {
	t.Parallel()

	s := NewScorer()
	text := "new file: main.go"
	tier := s.Score(text, ContentGitDiff)
	if tier != TierCritical {
		t.Errorf("expected Critical, got %v", tier)
	}
}

func TestScorer_GitDiffContext(t *testing.T) {
	t.Parallel()

	s := NewScorer()
	text := "@@ -1,5 +1,5 @@"
	tier := s.Score(text, ContentGitDiff)
	if tier != TierContext {
		t.Errorf("expected Context, got %v", tier)
	}
}

func TestScorer_BuildContext(t *testing.T) {
	t.Parallel()

	s := NewScorer()
	text := "compilation successful"
	tier := s.Score(text, ContentBuildOutput)
	if tier != TierContext {
		t.Errorf("expected Context, got %v", tier)
	}
}

func TestScorer_TestContext(t *testing.T) {
	t.Parallel()

	s := NewScorer()
	text := "running tests..."
	tier := s.Score(text, ContentTestOutput)
	if tier != TierContext {
		t.Errorf("expected Context, got %v", tier)
	}
}

func TestScorer_LogInfo(t *testing.T) {
	t.Parallel()

	s := NewScorer()
	text := "INFO: server started on :8080"
	tier := s.Score(text, ContentLogOutput)
	if tier != TierContext {
		t.Errorf("expected Context, got %v", tier)
	}
}

func TestScorer_LogNoise(t *testing.T) {
	t.Parallel()

	s := NewScorer()
	text := "TRACE: entering function foo"
	tier := s.Score(text, ContentLogOutput)
	if tier != TierNoise {
		t.Errorf("expected Noise, got %v", tier)
	}
}

func TestScorer_Unknown(t *testing.T) {
	t.Parallel()

	s := NewScorer()
	tier := s.Score("random text", ContentUnknown)
	if tier != TierNoise {
		t.Errorf("expected Noise, got %v", tier)
	}
}

func TestCollapse_Empty(t *testing.T) {
	t.Parallel()

	c := NewCollapse()
	result := c.Compress("")
	if result != "" {
		t.Errorf("expected empty, got %q", result)
	}
}

func TestCollapse_SingleLine(t *testing.T) {
	t.Parallel()

	c := NewCollapse()
	result := c.Compress("single line")
	if result != "single line" {
		t.Errorf("expected 'single line', got %q", result)
	}
}

func TestCollapse_ExactThreeRepeats(t *testing.T) {
	t.Parallel()

	c := NewCollapse()
	text := "line\nline\nline\nend"
	result := c.Compress(text)
	if !strings.Contains(result, "omitted") {
		t.Errorf("expected omission, got %s", result)
	}
}

func TestComposer_DefaultThreshold(t *testing.T) {
	t.Parallel()

	co := NewComposer()
	if co.Threshold != TierNoise {
		t.Errorf("expected TierNoise default, got %v", co.Threshold)
	}
	lines := []string{"noise"}
	tiers := []SignalTier{TierNoise}
	result, dropped := co.Compose(lines, tiers)
	if dropped != 0 {
		t.Errorf("expected 0 dropped (noise >= noise threshold), got %d", dropped)
	}
	if result != "noise" {
		t.Errorf("expected 'noise', got %q", result)
	}
}

func TestScorer_TabularData(t *testing.T) {
	t.Parallel()

	s := NewScorer()
	tier := s.Score("table data", ContentTabularData)
	if tier != TierContext {
		t.Errorf("expected Context, got %v", tier)
	}
}

func TestScorer_StructuredData(t *testing.T) {
	t.Parallel()

	s := NewScorer()
	tier := s.Score(`{"key":"value"}`, ContentStructuredData)
	if tier != TierImportant {
		t.Errorf("expected Important, got %v", tier)
	}
}

func TestScorer_InfraOutput(t *testing.T) {
	t.Parallel()

	s := NewScorer()
	tier := s.Score("kubectl apply", ContentInfraOutput)
	if tier != TierImportant {
		t.Errorf("expected Important, got %v", tier)
	}
}

func TestScorer_GitDiffDeleted(t *testing.T) {
	t.Parallel()

	s := NewScorer()
	tier := s.Score("deleted file: old.go", ContentGitDiff)
	if tier != TierCritical {
		t.Errorf("expected Critical, got %v", tier)
	}
}

func TestCollapse_TwoRepeats(t *testing.T) {
	t.Parallel()

	c := NewCollapse()
	text := "line\nline\nend"
	result := c.Compress(text)
	if strings.Contains(result, "omitted") {
		t.Errorf("expected no omission for 2 repeats, got %s", result)
	}
}

func TestCollapse_AllBlank(t *testing.T) {
	t.Parallel()

	c := NewCollapse()
	text := "\n\n\n\n"
	result := c.Compress(text)
	lines := strings.Split(result, "\n")
	blankCount := 0
	for _, l := range lines {
		if strings.TrimSpace(l) == "" {
			blankCount++
		}
	}
	if blankCount > 1 {
		t.Errorf("expected at most 1 blank line, got %d", blankCount)
	}
}

func TestClassifier_InfraOutput_NotInfra(t *testing.T) {
	t.Parallel()

	c := NewClassifier()
	text := "just some random text about docker without any action verbs"
	if ct := c.Classify(text); ct != ContentUnknown {
		t.Errorf("expected Unknown, got %v", ct)
	}
}

func TestClassifier_TabularData_NotTabular(t *testing.T) {
	t.Parallel()

	c := NewClassifier()
	text := "no pipes here\njust two lines"
	if ct := c.Classify(text); ct != ContentUnknown {
		t.Errorf("expected Unknown, got %v", ct)
	}
}

func TestClassifier_StructuredData_NotJSON(t *testing.T) {
	t.Parallel()

	c := NewClassifier()
	text := "{invalid json}"
	if ct := c.Classify(text); ct != ContentUnknown {
		t.Errorf("expected Unknown, got %v", ct)
	}
}

func TestScorer_GitDiffContextFallback(t *testing.T) {
	t.Parallel()

	s := NewScorer()
	text := "some random diff text without + or - or @@"
	tier := s.Score(text, ContentGitDiff)
	if tier != TierContext {
		t.Errorf("expected Context, got %v", tier)
	}
}
