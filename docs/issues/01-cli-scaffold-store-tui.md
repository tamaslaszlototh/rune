# 01 — CLI scaffold + store + basic TUI

## What to build

The foundational vertical slice of `rune`. Set up the Go module, CLI entry point with basic arg parsing, the `store` module for reading/writing entries to `~/.rune/entries/YYYY-MM-DD.md`, and a minimal Bubble Tea TUI that reads today's entries and renders them as a list. At this stage entries can only be added by editing the markdown file directly — the TUI is read-only.

## Acceptance criteria

- [ ] `go mod init` with module name, all Go dependencies declared
- [ ] CLI entry point at `cmd/rune/main.go` that dispatches `rune` (no args) to open the TUI, and prints usage for unknown flags
- [ ] `store` module implements `ReadDay(date) → []Entry`, `ReadRange(from, to) → []Entry`, `AppendEntry(date, Entry)`
- [ ] Entry struct parsed from markdown: timestamp, project, body, tags, links, branch
- [ ] `~/.rune/entries/` directory created automatically on first write
- [ ] Bubble Tea TUI shows today's entries (or empty state if none)
- [ ] store tests: write entries to temp dir, read back, assert fields
- [ ] Entry markdown format follows the spec: `- [@HH:MM] [project-slug] Body text (branch: name)`

## Blocked by

None — can start immediately
