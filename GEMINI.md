# AGENTS.md — Development Guidelines for rlink

## ⚠️ Critical Binary Deployment Rule

> [!CAUTION]
> **NEVER** copy, build, overwrite, or delete the binary at `~/.local/bin/rlink` or `/usr/local/bin/rlink` directly during agent tasks.
>
> **Reason**: Overwriting a live executable binary on macOS while active terminal panes or background processes are running can corrupt running process mappings and trigger kernel SIGKILL.

### ✅ Allowed Build & Test Procedures
1. **Local Build & Test**:
   ```bash
   go build ./...
   go test -v ./...
   ```
2. **Local Binary in Repository**:
   ```bash
   go build -o ./rlink .
   ```
3. **Go Install (Standard)**:
   ```bash
   go install .
   ```
   *(Installs safely to `$GOPATH/bin/rlink` without touching system-critical paths).*

---

## 🛠️ Project Architecture & Standards

- **Entrypoint**: `main.go` at the repository root calling `cmd.Execute()`.
- **Commands**: `cmd/` directory holds all Cobra subcommands (`root.go`, `setup.go`, `version.go`).
- **Packages**: Reusable logic belongs in `pkg/`:
  - `pkg/config`: OpenSSH client `~/.ssh/config` parsing and `RemoteForward` injection.
  - `pkg/detector`: Local GUI editor discovery (Zed, VS Code, Cursor) and network discovery (Tailscale, LAN).
  - `pkg/template`: Embedded POSIX `/bin/sh` wrapper generator (`zr`, `cr`, etc.).
  - `pkg/remote`: Remote server SSH execution, binary path discovery, and wrapper deployment.
  - `pkg/version`: Version string injected via GoReleaser `-ldflags`.
- **Zero Remote Dependencies**: The wrapper script deployed on remote machines MUST be 100% pure POSIX `/bin/sh`.
- **Quality Assurance**:
  - Always run `go test ./...` and ensure 100% test passing before completing tasks.
  - Keep `go vet ./...` completely clean without warnings.
