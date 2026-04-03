package healing

import (
	"context"
	"errors"
	"fmt"
	"testing"
)

func TestErrorDetector_Detect(t *testing.T) {
	t.Parallel()

	d := &ErrorDetector{}
	et, _ := d.Detect(context.DeadlineExceeded)
	if et != ToolExecutionError {
		t.Errorf("expected ToolExecutionError for deadline exceeded, got %v", et)
	}
	et, _ = d.Detect(nil)
	if et != ToolExecutionError {
		t.Errorf("expected ToolExecutionError for nil, got %v", et)
	}
	et, _ = d.Detect(fmt.Errorf("timeout exceeded"))
	if et != TimeoutError {
		t.Errorf("expected TimeoutError, got %v", et)
	}
}

func TestErrorType_String(t *testing.T) {
	t.Parallel()

	if ToolExecutionError.String() != "ToolExecutionError" {
		t.Errorf("unexpected: %s", ToolExecutionError.String())
	}
	if TimeoutError.String() != "TimeoutError" {
		t.Errorf("unexpected: %s", TimeoutError.String())
	}
	if ErrorType(99).String() != "UnknownError" {
		t.Errorf("unexpected: %s", ErrorType(99).String())
	}
}

func TestErrorAnalyzer_Analyze(t *testing.T) {
	t.Parallel()

	a := &ErrorAnalyzer{Detector: &ErrorDetector{}}
	analysis, err := a.Analyze(errors.New("test error"))
	if err != nil {
		t.Fatal(err)
	}
	if analysis == nil {
		t.Fatal("expected analysis")
	}
	if analysis.RootCause == "" {
		t.Error("expected root cause")
	}
}

func TestErrorAnalyzer_NilDetector(t *testing.T) {
	t.Parallel()

	a := &ErrorAnalyzer{}
	analysis, err := a.Analyze(errors.New("test"))
	if err != nil {
		t.Fatal(err)
	}
	if analysis == nil {
		t.Fatal("expected analysis")
	}
}

func TestRetryManager_ShouldRetry(t *testing.T) {
	t.Parallel()

	rm := NewDefaultRetryManager()
	if !rm.ShouldRetry(NetworkError, 1) {
		t.Error("expected retry for network error")
	}
	if rm.ShouldRetry(NetworkError, 6) {
		t.Error("expected no retry after max")
	}
	if rm.ShouldRetry(PermissionError, 2) {
		t.Error("expected no retry for permission")
	}
}

func TestRetryManager_GetDelay(t *testing.T) {
	t.Parallel()

	rm := NewDefaultRetryManager()
	d1 := rm.GetDelay(NetworkError, 1)
	d2 := rm.GetDelay(NetworkError, 2)
	if d2 <= d1 {
		t.Error("expected increasing delay")
	}
}

func TestRetryManager_GetDelay_Invalid(t *testing.T) {
	t.Parallel()

	rm := NewDefaultRetryManager()
	d := rm.GetDelay(ErrorType(99), 1)
	if d <= 0 {
		t.Error("expected positive delay")
	}
}

func TestGetRecoveryActions(t *testing.T) {
	t.Parallel()

	actions := GetRecoveryActions(TimeoutError)
	if len(actions) == 0 {
		t.Error("expected actions")
	}
	actions = GetRecoveryActions(ErrorType(99))
	if len(actions) == 0 {
		t.Error("expected default actions")
	}
}

func TestAutoFix(t *testing.T) {
	t.Parallel()

	tests := []struct {
		et ErrorType
	}{
		{TimeoutError},
		{RateLimitError},
		{NetworkError},
		{ToolExecutionError},
		{SyntaxError},
		{PermissionError},
		{ResourceError},
		{RuntimeError},
	}
	for _, tt := range tests {
		t.Run(tt.et.String(), func(t *testing.T) {
			fix, err := AutoFix(tt.et, nil)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if fix == "" {
				t.Error("expected fix applied")
			}
		})
	}
	_, err := AutoFix(ErrorType(99), nil)
	if err == nil {
		t.Error("expected error for unknown type")
	}
}

