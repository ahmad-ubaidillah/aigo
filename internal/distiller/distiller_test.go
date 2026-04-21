package distiller

import (
	"fmt"
	"strings"
	"testing"
)

func TestCleanHTML(t *testing.T) {
	rawHTML := `<!DOCTYPE html>
<html>
<head>
	<title>Test Page</title>
	<script>alert('ads')</script>
	<style>body{color:red}</style>
</head>
<body>
	<nav class="main-menu">Menu items here</nav>
	<header>Site Header</header>
	<div class="sidebar">Sidebar content</div>
	<article>
		<h1>Article Title</h1>
		<p>This is the main content of the article.</p>
		<p>Go is a statically typed, compiled language.</p>
		<p>It has garbage collection and CSP-style concurrency.</p>
	</article>
	<aside>Related links</aside>
	<footer>Copyright 2024</footer>
	<!-- This is a comment -->
	<div class="cookie-consent">Accept cookies?</div>
</body>
</html>`

	cleaned := CleanHTML(rawHTML)

	// Should NOT contain noise
	if strings.Contains(cleaned, "alert('ads')") {
		t.Error("script content not removed")
	}
	if strings.Contains(cleaned, "body{color:red}") {
		t.Error("style content not removed")
	}
	if strings.Contains(cleaned, "Menu items here") {
		t.Error("nav content not removed")
	}
	if strings.Contains(cleaned, "Accept cookies") {
		t.Error("cookie consent not removed")
	}
	if strings.Contains(cleaned, "Copyright 2024") {
		t.Error("footer not removed")
	}

	// Should contain content
	if !strings.Contains(cleaned, "main content") {
		t.Error("main content was removed")
	}
	if !strings.Contains(cleaned, "Article Title") {
		t.Error("article title was removed")
	}
}

func TestExtractContent(t *testing.T) {
	rawHTML := `<html>
<body>
	<nav>Navigation</nav>
	<div class="content">
		<h1>How to Learn Go</h1>
		<article>
			<p>Go was designed at Google by Robert Griesemer, Rob Pike, and Ken Thompson.</p>
			<p>Go is statically typed and compiled.</p>
			<p>It has garbage collection and supports concurrent programming.</p>
		</article>
	</div>
	<footer>Footer</footer>
</body>
</html>`

	title, content := ExtractContent(rawHTML)

	fmt.Printf("Title: %s\n", title)
	fmt.Printf("Content: %s\n", content)

	if title == "" {
		t.Error("title not extracted")
	}

	if !strings.Contains(content, "Robert Griesemer") {
		t.Error("main content not extracted")
	}
}

func TestDistillSmart(t *testing.T) {
	text := `Go is a statically typed, compiled language designed at Google.
It has garbage collection, type safety, and CSP-style concurrency.
Python is an interpreted, high-level language.
JavaScript runs in the browser and Node.js.
Go's concurrency model uses goroutines and channels.
The Go compiler produces fast native binaries.
React is a JavaScript library for building user interfaces.
Docker is written in Go for container orchestration.`

	result := Distill(text, "Go concurrency", ModeSmart)

	fmt.Printf("Mode: %s\n", result.Mode)
	fmt.Printf("Sentences: %d→%d\n", result.SentencesIn, result.SentencesOut)
	fmt.Printf("Content: %s\n", result.Content)

	// Should keep Go-related sentences
	if !strings.Contains(result.Content, "goroutines") {
		t.Error("Go concurrency sentence filtered out")
	}
	if !strings.Contains(result.Content, "Robert") && !strings.Contains(result.Content, "Google") {
		// It's OK if this is filtered, not a keyword match
	}

	// Should filter out unrelated sentences
	if strings.Contains(result.Content, "React is") {
		t.Error("React sentence should be filtered (not related to Go)")
	}
}

func TestDistillCompact(t *testing.T) {
	text := `Go is a statically typed, compiled language.
Python is interpreted and high-level.
Go has goroutines and channels for concurrency.
JavaScript runs in browsers.
Go compiles to fast native binaries.
Ruby is a dynamic language.`

	result := Distill(text, "Go concurrency", ModeCompact)

	fmt.Printf("Mode: %s\n", result.Mode)
	fmt.Printf("Sentences: %d→%d\n", result.SentencesIn, result.SentencesOut)
	fmt.Printf("Content: %s\n", result.Content)

	if result.SentencesOut >= result.SentencesIn {
		t.Error("compact mode should filter more aggressively")
	}
}

func TestDistillResearch(t *testing.T) {
	text := `Sentence one. Sentence two. Sentence three.`
	result := Distill(text, "unrelated query", ModeResearch)

	if result.SentencesOut != result.SentencesIn {
		t.Error("research mode should keep all sentences")
	}
}

func TestFullPipeline(t *testing.T) {
	rawHTML := `<html>
<head><script>ads</script><style>css</style></head>
<body>
	<nav>Menu</nav>
	<article>
		<h1>Go Programming Guide</h1>
		<p>Go is a programming language created at Google.</p>
		<p>It features goroutines for concurrent execution.</p>
		<p>Python is also popular for web development.</p>
		<p>Go's garbage collector handles memory management.</p>
	</article>
	<footer>Footer</footer>
</body>
</html>`

	result := FullPipeline(rawHTML, "Go programming", ModeSmart)

	fmt.Printf("Pipeline result:\n")
	fmt.Printf("  Title: %s\n", result.Title)
	fmt.Printf("  Mode: %s\n", result.Mode)
	fmt.Printf("  Size: %d→%d bytes (%.0f%% compressed)\n",
		result.OriginalSize, result.FinalSize,
		(1.0-float64(result.FinalSize)/float64(result.OriginalSize))*100)
	fmt.Printf("  Sentences: %d→%d\n", result.SentencesIn, result.SentencesOut)
	fmt.Printf("  Content:\n%s\n", result.Content)

	if result.Title == "" {
		t.Error("title not extracted")
	}
	if !strings.Contains(result.Content, "goroutines") {
		t.Error("Go content not in result")
	}
	if result.FinalSize >= result.OriginalSize {
		t.Error("should be compressed")
	}
}

func TestExtractKeywords(t *testing.T) {
	keywords := extractKeywords("Go concurrency goroutines channels")

	fmt.Printf("Keywords: %v\n", keywords)

	foundGo := false
	foundConcurrency := false
	for _, kw := range keywords {
		if kw == "go" {
			foundGo = true
		}
		if kw == "concurrency" {
			foundConcurrency = true
		}
	}
	if !foundGo {
		t.Error("'go' not found in keywords")
	}
	if !foundConcurrency {
		t.Error("'concurrency' not found in keywords")
	}
}
