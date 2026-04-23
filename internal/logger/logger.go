// Package logger provides structured JSON logging for Aigo.
package logger

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"
)

// Level represents log severity.
type Level string

const (
	LevelDebug Level = "debug"
	LevelInfo  Level = "info"
	LevelWarn  Level = "warn"
	LevelError Level = "error"
)

// Logger writes structured JSON logs.
type Logger struct {
	service string
	output  *log.Logger
}

// New creates a new structured logger.
func New(service string) *Logger {
	return &Logger{
		service: service,
		output:  log.New(os.Stderr, "", 0),
	}
}

// NewWithOutput creates a logger with a custom output.
func NewWithOutput(service string, out *log.Logger) *Logger {
	return &Logger{service: service, output: out}
}

func (l *Logger) log(level Level, msg string, fields map[string]interface{}) {
	entry := map[string]interface{}{
		"time":    time.Now().UTC().Format(time.RFC3339),
		"level":   string(level),
		"service": l.service,
		"message": msg,
	}
	for k, v := range fields {
		entry[k] = v
	}
	b, _ := json.Marshal(entry)
	l.output.Println(string(b))
}

// Debug logs a debug message.
func (l *Logger) Debug(msg string, fields map[string]interface{}) {
	l.log(LevelDebug, msg, fields)
}

// Info logs an info message.
func (l *Logger) Info(msg string, fields map[string]interface{}) {
	l.log(LevelInfo, msg, fields)
}

// Warn logs a warning message.
func (l *Logger) Warn(msg string, fields map[string]interface{}) {
	l.log(LevelWarn, msg, fields)
}

// Error logs an error message.
func (l *Logger) Error(msg string, err error, fields map[string]interface{}) {
	if fields == nil {
		fields = make(map[string]interface{})
	}
	if err != nil {
		fields["error"] = err.Error()
	}
	l.log(LevelError, msg, fields)
}

// Fatal logs an error and exits.
func (l *Logger) Fatal(msg string, err error) {
	l.Error(msg, err, nil)
	os.Exit(1)
}

// Simple wrappers for common usage
func (l *Logger) Printf(format string, v ...interface{}) {
	l.Info(fmt.Sprintf(format, v...), nil)
}
