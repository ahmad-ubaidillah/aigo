package git

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type Repo struct {
	Dir string
}

type Status struct {
	Branch     string
	Staged     []string
	Modified   []string
	Untracked  []string
	Conflicts  []string
}

type Diff struct {
	File    string
	Content string
	Stats  DiffStats
}

type DiffStats struct {
	Added   int
	Removed int
	Files   int
}

type Commit struct {
	Hash    string
	Message string
	Author  string
	Date    time.Time
}

func New(dir string) (*Repo, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(filepath.Join(absDir, ".git")); err != nil {
		return nil, fmt.Errorf("not a git repo: %w", err)
	}

	return &Repo{Dir: absDir}, nil
}

func (r *Repo) Status() (*Status, error) {
	cmd := exec.Command("git", "status", "--porcelain", "-b")
	cmd.Dir = r.Dir
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	status := &Status{
		Branch:    "main",
		Staged:    []string{},
		Modified:  []string{},
		Untracked:  []string{},
		Conflicts:  []string{},
	}

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if len(line) < 3 {
			continue
		}

		if strings.HasPrefix(line, "## ") {
			branch := strings.TrimPrefix(line, "## ")
			branch = strings.Split(branch, "...")[0]
			branch = strings.TrimSpace(branch)
			status.Branch = branch
			continue
		}

		index := line[:2]
		file := strings.TrimSpace(line[3:])

		switch index {
		case "M ", "A ", "D ", "R ", "C ":
			status.Staged = append(status.Staged, file)
		case " M", " D":
			status.Modified = append(status.Modified, file)
		case "??":
			status.Untracked = append(status.Untracked, file)
		case "UU", "AA", "DD":
			status.Conflicts = append(status.Conflicts, file)
		}
	}

	return status, nil
}

func (r *Repo) Diff(staged bool) ([]Diff, error) {
	var cmd *exec.Cmd
	if staged {
		cmd = exec.Command("git", "diff", "--cached", "--stat")
	} else {
		cmd = exec.Command("git", "diff", "--stat")
	}
	cmd.Dir = r.Dir

	statOut, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var diffs []Diff

	statLines := strings.Split(string(statOut), "\n")
	for _, line := range statLines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "diff") || strings.HasPrefix(line, "index") {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) >= 3 {
			file := parts[0]
			if strings.HasSuffix(file, "|") {
				file = strings.TrimSuffix(file, "|")
			}
			file = strings.TrimSpace(file)

			diffs = append(diffs, Diff{File: file})
		}
	}

	if staged {
		cmd = exec.Command("git", "diff", "--cached")
	} else {
		cmd = exec.Command("git", "diff")
	}
	cmd.Dir = r.Dir

	fullDiff, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	fileDiffs := strings.Split(string(fullDiff), "diff --git")
	for _, fd := range fileDiffs {
		if fd == "" {
			continue
		}

		lines := strings.Split(fd, "\n")
		if len(lines) < 3 {
			continue
		}

		var file string
		for _, line := range lines {
			if strings.HasPrefix(line, "a/") {
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					file = strings.TrimPrefix(parts[1], "b/")
					break
				}
			}
		}

		for i := range diffs {
			if diffs[i].File == file || file == "" {
				diffs[i].Content = fd
				break
			}
		}
	}

	return diffs, nil
}

func (r *Repo) DiffFile(file string, staged bool) (string, error) {
	var cmd *exec.Cmd
	if staged {
		cmd = exec.Command("git", "diff", "--cached", "--", file)
	} else {
		cmd = exec.Command("git", "diff", "--", file)
	}
	cmd.Dir = r.Dir

	out, err := cmd.Output()
	return string(out), err
}

func (r *Repo) Stage(files ...string) error {
	if len(files) == 0 {
		cmd := exec.Command("git", "add", "-u")
		cmd.Dir = r.Dir
		_, err := cmd.Output()
		return err
	}

	args := append([]string{"add"}, files...)
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = r.Dir

	_, err := cmd.Output()
	return err
}

func (r *Repo) Unstage(files ...string) error {
	if len(files) == 0 {
		cmd := exec.Command("git", "reset", "HEAD")
		cmd.Dir = r.Dir
		_, err := cmd.Output()
		return err
	}

	args := append([]string{"reset", "HEAD", "--"}, files...)
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = r.Dir

	_, err := cmd.Output()
	return err
}

func (r *Repo) Commit(message string) (string, error) {
	cmd := exec.Command("git", "commit", "-m", message)
	cmd.Dir = r.Dir

	if _, err := cmd.Output(); err != nil {
		return "", err
	}

	hashCmd := exec.Command("git", "rev-parse", "HEAD")
	hashCmd.Dir = r.Dir
	hash, _ := hashCmd.Output()

	return strings.TrimSpace(string(hash)), nil
}

func (r *Repo) CommitAuto(message string) (bool, error) {
	status, err := r.Status()
	if err != nil {
		return false, err
	}

	if len(status.Staged) == 0 && len(status.Modified) == 0 && len(status.Untracked) == 0 {
		return false, nil
	}

	if len(status.Untracked) > 0 {
		r.Stage(status.Untracked...)
	}

	if len(status.Modified) > 0 {
		r.Stage(status.Modified...)
	}

	hash, err := r.Commit(message)
	if err != nil {
		return false, err
	}

	fmt.Printf("Committed: %s\n", hash[:8])
	return true, nil
}

