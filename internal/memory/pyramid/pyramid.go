// Package pyramid implements 5-tier memory compression inspired by Project Golem.
// Tier 0: Raw conversation logs (hourly)
// Tier 1: Daily summaries (~1500 words)
// Tier 2: Monthly highlights
// Tier 3: Yearly reviews
// Tier 4: Epoch milestones
//
// 50-year storage: ~3MB (vs 500MB uncompressed)
package pyramid

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Pyramid manages 5-tier compressed memory storage.
type Pyramid struct {
	baseDir string
}

// Entry represents a single memory entry at any tier.
type Entry struct {
	Tier     int       `json:"tier"`
	Date     string    `json:"date"`
	Content  string    `json:"content"`
	Created  time.Time `json:"created"`
}

// New creates a new Pyramid memory manager.
func New(baseDir string) *Pyramid {
	tiers := []string{"raw", "daily", "monthly", "yearly", "epoch"}
	for _, t := range tiers {
		os.MkdirAll(filepath.Join(baseDir, t), 0755)
	}
	return &Pyramid{baseDir: baseDir}
}

// WriteRaw saves a raw conversation turn (Tier 0).
func (p *Pyramid) WriteRaw(speaker, content string) error {
	now := time.Now()
	filename := fmt.Sprintf("%s_%02d.log",
		now.Format("20060102"), now.Hour())
	path := filepath.Join(p.baseDir, "raw", filename)

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	line := fmt.Sprintf("[%s] %s: %s\n",
		now.Format("15:04:05"), speaker, content)
	_, err = f.WriteString(line)
	return err
}

// ReadRecent reads recent raw logs (last N hours).
func (p *Pyramid) ReadRecent(hours int) (string, error) {
	rawDir := filepath.Join(p.baseDir, "raw")
	files, err := listFilesByDate(rawDir, ".log")
	if err != nil {
		return "", nil
	}

	var result strings.Builder
	cutoff := time.Now().Add(-time.Duration(hours) * time.Hour)

	for _, f := range files {
		info, _ := os.Stat(filepath.Join(rawDir, f))
		if info != nil && info.ModTime().After(cutoff) {
			data, err := os.ReadFile(filepath.Join(rawDir, f))
			if err == nil {
				result.WriteString(fmt.Sprintf("=== %s ===\n%s\n", f, data))
			}
		}
	}
	return result.String(), nil
}

// WriteSummary saves a summary at the given tier level.
func (p *Pyramid) WriteSummary(tier int, date string, content string) error {
	tierDir := p.tierDir(tier)
	filename := fmt.Sprintf("%s.md", date)
	path := filepath.Join(tierDir, filename)

	header := fmt.Sprintf("# Tier %d Summary: %s\n\n", tier, date)
	return os.WriteFile(path, []byte(header+content), 0644)
}

// ReadTier reads summaries from a specific tier (most recent N).
func (p *Pyramid) ReadTier(tier int, count int) ([]Entry, error) {
	tierDir := p.tierDir(tier)
	files, err := listFilesByDate(tierDir, ".md")
	if err != nil {
		return nil, nil
	}

	// Take most recent N
	if len(files) > count {
		files = files[len(files)-count:]
	}

	var entries []Entry
	for _, f := range files {
		data, err := os.ReadFile(filepath.Join(tierDir, f))
		if err != nil {
			continue
		}
		date := strings.TrimSuffix(f, ".md")
		entries = append(entries, Entry{
			Tier:    tier,
			Date:    date,
			Content: string(data),
		})
	}
	return entries, nil
}

// InjectContext builds a multi-tier memory context string for prompt injection.
// Returns context with max totalChars budget.
func (p *Pyramid) InjectContext(totalChars int) string {
	var parts []string

	// Load from highest to lowest tier (epoch → yearly → monthly → daily)
	// Priority: epoch and yearly give most context
	tiers := []struct {
		tier  int
		count int
		label string
	}{
		{4, 1, "Epoch"},
		{3, 1, "Yearly"},
		{2, 3, "Monthly"},
		{1, 7, "Daily"},
	}

	used := 0
	for _, t := range tiers {
		entries, _ := p.ReadTier(t.tier, t.count)
		for _, e := range entries {
			chunk := fmt.Sprintf("\n=== [%s: %s] ===\n%s\n", t.label, e.Date, e.Content)
			if used+len(chunk) > totalChars {
				break
			}
			parts = append(parts, chunk)
			used += len(chunk)
		}
	}

	// Fallback: if no compressed summaries, load raw logs
	if len(parts) == 0 {
		raw, _ := p.ReadRecent(48)
		if raw != "" {
			if len(raw) > totalChars {
				raw = raw[len(raw)-totalChars:]
			}
			parts = append(parts, fmt.Sprintf("\n=== [Recent Conversations] ===\n%s", raw))
		}
	}

	if len(parts) == 0 {
		return ""
	}

	return "【Long-term Memory】\n" + strings.Join(parts, "\n")
}

