# 03 — Git auto-detection + project tagging

## What to build

The `git` module that detects the current repo slug and branch name. Every new entry is automatically tagged with its project. The TUI prompt shows which project the next entry will be tagged under. When outside a git repo, fall back to the current directory name. Entries from different projects are grouped visually in the TUI (ordered chronologically, grouped by project section).

## Acceptance criteria

- [ ] `git.Detect()` returns `(project, branch, error)` by shelling out to git
- [ ] Repo slug extracted from `git remote get-url origin` (e.g. `owner/repo` → `repo`)
- [ ] Current branch from `git rev-parse --abbrev-ref HEAD`
- [ ] Outside git: falls back to current directory basename, branch is empty string
- [ ] Result cached per `rune` session
- [ ] New entries saved with `[project]` tag in the markdown line and `(branch: name)` suffix
- [ ] TUI prompt shows current project in brackets (e.g. `> [idea001] _`)
- [ ] Today's entries grouped by project with project headers
- [ ] Tests: temporary git repo fixture with remote, assert slug + branch extraction; non-git directory test

## Blocked by

- 02 — TUI: input new entries + auto-save
