# radar

A lightweight local process observer for coding agents.

## Binary

`rdr`

## Commands

- `rdr` – live TUI dashboard
- `rdr once` – print one snapshot and exit
- `rdr list` – list currently matched processes
- `rdr history --limit 20 --agent Codex` – show saved sessions
- `rdr config init` – write default config
- `rdr config path` – print config path

## Paths

- Config: `~/.config/radar/config.yaml`
- DB: `~/.local/share/radar/radar.sqlite3`

## Build & run

```bash
go build ./cmd/rdr
./rdr once
```
