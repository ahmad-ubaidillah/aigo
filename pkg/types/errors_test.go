package types

import (
	"errors"
	"testing"
	"time"
)

func TestError_Error(t *testing.T) {
	t.Parallel()
	err := NewError(1, "test")
	if err.Error() != "[1] test" {
		t.Errorf("unexpected error: %s", err.Error())
	}
}

func TestError_ErrorWithDetails(t *testing.T) {
	t.Parallel()
	err := NewError(1, "test").WithDetails("details")
	if err.Error() != "[1] test: details" {
		t.Errorf("unexpected error: %s", err.Error())
	}
}

func TestError_Unwrap(t *testing.T) {
	t.Parallel()
	inner := errors.New("inner")
	err := NewError(1, "test").WithCause(inner)
	if err.Unwrap() != inner {
		t.Error("expected inner error")
	}
}

func TestError_WithCause(t *testing.T) {
	t.Parallel()
	inner := errors.New("inner")
	err := NewError(1, "test").WithCause(inner)
	if err.Details != "inner" {
		t.Errorf("expected 'inner', got %s", err.Details)
	}
}

func TestError_SetRetryable(t *testing.T) {
	t.Parallel()
	err := NewError(1, "test").SetRetryable()
	if !err.Retryable {
		t.Error("expected retryable")
	}
}

func TestErrNotFound(t *testing.T) {
	t.Parallel()
	err := ErrNotFound("user")
	if err.Code != ErrCodeNotFound {
		t.Errorf("expected ErrCodeNotFound, got %d", err.Code)
	}
	if err.Details != "user" {
		t.Errorf("expected 'user', got %s", err.Details)
	}
}

func TestErrInvalidInput(t *testing.T) {
	t.Parallel()
	err := ErrInvalidInput("name", "too short")
	if err.Code != ErrCodeInvalidInput {
		t.Errorf("expected ErrCodeInvalidInput, got %d", err.Code)
	}
	if err.Details != "name: too short" {
		t.Errorf("expected 'name: too short', got %s", err.Details)
	}
}

func TestErrTimeout(t *testing.T) {
	t.Parallel()
	err := ErrTimeout("query")
	if err.Code != ErrCodeTimeout {
		t.Errorf("expected ErrCodeTimeout, got %d", err.Code)
	}
	if !err.Retryable {
		t.Error("expected retryable")
	}
}

func TestErrExternal(t *testing.T) {
	t.Parallel()
	inner := errors.New("connection refused")
	err := ErrExternal("api", inner)
	if err.Code != ErrCodeExternal {
		t.Errorf("expected ErrCodeExternal, got %d", err.Code)
	}
	if !err.Retryable {
		t.Error("expected retryable")
	}
}

func TestErrConfig(t *testing.T) {
	t.Parallel()
	inner := errors.New("missing key")
	err := ErrConfig("api_key", inner)
	if err.Code != ErrCodeConfig {
		t.Errorf("expected ErrCodeConfig, got %d", err.Code)
	}
}

func TestErrInternal(t *testing.T) {
	t.Parallel()
	inner := errors.New("panic")
	err := ErrInternal("process", inner)
	if err.Code != ErrCodeInternal {
		t.Errorf("expected ErrCodeInternal, got %d", err.Code)
	}
}

func TestProgressState(t *testing.T) {
	t.Parallel()
	ps := NewProgressState("test task")
	if ps.State != ProgressStatePending {
		t.Errorf("expected pending, got %s", ps.State)
	}
	if ps.Task != "test task" {
		t.Errorf("expected 'test task', got %s", ps.Task)
	}
}

func TestProgressState_Transitions(t *testing.T) {
	t.Parallel()
	ps := NewProgressState("test")
	ps.Thinking("analyzing")
	if ps.State != ProgressStateThinking {
		t.Errorf("expected thinking, got %s", ps.State)
	}
	ps.Executing("running")
	if ps.State != ProgressStateExecuting {
		t.Errorf("expected executing, got %s", ps.State)
	}
	if ps.Percent != 50 {
		t.Errorf("expected 50%%, got %d", ps.Percent)
	}
	ps.Waiting("waiting for response")
	if ps.State != ProgressStateWaiting {
		t.Errorf("expected waiting, got %s", ps.State)
	}
	ps.Complete()
	if ps.State != ProgressStateCompleted {
		t.Errorf("expected completed, got %s", ps.State)
	}
	if ps.Percent != 100 {
		t.Errorf("expected 100%%, got %d", ps.Percent)
	}
}

