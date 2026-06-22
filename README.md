# worker

`worker` is a small Go CLI that creates a [git worktree](https://git-scm.com/docs/git-worktree)
for a branch and opens it inside a new [zellij](https://zellij.dev/) tab. Each
tab contains a terminal, `lazygit`, and a `claude` pane that can be primed with
an initial instruction.

## Install

### Binary

```sh
go install github.com/satoshin21/worker/cmd/worker@latest
# or, from a checkout:
make install
```

The `worker` binary, `git`, `zellij`, `claude`, and `lazygit` must all be on
your `PATH`. `worker` must be invoked from inside a zellij session.

### Claude Code plugin (optional)

The repository also ships as a Claude Code marketplace that exposes a skill
with usage guidance. Inside Claude Code:

```
/plugin marketplace add satoshin21/worker
/plugin install worker@worker
```

- `marketplace add satoshin21/worker` registers this repository as a
  marketplace by reading `.claude-plugin/marketplace.json`.
- `install worker@worker` installs the `worker` plugin (the skill) from that
  marketplace.

The plugin only installs the skill — the `worker` binary itself still has to
be installed separately (see above).

## Configuration

`worker` reads its settings from `git config`. Global defaults live in
`~/.gitconfig`; per-repository overrides live in the repository's `.git/config`
(which is never committed). Standard git config precedence (`local` > `global`)
applies automatically.

```ini
# ~/.gitconfig
[worker]
    # Required. Worktrees are created at:
    #   <worktreeBase>/<repo-name>.<branch-name-with-slashes-as-dashes>
    worktreeBase = ~/dev/worktrees
```

```sh
# Per-repository overrides (written to .git/config)
git config --local worker.worktreeBase ~/dev/scratch-worktrees
git config --local --add worker.postCreate "./scripts/bootstrap.sh"
git config --local --add worker.postCreate "make deps"
```

### Recognized keys

| Key                   | Required | Description                                                                                                  |
| --------------------- | -------- | ------------------------------------------------------------------------------------------------------------ |
| `worker.worktreeBase` | no       | Directory under which worktrees are created. `~` is expanded. Defaults to the repository's parent directory. |
| `worker.postCreate`   | no       | Shell command run after a worktree is created. Set multiple times (`--add`) to run several scripts in order. |

## Usage

```sh
worker create <branch-name> [--instruction <text>] [--description <text>] [--base <ref>] [--rerun-post-create]
```

### Behavior

1. Resolves the repository root from the current directory.
2. If `<branch-name>` is already checked out in an existing worktree, that
   worktree is reused (no `post-create` run, unless `--rerun-post-create` is
   set).
3. Otherwise a worktree is created at
   `<worktreeBase>/<repo-name>.<branch-name>` (with `/` in the branch name
   replaced by `-` in the path component only). A new branch is created from
   `--base` if provided, or from current `HEAD` otherwise.
4. `worker.postCreate` scripts run with `sh -c`, cwd set to the worktree, and
   the following environment variables exported:
   - `WORKER_BRANCH`
   - `WORKER_WORKTREE_PATH`
   - `WORKER_INSTRUCTION`
   - `WORKER_DESCRIPTION`
   - `WORKER_REPO_ROOT`
5. A zellij tab is created (or refocused, if a tab with the same name already
   exists). The tab name is `<branch-name>` or `<branch-name>_<description>`
   when `--description` is supplied.

### Layout

`worker` ships with a built-in zellij layout: a left column with a terminal on
top and `lazygit` on the bottom, and a right column running `claude`. When
`--instruction` is provided, it is passed to `claude` as its first argument.

### Flags

| Flag                   | Description                                                                          |
| ---------------------- | ------------------------------------------------------------------------------------ |
| `--instruction`        | Passed to `claude` in the new tab as a single argument.                              |
| `--description`        | Appended to the tab name as `<branch>_<description>`.                                |
| `--base`               | Base ref for new branches. Defaults to current `HEAD`.                               |
| `--rerun-post-create`  | Re-run `worker.postCreate` scripts even when reusing an existing worktree.           |

## Development

```sh
go build ./...
go test ./...
```
