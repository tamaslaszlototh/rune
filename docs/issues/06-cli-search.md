# 06 — CLI: `rune search` command

## What to build

The `search` module (pure function) and the `rune search` CLI command. Scans all entry files in `~/.rune/entries/`, fuzzy-matches against the query string, and prints matching entries with date context. Supports `-p <project>` to scope search to a single project.

## Acceptance criteria

- [ ] `rune search <query>` scans all entry files and returns matching entries
- [ ] `rune search -p <project> <query>` scopes to one project
- [ ] Results printed with date, project, timestamp, and matching line
- [ ] Fuzzy matching (case-insensitive substring or simple fuzzy — matching chars in order)
- [ ] Empty result set if no matches
- [ ] `search.Search(entries, query)` is a pure function — trivially testable
- [ ] search tests: feed known entries, assert match/no-match for various queries

## Blocked by

- 01 — CLI scaffold + store + basic TUI