func TestProgressState_Fail(t *testing.T) {
	t.Parallel()
	ps := NewProgressState("test")
	ps.Fail("error occurred")
	if ps.State != ProgressStateFailed {
		t.Errorf("expected failed, got %s", ps.State)
	}
	if ps.Message != "error occurred" {
		t.Errorf("expected 'error occurred', got %s", ps.Message)
	}
}

func TestProgressState_Duration(t *testing.T) {
	t.Parallel()
	ps := NewProgressState("test")
	ps.Complete()
	d := ps.Duration()
	if d < 0 {
		t.Error("expected positive duration")
	}
}

func TestProgressState_DurationIncomplete(t *testing.T) {
	t.Parallel()
	ps := NewProgressState("test")
	d := ps.Duration()
	if d < 0 {
		t.Error("expected positive duration")
	}
}

func TestErrorConstants(t *testing.T) {
	t.Parallel()
	if ErrCodeOK != 0 {
		t.Errorf("expected 0, got %d", ErrCodeOK)
	}
	if ErrCodeUnknown != 1 {
		t.Errorf("expected 1, got %d", ErrCodeUnknown)
	}
	if ErrCodeNotFound != 2 {
		t.Errorf("expected 2, got %d", ErrCodeNotFound)
	}
	if ErrCodeInvalidInput != 3 {
		t.Errorf("expected 3, got %d", ErrCodeInvalidInput)
	}
	if ErrCodeTimeout != 4 {
		t.Errorf("expected 4, got %d", ErrCodeTimeout)
	}
	if ErrCodeUnauthorized != 5 {
		t.Errorf("expected 5, got %d", ErrCodeUnauthorized)
	}
	if ErrCodeRateLimit != 6 {
		t.Errorf("expected 6, got %d", ErrCodeRateLimit)
	}
	if ErrCodeExternal != 7 {
		t.Errorf("expected 7, got %d", ErrCodeExternal)
	}
	if ErrCodeConfig != 8 {
		t.Errorf("expected 8, got %d", ErrCodeConfig)
	}
	if ErrCodeInternal != 9 {
		t.Errorf("expected 9, got %d", ErrCodeInternal)
	}
}

func TestProgressStateConstants(t *testing.T) {
	t.Parallel()
	if ProgressStatePending != "pending" {
		t.Errorf("expected pending, got %s", ProgressStatePending)
	}
	if ProgressStateThinking != "thinking" {
		t.Errorf("expected thinking, got %s", ProgressStateThinking)
	}
	if ProgressStateExecuting != "executing" {
		t.Errorf("expected executing, got %s", ProgressStateExecuting)
	}
	if ProgressStateWaiting != "waiting" {
		t.Errorf("expected waiting, got %s", ProgressStateWaiting)
	}
	if ProgressStateCompleted != "completed" {
		t.Errorf("expected completed, got %s", ProgressStateCompleted)
	}
	if ProgressStateFailed != "failed" {
		t.Errorf("expected failed, got %s", ProgressStateFailed)
	}
}

func TestError_Is(t *testing.T) {
	t.Parallel()
	err := NewError(1, "test")
	target := NewError(1, "other")
	if err.Code != target.Code {
		t.Error("expected Is to match same code")
	}
}

func TestError_IsDifferentCode(t *testing.T) {
	t.Parallel()
	err := NewError(1, "test")
	target := NewError(2, "test")
	if err.Code == target.Code {
		t.Error("expected Is to not match different codes")
	}
}

func TestProgressState_CompletedAt(t *testing.T) {
	t.Parallel()
	ps := NewProgressState("test")
	if !ps.CompletedAt.IsZero() {
		t.Error("expected zero CompletedAt")
	}
	ps.Complete()
	if ps.CompletedAt.IsZero() {
		t.Error("expected non-zero CompletedAt")
	}
}

func TestProgressState_FailCompletedAt(t *testing.T) {
	t.Parallel()
	ps := NewProgressState("test")
	ps.Fail("error")
	if ps.CompletedAt.IsZero() {
		t.Error("expected non-zero CompletedAt after fail")
	}
}

func TestError_NilCause(t *testing.T) {
	t.Parallel()
	err := NewError(1, "test").WithCause(nil)
	if err.Details != "" {
		t.Errorf("expected empty details, got %s", err.Details)
	}
}

