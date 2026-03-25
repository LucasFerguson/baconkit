# baconkit

A Linux forensic TUI for identifying malware and rootkits. It inspects running processes with a suite of scanners and surfaces suspicious indicators in an interactive terminal interface.

## Structure

```
baconkit/
├── main.go        # TUI entry point (bubbletea)
├── tools/         # per-process scanners
└── scans/         # system-wide scans
    └── deb.go     # finds ELF binaries not tracked by dpkg
```

## Scans

### deb
Requires root:
```bash
sudo go run . deb
```

## Prerequisites

- Go 1.25+
- Linux (Debian-based for `scans/deb`)

## Commands

```bash
go run .          # run the TUI
go build          # compile binary
go fmt ./...      # format
go vet ./...      # lint
go test ./...     # test
go mod tidy       # sync dependencies
```
