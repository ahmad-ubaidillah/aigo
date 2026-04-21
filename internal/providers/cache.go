package providers

import (
	"strings"
)

func isCachedModel(model string) bool {
	cachedModels := []string{
		"o1",
		"o1-mini",
		"o1-preview",
		"o3",
		"o3-mini",
		"claude-3-5-sonnet-20241022",
		"claude-3-opus-20240229",
		"claude-3-sonnet-20240229",
		"claude-3-haiku-20240307",
	}
	modelLower := strings.ToLower(model)
	for _, m := range cachedModels {
		if strings.Contains(modelLower, strings.ToLower(m)) {
			return true
		}
	}
	return false
}

func calculateCacheTokens(messages []Message) int {
	total := 0
	for _, m := range messages {
		total += len(m.Content) / 4
	}
	return total
}