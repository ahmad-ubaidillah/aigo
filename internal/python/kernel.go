package python

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Kernel struct {
	mu        sync.RWMutex
	cmd       *exec.Cmd
	stdin     io.WriteCloser
	stdout    io.Reader
	stderr    io.Reader
	isRunning bool
	timeout   time.Duration
	packages  map[string]bool
	workDir   string
}

type ExecResult struct {
	Success     bool
	Output      string
	Error       string
	ReturnValue string
	Duration    time.Duration
}

type KernelOptions struct {
	Timeout  time.Duration
	WorkDir  string
	Packages []string
}

func NewKernel(opts *KernelOptions) *Kernel {
	k := &Kernel{
		timeout:  60 * time.Second,
		packages: make(map[string]bool),
		workDir:  opts.WorkDir,
	}
	if opts.Timeout > 0 {
		k.timeout = opts.Timeout
	}
	if k.workDir == "" {
		k.workDir = "/tmp/aigo-python"
	}
	return k
}

func (k *Kernel) Start() error {
	k.mu.Lock()
	defer k.mu.Unlock()

	if k.isRunning {
		return nil
	}

	if err := os.MkdirAll(k.workDir, 0755); err != nil {
		return fmt.Errorf("create workdir: %w", err)
	}

	script := k.generateStartupScript()
	scriptPath := filepath.Join(k.workDir, "kernel.py")
	if err := os.WriteFile(scriptPath, []byte(script), 0644); err != nil {
		return fmt.Errorf("write kernel script: %w", err)
	}

	cmd := exec.Command("python3", "-u", scriptPath)
	cmd.Dir = k.workDir

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start python: %w", err)
	}

	k.cmd = cmd
	k.stdin = stdin
	k.stdout = stdout
	k.stderr = stderr
	k.isRunning = true

	go k.readOutput()

	return nil
}

func (k *Kernel) generateStartupScript() string {
	return `
import sys
import json
import traceback

def run_code(code):
    try:
        local_ns = {}
        exec(code, {}, local_ns)
        result = local_ns.get('_result_', None)
        return {'success': True, 'output': str(result) if result else '', 'error': ''}
    except Exception as e:
        return {'success': False, 'output': '', 'error': str(e), 'traceback': traceback.format_exc()}

while True:
    try:
        line = sys.stdin.readline()
        if not line:
            break
        
        data = json.loads(line.strip())
        result = run_code(data.get('code', ''))
        
        response = json.dumps(result) + '\n'
        sys.stdout.write(response)
        sys.stdout.flush()
    except Exception as e:
        error_resp = json.dumps({'success': False, 'output': '', 'error': str(e)}) + '\n'
        sys.stdout.write(error_resp)
        sys.stdout.flush()
`
}

func (k *Kernel) readOutput() {
	scanner := bufio.NewScanner(k.stdout)
	for scanner.Scan() {
	}
}

func (k *Kernel) Execute(code string) (*ExecResult, error) {
	k.mu.RLock()
	if !k.isRunning {
		k.mu.RUnlock()
		return nil, fmt.Errorf("kernel not running")
	}
	k.mu.RUnlock()

	start := time.Now()

	input := map[string]string{"code": code}
	inputJSON, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("marshal: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), k.timeout)
	defer cancel()

	inputCh := make(chan error, 1)
	go func() {
		_, err := k.stdin.Write(append(append(inputJSON, '\n'), []byte("\n")...))
		inputCh <- err
	}()

	select {
	case <-ctx.Done():
		return &ExecResult{
			Success: false,
			Error:   "timeout",
		}, ctx.Err()
	case err := <-inputCh:
		if err != nil {
			return nil, fmt.Errorf("write stdin: %w", err)
		}
	}

	respCh := make(chan string, 1)
	go func() {
		reader := bufio.NewReader(k.stdout)
		line, err := reader.ReadString('\n')
		if err != nil {
			respCh <- ""
			return
		}
		respCh <- line
	}()

	select {
	case <-ctx.Done():
		return &ExecResult{
			Success: false,
			Error:   "timeout waiting for result",
		}, ctx.Err()
	case resp := <-respCh:
		var result ExecResult
		if err := json.Unmarshal([]byte(resp), &result); err != nil {
			return &ExecResult{
				Success: false,
				Error:   fmt.Sprintf("parse response: %v", err),
			}, nil
		}
		result.Duration = time.Since(start)
		return &result, nil
	}
}

func (k *Kernel) InstallPackage(pkg string) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	if k.packages[pkg] {
		return nil
	}

	cmd := exec.Command("pip", "install", pkg)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("install %s: %w", pkg, err)
	}

	k.packages[pkg] = true
	return nil
}