func TestAutoFix_ToolExecutionWithNotFound(t *testing.T) {
	t.Parallel()

	fix, err := AutoFix(ToolExecutionError, map[string]string{"error": "tool not found"})
	if err != nil {
		t.Fatal(err)
	}
	if fix == "" {
		t.Error("expected fix applied")
	}
}

func TestHealingLog(t *testing.T) {
	t.Parallel()

	log := NewHealingLog()
	log.Log(HealingAttempt{ErrorType: "test", Success: true})
	log.Log(HealingAttempt{ErrorType: "test", Success: false})

	if len(log.GetAttempts()) != 2 {
		t.Errorf("expected 2 attempts, got %d", len(log.GetAttempts()))
	}
	if len(log.GetRecent(1)) != 1 {
		t.Error("expected 1 recent")
	}
	if len(log.GetRecent(100)) != 2 {
		t.Error("expected 2 recent when n > total")
	}
}

func TestHealingStats(t *testing.T) {
	t.Parallel()

	s := NewHealingStats()
	s.RecordAttempt(HealingAttempt{ErrorType: "timeout", Success: true})
	s.RecordAttempt(HealingAttempt{ErrorType: "timeout", Success: false})
	s.RecordAttempt(HealingAttempt{ErrorType: "network", Success: true})

	if s.Attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", s.Attempts)
	}
	if s.Successes != 2 {
		t.Errorf("expected 2 successes, got %d", s.Successes)
	}
	if s.Failures != 1 {
		t.Errorf("expected 1 failure, got %d", s.Failures)
	}
	if s.SuccessRate() != 66.66666666666666 {
		t.Errorf("expected ~66.7%%, got %.1f", s.SuccessRate())
	}
	if s.SuccessRateByType("timeout") != 50 {
		t.Errorf("expected 50%%, got %.1f", s.SuccessRateByType("timeout"))
	}
	if s.SuccessRateByType("unknown") != 0 {
		t.Error("expected 0%")
	}
	if s.Attempts == 0 {
		t.Error("expected positive avg duration")
	}
}

func TestHealingStats_Empty(t *testing.T) {
	t.Parallel()

	s := NewHealingStats()
	if s.SuccessRate() != 0 {
		t.Error("expected 0%")
	}
	if s.AvgDuration() != 0 {
		t.Error("expected 0 duration")
	}
}

func TestHealingStats_Summary(t *testing.T) {
	t.Parallel()

	s := NewHealingStats()
	s.RecordAttempt(HealingAttempt{ErrorType: "test", Success: true})
	summary := s.Summary()
	if summary == "" {
		t.Error("expected summary")
	}
}

func TestHealingLog_Report(t *testing.T) {
	t.Parallel()

	log := NewHealingLog()
	stats := NewHealingStats()
	stats.RecordAttempt(HealingAttempt{ErrorType: "timeout", Success: true})
	report := log.Report(stats)
	if report == "" {
		t.Error("expected report")
	}
}

func TestTracebackParser_Python(t *testing.T) {
	t.Parallel()

	p := NewTracebackParser()
	traceback := `Traceback (most recent call last):
  File "test.py", line 10, in main
    result = func()
ValueError: invalid literal`

	result := p.Parse(traceback)
	if result.Language != "python" {
		t.Errorf("expected python, got %s", result.Language)
	}
	if len(result.Frames) == 0 {
		t.Error("expected frames")
	}
}

func TestTracebackParser_Go(t *testing.T) {
	t.Parallel()

	p := NewTracebackParser()
	traceback := `goroutine 1 [running]:
main.main()
	/test.go:15 +0x1
panic: runtime error`

	result := p.Parse(traceback)
	if result.Language != "go" {
		t.Errorf("expected go, got %s", result.Language)
	}
}