func TestError_Timestamp(t *testing.T) {
	t.Parallel()
	err := NewError(1, "test")
	if err.Timestamp.IsZero() {
		t.Error("expected non-zero timestamp")
	}
}

func TestProgressState_InitialState(t *testing.T) {
	t.Parallel()
	ps := NewProgressState("test")
	if ps.Percent != 0 {
		t.Errorf("expected 0%%, got %d", ps.Percent)
	}
	if ps.StartedAt.IsZero() {
		t.Error("expected non-zero StartedAt")
	}
}

func TestError_ErrorWithNilCause(t *testing.T) {
	t.Parallel()
	err := NewError(1, "test").WithCause(nil)
	if err.Error() != "[1] test" {
		t.Errorf("expected '[1] test', got %s", err.Error())
	}
}

func TestError_ErrorWithCause(t *testing.T) {
	t.Parallel()
	err := NewError(1, "test").WithCause(errors.New("cause"))
	if err.Error() != "[1] test: cause" {
		t.Errorf("expected '[1] test: cause', got %s", err.Error())
	}
}

func TestError_UnwrapNil(t *testing.T) {
	t.Parallel()
	err := NewError(1, "test")
	if err.Unwrap() != nil {
		t.Error("expected nil")
	}
}

func TestErrNotFound_Is(t *testing.T) {
	t.Parallel()
	err := ErrNotFound("user")
	target := NewError(ErrCodeNotFound, "other")
	if err.Code != target.Code {
		t.Error("expected Is to match")
	}
}

func TestErrTimeout_Is(t *testing.T) {
	t.Parallel()
	err := ErrTimeout("query")
	target := NewError(ErrCodeTimeout, "other")
	if err.Code != target.Code {
		t.Error("expected Is to match")
	}
}

func TestErrExternal_Is(t *testing.T) {
	t.Parallel()
	err := ErrExternal("api", nil)
	target := NewError(ErrCodeExternal, "other")
	if err.Code != target.Code {
		t.Error("expected Is to match")
	}
}

func TestErrConfig_Is(t *testing.T) {
	t.Parallel()
	err := ErrConfig("key", nil)
	target := NewError(ErrCodeConfig, "other")
	if err.Code != target.Code {
		t.Error("expected Is to match")
	}
}

func TestErrInternal_Is(t *testing.T) {
	t.Parallel()
	err := ErrInternal("op", nil)
	target := NewError(ErrCodeInternal, "other")
	if err.Code != target.Code {
		t.Error("expected Is to match")
	}
}

func TestErrInvalidInput_Is(t *testing.T) {
	t.Parallel()
	err := ErrInvalidInput("field", "reason")
	target := NewError(ErrCodeInvalidInput, "other")
	if err.Code != target.Code {
		t.Error("expected Is to match")
	}
}

func TestProgressState_DurationAfterFail(t *testing.T) {
	t.Parallel()
	ps := NewProgressState("test")
	time.Sleep(10 * time.Millisecond)
	ps.Fail("error")
	d := ps.Duration()
	if d < 100*time.Nanosecond {
		t.Errorf("expected duration >= 1ms, got %v", d)
	}
}

func TestProgressState_DurationAfterComplete(t *testing.T) {
	t.Parallel()
	ps := NewProgressState("test")
	ps.Complete()
	d := ps.Duration()
	if d < 0 {
		t.Errorf("expected positive duration, got %v", d)
	}
}

func TestError_WithDetailsChain(t *testing.T) {
	t.Parallel()
	err := NewError(1, "test").WithDetails("detail1").WithDetails("detail2")
	if err.Details != "detail2" {
		t.Errorf("expected 'detail2', got %s", err.Details)
	}
}

func TestError_WithCauseChain(t *testing.T) {
	t.Parallel()
	err := NewError(1, "test").WithCause(errors.New("cause1")).WithCause(errors.New("cause2"))
	if err.Cause.Error() != "cause2" {
		t.Errorf("expected 'cause2', got %s", err.Cause.Error())
	}
}

func TestError_SetRetryableIdempotent(t *testing.T) {
	t.Parallel()
	err := NewError(1, "test").SetRetryable().SetRetryable()
	if !err.Retryable {
		t.Error("expected retryable")
	}
}

func TestProgressState_MultipleTransitions(t *testing.T) {
	t.Parallel()
	ps := NewProgressState("test")
	ps.Thinking("thinking")
	ps.Waiting("waiting")
	ps.Executing("executing")
	ps.Complete()
	if ps.State != ProgressStateCompleted {
		t.Errorf("expected completed, got %s", ps.State)
	}
}