// NeedsCompression checks if raw logs exceed threshold for compression.
func (p *Pyramid) NeedsCompression() (bool, string) {
	rawDir := filepath.Join(p.baseDir, "raw")
	files, err := listFilesByDate(rawDir, ".log")
	if err != nil || len(files) == 0 {
		return false, ""
	}

	// Get yesterday's date
	yesterday := time.Now().AddDate(0, 0, -1).Format("20060102")
	count := 0
	for _, f := range files {
		if strings.HasPrefix(f, yesterday) {
			count++
		}
	}

	// Threshold: 6+ hours of logs for yesterday = compress
	if count >= 6 {
		return true, yesterday
	}
	return false, ""
}

// CompressDaily is called by the autonomous agent to compress raw → daily.
// summaryContent is the AI-generated summary text.
func (p *Pyramid) CompressDaily(date string, summaryContent string) error {
	// Write daily summary
	err := p.WriteSummary(1, date, summaryContent)
	if err != nil {
		return err
	}

	// Delete raw logs for that date (save space)
	rawDir := filepath.Join(p.baseDir, "raw")
	files, _ := listFilesByDate(rawDir, ".log")
	for _, f := range files {
		if strings.HasPrefix(f, date) {
			os.Remove(filepath.Join(rawDir, f))
		}
	}
	return nil
}

// CompressMonthly merges daily summaries into monthly highlight.
func (p *Pyramid) CompressMonthly(yearMonth string, summaryContent string) error {
	err := p.WriteSummary(2, yearMonth, summaryContent)
	if err != nil {
		return err
	}

	// Delete daily summaries for that month
	dailyDir := filepath.Join(p.baseDir, "daily")
	files, _ := listFilesByDate(dailyDir, ".md")
	for _, f := range files {
		if strings.HasPrefix(f, yearMonth) {
			os.Remove(filepath.Join(dailyDir, f))
		}
	}
	return nil
}

// CompressYearly merges monthly summaries into yearly review.
func (p *Pyramid) CompressYearly(year string, summaryContent string) error {
	err := p.WriteSummary(3, year, summaryContent)
	if err != nil {
		return err
	}

	monthlyDir := filepath.Join(p.baseDir, "monthly")
	files, _ := listFilesByDate(monthlyDir, ".md")
	for _, f := range files {
		if strings.HasPrefix(f, year) {
			os.Remove(filepath.Join(monthlyDir, f))
		}
	}
	return nil
}

// CompressEpoch merges yearly reviews into epoch milestone.
func (p *Pyramid) CompressEpoch(label string, summaryContent string) error {
	return p.WriteSummary(4, label, summaryContent)
}

// Stats returns storage statistics.
func (p *Pyramid) Stats() map[string]interface{} {
	stats := make(map[string]interface{})
	tiers := []string{"raw", "daily", "monthly", "yearly", "epoch"}

	totalSize := int64(0)
	for _, t := range tiers {
		dir := filepath.Join(p.baseDir, t)
		files, _ := listFilesByDate(dir, "")
		var size int64
		for _, f := range files {
			info, err := os.Stat(filepath.Join(dir, f))
			if err == nil {
				size += info.Size()
			}
		}
		stats[t] = map[string]interface{}{
			"files": len(files),
			"bytes": size,
		}
		totalSize += size
	}
	stats["total_bytes"] = totalSize
	return stats
}

func (p *Pyramid) tierDir(tier int) string {
	names := map[int]string{0: "raw", 1: "daily", 2: "monthly", 3: "yearly", 4: "epoch"}
	return filepath.Join(p.baseDir, names[tier])
}

func listFilesByDate(dir string, ext string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if ext == "" || strings.HasSuffix(e.Name(), ext) {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)
	return files, nil
}

// ScanTier0Raw scans raw logs and returns as lines for LLM summarization.
func (p *Pyramid) ScanTier0Raw(date string) (string, error) {
	rawDir := filepath.Join(p.baseDir, "raw")
	pattern := date + "*.log"
	matches, err := filepath.Glob(filepath.Join(rawDir, pattern))
	if err != nil {
		return "", err
	}

	var lines []string
	for _, m := range matches {
		f, err := os.Open(m)
		if err != nil {
			continue
		}
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}
		f.Close()
	}
	return strings.Join(lines, "\n"), nil
}
