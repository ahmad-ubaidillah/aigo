// Package distiller provides 3-layer web content distillation:
// Layer 1: HTML Cleaner  — strip noise (script, style, nav, footer, ads, comments)
// Layer 2: Content Extractor — extract main content (Readability-like)
// Layer 3: Sentence Distiller — filter by keyword relevance
package distiller

import (
	"bytes"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"

	"golang.org/x/net/html"
)

// DistillMode controls how aggressive distillation is.
type DistillMode int

const (
	ModeResearch DistillMode = iota // Full content, minimal filtering
	ModeSmart                       // Query-aware, keyword-filtered
	ModeCompact                     // Aggressive: only highest-relevance sentences
)

// DistillResult contains the distilled content.
type DistillResult struct {
	Title        string
	Content      string
	OriginalSize int
	FinalSize    int
	SentencesIn  int
	SentencesOut int
	Mode         string
}

// noiseTags are elements that are NEVER content — always strip entirely.
// These are layout/navigation/ads/tracking elements.
var noiseTags = map[string]bool{
	"script": true, "style": true, "noscript": true,
	"nav": true, "footer": true, "header": true,
	"aside": true,
	"iframe": true, "embed": true, "object": true,
	"svg": true, "canvas": true, "video": true, "audio": true,
	"form": true, "input": true, "button": true, "select": true, "textarea": true,
	// Semantic but non-content
	"meta": true, "link": true, "head": true,
}

// noiseClasses/IDs — elements matching these are stripped.
var noisePatterns = []string{
	"nav", "menu", "sidebar", "footer", "header", "banner",
	"ad-", "ads-", "advert", "sponsor", "promo", "popup",
	"cookie", "consent", "gdpr", "newsletter", "subscribe",
	"social", "share", "comment", "reply", "related",
	"widget", "breadcrumb", "pagination", "pager",
	"toolbar", "tab-", "modal", "overlay", "toast",
	"search", "filter", "sort",
}

// contentSignals — elements/attributes that SIGNAL main content.
var contentSignals = []string{
	"article", "main", "content", "post", "entry",
	"story", "body", "text", "article-body",
}

// --- Layer 1: HTML Cleaner ---

// CleanHTML strips noise elements and returns clean HTML.
func CleanHTML(rawHTML string) string {
	doc, err := html.Parse(strings.NewReader(rawHTML))
	if err != nil {
		// Fallback: regex-based cleaning
		return regexClean(rawHTML)
	}

	// First pass: remove noise nodes
	removeNoiseNodes(doc)

	// Render back to HTML
	var buf bytes.Buffer
	html.Render(&buf, doc)
	return buf.String()
}

