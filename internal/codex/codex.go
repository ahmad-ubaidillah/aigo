package codex

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

type Symbol struct {
	Name      string `json:"name"`
	Kind      string `json:"kind"`
	File      string `json:"file"`
	Line      int    `json:"line"`
	Package   string `json:"package"`
	Export    bool   `json:"export"`
	Signature string `json:"signature,omitempty"`
	Refs      int    `json:"refs"`
}

type CodeIndex struct {
	mu         sync.RWMutex
	symbols    map[string][]Symbol
	byFile     map[string][]Symbol
	byPackage  map[string][]Symbol
	projectDir string
}

func New(projectDir string) (*CodeIndex, error) {
	absDir, err := filepath.Abs(projectDir)
	if err != nil {
		return nil, err
	}

	c := &CodeIndex{
		symbols:    make(map[string][]Symbol),
		byFile:     make(map[string][]Symbol),
		byPackage:  make(map[string][]Symbol),
		projectDir: absDir,
	}

	return c, nil
}

func (c *CodeIndex) Index() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	extensions := []string{".go", ".js", ".ts", ".tsx", ".jsx", ".py", ".java", ".rs", ".cpp", ".c", ".h"}

	err := filepath.Walk(c.projectDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() {
			// Skip common non-code dirs
			skip := []string{".git", "node_modules", "vendor", "dist", "build", "__pycache__", ".venv", "target"}
			for _, s := range skip {
				if info.Name() == s {
					return filepath.SkipDir
				}
			}
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		for _, e := range extensions {
			if ext == e {
				c.indexFile(path)
				break
			}
		}

		return nil
	})

	return err
}

func (c *CodeIndex) indexFile(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	content := string(data)
	relPath, _ := filepath.Rel(c.projectDir, path)
	lang := detectLanguage(path)

	var symbols []Symbol

	switch lang {
	case "go":
		symbols = c.extractGoSymbols(content, relPath)
	case "javascript", "typescript":
		symbols = c.extractJSSymbols(content, relPath)
	case "python":
		symbols = c.extractPythonSymbols(content, relPath)
	case "rust":
		symbols = c.extractRustSymbols(content, relPath)
	}

	for _, s := range symbols {
		c.symbols[s.Name] = append(c.symbols[s.Name], s)
		c.byFile[relPath] = append(c.byFile[relPath], s)

		if s.Package != "" {
			c.byPackage[s.Package] = append(c.byPackage[s.Package], s)
		}
	}
}

func detectLanguage(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	m := map[string]string{
		".go":     "go",
		".js":     "javascript",
		".ts":     "typescript",
		".tsx":    "typescript",
		".jsx":    "javascript",
		".py":     "python",
		".java":   "java",
		".rs":     "rust",
		".cpp":    "cpp",
		".c":      "c",
		".h":      "c",
		".rb":     "ruby",
		".php":    "php",
	}
	return m[ext]
}

func (c *CodeIndex) extractGoSymbols(content, file string) []Symbol {
	var symbols []Symbol
	lines := strings.Split(content, "\n")

	// Package
	pkgMatch := regexp.MustCompile(`^package (\w+)`).FindStringSubmatch(content)
	pkg := ""
	if len(pkgMatch) > 1 {
		pkg = pkgMatch[1]
	}

	// Type definitions
	typeRe := regexp.MustCompile(`type\s+(\w+)\s+(?:struct|interface)`)
	for _, match := range typeRe.FindAllStringSubmatch(content, -1) {
		if len(match) > 1 {
			line := c.findLineNumber(lines, match[0])
			symbols = append(symbols, Symbol{
				Name:    match[1],
				Kind:    "type",
				File:    file,
				Line:    line,
				Package: pkg,
				Export:  strings.HasPrefix(match[1], strings.ToUpper(match[1][:1])),
			})
		}
	}

	// Function definitions
	funcRe := regexp.MustCompile(`(?:func|func\s+)\s*(?:\((\w+)\s+\*?(\w+)\)\s+)?(\w+)\s*\(([^)]*)\)`)
	for _, match := range funcRe.FindAllStringSubmatch(content, -1) {
		if len(match) > 3 {
			line := c.findLineNumber(lines, match[0])
			symbols = append(symbols, Symbol{
				Name:      match[3],
				Kind:      "function",
				File:      file,
				Line:      line,
				Package:   pkg,
				Export:    strings.HasPrefix(match[3], strings.ToUpper(match[3][:1])),
				Signature: match[0],
			})
		}
	}

	// Variable declarations
	varRe := regexp.MustCompile(`(?:var|const)\s+(\w+)`)
	for _, match := range varRe.FindAllStringSubmatch(content, -1) {
		if len(match) > 1 {
			line := c.findLineNumber(lines, match[0])
			symbols = append(symbols, Symbol{
				Name:    match[1],
				Kind:    "variable",
				File:    file,
				Line:    line,
				Package: pkg,
				Export:  strings.HasPrefix(match[1], strings.ToUpper(match[1][:1])),
			})
		}
	}

	return symbols
}

