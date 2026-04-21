package memory

import (
	"strings"
)

type ExtractedItem struct {
	Content   string
	Category string
	Confidence float64
}

type Extractor struct{}

func NewExtractor() *Extractor {
	return &Extractor{}
}

func (e *Extractor) ExtractProfile(text string) *ExtractedItem {
	profileIndicators := []string{"prefer", "like", "hate", "use", "uses"}
	for _, ind := range profileIndicators {
		if strings.Contains(strings.ToLower(text), ind) {
			return &ExtractedItem{
				Content:   text,
				Category:  "profile",
				Confidence: 0.7,
			}
		}
	}
	return &ExtractedItem{
		Content:   text,
		Category:  "profile",
		Confidence: 0.3,
	}
}

func (e *Extractor) ExtractPreferences(text string) *ExtractedItem {
	prefIndicators := []string{"prefer", "like", "hate"}
	for _, ind := range prefIndicators {
		if strings.Contains(strings.ToLower(text), ind) {
			return &ExtractedItem{
				Content:   text,
				Category:  "preferences",
				Confidence: 0.6,
			}
		}
	}
	return nil
}

func (e *Extractor) ExtractEntities(text string) *ExtractedItem {
	entityIndicators := []string{"at /", "table", "database", "api", "endpoint", "struct", "function"}
	for _, ind := range entityIndicators {
		if strings.Contains(strings.ToLower(text), ind) {
			return &ExtractedItem{
				Content:   text,
				Category:  "entities",
				Confidence: 0.7,
			}
		}
	}
	return nil
}

func (e *Extractor) ExtractEvents(text string) *ExtractedItem {
	eventIndicators := []string{"found", "scheduled", "meeting", "bug", "error", "created", "updated"}
	for _, ind := range eventIndicators {
		if strings.Contains(strings.ToLower(text), ind) {
			return &ExtractedItem{
				Content:   text,
				Category:  "events",
				Confidence: 0.6,
			}
		}
	}
	return nil
}

func (e *Extractor) ExtractCases(text string) *ExtractedItem {
	caseIndicators := []string{"handle", "if", "when", "error", "timeout", "retry"}
	for _, ind := range caseIndicators {
		if strings.Contains(strings.ToLower(text), ind) {
			return &ExtractedItem{
				Content:   text,
				Category:  "cases",
				Confidence: 0.5,
			}
		}
	}
	return nil
}

func (e *Extractor) ExtractPatterns(text string) *ExtractedItem {
	patternIndicators := []string{"always", "never", "usually", "always check"}
	for _, ind := range patternIndicators {
		if strings.Contains(strings.ToLower(text), ind) {
			return &ExtractedItem{
				Content:   text,
				Category:  "patterns",
				Confidence: 0.8,
			}
		}
	}
	return nil
}