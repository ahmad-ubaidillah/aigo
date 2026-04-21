package git

import (
	"context"
	"fmt"

	"github.com/hermes-v2/aigo/internal/tools"
)

type GitTool struct {
	repo *Repo
}

func NewGitTool(dir string) (*GitTool, error) {
	repo, err := New(dir)
	if err != nil {
		return nil, err
	}
	return &GitTool{repo: repo}, nil
}

func (t *GitTool) Name() string { return "git_status" }
func (t *GitTool) Description() string {
	return "Get git repository status (branch, staged, modified, untracked files)"
}

func (t *GitTool) Annotations() tools.Annotations {
	return tools.Annotations{ReadOnly: true}
}

// Schema implements tools.Tool.
func (t *GitTool) Schema() tools.ToolSchema {
	return tools.ToolSchema{
		Type: "function",
		Function: tools.ToolFunctionSchema{
			Name:        "git_status",
			Description: t.Description(),
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{},
			},
		},
	}
}

// Execute implements tools.Tool.
func (t *GitTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	status, err := t.repo.Status()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Branch: %s\nStaged: %v\nModified: %v\nUntracked: %v",
		status.Branch, status.Staged, status.Modified, status.Untracked), nil
}

// GitDiffTool shows git diff.
type GitDiffTool struct {
	repo *Repo
}

// NewGitDiffTool creates a new git diff tool.
func NewGitDiffTool(dir string) (*GitDiffTool, error) {
	repo, err := New(dir)
	if err != nil {
		return nil, err
	}
	return &GitDiffTool{repo: repo}, nil
}

func (t *GitDiffTool) Name() string { return "git_diff" }
func (t *GitDiffTool) Description() string {
	return "Get git diff for modified files"
}

func (t *GitDiffTool) Annotations() tools.Annotations {
	return tools.Annotations{ReadOnly: true}
}

func (t *GitDiffTool) Schema() tools.ToolSchema {
	return tools.ToolSchema{
		Type: "function",
		Function: tools.ToolFunctionSchema{
			Name:        "git_diff",
			Description: t.Description(),
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"staged": map[string]interface{}{
						"type":        "boolean",
						"description": "Show staged diff",
					},
				},
			},
		},
	}
}

func (t *GitDiffTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	staged, _ := args["staged"].(bool)
	diffs, err := t.repo.Diff(staged)
	if err != nil {
		return "", err
	}
	var result string
	for _, d := range diffs {
		result += fmt.Sprintf("=== %s ===\n%s\n", d.File, d.Content)
	}
	if result == "" {
		return "No changes", nil
	}
	return result, nil
}

// GitCommitTool commits changes.
type GitCommitTool struct {
	repo *Repo
}

// NewGitCommitTool creates a new git commit tool.
func NewGitCommitTool(dir string) (*GitCommitTool, error) {
	repo, err := New(dir)
	if err != nil {
		return nil, err
	}
	return &GitCommitTool{repo: repo}, nil
}

func (t *GitCommitTool) Name() string { return "git_commit" }
func (t *GitCommitTool) Description() string {
	return "Commit staged changes with a message"
}

func (t *GitCommitTool) Annotations() tools.Annotations {
	return tools.Annotations{Destructive: true}
}

func (t *GitCommitTool) Schema() tools.ToolSchema {
	return tools.ToolSchema{
		Type: "function",
		Function: tools.ToolFunctionSchema{
			Name:        "git_commit",
			Description: t.Description(),
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"message": map[string]interface{}{
						"type":        "string",
						"description": "Commit message",
					},
				},
				"required": []string{"message"},
			},
		},
	}
}

func (t *GitCommitTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	msg, ok := args["message"].(string)
	if !ok {
		return "", fmt.Errorf("message is required")
	}
	hash, err := t.repo.Commit(msg)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Committed: %s", hash[:8]), nil
}

// GitAutoCommitTool auto-commits all changes.
type GitAutoCommitTool struct {
	repo *Repo
}

// NewGitAutoCommitTool creates a new git auto-commit tool.
func NewGitAutoCommitTool(dir string) (*GitAutoCommitTool, error) {
	repo, err := New(dir)
	if err != nil {
		return nil, err
	}
	return &GitAutoCommitTool{repo: repo}, nil
}

func (t *GitAutoCommitTool) Name() string { return "git_commit_auto" }
func (t *GitAutoCommitTool) Description() string {
	return "Auto-commit all changes with an auto-generated message"
}

func (t *GitAutoCommitTool) Annotations() tools.Annotations {
	return tools.Annotations{Destructive: true}
}

func (t *GitAutoCommitTool) Schema() tools.ToolSchema {
	return tools.ToolSchema{
		Type: "function",
		Function: tools.ToolFunctionSchema{
			Name:        "git_commit_auto",
			Description: t.Description(),
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{},
			},
		},
	}
}

func (t *GitAutoCommitTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	committed, err := t.repo.CommitAuto("Auto-commit by Aigo")
	if err != nil {
		return "", err
	}
	if !committed {
		return "Nothing to commit", nil
	}
	return "Auto-committed", nil
}

// GitUndoTool undoes last N commits.
type GitUndoTool struct {
	repo *Repo
}

// NewGitUndoTool creates a new git undo tool.
func NewGitUndoTool(dir string) (*GitUndoTool, error) {
	repo, err := New(dir)
	if err != nil {
		return nil, err
	}
	return &GitUndoTool{repo: repo}, nil
}

