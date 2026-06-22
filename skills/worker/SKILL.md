---
name: worker
description: Use the `worker` CLI to create a git worktree for a branch and open it in a new zellij tab (terminal + lazygit + claude). Load when the user asks to "create a worker", "open a worktree in zellij", "run worker create", configures `worker.worktreeBase` or `worker.postCreate` in gitconfig, or troubleshoots worker behavior.
---

# worker

`worker` is a Go CLI that automates a single workflow:

1. Create a git worktree for a branch (or reuse an existing one).
2. Run user-defined post-create scripts inside the worktree.
3. Open a new zellij tab pre-populated with `terminal`, `lazygit`, and `claude` panes.

The tool must be invoked from **inside a zellij session** (it reads `$ZELLIJ`).

## Command

```sh
worker create <branch-name> [--instruction <text>] [--description <text>] [--base <ref>] [--rerun-post-create]
```

| Flag                  | Meaning                                                                                   |
| --------------------- | ----------------------------------------------------------------------------------------- |
| `--instruction`       | Passed to `claude` as its first argument in the new tab.                                  |
| `--description`       | Appended to the tab name as `<branch>_<description>`. Whitespace is preserved as-is.       |
| `--base`              | Base ref used when creating a new branch. Defaults to the current `HEAD`.                 |
| `--rerun-post-create` | Re-run `worker.postCreate` scripts even when reusing an existing worktree.                |

## Configuration

`worker` reads everything from `git config`. Global defaults live in `~/.gitconfig`;
per-repository overrides live in `.git/config` (never committed). Standard git config
precedence (`local` > `global`) applies.

```ini
# ~/.gitconfig
[worker]
    # Required. Worktrees are created at:
    #   <worktreeBase>/<repo-name>.<branch-name-with-slashes-as-dashes>
    worktreeBase = ~/dev/worktrees
```

Per-repository overrides:

```sh
git config --local worker.worktreeBase ~/dev/scratch-worktrees
git config --local --add worker.postCreate "./scripts/bootstrap.sh"
git config --local --add worker.postCreate "make deps"
```

### Recognized keys

| Key                   | Required | Notes                                                                                                         |
| --------------------- | -------- | ------------------------------------------------------------------------------------------------------------- |
| `worker.worktreeBase` | no       | Directory under which worktrees are created. `~` is expanded. Defaults to the repository's parent directory (`..`). |
| `worker.postCreate`   | no       | Shell command run after a worktree is created. Set multiple times with `--add` to run several scripts in order. |

## Behavior

1. Resolves the repository root from the current directory (`git rev-parse --git-common-dir`).
2. If the branch is already checked out in an existing worktree, `worker` reuses
   that worktree (no `post-create` run, unless `--rerun-post-create` is set).
3. Otherwise a worktree is created at
   `<worktreeBase>/<repo-name>.<branch-name>` (with `/` in the branch name replaced
   by `-` in the path component only — the branch name and tab name keep their slashes).
   A new branch is created from `--base` (or current `HEAD` if `--base` is omitted).
4. `worker.postCreate` scripts run with `sh -c`, cwd set to the worktree, and the
   following environment variables exported:
   - `WORKER_BRANCH`
   - `WORKER_WORKTREE_PATH`
   - `WORKER_INSTRUCTION`
   - `WORKER_DESCRIPTION`
   - `WORKER_REPO_ROOT`
   A non-zero exit code aborts the workflow before the zellij tab is created. The
   worktree itself is left in place so the user can fix the issue and retry with
   `--rerun-post-create`.
5. A zellij tab is created (or refocused, if a tab with the same name already
   exists). The tab name is `<branch-name>` or `<branch-name>_<description>` when
   `--description` is supplied.

## Layout

`worker` ships with a built-in zellij layout (a left column with a terminal on
top and `lazygit` on the bottom, and a right column running `claude`). There is
no user-customizable layout file. When `--instruction` is provided it is passed
to `claude` as its first argument; when omitted, `claude` starts with no
arguments.

## Common requests

- **"Set up worker"** — Ensure `git config --global worker.worktreeBase <path>` is
  set, then verify the binary is on `PATH` (`make install` puts it in `~/go/bin/worker`).
- **"Run worker for branch X with instruction Y"** —
  `worker create X --instruction "Y"`. Must be inside a zellij session.
- **"Add a setup script that runs after worktree creation"** —
  `git config --local --add worker.postCreate "<command>"`. Use `--add` to keep
  existing entries; setting without `--add` replaces them.
- **"Reopen the worker tab"** — Run the same `worker create <branch>` again. If
  the worktree and tab still exist, the existing tab is refocused.

## Troubleshooting

| Symptom                                                                         | Cause / Fix                                                                                                                  |
| ------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------- |
| `worker must be run inside a zellij session ($ZELLIJ is not set)`              | Launch zellij first, then run `worker create` from a pane.                                                                   |
| `worktree path already exists but is not registered with git`                   | Stale directory at `<worktreeBase>/<repo>.<branch>`. Remove it manually (`trash` / `rm -rf`) and re-run.                      |
| `post-create script failed`                                                     | Fix the script, then re-run with `--rerun-post-create` to apply it to the existing worktree without recreating it.            |
| Existing tab is not refocused                                                   | Check that the tab name matches exactly (`<branch>` or `<branch>_<description>`). Tab names are case-sensitive in zellij.    |

## Source

The repository ships the CLI source plus this skill. Install the binary with:

```sh
make install            # → $GOBIN or ~/go/bin
# or
go install github.com/satoshin21/worker/cmd/worker@latest
```
