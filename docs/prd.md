# PRD: rune — Engineering Daily Journal

## Problem Statement

Software engineers accumulate a lot of ephemeral knowledge day-to-day: decisions made during debugging, links to relevant PRs, root causes of incidents, quick notes about architectural choices. This information is typically lost because there's no friction-free capture mechanism. The result is forgotten context, rambling standup updates, and hours of re-digging when a similar issue resurfaces months later. Existing solutions (wiki pages, Notion docs, Slack threads) are too heavy for real-time capture and don't integrate with the engineer's terminal workflow.

## Solution

**rune** — a terminal UI (TUI) daily engineering journal. It lets you log timestamped entries grouped by project with zero overhead. Auto-detects the project you're working on via git, saves everything as plain markdown in `~/.rune/`, and provides a `rune standup` command that formats the last 24 hours into a ready-to-paste standup summary. Full-text search across all entries means nothing is ever truly lost.

## User Stories

1. As a software engineer, I want to open `rune` in my terminal and see today's entries grouped by project, so that I can quickly review what I've done so far.
2. As a software engineer, I want to type a new entry and have it auto-tagged with the current git project and branch, so that I don't have to manually annotate context.
3. As a software engineer, I want entries to save automatically when I stop typing, so that I never lose a note.
4. As a software engineer, I want to filter the TUI view by project, so that I can focus on a single project's timeline.
5. As a software engineer on standup, I want to run `rune standup` and get a formatted summary of the last 24 hours grouped by project, so that I can paste it directly into Slack.
6. As a software engineer on standup, I want `rune standup --since friday` to cover the weekend, so that Monday standups include Friday's work.
7. As a software engineer, I want to run `rune search <query>` to fuzzy-find entries across all dates, so that I can recover context from weeks or months ago.
8. As a software engineer, I want `rune search --project idea001 <query>` to scope search to a single project, so that I can narrow results.
9. As a software engineer, I want the TUI to have an inline search mode (triggered by `/`), so that I can filter today's entries without leaving the TUI.
10. As a software engineer working across multiple repos, I want each entry to be automatically associated with the correct project, so that multi-project days are organized without extra effort.
11. As a software engineer, I want entries to include tag and link syntax (`#tag`, `@pr/123`), so that I can cross-reference entries with issues and PRs.
12. As a software engineer, I want `~/.rune/` to be plain markdown files, so that I can read, edit, and grep them with standard Unix tools.
13. As a software engineer, I want `rune` to work without any setup, so that I can start logging immediately after install.
14. As a software engineer, I want to see a cursor in the TUI prompt showing which project the next entry will be tagged under, so that I have confidence the auto-detection is correct.
15. As a software engineer, I want `rune` to gracefully handle being run outside a git repo (falling back to the directory name), so that I can still log entries for non-git work.
16. As a software engineer, I want `rune` to save entries with a timestamp, so that I have an accurate timeline of my day.

## Implementation Decisions

### Module architecture

Seven modules following deep vs. shallow separation:

| Module | Depth | Interface |
|---|---|---|
| `store` | Deep | `ReadDay(date) → []Entry`, `AppendEntry(date, Entry)`, `ReadRange(from, to) → []Entry` |
| `git` | Deep | `Detect() → (project, branch, error)` (cached per call) |
| `search` | Deep | `Search(entries []Entry, query string) → []Match` (pure function) |
| `standup` | Deep | `FormatStandup(entries []Entry, since time.Time) → string` (pure function) |
| `config` | Deep | `Load() → Config` |
| `tui` | Shallow | Bubble Tea model — delegates to store, git, search |
| `cli` | Shallow | `os.Args` dispatch to TUI or commands |

### Data storage

- **Path**: `~/.rune/entries/YYYY-MM-DD.md`
- **Format**: Markdown with a structured convention. Each line starts with `- [@HH:MM] [project-slug] Body text (branch: name)`. Tags use `#tag` syntax, links use `@resource/123` syntax.
- **No database**: Plain files are grep-friendly, git-friendly, and require zero migration. The trade-off (no relational queries) is irrelevant for this use case.

