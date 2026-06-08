# 07 — TUI: inline search mode

## What to build

Add a `/` key shortcut in the TUI that opens an inline search/filter prompt. As the user types, the entry list filters in real-time to show only matching entries. Esc exits search mode and restores the full view.

## Acceptance criteria

- [ ] `/` key opens a search prompt in the TUI (input replaces or overlays the entry input area)
- [ ] Real-time filtering as the user types (fuzzy match against entry body text)
- [ ] Matches highlighted in the results
- [ ] Esc exits search mode and restores the last active filter (project filter, if any)
- [ ] Empty results shown if no match
- [ ] Search works within the current TUI session only (not persistent)

## Blocked by

- 04 — TUI: project filter tabs
