package token

import (
	"strings"
	"testing"

	"github.com/ahmad-ubaidillah/aigo/pkg/types"
)

func TestDistiller_Process(t *testing.T) {
	cfg := &types.Config{}
	d := NewDistiller(cfg)

	tests := []struct {
		name     string
		input    string
		wantType string
	}{
		{
			name:     "error message",
			input:    "Error: Failed to connect to database",
			wantType: "error",
		},
		{
			name:     "success message",
			input:    "Successfully created file test.txt",
			wantType: "success",
		},
		{
			name:     "progress message",
			input:    "Processing files...",
			wantType: "progress",
		},
		{
			name:     "info message",
			input:    "The configuration file is located at /etc/config.yaml",
			wantType: "info",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := d.Process(tt.input)
			if got == "" {
				t.Errorf("Process() returned empty string for input: %s", tt.input)
			}

			// Verify that output is not longer than input (compression goal)
			if len(got) > len(tt.input) {
				t.Logf("Warning: Process() output longer than input for %s", tt.name)
			}
		})
	}
}

func TestDistiller_Classify(t *testing.T) {
	cfg := &types.Config{}
	d := NewDistiller(cfg)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "error type",
			input:    "Error: something went wrong",
			expected: "error",
		},
		{
			name:     "success type",
			input:    "Task completed successfully",
			expected: "success",
		},
		{
			name:     "progress type",
			input:    "Processing your request...",
			expected: "progress",
		},
		{
			name:     "info type",
			input:    "The sky is blue",
			expected: "info",
		},
		{
			name:     "error with panic",
			input:    "PANIC: system failure",
			expected: "error",
		},
		{
			name:     "error with fatal",
			input:    "FATAL: unrecoverable error",
			expected: "error",
		},
		{
			name:     "success with checkmark",
			input:    "✓ File uploaded",
			expected: "success",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := d.classify(tt.input)
			if got != tt.expected {
				t.Errorf("classify() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestDistiller_Score(t *testing.T) {
	cfg := &types.Config{}
	d := NewDistiller(cfg)

	tests := []struct {
		name        string
		input       string
		contentType string
		minExpected float64
		maxExpected float64
	}{
		{
			name:        "error high score",
			input:       "Error: critical security vulnerability",
			contentType: "error",
			minExpected: 0.8,
			maxExpected: 1.0,
		},
		{
			name:        "success low score",
			input:       "Successfully completed",
			contentType: "success",
			minExpected: 0.0,
			maxExpected: 0.5,
		},
		{
			name:        "progress very low score",
			input:       "Processing...",
			contentType: "progress",
			minExpected: 0.0,
			maxExpected: 0.4,
		},
		{
			name:        "info medium score",
			input:       "This is informational",
			contentType: "info",
			minExpected: 0.3,
			maxExpected: 0.7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := d.score(tt.input, tt.contentType)
			if got < tt.minExpected || got > tt.maxExpected {
				t.Errorf("score() = %v, want between %v and %v", got, tt.minExpected, tt.maxExpected)
			}
		})
	}
}

func TestDistiller_IsRepetitive(t *testing.T) {
	cfg := &types.Config{}
	d := NewDistiller(cfg)

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "repetitive lines",
			input:    "test\ntest\ntest",
			expected: true,
		},
		{
			name:     "non-repetitive",
			input:    "line1\nline2\nline3",
			expected: false,
		},
		{
			name:     "short text",
			input:    "short",
			expected: false,
		},
		{
			name:     "mixed repetition",
			input:    "unique1\nrepeat\nunique2\nrepeat\nunique3\nrepeat",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := d.isRepetitive(tt.input)
			if got != tt.expected {
				t.Errorf("isRepetitive() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestDistiller_DistillBatch(t *testing.T) {
	cfg := &types.Config{}
	d := NewDistiller(cfg)

	inputs := []string{
		"Error: failed to connect",
		"Successfully created file",
		"Processing...",
	}

	results := d.DistillBatch(inputs)

	if len(results) != len(inputs) {
		t.Errorf("DistillBatch() returned %d results, want %d", len(results), len(inputs))
	}

	for i, result := range results {
		if result == "" {
			t.Errorf("DistillBatch()[%d] is empty", i)
		}
	}
}

func TestDistiller_EstimateTokenSavings(t *testing.T) {
	cfg := &types.Config{}
	d := NewDistiller(cfg)

	tests := []struct {
		name          string
		original      string
		distilled     string
		minSavingsPct float64
	}{
		{
			name:          "significant savings",
			original:      strings.Repeat("This is a test message. ", 10),
			distilled:     "Test message.",
			minSavingsPct: 80.0,
		},
		{
			name:          "no savings",
			original:      "short",
			distilled:     "short",
			minSavingsPct: 0.0,
		},
		{
			name:          "empty original",
			original:      "",
			distilled:     "",
			minSavingsPct: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			savings := d.EstimateTokenSavings(tt.original, tt.distilled)
			if savings < tt.minSavingsPct {
				t.Errorf("EstimateTokenSavings() = %v%%, want at least %v%%", savings, tt.minSavingsPct)
			}
		})
	}
}