func TestTracebackParser_JavaScript(t *testing.T) {
	t.Parallel()

	p := NewTracebackParser()
	traceback := `TypeError: Cannot read property 'x' of undefined
    at Object.test (/app/test.js:10:5)
    at Module._compile (internal/modules/cjs/loader.js:100:10)`

	result := p.Parse(traceback)
	if result.Language != "javascript" {
		t.Errorf("expected javascript, got %s", result.Language)
	}
}

func TestTracebackParser_Rust(t *testing.T) {
	t.Parallel()

	p := NewTracebackParser()
	traceback := `thread 'main' panicked at 'index out of bounds', src/main.rs:5:10
note: run with RUST_BACKTRACE=1`

	result := p.Parse(traceback)
	if result.Language != "rust" {
		t.Errorf("expected rust, got %s", result.Language)
	}
}

func TestTracebackParser_Java(t *testing.T) {
	t.Parallel()

	p := NewTracebackParser()
	traceback := `Exception in thread "main" java.lang.NullPointerException
	at com.example.App.main(App.java:15)`

	result := p.Parse(traceback)
	if result.Language != "java" {
		t.Errorf("expected java, got %s", result.Language)
	}
}

func TestTracebackParser_Unknown(t *testing.T) {
	t.Parallel()

	p := NewTracebackParser()
	result := p.Parse("some random error message")
	if result.Language != "unknown" {
		t.Errorf("expected unknown, got %s", result.Language)
	}
}

func TestRootCauseAnalyzer_IdentifyCauseType(t *testing.T) {
	t.Parallel()

	a := &RootCauseAnalyzer{}
	tests := []struct {
		input    string
		expected string
	}{
		{"import not found", "MissingImport"},
		{"type mismatch error", "TypeMismatch"},
		{"null pointer exception", "NullPointer"},
		{"index out of range", "IndexOutOfRange"},
		{"permission denied", "PermissionDenied"},
		{"timeout exceeded", "Timeout"},
		{"connection refused", "ConnectionFailed"},
		{"syntax error", "SyntaxError"},
		{"random error", "Unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := a.IdentifyCauseType(tt.input)
			if got != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, got)
			}
		})
	}
}

func TestRootCauseAnalyzer_SuggestFix(t *testing.T) {
	t.Parallel()

	a := &RootCauseAnalyzer{}
	fix := a.SuggestFix("MissingImport", "")
	if fix == "" {
		t.Error("expected fix suggestion")
	}
	fix = a.SuggestFix("UnknownType", "")
	if fix == "" {
		t.Error("expected default fix suggestion")
	}
}

func TestRootCauseAnalyzer_AnalyzeHeuristic(t *testing.T) {
	t.Parallel()

	a := &RootCauseAnalyzer{}
	result, err := a.AnalyzeRootCause("import foo not found")
	if err != nil {
		t.Fatal(err)
	}
	if result.Type != "MissingImport" {
		t.Errorf("expected MissingImport, got %s", result.Type)
	}
}

func TestHealingLoop_Execute_Success(t *testing.T) {
	t.Parallel()

	h := NewHealingLoop()
	err := h.Execute(context.Background(), func() error { return nil })
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestHealingLoop_Execute_Cancel(t *testing.T) {
	t.Parallel()

	h := NewHealingLoop()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := h.Execute(ctx, func() error { return errors.New("test") })
	if err == nil {
		t.Error("expected error")
	}
}

func TestHealingLoop_GetReport(t *testing.T) {
	t.Parallel()

	h := NewHealingLoop()
	report := h.GetReport()
	if report == "" {
		t.Error("expected report")
	}
}

func TestHealingLoop_Execute_Retry(t *testing.T) {
	t.Parallel()

	h := NewHealingLoop()
	count := 0
	err := h.Execute(context.Background(), func() error {
		count++
		if count < 2 {
			return errors.New("temporary error")
		}
		return nil
	})
	if err != nil {
		t.Errorf("expected nil after retry, got %v", err)
	}
}
