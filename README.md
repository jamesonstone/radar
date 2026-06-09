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

- Default config: `./.radar/config.yaml` (relative to current working directory)
- Default DB: `./.radar/radar.sqlite3` (relative to current working directory)
- Overrides:
  - `RADAR_CONFIG_PATH=/absolute/or/relative/path/config.yaml`
  - `RADAR_DB_PATH=/absolute/or/relative/path/radar.sqlite3`

## Build & run

```bash
go build ./cmd/rdr
./rdr once
```
