```text
██████╗  █████╗ ██████╗  █████╗ ██████╗
██╔══██╗██╔══██╗██╔══██╗██╔══██╗██╔══██╗
██████╔╝███████║██║  ██║███████║██████╔╝
██╔══██╗██╔══██║██║  ██║██╔══██║██╔══██╗
██║  ██║██║  ██║██████╔╝██║  ██║██║  ██║
╚═╝  ╚═╝╚═╝  ╚═╝╚═════╝ ╚═╝  ╚═╝╚═╝  ╚═╝
```

**Local-first process radar for coding agents**

## ⚙️ Installation

```bash
go install github.com/jamesonstone/radar/cmd/rdr@latest
```

Or build from source:

```bash
git clone https://github.com/jamesonstone/radar.git
cd radar
make build
```

## 🚀 Quick Start

```bash
rdr config init
rdr once
rdr
```

## 🧰 Commands

- `rdr` – live TUI dashboard
- `rdr once` – print one snapshot and exit
- `rdr list` – list currently matched processes
- `rdr history --limit 20 --agent Codex` – show saved sessions
- `rdr config init` – write default config
- `rdr config path` – print config path
- `rdr version` – print installed version

## 📁 Paths

- Default config: `./.radar/config.yaml` (relative to current working directory)
- Default DB: `./.radar/radar.sqlite3` (relative to current working directory)
- Overrides:
  - `RADAR_CONFIG_PATH=/absolute/or/relative/path/config.yaml`
  - `RADAR_DB_PATH=/absolute/or/relative/path/radar.sqlite3`

## 🛠️ Developer Ergonomics

Use the provided Makefile targets:

- `make build` – build `bin/rdr` with version injection
- `make install` – install `rdr` locally with version injection
- `make test` – run tests
- `make fmt` – run `go fmt ./...`
- `make vet` – run `go vet ./...`
- `make lint` – run `golangci-lint run ./...`
- `make tidy` – run `go mod tidy`
- `make all` – run `fmt`, `vet`, `test`, and `build`
