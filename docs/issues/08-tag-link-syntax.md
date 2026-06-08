# 08 — Tag & link syntax

## What to build

Parse and visually distinguish `#tag` and `@pr/123` / `@issue/42` syntax within entry bodies. Tags and links are extracted into structured fields on the Entry struct and rendered with distinct styling (e.g. color) in the TUI display.

## Acceptance criteria

- [ ] `#tag` syntax parsed from entry body into `Entry.Tags`
- [ ] `@resource/123` syntax parsed into `Entry.Links` with type (`pr`, `issue`, etc.) and ID
- [ ] Tags and links rendered with distinct lipgloss colors in the TUI
- [ ] Store round-trips correctly: write entry with tags → read back → tags preserved
- [ ] Tags and links are searchable via the search module
- [ ] Tests: parse known strings, assert tag/link extraction

## Blocked by

- 02 — TUI: input new entries + auto-save
