package opencode

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/ahmad-ubaidillah/aigo/pkg/types"
)

type Client struct {
	binary  string
	timeout time.Duration
	workdir string
}

func NewClient(binary string, timeoutSec int, workdir string) (*Client, error) {
	if binary == "" {
		var err error
		binary, err = LookPath("opencode")
		if err != nil {
			return nil, fmt.Errorf("detect opencode binary: %w", err)
		}
	}
	if !IsExecutable(binary) {
		return nil, fmt.Errorf("opencode binary not executable: %s", binary)
	}
	return &Client{
		binary:  binary,
		timeout: time.Duration(timeoutSec) * time.Second,
		workdir: workdir,
	}, nil
}

func DetectBinary(hint string) (string, error) {
	if hint != "" {
		if IsExecutable(hint) {
			return hint, nil
		}
	}
	path, err := LookPath("opencode")
	if err == nil {
		return path, nil
	}
	common := []string{"/usr/local/bin/opencode", "/usr/bin/opencode"}
	for _, p := range common {
		if IsExecutable(p) {
			return p, nil
		}
	}
	return "", fmt.Errorf("opencode binary not found")
}

func (c *Client) Run(ctx context.Context, prompt string, sessionID string) (*types.ToolResult, error) {
	return c.execCommand(ctx, c.buildArgs(sessionID, prompt, nil))
}

func (c *Client) RunWithFiles(ctx context.Context, prompt string, sessionID string, files []string) (*types.ToolResult, error) {
	return c.execCommand(ctx, c.buildArgs(sessionID, prompt, files))
}

func (c *Client) CheckVersion() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, c.binary, "--version")
	cmd.Dir = c.workdir
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("check version: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func (c *Client) HealthCheck() (*types.ToolResult, error) {
	version, err := c.CheckVersion()
	if err != nil {
		return &types.ToolResult{
			Success: false,
			Error:   fmt.Sprintf("version check failed: %v", err),
		}, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, c.binary, "run", "--session=test", "--prompt=say hello")
	cmd.Dir = c.workdir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return &types.ToolResult{
			Success:  false,
			Output:   string(out),
			Error:    fmt.Sprintf("health test failed: %v", err),
			Metadata: map[string]string{"version": version},
		}, nil
	}
	return &types.ToolResult{
		Success:  true,
		Output:   string(out),
		Metadata: map[string]string{"version": version},
	}, nil
}

func (c *Client) StreamOutput(ctx context.Context, prompt string, sessionID string, callback func(line string)) error {
	args := c.buildArgs(sessionID, prompt, nil)
	cmd := exec.CommandContext(ctx, c.binary, args...)
	cmd.Dir = c.workdir
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("create stdout pipe: %w", err)
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start command: %w", err)
	}
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		callback(scanner.Text())
	}
	return cmd.Wait()
}

func (c *Client) buildArgs(sessionID string, prompt string, files []string) []string {
	args := []string{
		"run",
		"--session=" + sessionID,
		"--prompt=" + prompt,
	}
	if len(files) > 0 {
		args = append(args, "--files="+strings.Join(files, ","))
	}
	return args
}

func (c *Client) execCommand(ctx context.Context, args []string) (*types.ToolResult, error) {
	if c.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.timeout)
		defer cancel()
	}
	cmd := exec.CommandContext(ctx, c.binary, args...)
	cmd.Dir = c.workdir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	result := &types.ToolResult{
		Success: err == nil,
		Output:  stdout.String(),
	}
	if err != nil {
		result.Error = stderr.String()
		if result.Error == "" {
			result.Error = err.Error()
		}
	}
	return result, nil
}

func LookPath(name string) (string, error) {
	return exec.LookPath(name)
}

func IsExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	if info.IsDir() {
		return false
	}
	mode := info.Mode()
	if mode&0111 == 0 {
		return false
	}
	return filepath.IsAbs(path) || LookPathMust(path)
}

func LookPathMust(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}
