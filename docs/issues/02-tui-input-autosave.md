# 02 — TUI: input new entries + auto-save

## What to build

Add the text input prompt to the TUI. The user types at the bottom of the screen, presses Enter, and the entry is saved to today's file with a timestamp. A debounced auto-save fires on idle (e.g. 2s after last keystroke) so entries are never lost on crash. The input area shows the current entry being composed.

## Acceptance criteria

- [ ] Text input area at the bottom of the TUI (using `bubbles/textinput`)
- [ ] Pressing Enter saves the entry to today's file with an `@HH:MM` timestamp and clears the input
- [ ] Auto-save on 2s idle debounce (saves partial draft — timestamp is assigned on final save)
- [ ] Saved entry immediately appears in the entries list above
- [ ] `Ctrl+W` deletes the last word, `Ctrl+U` clears the line, `Esc` clears input
- [ ] Tests: simulate input and assert entries appended to store

## Blocked by

- 01 — CLI scaffold + store + basic TUI
