package distill

import (
	"strconv"
	"strings"
)

// SignalTier represents the importance level of content.
type SignalTier float64

const (
	TierCritical  SignalTier = 0.9
	TierImportant SignalTier = 0.7
	TierContext   SignalTier = 0.4
	TierNoise     SignalTier = 0.05
)

// Scorer assigns signal tiers to content based on type and content.
type Scorer struct{}

// NewScorer creates a new Scorer instance.
func NewScorer() *Scorer {
	return &Scorer{}
}

// Score analyzes text and returns its SignalTier based on content type.
func (s *Scorer) Score(text string, ct ContentType) SignalTier {
	switch ct {
	case ContentGitDiff:
		return s.scoreGitDiff(text)
	case ContentBuildOutput:
		return s.scoreBuildOutput(text)
	case ContentTestOutput:
		return s.scoreTestOutput(text)
	case ContentLogOutput:
		return s.scoreLogOutput(text)
	case ContentStructuredData:
		return TierImportant
	case ContentTabularData:
		return TierContext
	case ContentInfraOutput:
		return TierImportant
	default:
		return TierNoise
	}
}

func (s *Scorer) scoreGitDiff(text string) SignalTier {
	if strings.Contains(text, "new file") || strings.Contains(text, "deleted file") {
		return TierCritical
	}
	if strings.Contains(text, "@@") {
		return TierContext
	}
	return TierContext
}

func (s *Scorer) scoreBuildOutput(text string) SignalTier {
	textLower := strings.ToLower(text)
	if strings.Contains(textLower, "error:") || strings.Contains(textLower, "fatal") {
		return TierCritical
	}
	if strings.Contains(textLower, "warning:") {
		return TierImportant
	}
	return TierContext
}

func (s *Scorer) scoreTestOutput(text string) SignalTier {
	if strings.Contains(text, "--- FAIL") || strings.Contains(text, "FAIL\t") {
		return TierCritical
	}
	if strings.Contains(text, "--- PASS") || strings.Contains(text, "PASS") {
		return TierImportant
	}
	return TierContext
}

func (s *Scorer) scoreLogOutput(text string) SignalTier {
	textUpper := strings.ToUpper(text)
	if strings.Contains(textUpper, "ERROR") || strings.Contains(textUpper, "FATAL") {
		return TierCritical
	}
	if strings.Contains(textUpper, "WARN") {
		return TierImportant
	}
	if strings.Contains(textUpper, "INFO") {
		return TierContext
	}
	return TierNoise
}

// Collapse compresses repetitive content in text.
type Collapse struct{}

// NewCollapse creates a new Collapse instance.
func NewCollapse() *Collapse {
	return &Collapse{}
}

// Compress replaces repetitive lines and collapses blank lines.
func (c *Collapse) Compress(text string) string {
	lines := strings.Split(text, "\n")
	if len(lines) == 0 {
		return text
	}

	result := make([]string, 0, len(lines))
	repeatCount := 1
	blankCount := 0

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		isBlank := strings.TrimSpace(line) == ""

		if isBlank {
			blankCount++
			if blankCount <= 1 {
				result = append(result, line)
			}
			repeatCount = 1
			continue
		}

		blankCount = 0
		if i+1 < len(lines) && lines[i+1] == line {
			repeatCount++
			continue
		}

		if repeatCount >= 3 {
			result = append(result, line)
			result = append(result, "  ... ("+strconv.Itoa(repeatCount)+" repeated lines omitted)")
		} else {
			for j := 0; j < repeatCount; j++ {
				result = append(result, line)
			}
		}
		repeatCount = 1
	}

	return strings.Join(result, "\n")
}

// Composer filters and assembles lines based on signal tiers.
type Composer struct {
	Threshold SignalTier
}

// NewComposer creates a new Composer with default threshold.
func NewComposer() *Composer {
	return &Composer{Threshold: TierNoise}
}

// Compose filters lines by tier threshold and joins them.
func (c *Composer) Compose(lines []string, tiers []SignalTier) (string, int) {
	if len(lines) != len(tiers) {
		return strings.Join(lines, "\n"), 0
	}

	kept := make([]string, 0, len(lines))
	dropped := 0

	for i, line := range lines {
		if tiers[i] >= c.Threshold {
			kept = append(kept, line)
		} else {
			dropped++
		}
	}

	return strings.Join(kept, "\n"), dropped
}

// ArchiveFunc is called for content that gets dropped during distillation.
type ArchiveFunc func(content, contentType, sessionID string) string

// Pipeline orchestrates the full distillation process.
type Pipeline struct {
	classifier *Classifier
	scorer     *Scorer
	collapse   *Collapse
	composer   *Composer
	archive    ArchiveFunc
	sessionID  string
}

// NewPipeline creates a new distillation pipeline.
func NewPipeline() *Pipeline {
	return &Pipeline{
		classifier: NewClassifier(),
		scorer:     NewScorer(),
		collapse:   NewCollapse(),
		composer:   NewComposer(),
	}
}

// WithArchive sets the archive function for dropped content.
func (p *Pipeline) WithArchive(fn ArchiveFunc) *Pipeline {
	p.archive = fn
	return p
}

// WithSessionID sets the session ID for this pipeline run.
func (p *Pipeline) WithSessionID(sid string) *Pipeline {
	p.sessionID = sid
	return p
}

// Process runs the full distillation pipeline on text.
func (p *Pipeline) Process(text string) (output string, tokensSaved int, originalTokens int) {
	originalTokens = estimateTokens(text)

	ct := p.classifier.Classify(text)
	lines := strings.Split(text, "\n")
	tiers := make([]SignalTier, len(lines))
	for i, line := range lines {
		tiers[i] = p.scorer.Score(line, ct)
	}

	filtered, droppedCount := p.composer.Compose(lines, tiers)

	// Archive dropped content if archive function is set
	if p.archive != nil && droppedCount > 0 {
		p.archiveDropped(lines, tiers, ct)
	}

	collapsed := p.collapse.Compress(filtered)

	output = collapsed
	finalTokens := estimateTokens(output)
	tokensSaved = originalTokens - finalTokens

	return output, tokensSaved, originalTokens
}

// archiveDropped sends dropped lines to the archive function.
func (p *Pipeline) archiveDropped(lines []string, tiers []SignalTier, ct ContentType) {
	var dropped []string
	for i, line := range lines {
		if tiers[i] < p.composer.Threshold && strings.TrimSpace(line) != "" {
			dropped = append(dropped, line)
		}
	}

	if len(dropped) > 0 {
		content := strings.Join(dropped, "\n")
		contentType := contentTypeToString(ct)
		p.archive(content, contentType, p.sessionID)
	}
}

func contentTypeToString(ct ContentType) string {
	switch ct {
	case ContentGitDiff:
		return "git_diff"
	case ContentBuildOutput:
		return "build_output"
	case ContentTestOutput:
		return "test_output"
	case ContentInfraOutput:
		return "infra_output"
	case ContentLogOutput:
		return "log_output"
	case ContentTabularData:
		return "tabular_data"
	case ContentStructuredData:
		return "structured_data"
	default:
		return "unknown"
	}
}

func estimateTokens(text string) int {
	return len(text)/4 + 1
}
