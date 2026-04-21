package evolution

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ContractResult holds the outcome of a contract validation run.
type ContractResult struct {
	Passed  bool            `json:"passed"`
	Steps   []ContractStep  `json:"steps"`
	Summary string          `json:"summary"`
}

// ContractStep represents one validation step.
type ContractStep struct {
	Name    string `json:"name"`
	Passed  bool   `json:"passed"`
	Output  string `json:"output,omitempty"`
	Error   string `json:"error,omitempty"`
}

// RunContract executes validation checks after code changes.
// If target is "all" or empty, the full project is validated.
func RunContract(projectDir string, target string) (ContractResult, error) {
	projectDir = filepath.Clean(projectDir)
	result := ContractResult{Passed: true}

	goPath := findGoPath()

	// Step 1: go build
	step1 := runStep("go_build", func() (string, error) {
		return execGo(projectDir, goPath, "build", "-o", "/dev/null", "./cmd/aigo/")
	})
	result.Steps = append(result.Steps, step1)
	if !step1.Passed {
		result.Passed = false
		result.Summary = "Build failed"
		return result, nil
	}

	// Step 2: go vet
	step2 := runStep("go_vet", func() (string, error) {
		return execGo(projectDir, goPath, "vet", "./...")
	})
	result.Steps = append(result.Steps, step2)
	if !step2.Passed {
		result.Passed = false
		result.Summary = "go vet found issues"
		return result, nil
	}

	// Step 3: Verify tool registration functions compile (already covered by go build,
	// but we do a targeted check for Register*Tools functions)
	step3 := runStep("register_check", func() (string, error) {
		return checkRegisterFunctions(projectDir)
	})
	result.Steps = append(result.Steps, step3)
	if !step3.Passed {
		result.Passed = false
		result.Summary = "Tool registration check failed"
		return result, nil
	}

	result.Summary = "All checks passed"
	return result, nil
}

func runStep(name string, fn func() (string, error)) ContractStep {
	output, err := fn()
	step := ContractStep{Name: name}
	if err != nil {
		step.Passed = false
		step.Output = output
		step.Error = err.Error()
	} else {
		step.Passed = true
		step.Output = output
	}
	return step
}

func execGo(dir, goPath string, args ...string) (string, error) {
	cmd := exec.Command(goPath, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "PATH="+os.Getenv("PATH")+":/usr/local/go/bin")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	output := stdout.String() + stderr.String()
	if err != nil {
		return output, fmt.Errorf("%s", strings.TrimSpace(output))
	}
	return strings.TrimSpace(output), nil
}

func findGoPath() string {
	// Check common locations
	candidates := []string{
		"/usr/local/go/bin/go",
		"/usr/bin/go",
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	return "go" // fallback to PATH
}

func checkRegisterFunctions(projectDir string) (string, error) {
	// Look for Register*Tools functions and ensure they reference valid tool structs
	// by compiling the package that contains them.
	// go build already covers this, so we just list them.
	toolsDir := filepath.Join(projectDir, "internal")
	entries, err := os.ReadDir(toolsDir)
	if err != nil {
		return "", fmt.Errorf("reading internal dir: %w", err)
	}

	var registerPkgs []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		pkgDir := filepath.Join(toolsDir, e.Name())
		files, _ := os.ReadDir(pkgDir)
		for _, f := range files {
			if strings.HasSuffix(f.Name(), "tools.go") {
				registerPkgs = append(registerPkgs, "github.com/hermes-v2/aigo/internal/"+e.Name())
			}
		}
	}

	if len(registerPkgs) == 0 {
		return "No tool packages found", nil
	}

	// Verify each tool package compiles
	var msgs []string
	for _, pkg := range registerPkgs {
		rel := strings.TrimPrefix(pkg, "github.com/hermes-v2/aigo/")
		cmd := exec.Command(findGoPath(), "build", "-o", "/dev/null", "./"+rel+"/")
		cmd.Dir = projectDir
		cmd.Env = append(os.Environ(), "PATH="+os.Getenv("PATH")+":/usr/local/go/bin")
		out, err := cmd.CombinedOutput()
		if err != nil {
			return string(out), fmt.Errorf("compiling %s: %w", rel, err)
		}
		msgs = append(msgs, fmt.Sprintf("  ✓ %s", rel))
	}

	return "All tool packages compile:\n" + strings.Join(msgs, "\n"), nil
}