func (c *CodeIndex) extractJSSymbols(content, file string) []Symbol {
	var symbols []Symbol
	lines := strings.Split(content, "\n")

	// Class definitions
	classRe := regexp.MustCompile(`class\s+(\w+)(?:\s+extends\s+(\w+))?`)
	for _, match := range classRe.FindAllStringSubmatch(content, -1) {
		if len(match) > 1 {
			line := c.findLineNumber(lines, match[0])
			symbols = append(symbols, Symbol{
				Name:    match[1],
				Kind:    "class",
				File:    file,
				Line:    line,
				Export:  true,
			})
		}
	}

	// Function definitions
	funcRe := regexp.MustCompile(`(?:function\s+(\w+)|(?:const|let|var)\s+(\w+)\s*=\s*(?:async\s*)?\([^)]*\)\s*=>|(\w+)\s*\([^)]*\)\s*\{)`)
	for _, match := range funcRe.FindAllStringSubmatch(content, -1) {
		for i := 1; i < len(match); i++ {
			if match[i] != "" && !strings.HasPrefix(match[i], " ") {
				line := c.findLineNumber(lines, match[0])
				symbols = append(symbols, Symbol{
					Name:   match[i],
					Kind:   "function",
					File:   file,
					Line:   line,
					Export: true,
				})
			}
		}
	}

	// Arrow functions
	arrowRe := regexp.MustCompile(`(?:const|let|var)\s+(\w+)\s*=\s*(?:async\s*)?\([^)]*\)\s*=>`)
	for _, match := range arrowRe.FindAllStringSubmatch(content, -1) {
		if len(match) > 1 {
			line := c.findLineNumber(lines, match[0])
			symbols = append(symbols, Symbol{
				Name:   match[1],
				Kind:   "function",
				File:   file,
				Line:   line,
				Export: true,
			})
		}
	}

	return symbols
}

func (c *CodeIndex) extractPythonSymbols(content, file string) []Symbol {
	var symbols []Symbol
	lines := strings.Split(content, "\n")

	// Class definitions
	classRe := regexp.MustCompile(`^class\s+(\w+)(?:\(([^)]+)\))?:`)
	for _, match := range classRe.FindAllStringSubmatch(content, -1) {
		if len(match) > 1 {
			line := c.findLineNumber(lines, match[0])
			symbols = append(symbols, Symbol{
				Name:   match[1],
				Kind:   "class",
				File:   file,
				Line:   line,
				Export: true,
			})
		}
	}

	// Function definitions
	funcRe := regexp.MustCompile(`^def\s+(\w+)\s*\(([^)]*)\):`)
	for _, match := range funcRe.FindAllStringSubmatch(content, -1) {
		if len(match) > 1 {
			line := c.findLineNumber(lines, match[0])
			symbols = append(symbols, Symbol{
				Name:      match[1],
				Kind:      "function",
				File:      file,
				Line:      line,
				Export:    !strings.HasPrefix(match[1], "_"),
				Signature: match[0],
			})
		}
	}

	return symbols
}

