// Package command holds cobra subcommands for the worker CLI.
package command

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/satoshin21/worker/internal/git"
	"github.com/satoshin21/worker/internal/layout"
	"github.com/satoshin21/worker/internal/zellij"
)

// NewCreateCommand returns the `worker create` subcommand.
func NewCreateCommand() *cobra.Command {
	var (
		instruction       string
		description       string
		base              string
		rerunPostCreate   bool
	)
	cmd := &cobra.Command{
		Use:   "create <branch-name>",
		Short: "Create a git worktree and open it in a new zellij tab",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCreate(createOptions{
				branch:          strings.TrimSpace(args[0]),
				instruction:     instruction,
				description:     description,
				base:            base,
				rerunPostCreate: rerunPostCreate,
			})
		},
	}
	cmd.Flags().StringVar(&instruction, "instruction", "", "Instruction passed to claude as its first argument")
	cmd.Flags().StringVar(&description, "description", "", "Description appended to the zellij tab name")
	cmd.Flags().StringVar(&base, "base", "", "Base ref used when creating a new branch (defaults to current HEAD)")
	cmd.Flags().BoolVar(&rerunPostCreate, "rerun-post-create", false, "Re-run post-create scripts when reopening an existing worktree")
	return cmd
}

type createOptions struct {
	branch          string
	instruction     string
	description     string
	base            string
	rerunPostCreate bool
}

func runCreate(opt createOptions) error {
	if opt.branch == "" {
		return errors.New("branch name is required")
	}
	if !zellij.InSession() {
		return errors.New("worker must be run inside a zellij session ($ZELLIJ is not set)")
	}

	repoRoot, err := git.CommonDir()
	if err != nil {
		return fmt.Errorf("resolve repository root: %w", err)
	}
	repoName := filepath.Base(repoRoot)

	worktreeBase, err := git.ConfigGet("worker.worktreeBase")
	if err != nil {
		return fmt.Errorf("read worker.worktreeBase: %w", err)
	}
	if worktreeBase == "" {
		worktreeBase = filepath.Dir(repoRoot)
	} else {
		worktreeBase, err = expandHome(worktreeBase)
		if err != nil {
			return err
		}
	}

	worktreePath := filepath.Join(worktreeBase, fmt.Sprintf("%s.%s", repoName, sanitizeForPath(opt.branch)))
	tabName := opt.branch
	if opt.description != "" {
		tabName = fmt.Sprintf("%s_%s", opt.branch, opt.description)
	}

	existing, err := git.WorktreeForBranch(opt.branch)
	if err != nil {
		return fmt.Errorf("inspect existing worktrees: %w", err)
	}

	createdNow := false
	if existing != nil {
		worktreePath = existing.Path
		fmt.Fprintf(os.Stderr, "worker: reusing existing worktree at %s\n", worktreePath)
	} else {
		if _, statErr := os.Stat(worktreePath); statErr == nil {
			return fmt.Errorf("worktree path already exists but is not registered with git: %s", worktreePath)
		} else if !errors.Is(statErr, os.ErrNotExist) {
			return fmt.Errorf("stat %s: %w", worktreePath, statErr)
		}

		branchExists, err := git.BranchExists(opt.branch)
		if err != nil {
			return fmt.Errorf("check branch existence: %w", err)
		}
		if err := git.AddWorktree(worktreePath, opt.branch, opt.base, !branchExists); err != nil {
			return err
		}
		createdNow = true
		fmt.Fprintf(os.Stderr, "worker: created worktree at %s\n", worktreePath)
	}

	if createdNow || opt.rerunPostCreate {
		if err := runPostCreate(worktreePath, repoRoot, opt); err != nil {
			return err
		}
	}

	tabExists, err := zellij.TabExists(tabName)
	if err != nil {
		return err
	}
	if tabExists {
		fmt.Fprintf(os.Stderr, "worker: focusing existing zellij tab %q\n", tabName)
		return zellij.FocusTab(tabName)
	}

	layoutKDL := layout.Render(opt.instruction)
	if err := zellij.NewTab(tabName, worktreePath, layoutKDL); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "worker: opened zellij tab %q\n", tabName)
	return nil
}

func runPostCreate(worktreePath, repoRoot string, opt createOptions) error {
	scripts, err := git.ConfigGetAll("worker.postCreate")
	if err != nil {
		return fmt.Errorf("read worker.postCreate: %w", err)
	}
	if len(scripts) == 0 {
		return nil
	}
	env := append(os.Environ(),
		"WORKER_BRANCH="+opt.branch,
		"WORKER_WORKTREE_PATH="+worktreePath,
		"WORKER_INSTRUCTION="+opt.instruction,
		"WORKER_DESCRIPTION="+opt.description,
		"WORKER_REPO_ROOT="+repoRoot,
	)
	for _, script := range scripts {
		fmt.Fprintf(os.Stderr, "worker: running post-create: %s\n", script)
		cmd := exec.Command("sh", "-c", script)
		cmd.Dir = worktreePath
		cmd.Env = env
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("post-create script failed: %s: %w", script, err)
		}
	}
	return nil
}

func sanitizeForPath(branch string) string {
	return strings.ReplaceAll(branch, "/", "-")
}

func expandHome(p string) (string, error) {
	if !strings.HasPrefix(p, "~") {
		return p, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve $HOME: %w", err)
	}
	if p == "~" {
		return home, nil
	}
	if strings.HasPrefix(p, "~/") {
		return filepath.Join(home, p[2:]), nil
	}
	return p, nil
}