func removeNoiseNodes(n *html.Node) {
	var remove []*html.Node

	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode {
			tag := strings.ToLower(n.Data)

			// Remove noise tags
			if noiseTags[tag] {
				remove = append(remove, n)
				return
			}

			// Check class/id for noise patterns
			if hasNoisePattern(n) {
				remove = append(remove, n)
				return
			}

			// Remove HTML comments
			for c := n.FirstChild; c != nil; {
				next := c.NextSibling
				if c.Type == html.CommentNode {
					n.RemoveChild(c)
				}
				c = next
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)

	// Remove marked nodes
	for _, node := range remove {
		if node.Parent != nil {
			node.Parent.RemoveChild(node)
		}
	}
}

func hasNoisePattern(n *html.Node) bool {
	for _, attr := range n.Attr {
		val := strings.ToLower(attr.Val)
		key := strings.ToLower(attr.Key)
		if key == "class" || key == "id" {
			for _, pattern := range noisePatterns {
				if strings.Contains(val, pattern) {
					return true
				}
			}
		}
		// role="navigation", role="banner", etc.
		if key == "role" {
			role := strings.ToLower(attr.Val)
			if role == "navigation" || role == "banner" || role == "complementary" ||
				role == "contentinfo" || role == "search" || role == "form" {
				return true
			}
		}
	}
	return false
}

func regexClean(rawHTML string) string {
	// Remove script/style blocks
	scriptRe := regexp.MustCompile(`(?is)<script[^>]*>.*?</script>`)
	rawHTML = scriptRe.ReplaceAllString(rawHTML, "")
	styleRe := regexp.MustCompile(`(?is)<style[^>]*>.*?</style>`)
	rawHTML = styleRe.ReplaceAllString(rawHTML, "")
	// Remove HTML comments
	commentRe := regexp.MustCompile(`(?is)<!--.*?-->`)
	rawHTML = commentRe.ReplaceAllString(rawHTML, "")
	// Remove noise elements
	for _, tag := range []string{"nav", "footer", "header", "aside", "noscript", "iframe", "svg", "canvas"} {
		re := regexp.MustCompile(fmt.Sprintf(`(?is)<%s[^>]*>.*?</%s>`, tag, tag))
		rawHTML = re.ReplaceAllString(rawHTML, "")
	}
	return rawHTML
}

// --- Layer 2: Content Extractor (Readability-like) ---

// ExtractContent extracts the main content from HTML using a scoring algorithm.
// Inspired by Mozilla's Readability: scores elements by text density and content signals.
func ExtractContent(rawHTML string) (title, content string) {
	doc, err := html.Parse(strings.NewReader(rawHTML))
	if err != nil {
		return extractTitle(rawHTML), stripTags(rawHTML)
	}

	title = extractTitleFromDoc(doc)
	if title == "" {
		title = extractTitle(rawHTML)
	}

	// Score all elements
	bestNode := findBestContentNode(doc)
	if bestNode == nil {
		return title, stripTags(rawHTML)
	}

	content = extractText(bestNode)
	content = cleanWhitespace(content)
	return title, content
}

// contentScore holds a node's content score.
type contentScore struct {
	node  *html.Node
	score float64
}

func findBestContentNode(doc *html.Node) *html.Node {
	var candidates []contentScore

	var walk func(*html.Node, int)
	walk = func(n *html.Node, depth int) {
		if n.Type == html.ElementNode {
			tag := strings.ToLower(n.Data)

			// Skip noise
			if noiseTags[tag] || hasNoisePattern(n) {
				return
			}

			score := scoreNode(n, depth)
			if score > 0 {
				candidates = append(candidates, contentScore{n, score})
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c, depth+1)
		}
	}
	walk(doc, 0)

	if len(candidates) == 0 {
		return nil
	}

	// Sort by score descending
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].score > candidates[j].score
	})

	return candidates[0].node
}

func scoreNode(n *html.Node, depth int) float64 {
	tag := strings.ToLower(n.Data)

	// Base score by tag
	score := 0.0
	switch tag {
	case "article":
		score += 10.0
	case "main":
		score += 8.0
	case "div", "section":
		score += 5.0
	case "p":
		score += 3.0
	case "pre", "code":
		score += 4.0 // Code blocks are valuable content
	default:
		score += 1.0
	}

	// Content signal boost
	for _, attr := range n.Attr {
		val := strings.ToLower(attr.Val)
		for _, signal := range contentSignals {
			if strings.Contains(val, signal) {
				score += 5.0
				break
			}
		}
	}

	// Text density: more text = higher score
	textLen := float64(utf8.RuneCountInString(extractText(n)))
	tagCount := float64(countTags(n))
	density := 0.0
	if tagCount > 0 {
		density = textLen / tagCount
	}
	score += math.Log2(density+1) * 2.0

	// Paragraph count bonus
	pCount := float64(countParagraphs(n))
	score += pCount * 2.0

	// Depth penalty — prefer shallower content nodes
	score -= float64(depth) * 0.3

	return score
}