func (k *Kernel) InstallPackages(packages []string) error {
	for _, pkg := range packages {
		if err := k.InstallPackage(pkg); err != nil {
			return err
		}
	}
	return nil
}

func (k *Kernel) Stop() error {
	k.mu.Lock()
	defer k.mu.Unlock()

	if !k.isRunning {
		return nil
	}

	if k.cmd != nil && k.cmd.Process != nil {
		k.cmd.Process.Kill()
		k.cmd.Wait()
	}

	k.isRunning = false
	return nil
}

func (k *Kernel) IsRunning() bool {
	k.mu.RLock()
	defer k.mu.RUnlock()
	return k.isRunning
}

type SelfHealingExecutor struct {
	kernel     *Kernel
	maxRetries int
}

func NewSelfHealingExecutor(maxRetries int) *SelfHealingExecutor {
	return &SelfHealingExecutor{
		kernel:     NewKernel(&KernelOptions{}),
		maxRetries: maxRetries,
	}
}

func (s *SelfHealingExecutor) ExecuteWithHealing(code string) (*ExecResult, error) {
	if err := s.kernel.Start(); err != nil {
		return nil, err
	}
	defer s.kernel.Stop()

	var lastResult *ExecResult

	for attempt := 0; attempt <= s.maxRetries; attempt++ {
		result, err := s.kernel.Execute(code)

		if err != nil {
			return &ExecResult{
				Success: false,
				Error:   fmt.Sprintf("execution error: %v", err),
			}, err
		}

		if result.Success {
			return result, nil
		}

		lastResult = result

		if attempt < s.maxRetries {
			fixedCode := s.fixCode(code, result.Error)
			if fixedCode == code {
				break
			}
			code = fixedCode
		}
	}

	return &ExecResult{
		Success: false,
		Output:  lastResult.Output,
		Error:   fmt.Sprintf("failed after %d attempts: %s", s.maxRetries+1, lastResult.Error),
	}, nil
}

func (s *SelfHealingExecutor) fixCode(code, errorMsg string) string {
	errorMsg = strings.ToLower(errorMsg)

	if strings.Contains(errorMsg, "modulenotfounderror") {
		parts := strings.Split(errorMsg, "modulenotfounderror: no module named '")
		if len(parts) > 1 {
			module := strings.Split(parts[1], "'")[0]
			s.kernel.InstallPackage(module)
			return fmt.Sprintf("import %s\n%s", module, code)
		}
	}

	if strings.Contains(errorMsg, "syntaxerror") {
		return s.fixSyntaxError(code, errorMsg)
	}

	if strings.Contains(errorMsg, "attributeerror") {
		return s.fixAttributeError(code, errorMsg)
	}

	if strings.Contains(errorMsg, "nameerror") {
		return s.fixNameError(code, errorMsg)
	}

	return code
}

func (s *SelfHealingExecutor) fixSyntaxError(code, errorMsg string) string {
	lines := strings.Split(code, "\n")
	var fixed []string

	for _, line := range lines {
		if strings.Contains(line, "==") && strings.Contains(line, "=") && !strings.Contains(line, "==") {
			idx := strings.Index(line, "=")
			if idx > 0 && line[idx-1] != '=' && line[idx+1] != '=' {
				line = line[:idx] + "==" + line[idx+1:]
			}
		}
		fixed = append(fixed, line)
	}

	return strings.Join(fixed, "\n")
}

func (s *SelfHealingExecutor) fixAttributeError(code, errorMsg string) string {
	parts := strings.Split(errorMsg, "'")
	if len(parts) > 2 {
		obj := parts[1]
		method := parts[3]

		if method == "append" && strings.Contains(code, obj+"+") {
			code = strings.ReplaceAll(code, obj+"+", obj+".append(")
			code = strings.ReplaceAll(code, "\n)", ")")
		}
	}

	return code
}

func (s *SelfHealingExecutor) fixNameError(code, errorMsg string) string {
	parts := strings.Split(errorMsg, "name '")
	if len(parts) > 1 {
		varName := strings.Split(parts[1], "'")[0]
		code = fmt.Sprintf("%s = None\n%s", varName, code)
	}

	return code
}

func QuickExec(code string) (*ExecResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "python3", "-c", code)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	if err != nil {
		return &ExecResult{
			Success: false,
			Output:  stdout.String(),
			Error:   stderr.String(),
		}, nil
	}

	return &ExecResult{
		Success: true,
		Output:  stdout.String(),
	}, nil
}