### Git detection

- Calls `git rev-parse --show-toplevel` to verify we're in a repo
- Calls `git remote get-url origin` → extracts short name from URL
- Calls `git rev-parse --abbrev-ref HEAD` for branch
- Cached for the session duration (re-detected on `rune` relaunch)

### Tech stack

- **Language**: Go
- **TUI framework**: Bubble Tea (charmbracelet/bubbletea)
- **Widgets**: bubbles/textinput for the entry prompt
- **Styling**: lipgloss
- **Config**: gopkg.in/yaml.v3
- **Distribution**: single static binary

### CLI interface

```
rune                    → Open TUI (today's entries)
rune standup            → Print last 24h standup summary
rune standup --since friday → Custom cutoff
rune search <query>     → Fuzzy search all entries → stdout
rune search -p <project> <query> → Scoped search
rune -p <project>       → Open TUI pre-filtered to project
```

### Standup format

```
## Standup — 2025-06-08

### project-a (main)
• Fixed rate limiting bug in API gateway (#api-gateway @pr/142)
• Prod latency spike — noisy neighbor, scaled workers

### project-b (feature/x)
• Reviewed PR #42 — data migration script
```

### TUI key bindings

| Key | Action |
|---|---|
| `Enter` | Save input as new entry |
| `Tab` / `Shift+Tab` | Cycle project filters |
| `Ctrl+S` | Force save |
| `/` | Enter inline search mode |
| `Esc` | Clear input / exit search |
| `Ctrl+C` | Quit |
| `Ctrl+W` | Delete word in input |
| `Ctrl+U` | Clear line |

## Testing Decisions

### Testing philosophy

Tests should verify external behavior (what the module does), not internal implementation (how it does it). Pure functions are preferred because they are trivially testable. File I/O and shell-out modules are tested via fixtures and fakes, not mocks.

### Modules to test

| Module | Test approach | Priority |
|---|---|---|
| `store` | Write entries to a temp dir, read them back, assert parsed fields. Test AppendEntry idempotency, ReadDay for missing dates, ReadRange boundary conditions. | High |
| `search` | Pure function tests. Feed known entries, assert match/no-match for various queries. Test fuzzy matching, case insensitivity, tag/link syntax. | High |
| `standup` | Pure function tests. Feed known entries, assert formatted output matches expected string. Test multi-project grouping, `--since` cutoff, empty input. | High |
| `config` | Write config YAML to temp file, load it, assert struct fields. Test missing file returns defaults. | Medium |
| `git` | Shell out to a temporary git repo fixture (init, add remote, commit). Assert correct slug and branch extraction. | Medium |
| `tui` | Lower priority — if tested, use Bubble Tea's testing utilities to simulate key presses and assert model state changes. | Low |
| `cli` | Lower priority — integration test that flags produce expected dispatch. | Low |

### Prior art

Since this is a greenfield project, there is no prior art in the repo. Tests will follow idiomatic Go conventions: `*_test.go` files, `testing.T`, table-driven tests for pure functions, `os.TempDir` for file I/O tests.

## Out of Scope

- Cloud sync or multi-machine support (no daemon, no server)
- Editing or deleting past entries from the TUI (use raw markdown files)
- Rich markdown rendering of entry bodies (plain text is sufficient)
- Mouse support in the TUI (keyboard-only for MVP)
- Multiple users or team features (single-user tool)
- Integration with project management tools (Jira, Linear, GitHub Issues)
- Automatic git commit/push of the rune directory
- Mobile app or web UI
- Plugins or extensibility API

## Further Notes

- The rune format is intentionally grep-friendly. Users who outgrow the TUI can always fall back to `grep -r` or `fzf` on `~/.rune/`.
- The `@pr/123` and `#tag` syntax is display-only for MVP — no link validation or auto-fetching is planned.
- Binary name `rune` is short and unlikely to conflict. If it does, a single-word alias is the fallback.
- Config file `~/.rune/config.yml` is optional — all features work out of the box with zero configuration.
