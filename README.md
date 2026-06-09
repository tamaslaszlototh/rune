# rune — Engineering Daily Journal

A terminal UI for logging daily engineering notes with zero overhead. Entries are auto-tagged with your current git project and branch, stored as plain markdown, and searchable from the CLI.

## Installation

### macOS

```bash
brew install tamaslaszlototh/tap/rune
```

### Linux

```bash
brew install tamaslaszlototh/tap/rune
```

Or download the `.deb` / `.rpm` from the [latest release](https://github.com/tamaslaszlototh/rune/releases/latest).

### Windows

Download the ZIP from the [latest release](https://github.com/tamaslaszlototh/rune/releases/latest) and extract `rune.exe` to a directory on your `PATH`.

> Requires [Git for Windows](https://git-scm.com) for project auto-detection. Works best in Windows Terminal, PowerShell, or any modern terminal with VT support.

### Any platform (requires Go)

```bash
go install github.com/tamaslaszlototh/rune/cmd/rune@latest
```

### Direct binary

Download the pre-built binary for your platform from the [latest release](https://github.com/tamaslaszlototh/rune/releases/latest).

## Usage

```
rune                    Open TUI (today's entries)
rune standup            Print last 24h standup summary
rune standup --since friday  Custom cutoff (ISO date, day name, or -Nd)
rune search <query>     Fuzzy search all entries
rune search -p <project> <query>  Scoped search
rune -p <project>       Open TUI pre-filtered to project
rune config             Open config file in $EDITOR
```

### TUI key bindings

| Key | Action |
|---|---|
| `Enter` | Save entry |
| `Tab` / `Shift+Tab` | Cycle project filters |
| `Ctrl+S` | Save draft |
| `/` | Inline search |
| `Esc` | Clear / exit search |
| `Ctrl+C` | Quit |
| `Ctrl+W` | Delete word |
| `Ctrl+U` | Clear line |

## Data storage

Entries are stored in `~/.rune/entries/YYYY-MM-DD.md` as plain markdown. Each line follows the format:

```
- [@HH:MM] [project] Body text (branch: name)
```

Tags and links (`#bug`, `@pr/142`) are highlighted in the TUI. Files are grep-friendly — use standard Unix tools when you outgrow the TUI.

## Standup output

```
## Standup — 2025-06-09

### project-a (main)
• Fixed rate limiting bug in API gateway
• Prod latency spike — noisy neighbor, scaled workers

### project-b (feature/x)
• Reviewed PR #42 — data migration script
```

## Configuration

Optional config at `~/.rune/config.yml`:

```yaml
editor: code
projects:
  idea001: Idea Project One
```

All features work with zero configuration.

## How it works

- **git detection**: Extracts project name from remote URL and current branch via `git rev-parse`. Falls back to directory name outside a repo.
- **search**: Case-insensitive substring match across all entries, with `-p` flag to scope to a project.
- **standup**: Groups entries by project since a given cutoff (default 24h), prints a Slack-ready summary.
- **draft auto-save**: Unsaved input is written to `~/.rune/drafts/` after 2s of inactivity.
