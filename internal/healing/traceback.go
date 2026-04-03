package healing

import (
	"fmt"
	"regexp"
	"strings"
)

// StackFrame represents a single frame in a stack trace.
type StackFrame struct {
	Function string
	File     string
	Line     int
	Column   int
	Code     string
}

// ParsedTraceback holds the result of parsing a stack trace.
type ParsedTraceback struct {
	Language     string
	ErrorMessage string
	ErrorType    string
	Frames       []StackFrame
	Raw          string
}

// TracebackParser parses stack traces from multiple languages.
type TracebackParser struct {
	pythonRE *regexp.Regexp
	goRE     *regexp.Regexp
	jsRE     *regexp.Regexp
	rustRE   *regexp.Regexp
	javaRE   *regexp.Regexp
}

// NewTracebackParser creates a new traceback parser.
func NewTracebackParser() *TracebackParser {
	return &TracebackParser{
		pythonRE: regexp.MustCompile(`File "([^"]+)", line (\d+)(?:, in (\w+))?`),
		goRE:     regexp.MustCompile(`(\S+)\n\t([^:]+):(\d+)`),
		jsRE:     regexp.MustCompile(`at\s+(?:(\S+)\s+\()?([^:]+):(\d+):(\d+)\)?`),
		rustRE:   regexp.MustCompile(`at\s+([^:]+):(\d+):(\d+)`),
		javaRE:   regexp.MustCompile(`at\s+([\w.$]+)\(([^:]+):(\d+)\)`),
	}
}

// Parse detects the language and parses a stack trace.
func (p *TracebackParser) Parse(traceback string) *ParsedTraceback {
	result := &ParsedTraceback{Raw: traceback}

	if strings.Contains(traceback, "Traceback (most recent call last)") {
		result.Language = "python"
		p.parsePython(traceback, result)
	} else if strings.Contains(traceback, "goroutine") || strings.Contains(traceback, ".go:") {
		result.Language = "go"
		p.parseGo(traceback, result)
	} else if strings.Contains(traceback, "at ") && strings.Contains(traceback, ".js:") {
		result.Language = "javascript"
		p.parseJS(traceback, result)
	} else if strings.Contains(traceback, "thread 'main' panicked") || strings.Contains(traceback, "src/") {
		result.Language = "rust"
		p.parseRust(traceback, result)
	} else if strings.Contains(traceback, "Exception in thread") || strings.Contains(traceback, ".java:") {
		result.Language = "java"
		p.parseJava(traceback, result)
	} else {
		result.Language = "unknown"
		p.extractGenericError(traceback, result)
	}

	return result
}

func (p *TracebackParser) parsePython(traceback string, result *ParsedTraceback) {
	lines := strings.Split(traceback, "\n")
	for _, line := range lines {
		matches := p.pythonRE.FindStringSubmatch(line)
		if len(matches) >= 3 {
			frame := StackFrame{File: matches[1]}
			fmt.Sscanf(matches[2], "%d", &frame.Line)
			if len(matches) > 3 && matches[3] != "" {
				frame.Function = matches[3]
			}
			result.Frames = append(result.Frames, frame)
		}
	}
	if len(lines) > 0 {
		lastLine := lines[len(lines)-1]
		if strings.Contains(lastLine, ":") {
			parts := strings.SplitN(lastLine, ":", 2)
			if len(parts) == 2 {
				result.ErrorType = strings.TrimSpace(parts[0])
				result.ErrorMessage = strings.TrimSpace(parts[1])
			}
		}
	}
}

func (p *TracebackParser) parseGo(traceback string, result *ParsedTraceback) {
	lines := strings.Split(traceback, "\n")
	for i := 0; i < len(lines)-1; i++ {
		if strings.HasPrefix(lines[i], "goroutine") {
			continue
		}
		if strings.HasPrefix(lines[i], "\t") && i > 0 {
			matches := p.goRE.FindStringSubmatch(lines[i])
			if len(matches) >= 4 {
				frame := StackFrame{
					Function: lines[i-1],
					File:     strings.TrimSpace(matches[1]),
				}
				fmt.Sscanf(matches[2], "%d", &frame.Line)
				result.Frames = append(result.Frames, frame)
			}
		}
	}
	lines = strings.Split(traceback, "\n")
	for _, line := range lines {
		if strings.Contains(line, "panic:") || strings.Contains(line, "error:") {
			result.ErrorMessage = strings.TrimSpace(line)
			break
		}
	}
}

func (p *TracebackParser) parseJS(traceback string, result *ParsedTraceback) {
	matches := p.jsRE.FindAllStringSubmatch(traceback, -1)
	for _, m := range matches {
		frame := StackFrame{}
		if len(m) > 1 && m[1] != "" {
			frame.Function = m[1]
		}
		if len(m) > 2 {
			frame.File = m[2]
		}
		if len(m) > 3 {
			fmt.Sscanf(m[3], "%d", &frame.Line)
		}
		if len(m) > 4 {
			fmt.Sscanf(m[4], "%d", &frame.Column)
		}
		result.Frames = append(result.Frames, frame)
	}
	lines := strings.Split(traceback, "\n")
	for _, line := range lines {
		if strings.Contains(line, "Error:") || strings.Contains(line, "TypeError:") || strings.Contains(line, "ReferenceError:") {
			result.ErrorMessage = strings.TrimSpace(line)
			break
		}
	}
}

func (p *TracebackParser) parseRust(traceback string, result *ParsedTraceback) {
	matches := p.rustRE.FindAllStringSubmatch(traceback, -1)
	for _, m := range matches {
		frame := StackFrame{File: m[1]}
		fmt.Sscanf(m[2], "%d", &frame.Line)
		if len(m) > 3 {
			fmt.Sscanf(m[3], "%d", &frame.Column)
		}
		result.Frames = append(result.Frames, frame)
	}
	if strings.Contains(traceback, "panicked at") {
		idx := strings.Index(traceback, "panicked at")
		rest := traceback[idx+len("panicked at"):]
		if endIdx := strings.Index(rest, "\n"); endIdx > 0 {
			result.ErrorMessage = strings.TrimSpace(rest[:endIdx])
		}
	}
}

func (p *TracebackParser) parseJava(traceback string, result *ParsedTraceback) {
	matches := p.javaRE.FindAllStringSubmatch(traceback, -1)
	for _, m := range matches {
		frame := StackFrame{Function: m[1], File: m[2]}
		fmt.Sscanf(m[3], "%d", &frame.Line)
		result.Frames = append(result.Frames, frame)
	}
	lines := strings.Split(traceback, "\n")
	for _, line := range lines {
		if strings.Contains(line, "Exception") || strings.Contains(line, "Error:") {
			result.ErrorMessage = strings.TrimSpace(line)
			break
		}
	}
}

func (p *TracebackParser) extractGenericError(traceback string, result *ParsedTraceback) {
	lines := strings.Split(traceback, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "error") || strings.Contains(line, "Error") || strings.Contains(line, "failed") {
			result.ErrorMessage = line
			break
		}
	}
}
