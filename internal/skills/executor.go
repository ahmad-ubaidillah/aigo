package skills

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/ahmad-ubaidillah/aigo/pkg/types"
)

type Executor struct {
	workspace string
	timeout   time.Duration
}

func NewExecutor(workspace string, timeout time.Duration) *Executor {
	if timeout == 0 {
		timeout = 60 * time.Second
	}
	return &Executor{
		workspace: workspace,
		timeout:   timeout,
	}
}

func (e *Executor) Execute(skillName, args string) (*types.SkillResult, error) {
	cmd := e.buildCommand(skillName, args)
	if cmd == nil {
		return nil, fmt.Errorf("unknown skill: %s", skillName)
	}

	if e.workspace != "" {
		cmd.Dir = e.workspace
	}

	cmd.Env = append(os.Environ(),
		"AIGO_SKILL="+skillName,
		"AIGO_ARGS="+args,
	)

	done := make(chan error, 1)
	var output string
	var err error

	go func() {
		out, runErr := cmd.CombinedOutput()
		output = string(out)
		done <- runErr
	}()

	select {
	case <-time.After(e.timeout):
		cmd.Process.Kill()
		return nil, fmt.Errorf("skill execution timed out after %v", e.timeout)
	case runErr := <-done:
		if runErr != nil {
			err = runErr
		}
	}

	return &types.SkillResult{
		Output:   output,
		Metadata: map[string]string{"skill": skillName, "args": args},
	}, err
}

func (e *Executor) buildCommand(skillName, args string) *exec.Cmd {
	skillName = strings.ToLower(skillName)

	switch skillName {
	case "git-master", "git":
		return e.buildGitCommand(args)
	case "playwright":
		return e.buildPlaywrightCommand(args)
	case "dev-browser", "browser":
		return e.buildBrowserCommand(args)
	case "code-review":
		return e.buildCodeReviewCommand(args)
	case "web-search":
		return e.buildWebSearchCommand(args)
	case "code-search":
		return e.buildCodeSearchCommand(args)
	default:
		return exec.Command("sh", "-c", skillName+" "+args)
	}
}

func (e *Executor) buildGitCommand(args string) *exec.Cmd {
	args = strings.TrimSpace(args)

	switch {
	case strings.HasPrefix(args, "commit"):
		msg := strings.TrimPrefix(args, "commit")
		msg = strings.TrimSpace(msg)
		if msg == "" {
			msg = "Auto commit via aigo"
		}
		return exec.Command("sh", "-c", fmt.Sprintf("git add -A && git commit -m '%s'", msg))
	case strings.HasPrefix(args, "push"):
		return exec.Command("git", "push")
	case strings.HasPrefix(args, "pull"):
		return exec.Command("git", "pull")
	case strings.HasPrefix(args, "branch"):
		branch := strings.TrimPrefix(args, "branch")
		branch = strings.TrimSpace(branch)
		if branch == "" {
			return exec.Command("git", "branch", "-a")
		}
		return exec.Command("git", "checkout", branch)
	case strings.HasPrefix(args, "status"):
		return exec.Command("git", "status")
	case strings.HasPrefix(args, "log"):
		return exec.Command("sh", "-c", "git log --oneline -10")
	default:
		return exec.Command("sh", "-c", "git "+args)
	}
}

func (e *Executor) buildPlaywrightCommand(args string) *exec.Cmd {
	args = strings.TrimSpace(args)

	if strings.HasPrefix(args, "test") {
		tests := strings.TrimPrefix(args, "test")
		tests = strings.TrimSpace(tests)
		if tests == "" {
			return exec.Command("npx", "playwright", "test")
		}
		return exec.Command("npx", "playwright", "test", tests)
	}

	if strings.HasPrefix(args, "install") {
		return exec.Command("npx", "playwright", "install")
	}

	if strings.HasPrefix(args, "screenshot") {
		url := strings.TrimPrefix(args, "screenshot")
		url = strings.TrimSpace(url)
		return exec.Command("npx", "playwright", "screenshot", url, "screenshot.png")
	}

	parts := strings.Fields(args)
	if len(parts) == 0 {
		return exec.Command("npx", "playwright")
	}
	cmd := exec.Command("npx", "playwright")
	cmd.Args = append(cmd.Args, parts...)
	return cmd
}

func (e *Executor) buildBrowserCommand(args string) *exec.Cmd {
	args = strings.TrimSpace(args)

	switch {
	case strings.HasPrefix(args, "open"):
		url := strings.TrimPrefix(args, "open")
		url = strings.TrimSpace(url)
		return exec.Command("open", url)
	case strings.HasPrefix(args, "screenshot"):
		url := strings.TrimPrefix(args, "screenshot")
		url = strings.TrimSpace(url)
		return exec.Command("sh", "-c", "curl -s "+url+" > page.html && cat page.html")
	default:
		return exec.Command("sh", "-c", args)
	}
}

func (e *Executor) buildCodeReviewCommand(args string) *exec.Cmd {
	args = strings.TrimSpace(args)

	if args == "" || args == "." {
		return exec.Command("sh", "-c", "echo 'Running code review...' && golangci-lint run ./... 2>&1 || echo 'No linter configured'")
	}

	return exec.Command("sh", "-c", args)
}

func (e *Executor) buildWebSearchCommand(args string) *exec.Cmd {
	args = strings.TrimSpace(args)
	return exec.Command("sh", "-c", fmt.Sprintf("curl -s 'https://api.duckduckgo.com/?q=%s&format=json' | jq -r '.AbstractText' 2>/dev/null || echo 'Search: %s'", args, args))
}

func (e *Executor) buildCodeSearchCommand(args string) *exec.Cmd {
	args = strings.TrimSpace(args)
	return exec.Command("sh", "-c", fmt.Sprintf("curl -s 'https://grep.app/search?q=%s&limit=5' | jq -r '.[]' 2>/dev/null || echo 'Search: %s'", args, args))
}
