# 10 — CLI: `rune -p <project>` opens TUI pre-filtered

## What to build

Add a `-p <project>` flag so that `rune -p idea001` opens the TUI with the entry list pre-filtered to that project. Currently `rune` with no subcommand launches the TUI showing all entries; this flag narrows the initial view without requiring the user to Tab through filters.

## Acceptance criteria

- [ ] `rune -p <project>` launches the TUI with the entry list filtered to the given project
- [ ] The filter bar shows "All" and the filtered project as active
- [ ] Tab/Shift+Tab still cycle through all available filters
- [ ] `rune` with no args still opens the TUI unfiltered (no regression)
- [ ] Unknown flags produce a clear error message

## Implementation notes

- Parse `-p` in `main.go` before the `len(os.Args) < 2` check so the TUI or a usage message is dispatched accordingly
- Pass a `filterProject string` (or similar) through `tui.Run()` → `initialModel()` to set `filterIndex` based on the matching project name in `projects`
- Update the usage string to reflect the new flag: `rune [-p <project>] [config|standup|search]`

## Blocked by

- 04 — TUI: project filter tabs