func countTags(n *html.Node) int {
	count := 0
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode {
			count++
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return count
}

func countParagraphs(n *html.Node) int {
	count := 0
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && strings.ToLower(n.Data) == "p" {
			count++
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return count
}

func extractTitleFromDoc(doc *html.Node) string {
	var title string
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode {
			tag := strings.ToLower(n.Data)
			// h1 is usually the title
			if tag == "h1" && title == "" {
				title = extractText(n)
			}
			// og:title or <title> tag
			if tag == "meta" {
				for _, attr := range n.Attr {
					if strings.ToLower(attr.Key) == "property" &&
						strings.ToLower(attr.Val) == "og:title" {
						for _, a := range n.Attr {
							if strings.ToLower(a.Key) == "content" {
								title = a.Val
							}
						}
					}
				}
			}
		}
		if n.Type == html.ElementNode && strings.ToLower(n.Data) == "title" {
			t := extractText(n)
			if t != "" && title == "" {
				title = t
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
	return strings.TrimSpace(title)
}

func extractTitle(rawHTML string) string {
	re := regexp.MustCompile(`(?is)<title[^>]*>(.*?)</title>`)
	m := re.FindStringSubmatch(rawHTML)
	if len(m) > 1 {
		return strings.TrimSpace(m[1])
	}
	return ""
}

// --- Layer 3: Sentence Distiller ---

// Distill filters sentences by keyword relevance.
// query: the search/query keywords (comma or space separated)
// mode: how aggressive the filtering is
func Distill(text, query string, mode DistillMode) *DistillResult {
	originalSize := len(text)

	sentences := splitSentences(text)
	totalIn := len(sentences)

	if query == "" || mode == ModeResearch {
		// No filtering — return everything
		return &DistillResult{
			Content:      text,
			OriginalSize: originalSize,
			FinalSize:    len(text),
			SentencesIn:  totalIn,
			SentencesOut: totalIn,
			Mode:         mode.String(),
		}
	}

	keywords := extractKeywords(query)

	// Score each sentence
	type scored struct {
		text  string
		score float64
	}

	var scoredSentences []scored
	for _, s := range sentences {
		sc := scoreSentence(s, keywords)
		scoredSentences = append(scoredSentences, scored{s, sc})
	}

	// Determine threshold based on mode
	threshold := 0.0
	maxResults := totalIn
	switch mode {
	case ModeSmart:
		// Keep sentences with any keyword match + high-value sentences
		threshold = 0.0
		maxResults = int(float64(totalIn) * 0.6) // Keep top 60%
	case ModeCompact:
		// Only keep sentences with strong keyword match
		threshold = 0.1
		maxResults = int(math.Max(float64(totalIn)*0.3, float64(len(keywords)*2)))
	}

	// Filter
	var kept []string
	for _, s := range scoredSentences {
		if s.score > threshold {
			kept = append(kept, s.text)
		}
	}

	// If too many kept, take top N by score
	if len(kept) > maxResults {
		sort.Slice(scoredSentences, func(i, j int) bool {
			return scoredSentences[i].score > scoredSentences[j].score
		})
		kept = nil
		for i := 0; i < maxResults && i < len(scoredSentences); i++ {
			if scoredSentences[i].score > threshold {
				kept = append(kept, scoredSentences[i].text)
			}
		}
	}

	result := strings.Join(kept, " ")
	finalSize := len(result)

	return &DistillResult{
		Content:      result,
		OriginalSize: originalSize,
		FinalSize:    finalSize,
		SentencesIn:  totalIn,
		SentencesOut: len(kept),
		Mode:         mode.String(),
	}
}

func (m DistillMode) String() string {
	switch m {
	case ModeResearch:
		return "research"
	case ModeSmart:
		return "smart"
	case ModeCompact:
		return "compact"
	default:
		return "unknown"
	}
}

func extractKeywords(query string) []string {
	// Normalize
	query = strings.ToLower(query)

	// Split by common separators
	re := regexp.MustCompile(`[\s,;:|/&]+`)
	parts := re.Split(query, -1)

	var keywords []string
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "is": true, "are": true,
		"was": true, "were": true, "be": true, "been": true, "being": true,
		"have": true, "has": true, "had": true, "do": true, "does": true,
		"did": true, "will": true, "would": true, "could": true, "should": true,
		"may": true, "might": true, "shall": true, "can": true,
		"and": true, "or": true, "but": true, "not": true, "no": true,
		"for": true, "to": true, "of": true, "in": true, "on": true,
		"at": true, "by": true, "with": true, "from": true, "up": true,
		"about": true, "into": true, "over": true, "after": true,
		"this": true, "that": true, "these": true, "those": true,
		"it": true, "its": true, "he": true, "she": true, "they": true,
		"we": true, "you": true, "i": true, "me": true, "my": true,
		"what": true, "which": true, "who": true, "whom": true,
		"how": true, "when": true, "where": true, "why": true,
		"all": true, "each": true, "every": true, "both": true,
		"few": true, "more": true, "most": true, "other": true,
		"some": true, "such": true, "only": true, "own": true,
		"same": true, "so": true, "than": true, "too": true, "very": true,
		"just": true, "also": true, "any": true, "get": true, "use": true,
		"make": true, "like": true, "long": true, "look": true,
	}

	for _, p := range parts {
		p = strings.TrimSpace(p)
		if len(p) >= 2 && !stopWords[p] {
			keywords = append(keywords, p)
		}
	}

	// Also keep bigrams (2-word phrases)
	words := strings.Fields(query)
	for i := 0; i < len(words)-1; i++ {
		bigram := words[i] + " " + words[i+1]
		if len(bigram) >= 5 {
			keywords = append(keywords, bigram)
		}
	}

	return keywords
}

func scoreSentence(sentence string, keywords []string) float64 {
	lower := strings.ToLower(sentence)
	score := 0.0

	for _, kw := range keywords {
		if strings.Contains(lower, kw) {
			// Longer keyword match = higher score
			score += float64(len(kw)) * 1.5
		}
	}

	// Normalize by sentence length (avoid penalizing short sentences too much)
	words := float64(len(strings.Fields(sentence)))
	if words > 0 {
		score = score / math.Sqrt(words) * 10.0
	}

	// Boost for sentences with numbers (often factual)
	if regexp.MustCompile(`\d+`).MatchString(sentence) {
		score *= 1.2
	}

	// Boost for sentences starting with capital (likely proper sentences)
	if len(sentence) > 0 && sentence[0] >= 'A' && sentence[0] <= 'Z' {
		score *= 1.1
	}

	// Penalty for very short sentences
	if words < 5 {
		score *= 0.5
	}

	return score
}

// --- Utilities ---

// splitSentences splits text into sentences.
func splitSentences(text string) []string {
	// Split on sentence-ending punctuation followed by space/newline
	re := regexp.MustCompile(`([.!?])\s+`)
	parts := re.Split(text, -1)

	var sentences []string
	for _, s := range parts {
		s = strings.TrimSpace(s)
		if len(s) >= 10 { // Skip very short fragments
			sentences = append(sentences, s)
		}
	}
	return sentences
}

// extractText recursively extracts text content from an HTML node.
func extractText(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}
	if n.Type == html.ElementNode {
		tag := strings.ToLower(n.Data)
		// Skip noise
		if noiseTags[tag] {
			return ""
		}
		// Add newlines for block elements
		blockElements := map[string]bool{
			"p": true, "div": true, "br": true, "li": true,
			"h1": true, "h2": true, "h3": true, "h4": true, "h5": true, "h6": true,
			"tr": true, "blockquote": true, "pre": true,
		}
		var text strings.Builder
		isBlock := blockElements[tag]
		if isBlock && tag != "br" {
			text.WriteString("\n")
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			text.WriteString(extractText(c))
		}
		if isBlock {
			text.WriteString("\n")
		}
		return text.String()
	}
	var text strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		text.WriteString(extractText(c))
	}
	return text.String()
}

// stripTags removes all HTML tags and returns plain text.
func stripTags(rawHTML string) string {
	// Remove all tags
	re := regexp.MustCompile(`<[^>]+>`)
	text := re.ReplaceAllString(rawHTML, "")
	// Decode entities
	for old, new := range map[string]string{
		"&amp;": "&", "&lt;": "<", "&gt;": ">", "&quot;": `"`,
		"&nbsp;": " ", "&#39;": "'", "&apos;": "'",
	} {
		text = strings.ReplaceAll(text, old, new)
	}
	return cleanWhitespace(text)
}

// cleanWhitespace normalizes whitespace.
func cleanWhitespace(text string) string {
	// Replace multiple newlines with double newline
	re := regexp.MustCompile(`\n{3,}`)
	text = re.ReplaceAllString(text, "\n\n")
	// Replace tabs and multiple spaces
	re2 := regexp.MustCompile(`[ \t]{2,}`)
	text = re2.ReplaceAllString(text, " ")
	return strings.TrimSpace(text)
}

// --- Public API ---

// FullPipeline runs all 3 layers on raw HTML.
// query: search keywords for distillation (empty = no filtering)
// mode: distillation mode
func FullPipeline(rawHTML, query string, mode DistillMode) *DistillResult {
	originalSize := len(rawHTML)

	// Layer 1: Clean HTML
	cleaned := CleanHTML(rawHTML)

	// Layer 2: Extract content
	title, content := ExtractContent(cleaned)

	// Layer 3: Distill by query
	result := Distill(content, query, mode)
	result.Title = title
	result.OriginalSize = originalSize

	return result
}

// DistillText runs distillation on already-extracted text (no HTML parsing).
func DistillText(text, query string, mode DistillMode) *DistillResult {
	return Distill(text, query, mode)
}

// CompactDistill is a shorthand for aggressive distillation.
func CompactDistill(text, query string) string {
	result := Distill(text, query, ModeCompact)
	return result.Content
}

// SmartDistill is a shorthand for smart distillation.
func SmartDistill(text, query string) string {
	result := Distill(text, query, ModeSmart)
	return result.Content
}
