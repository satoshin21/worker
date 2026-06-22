// Package git wraps the git CLI calls used by worker.
package git

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// Worktree represents an entry from `git worktree list --porcelain`.
type Worktree struct {
	Path   string
	Branch string // e.g. "refs/heads/foo" — empty for detached HEAD
	Head   string
}

// CommonDir returns the top-level directory of the main repository
// (the original clone, not a worktree).
func CommonDir() (string, error) {
	out, err := run("git", "rev-parse", "--path-format=absolute", "--git-common-dir")
	if err != nil {
		return "", err
	}
	gitCommonDir := strings.TrimSpace(out)
	return filepath.Dir(gitCommonDir), nil
}

// CurrentBranch returns the symbolic name of HEAD, or empty if detached.
func CurrentBranch() (string, error) {
	out, err := run("git", "symbolic-ref", "--quiet", "--short", "HEAD")
	if err != nil {
		if _, ok := errors.AsType[*exec.ExitError](err); ok {
			return "", nil
		}
		return "", err
	}
	return strings.TrimSpace(out), nil
}

// BranchExists reports whether a local branch with the given name exists.
func BranchExists(name string) (bool, error) {
	cmd := exec.Command("git", "show-ref", "--verify", "--quiet", "refs/heads/"+name)
	err := cmd.Run()
	if err == nil {
		return true, nil
	}
	if _, ok := errors.AsType[*exec.ExitError](err); ok {
		return false, nil
	}
	return false, err
}

// ListWorktrees parses `git worktree list --porcelain`.
func ListWorktrees() ([]Worktree, error) {
	out, err := run("git", "worktree", "list", "--porcelain")
	if err != nil {
		return nil, err
	}
	var result []Worktree
	var cur Worktree
	flush := func() {
		if cur.Path != "" {
			result = append(result, cur)
		}
		cur = Worktree{}
	}
	for line := range strings.SplitSeq(out, "\n") {
		if line == "" {
			flush()
			continue
		}
		switch {
		case strings.HasPrefix(line, "worktree "):
			cur.Path = strings.TrimPrefix(line, "worktree ")
		case strings.HasPrefix(line, "HEAD "):
			cur.Head = strings.TrimPrefix(line, "HEAD ")
		case strings.HasPrefix(line, "branch "):
			cur.Branch = strings.TrimPrefix(line, "branch ")
		}
	}
	flush()
	return result, nil
}

// WorktreeForBranch returns the worktree linked to refs/heads/{branch}, or nil.
func WorktreeForBranch(branch string) (*Worktree, error) {
	wts, err := ListWorktrees()
	if err != nil {
		return nil, err
	}
	target := "refs/heads/" + branch
	for i := range wts {
		if wts[i].Branch == target {
			return &wts[i], nil
		}
	}
	return nil, nil
}

// AddWorktree creates a new worktree at path, optionally creating a new branch from base.
func AddWorktree(path, branch, base string, createBranch bool) error {
	args := []string{"worktree", "add"}
	if createBranch {
		args = append(args, "-b", branch, path)
		if base != "" {
			args = append(args, base)
		}
	} else {
		args = append(args, path, branch)
	}
	cmd := exec.Command("git", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}
	return nil
}

// ConfigGet returns the value of a git config key, or empty if unset.
func ConfigGet(key string) (string, error) {
	cmd := exec.Command("git", "config", "--get", key)
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := errors.AsType[*exec.ExitError](err); ok && exitErr.ExitCode() == 1 {
			return "", nil
		}
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// ConfigGetAll returns all values for a multi-valued git config key.
func ConfigGetAll(key string) ([]string, error) {
	cmd := exec.Command("git", "config", "--get-all", key)
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := errors.AsType[*exec.ExitError](err); ok && exitErr.ExitCode() == 1 {
			return nil, nil
		}
		return nil, err
	}
	var values []string
	for line := range strings.SplitSeq(strings.TrimRight(string(out), "\n"), "\n") {
		if line != "" {
			values = append(values, line)
		}
	}
	return values, nil
}

func run(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("%s %s: %w: %s", name, strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}
	return stdout.String(), nil
}