func TestProgressState_FailFromAnyState(t *testing.T) {
	t.Parallel()
	ps := NewProgressState("test")
	ps.Thinking("thinking")
	ps.Fail("error")
	if ps.State != ProgressStateFailed {
		t.Errorf("expected failed, got %s", ps.State)
	}
}

func TestErrExternal_WithCause(t *testing.T) {
	t.Parallel()
	inner := errors.New("connection refused")
	err := ErrExternal("api", inner)
	if err.Cause != inner {
		t.Error("expected inner error as cause")
	}
}

func TestErrConfig_WithCause(t *testing.T) {
	t.Parallel()
	inner := errors.New("invalid config")
	err := ErrConfig("db_url", inner)
	if err.Cause != inner {
		t.Error("expected inner error as cause")
	}
}

func TestErrInternal_WithCause(t *testing.T) {
	t.Parallel()
	inner := errors.New("panic")
	err := ErrInternal("process", inner)
	if err.Cause != inner {
		t.Error("expected inner error as cause")
	}
}

func TestProgressState_CompleteTimestamp(t *testing.T) {
	t.Parallel()
	ps := NewProgressState("test")
	before := time.Now()
	ps.Complete()
	after := time.Now()
	if ps.CompletedAt.Before(before) || ps.CompletedAt.After(after) {
		t.Error("expected CompletedAt to be between before and after")
	}
}

func TestProgressState_FailTimestamp(t *testing.T) {
	t.Parallel()
	ps := NewProgressState("test")
	before := time.Now()
	ps.Fail("error")
	after := time.Now()
	if ps.CompletedAt.Before(before) || ps.CompletedAt.After(after) {
		t.Error("expected CompletedAt to be between before and after")
	}
}

func TestError_NewError(t *testing.T) {
	t.Parallel()
	err := NewError(1, "message")
	if err.Code != 1 {
		t.Errorf("expected 1, got %d", err.Code)
	}
	if err.Message != "message" {
		t.Errorf("expected 'message', got %s", err.Message)
	}
	if err.Timestamp.IsZero() {
		t.Error("expected non-zero timestamp")
	}
}

func TestError_ErrorWithEmptyDetails(t *testing.T) {
	t.Parallel()
	err := NewError(1, "test")
	if err.Error() != "[1] test" {
		t.Errorf("expected '[1] test', got %s", err.Error())
	}
}

func TestProgressState_InitialStateTimestamp(t *testing.T) {
	t.Parallel()
	ps := NewProgressState("test")
	if ps.StartedAt.IsZero() {
		t.Error("expected non-zero StartedAt")
	}
}

func TestProgressState_ThinkingMessage(t *testing.T) {
	t.Parallel()
	ps := NewProgressState("test")
	ps.Thinking("thinking hard")
	if ps.Message != "thinking hard" {
		t.Errorf("expected 'thinking hard', got %s", ps.Message)
	}
}

func TestProgressState_ExecutingMessage(t *testing.T) {
	t.Parallel()
	ps := NewProgressState("test")
	ps.Executing("executing task")
	if ps.Message != "executing task" {
		t.Errorf("expected 'executing task', got %s", ps.Message)
	}
}

func TestProgressState_WaitingMessage(t *testing.T) {
	t.Parallel()
	ps := NewProgressState("test")
	ps.Waiting("waiting for input")
	if ps.Message != "waiting for input" {
		t.Errorf("expected 'waiting for input', got %s", ps.Message)
	}
}

func TestProgressState_FailMessage(t *testing.T) {
	t.Parallel()
	ps := NewProgressState("test")
	ps.Fail("fatal error")
	if ps.Message != "fatal error" {
		t.Errorf("expected 'fatal error', got %s", ps.Message)
	}
}

func TestProgressState_ExecutingPercent(t *testing.T) {
	t.Parallel()
	ps := NewProgressState("test")
	ps.Executing("running")
	if ps.Percent != 50 {
		t.Errorf("expected 50, got %d", ps.Percent)
	}
}

func TestProgressState_CompletePercent(t *testing.T) {
	t.Parallel()
	ps := NewProgressState("test")
	ps.Complete()
	if ps.Percent != 100 {
		t.Errorf("expected 100, got %d", ps.Percent)
	}
}
