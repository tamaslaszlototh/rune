# 05 — CLI: `rune standup` command

## What to build

The `standup` module (pure function) and the `rune standup` CLI command. Reads the last 24 hours of entries (or custom cutoff via `--since`), groups them by project, and formats them as a clean standup summary printed to stdout. Pipe-friendly for `pbcopy` / `xclip`.

## Acceptance criteria

- [ ] `rune standup` reads entries from the last 24 hours
- [ ] `rune standup --since friday` uses custom cutoff (parses natural day names, ISO dates, or `-3d` relative format)
- [ ] Output grouped by project, entries within each group ordered chronologically
- [ ] Output format matches spec: `## Standup — YYYY-MM-DD` header, `### project (branch)` per group, bullet points
- [ ] Empty output if no entries in the time range
- [ ] `standup.FormatStandup(entries, since)` is a pure function — trivially testable
- [ ] standup tests: feed known entries, assert exact output string

## Blocked by

- 03 — Git auto-detection + project tagging
