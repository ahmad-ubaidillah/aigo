package intent

import (
	"testing"

	"github.com/ahmad-ubaidillah/aigo/pkg/types"
)

func TestRuleBasedClassify(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedIntent string
		minConfidence  float64
	}{
		{
			name:           "coding intent with multiple keywords",
			input:          "fix the bug in the authentication function",
			expectedIntent: types.IntentCoding,
			minConfidence:  0.5,
		},
		{
			name:           "web search intent",
			input:          "search the web for Go best practices",
			expectedIntent: types.IntentWeb,
			minConfidence:  0.3,
		},
		{
			name:           "file operation intent",
			input:          "read the file and show me the contents",
			expectedIntent: types.IntentFile,
			minConfidence:  0.3,
		},
		{
			name:           "gateway messaging intent",
			input:          "send a message to telegram channel",
			expectedIntent: types.IntentGateway,
			minConfidence:  0.3,
		},
		{
			name:           "memory intent",
			input:          "remember this important fact for later",
			expectedIntent: types.IntentMemory,
			minConfidence:  0.3,
		},
		{
			name:           "automation intent",
			input:          "schedule a daily report every day at 9am",
			expectedIntent: types.IntentAutomation,
			minConfidence:  0.3,
		},
		{
			name:           "general conversation",
			input:          "hello how are you doing today",
			expectedIntent: types.IntentGeneral,
			minConfidence:  0.1,
		},
	}

	cfg := types.Config{}
	classifier := NewClassifier(cfg)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := classifier.Classify(tt.input)

			if result.Intent != tt.expectedIntent {
				t.Errorf("expected intent %s, got %s", tt.expectedIntent, result.Intent)
			}

			if result.Confidence < tt.minConfidence {
				t.Errorf("expected confidence >= %.2f, got %.2f", tt.minConfidence, result.Confidence)
			}

			if result.Confidence > 1.0 {
				t.Errorf("confidence %.2f exceeds 1.0", result.Confidence)
			}
		})
	}
}

func TestExtractWorkspace(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "workspace flag",
			input:    "fix bug --workspace /home/user/project",
			expected: "/home/user/project",
		},
		{
			name:     "in directory pattern",
			input:    "list files in /var/log",
			expected: "/var/log",
		},
		{
			name:     "at directory pattern",
			input:    "create file at /tmp/test.txt",
			expected: "/tmp/test.txt",
		},
		{
			name:     "no workspace",
			input:    "fix the bug in the code",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractWorkspace(tt.input)
			if result != tt.expected {
				t.Errorf("expected workspace %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestExtractTags(t *testing.T) {
	input := "fix the authentication bug"
	tags := extractTags(input)

	if len(tags) == 0 {
		t.Error("expected tags to be extracted")
	}
}

func TestClassifyWithoutLLM(t *testing.T) {
	cfg := types.Config{}
	classifier := NewClassifier(cfg)

	result := classifier.Classify("fix the bug in main.go")

	if result.Intent != types.IntentCoding {
		t.Errorf("expected coding intent, got %s", result.Intent)
	}
}

func TestStopWordsFiltering(t *testing.T) {
	input := "please can you help me with the code"
	tags := extractTags(input)

	stopWordCount := 0
	for _, tag := range tags {
		if stopWords[tag] {
			stopWordCount++
		}
	}

	if stopWordCount > 0 {
		t.Errorf("found %d stop words in tags: %v", stopWordCount, tags)
	}
}

func TestKeywordSetsCompleteness(t *testing.T) {
	requiredIntents := []string{
		types.IntentCoding,
		types.IntentWeb,
		types.IntentFile,
		types.IntentGateway,
		types.IntentMemory,
		types.IntentAutomation,
	}

	for _, intent := range requiredIntents {
		if _, exists := keywordSets[intent]; !exists {
			t.Errorf("missing keyword set for intent: %s", intent)
		}
	}
}

func TestConfidenceBounds(t *testing.T) {
	cfg := types.Config{}
	classifier := NewClassifier(cfg)

	inputs := []string{
		"fix bug",
		"search web",
		"read file",
		"send message",
		"remember this",
		"schedule task",
		"hello world",
	}

	for _, input := range inputs {
		result := classifier.Classify(input)
		if result.Confidence < 0.0 || result.Confidence > 1.0 {
			t.Errorf("confidence %.2f out of bounds [0.0, 1.0] for input: %s", result.Confidence, input)
		}
	}
}
