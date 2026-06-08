# 04 — TUI: project filter tabs

## What to build

Add a filter bar at the bottom of the TUI showing all distinct projects from today's entries. Tab/Shift+Tab cycles through project filters. Selecting a project narrows the view to show only that project's entries. An "All" option shows everything.

## Acceptance criteria

- [ ] Bottom bar shows pill-shaped project names extracted from today's entries (+ "All")
- [ ] Active filter is highlighted
- [ ] Tab cycles forward, Shift+Tab cycles backward through project filters
- [ ] Selecting a project filters the entry list to that project only
- [ ] Filter state resets when switching days (not in MVP scope — today-only view)
- [ ] Empty state shown if filtered project has no entries

## Blocked by

- 03 — Git auto-detection + project tagging