func (r *Repo) Log(limit int) ([]Commit, error) {
	if limit <= 0 {
		limit = 10
	}

	cmd := exec.Command("git", "log", fmt.Sprintf("-%d", limit), "--pretty=format:%H|%s|%an|%ai")
	cmd.Dir = r.Dir

	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var commits []Commit
	lines := strings.Split(string(out), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, "|", 4)
		if len(parts) < 4 {
			continue
		}

		date, _ := time.Parse("2006-01-02 15:04:05 -0700", parts[3])

		commits = append(commits, Commit{
			Hash:    parts[0],
			Message: parts[1],
			Author:  parts[2],
			Date:    date,
		})
	}

	return commits, nil
}

func (r *Repo) Undo(count int) error {
	if count <= 0 {
		count = 1
	}

	cmd := exec.Command("git", fmt.Sprintf("~%d", count))
	cmd.Dir = r.Dir

	_, err := cmd.Output()
	return err
}

func (r *Repo) Checkout(branch string) error {
	cmd := exec.Command("git", "checkout", branch)
	cmd.Dir = r.Dir

	_, err := cmd.Output()
	return err
}

func (r *Repo) Branch(name string, create bool) error {
	if create {
		cmd := exec.Command("git", "checkout", "-b", name)
		cmd.Dir = r.Dir
		_, err := cmd.Output()
		return err
	}

	cmd := exec.Command("git", "branch", "-d", name)
	cmd.Dir = r.Dir

	_, err := cmd.Output()
	return err
}

func (r *Repo) Branches() ([]string, error) {
	cmd := exec.Command("git", "branch", "-a")
	cmd.Dir = r.Dir

	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var branches []string
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		line = strings.TrimPrefix(line, "* ")
		branches = append(branches, line)
	}

	return branches, nil
}

func (r *Repo) CurrentBranch() (string, error) {
	cmd := exec.Command("git", "branch", "--show-current")
	cmd.Dir = r.Dir

	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(out)), nil
}

func (r *Repo) AddAll() error {
	cmd := exec.Command("git", "add", "-A")
	cmd.Dir = r.Dir

	_, err := cmd.Output()
	return err
}

func (r *Repo) Discard(file string) error {
	cmd := exec.Command("git", "checkout", "--", file)
	cmd.Dir = r.Dir

	_, err := cmd.Output()
	return err
}

func (r *Repo) DiscardAll() error {
	cmd := exec.Command("git", "checkout", "--", ".")
	cmd.Dir = r.Dir

	_, err := cmd.Output()
	return err
}

func (r *Repo) Stash() error {
	cmd := exec.Command("git", "stash", "-u")
	cmd.Dir = r.Dir

	_, err := cmd.Output()
	return err
}

func (r *Repo) StashPop() error {
	cmd := exec.Command("git", "stash", "pop")
	cmd.Dir = r.Dir

	_, err := cmd.Output()
	return err
}

func (r *Repo) IsDirty() (bool, error) {
	status, err := r.Status()
	if err != nil {
		return false, err
	}

	return len(status.Staged) > 0 || len(status.Modified) > 0 || len(status.Untracked) > 0, nil
}

func (r *Repo) Files() ([]string, error) {
	cmd := exec.Command("git", "ls-files")
	cmd.Dir = r.Dir

	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var files []string
	for _, f := range strings.Split(string(out), "\n") {
		f = strings.TrimSpace(f)
		if f != "" {
			files = append(files, f)
		}
	}

	return files, nil
}

func (r *Repo) LastCommit() (*Commit, error) {
	cmd := exec.Command("git", "log", "-1", "--pretty=format:%H|%s|%an|%ai")
	cmd.Dir = r.Dir

	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	parts := strings.SplitN(string(out), "|", 4)
	if len(parts) < 4 {
		return nil, fmt.Errorf("invalid commit format")
	}

	date, _ := time.Parse("2006-01-02 15:04:05 -0700", parts[3])

	return &Commit{
		Hash:    parts[0],
		Message: parts[1],
		Author:  parts[2],
		Date:    date,
	}, nil
}

func (r *Repo) HasUncommitted() (bool, error) {
	status, err := r.Status()
	if err != nil {
		return false, err
	}

	return len(status.Modified) > 0 || len(status.Untracked) > 0 || len(status.Staged) > 0, nil
}

func (r *Repo) Root() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Dir = r.Dir

	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(out)), nil
}

func (r *Repo) DiffSummary() (string, error) {
	cmd := exec.Command("git", "diff", "--stat")
	cmd.Dir = r.Dir

	out, err := cmd.Output()
	return string(out), err
}

func (r *Repo) StagedSummary() (string, error) {
	cmd := exec.Command("git", "diff", "--cached", "--stat")
	cmd.Dir = r.Dir

	out, err := cmd.Output()
	return string(out), err
}

func (r *Repo) Init() error {
	cmd := exec.Command("git", "init")
	cmd.Dir = r.Dir

	_, err := cmd.Output()
	return err
}

func InitRepo(dir string) (*Repo, error) {
	r := &Repo{Dir: dir}

	if _, err := os.Stat(filepath.Join(dir, ".git")); os.IsNotExist(err) {
		if err := r.Init(); err != nil {
			return nil, err
		}
	}

	return r, nil
}

func (r *Repo) Changes() (string, error) {
	cmd := exec.Command("git", "diff", "--stat", "--no-color")
	cmd.Dir = r.Dir

	var buf bytes.Buffer
	cmd.Stdout = &buf
	err := cmd.Run()

	return buf.String(), err
}