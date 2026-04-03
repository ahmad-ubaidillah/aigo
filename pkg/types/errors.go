package types

import (
	"fmt"
	"time"
)

const (
	ErrCodeOK           = 0
	ErrCodeUnknown      = 1
	ErrCodeNotFound     = 2
	ErrCodeInvalidInput = 3
	ErrCodeTimeout      = 4
	ErrCodeUnauthorized = 5
	ErrCodeRateLimit    = 6
	ErrCodeExternal     = 7
	ErrCodeConfig       = 8
	ErrCodeInternal     = 9
)

type Error struct {
	Code      int       `json:"code"`
	Message   string    `json:"message"`
	Details   string    `json:"details,omitempty"`
	Timestamp time.Time `json:"timestamp"`
	Retryable bool      `json:"retryable"`
	Cause     error     `json:"-"`
}

func (e *Error) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("[%d] %s: %s", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

func (e *Error) Unwrap() error {
	return e.Cause
}

func NewError(code int, message string) *Error {
	return &Error{
		Code:      code,
		Message:   message,
		Timestamp: time.Now(),
	}
}

func (e *Error) WithDetails(details string) *Error {
	e.Details = details
	return e
}

func (e *Error) WithCause(cause error) *Error {
	e.Cause = cause
	if cause != nil {
		e.Details = cause.Error()
	}
	return e
}

func (e *Error) SetRetryable() *Error {
	e.Retryable = true
	return e
}

func ErrNotFound(resource string) *Error {
	return NewError(ErrCodeNotFound, "resource not found").
		WithDetails(resource)
}

func ErrInvalidInput(field, reason string) *Error {
	return NewError(ErrCodeInvalidInput, "invalid input").
		WithDetails(fmt.Sprintf("%s: %s", field, reason))
}

func ErrTimeout(operation string) *Error {
	return NewError(ErrCodeTimeout, "operation timed out").
		WithDetails(operation).
		SetRetryable()
}

func ErrExternal(service string, err error) *Error {
	return NewError(ErrCodeExternal, "external service error").
		WithDetails(service).
		WithCause(err).
		SetRetryable()
}

func ErrConfig(key string, err error) *Error {
	return NewError(ErrCodeConfig, "configuration error").
		WithDetails(key).
		WithCause(err)
}

func ErrInternal(operation string, err error) *Error {
	return NewError(ErrCodeInternal, "internal error").
		WithDetails(operation).
		WithCause(err)
}

type ProgressState struct {
	State       string    `json:"state"`
	Message     string    `json:"message"`
	Percent     int       `json:"percent"`
	StartedAt   time.Time `json:"started_at"`
	CompletedAt time.Time `json:"completed_at,omitempty"`
	Agent       string    `json:"agent,omitempty"`
	Task        string    `json:"task,omitempty"`
}

const (
	ProgressStatePending   = "pending"
	ProgressStateThinking  = "thinking"
	ProgressStateExecuting = "executing"
	ProgressStateWaiting   = "waiting"
	ProgressStateCompleted = "completed"
	ProgressStateFailed    = "failed"
)

func NewProgressState(task string) *ProgressState {
	return &ProgressState{
		State:     ProgressStatePending,
		Task:      task,
		StartedAt: time.Now(),
		Percent:   0,
	}
}

func (p *ProgressState) Thinking(message string) {
	p.State = ProgressStateThinking
	p.Message = message
}

func (p *ProgressState) Executing(message string) {
	p.State = ProgressStateExecuting
	p.Message = message
	p.Percent = 50
}

func (p *ProgressState) Waiting(message string) {
	p.State = ProgressStateWaiting
	p.Message = message
}

func (p *ProgressState) Complete() {
	p.State = ProgressStateCompleted
	p.Percent = 100
	p.CompletedAt = time.Now()
}

func (p *ProgressState) Fail(message string) {
	p.State = ProgressStateFailed
	p.Message = message
	p.CompletedAt = time.Now()
}

func (p *ProgressState) Duration() time.Duration {
	if p.CompletedAt.IsZero() {
		return time.Since(p.StartedAt)
	}
	return p.CompletedAt.Sub(p.StartedAt)
}