func (c *CodeIndex) extractRustSymbols(content, file string) []Symbol {
	var symbols []Symbol
	lines := strings.Split(content, "\n")

	// Struct definitions
	structRe := regexp.MustCompile(`(?:pub\s+)?struct\s+(\w+)`)
	for _, match := range structRe.FindAllStringSubmatch(content, -1) {
		if len(match) > 1 {
			line := c.findLineNumber(lines, match[0])
			symbols = append(symbols, Symbol{
				Name:   match[1],
				Kind:   "struct",
				File:   file,
				Line:   line,
				Export: true,
			})
		}
	}

	// Function definitions
	funcRe := regexp.MustCompile(`(?:pub\s+)?(?:async\s+)?fn\s+(\w+)`)
	for _, match := range funcRe.FindAllStringSubmatch(content, -1) {
		if len(match) > 1 {
			line := c.findLineNumber(lines, match[0])
			symbols = append(symbols, Symbol{
				Name:   match[1],
				Kind:   "function",
				File:   file,
				Line:   line,
				Export: true,
			})
		}
	}

	return symbols
}

func (c *CodeIndex) findLineNumber(lines []string, text string) int {
	for i, line := range lines {
		if strings.Contains(line, text) {
			return i + 1
		}
	}
	return 0
}

func (c *CodeIndex) Find(name string) []Symbol {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.symbols[name]
}

func (c *CodeIndex) FindInFile(file string) []Symbol {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.byFile[file]
}

func (c *CodeIndex) FindByPackage(pkg string) []Symbol {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.byPackage[pkg]
}

func (c *CodeIndex) FindKind(kind string) []Symbol {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var results []Symbol
	for _, syms := range c.symbols {
		for _, s := range syms {
			if s.Kind == kind {
				results = append(results, s)
			}
		}
	}
	return results
}

func (c *CodeIndex) MapError(errorMsg string) []Symbol {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var results []Symbol

	// Extract potential symbol names from error
	patterns := []string{
		`undefined: (\w+)`,
		`cannot find (\w+)`,
		`(\w+) is not defined`,
		`no such (\w+)`,
		`undefined method '(\w+)'`,
		`undefined variable '(\w+)'`,
		`'(\w+)' not found`,
		`cannot resolve (\w+)`,
		`(\w+) has no (\w+)`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindAllStringSubmatch(errorMsg, -1)
		for _, match := range matches {
			if len(match) > 1 {
				if syms, ok := c.symbols[match[1]]; ok {
					results = append(results, syms...)
				}
			}
		}
	}

	return results
}

func (c *CodeIndex) Export() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return map[string]interface{}{
		"symbols": c.symbols,
		"byFile":  c.byFile,
		"byPackage": c.byPackage,
		"count": len(c.symbols),
	}
}

func (c *CodeIndex) Save(path string) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	data, err := json.MarshalIndent(c.symbols, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func (c *CodeIndex) Load(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var symbols map[string][]Symbol
	if err := json.Unmarshal(data, &symbols); err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.symbols = symbols
	c.rebuildIndexes()

	return nil
}

func (c *CodeIndex) rebuildIndexes() {
	c.byFile = make(map[string][]Symbol)
	c.byPackage = make(map[string][]Symbol)

	for _, syms := range c.symbols {
		for _, s := range syms {
			c.byFile[s.File] = append(c.byFile[s.File], s)
			if s.Package != "" {
				c.byPackage[s.Package] = append(c.byPackage[s.Package], s)
			}
		}
	}
}

func (c *CodeIndex) Stats() map[string]int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := make(map[string]int)
	stats["total_symbols"] = len(c.symbols)

	kindCounts := make(map[string]int)
	fileCount := make(map[string]int)

	for _, syms := range c.symbols {
		for _, s := range syms {
			kindCounts[s.Kind]++
			fileCount[s.File]++
		}
	}

	for k, v := range kindCounts {
		stats["kind_"+k] = v
	}

	stats["indexed_files"] = len(fileCount)

	return stats
}