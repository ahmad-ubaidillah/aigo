package memory

import (
	"sort"
	"time"
)

type Archive struct {
	items     map[string]string
	coldThreshold time.Duration
}

func NewArchive() *Archive {
	return &Archive{
		items:       make(map[string]string),
		coldThreshold: 30 * 24 * time.Hour,
	}
}

func (a *Archive) Store(data string) error {
	key := time.Now().Format("2006-01-02-15:04")
	a.items[key] = data
	return nil
}

func (a *Archive) IsCold(dateStr string) bool {
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return false
	}
	age := time.Since(date)
	return age > a.coldThreshold
}

func (a *Archive) Retrieve(key string) string {
	return a.items[key]
}

func (a *Archive) List() []string {
	keys := make([]string, 0, len(a.items))
	for k := range a.items {
		keys = append(keys, k)
	}
	return keys
}

func (a *Archive) Prune(maxItems int) int {
	if len(a.items) <= maxItems {
		return 0
	}

	removed := 0
	sorted := make([]string, 0, len(a.items))
	for k := range a.items {
		sorted = append(sorted, k)
	}
	sort.Strings(sorted)

	for i := 0; i < len(sorted)-maxItems; i++ {
		delete(a.items, sorted[i])
		removed++
	}
	return removed
}