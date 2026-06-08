# 09 — Config file support

## What to build

The `config` module that loads `~/.rune/config.yml`. All features work without the file existing — config is purely additive overrides. The initial config supports only a few optional keys (e.g. projects map for friendly names, editor command). Loading is done once at startup.

## Acceptance criteria

- [ ] `config.Load()` reads `~/.rune/config.yml` and returns a `Config` struct
- [ ] Missing file returns a default config (no error)
- [ ] Invalid YAML returns a parse error
- [ ] Config struct is accessible from the TUI and CLI commands
- [ ] Tests: write YAML to temp file, load, assert struct fields; missing file returns defaults; invalid YAML returns error

## Blocked by

- 01 — CLI scaffold + store + basic TUI