func (t *GitUndoTool) Name() string { return "git_undo" }
func (t *GitUndoTool) Description() string {
	return "Undo the last N commits (soft reset)"
}

func (t *GitUndoTool) Annotations() tools.Annotations {
	return tools.Annotations{Destructive: true}
}

func (t *GitUndoTool) Schema() tools.ToolSchema {
	return tools.ToolSchema{
		Type: "function",
		Function: tools.ToolFunctionSchema{
			Name:        "git_undo",
			Description: t.Description(),
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"count": map[string]interface{}{
						"type":        "number",
						"description": "Number of commits to undo (default: 1)",
					},
				},
			},
		},
	}
}

func (t *GitUndoTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	count := 1
	if c, ok := args["count"].(float64); ok {
		count = int(c)
	}
	if err := t.repo.Undo(count); err != nil {
		return "", err
	}
	return fmt.Sprintf("Undone %d commit(s)", count), nil
}

// GitLogTool shows commit log.
type GitLogTool struct {
	repo *Repo
}

// NewGitLogTool creates a new git log tool.
func NewGitLogTool(dir string) (*GitLogTool, error) {
	repo, err := New(dir)
	if err != nil {
		return nil, err
	}
	return &GitLogTool{repo: repo}, nil
}

func (t *GitLogTool) Name() string { return "git_log" }
func (t *GitLogTool) Description() string {
	return "Show recent commit history"
}

func (t *GitLogTool) Annotations() tools.Annotations {
	return tools.Annotations{ReadOnly: true}
}

func (t *GitLogTool) Schema() tools.ToolSchema {
	return tools.ToolSchema{
		Type: "function",
		Function: tools.ToolFunctionSchema{
			Name:        "git_log",
			Description: t.Description(),
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"limit": map[string]interface{}{
						"type":        "number",
						"description": "Number of commits to show (default: 10)",
					},
				},
			},
		},
	}
}

func (t *GitLogTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	limit := 10
	if l, ok := args["limit"].(float64); ok {
		limit = int(l)
	}
	commits, err := t.repo.Log(limit)
	if err != nil {
		return "", err
	}
	var result string
	for _, c := range commits {
		result += fmt.Sprintf("%s | %s | %s\n", c.Hash[:8], c.Author, c.Message)
	}
	return result, nil
}

// GitBranchTool manages branches.
type GitBranchTool struct {
	repo *Repo
}

// NewGitBranchTool creates a new git branch tool.
func NewGitBranchTool(dir string) (*GitBranchTool, error) {
	repo, err := New(dir)
	if err != nil {
		return nil, err
	}
	return &GitBranchTool{repo: repo}, nil
}

func (t *GitBranchTool) Name() string { return "git_branch" }
func (t *GitBranchTool) Description() string {
	return "Manage git branches (list, create, checkout)"
}

func (t *GitBranchTool) Annotations() tools.Annotations {
	return tools.Annotations{Destructive: true}
}

func (t *GitBranchTool) Schema() tools.ToolSchema {
	return tools.ToolSchema{
		Type: "function",
		Function: tools.ToolFunctionSchema{
			Name:        "git_branch",
			Description: t.Description(),
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"action": map[string]interface{}{
						"type":        "string",
						"description": "Action: list, create, checkout, delete",
					},
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Branch name",
					},
				},
			},
		},
	}
}

func (t *GitBranchTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	action, _ := args["action"].(string)
	name, _ := args["name"].(string)

	switch action {
	case "list":
		branches, err := t.repo.Branches()
		if err != nil {
			return "", err
		}
		result := "Branches:\n"
		for _, b := range branches {
			result += fmt.Sprintf("  %s\n", b)
		}
		return result, nil
	case "create":
		if name == "" {
			return "", fmt.Errorf("name is required")
		}
		if err := t.repo.Branch(name, true); err != nil {
			return "", err
		}
		return fmt.Sprintf("Created branch: %s", name), nil
	case "checkout":
		if name == "" {
			return "", fmt.Errorf("name is required")
		}
		if err := t.repo.Checkout(name); err != nil {
			return "", err
		}
		return fmt.Sprintf("Checked out: %s", name), nil
	case "delete":
		if name == "" {
			return "", fmt.Errorf("name is required")
		}
		if err := t.repo.Branch(name, false); err != nil {
			return "", err
		}
		return fmt.Sprintf("Deleted branch: %s", name), nil
	default:
		branch, err := t.repo.CurrentBranch()
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("Current branch: %s", branch), nil
	}
}

// RegisterGitTools registers all git tools to the registry.
func RegisterGitTools(reg *tools.Registry, dir string) error {
	gitTool, err := NewGitTool(dir)
	if err != nil {
		return err
	}
	reg.Register(gitTool)

	diffTool, err := NewGitDiffTool(dir)
	if err != nil {
		return err
	}
	reg.Register(diffTool)

	commitTool, err := NewGitCommitTool(dir)
	if err != nil {
		return err
	}
	reg.Register(commitTool)

	autoCommitTool, err := NewGitAutoCommitTool(dir)
	if err != nil {
		return err
	}
	reg.Register(autoCommitTool)

	undoTool, err := NewGitUndoTool(dir)
	if err != nil {
		return err
	}
	reg.Register(undoTool)

	logTool, err := NewGitLogTool(dir)
	if err != nil {
		return err
	}
	reg.Register(logTool)

	branchTool, err := NewGitBranchTool(dir)
	if err != nil {
		return err
	}
	reg.Register(branchTool)

	return nil
}