# radar

A lightweight local process observer for coding agents.

## Binary

`agent-watch`

## Commands

- `agent-watch` – live TUI dashboard
- `agent-watch once` – print one snapshot and exit
- `agent-watch list` – list currently matched processes
- `agent-watch history --limit 20 --agent Codex` – show saved sessions
- `agent-watch config init` – write default config
- `agent-watch config path` – print config path

## Paths

- Config: `~/.config/agent-watch/config.yaml`
- DB: `~/.local/share/agent-watch/agent-watch.sqlite3`

## Build & run

```bash
go build ./cmd/agent-watch
./agent-watch once
```
