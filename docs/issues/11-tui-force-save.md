# 11 — TUI: `Ctrl+S` force save

## What to build

Add a `Ctrl+S` keybinding that force-saves the current input as a draft immediately (without waiting for the 2s auto-save debounce). This gives users an explicit save gesture when they want to ensure a note is persisted before switching contexts or closing the terminal.

## Acceptance criteria

- [x] Pressing `Ctrl+S` saves the current input value as a draft to `~/.rune/drafts/` and shows no visible feedback (silent save)
- [x] Does not clear the input or modify the entry list (unlike Enter which finalises an entry)
- [x] Does not interfere with terminal flow control (`Ctrl+S` is normally XON/XOFF; Bubble Tea should swallow it)
- [x] Existing auto-save (2s idle debounce) continues to work unchanged
- [x] Force-save works both in normal input mode and search mode (saves the search string as draft — acceptable)

## Implementation notes

- Add a `case "ctrl+s"` to the `tea.KeyMsg` switch in `tui.go`'s `Update` method
- Reuse the existing `SaveDraft` store method
- No UI feedback needed beyond the draft being written to disk

## Blocked by

- 02 — TUI: input new entries + auto-save
