# Todo Checklist: `rlink`

## Phase 1: Foundation & Core Scaffolding
- [x] Task 1.1: Initialize Go module `github.com/quaywin/rlink` and Cobra CLI
- [x] Task 1.2: Implement `pkg/config` (OpenSSH config parser & `RemoteForward` injector with tests)
- [x] Task 1.3: Implement `pkg/detector` (Zed, VS Code, Cursor discovery & Tailscale/LAN IP detection with tests)
- [x] Task 1.4: Implement `pkg/template` (POSIX shell wrapper template & generator with tests)
- [x] Task 1.5: Implement `pkg/remote` (SSH executor, bin dir discovery, and script deployment)
- [x] Task 1.6: Implement `cmd/setup.go` (Interactive 4-step wizard using `charmbracelet/huh`)
- [x] Task 1.7: Align structure with `agys` (`main.go`, `cmd/`, `pkg/`, `install.sh`, `.goreleaser.yaml`, `AGENTS.md`, `GEMINI.md`, `LICENSE`)

## Phase 2: Security & Passwordless Reverse Auth
- [x] Task 2.1: Implement isolated Ed25519 key generation and local `~/.ssh/authorized_keys` provisioning (`pkg/auth/key.go`)
- [x] Task 2.2: Add remote private key deployment during setup wizard
- [x] Task 2.3: Implement local SSH daemon listening check (`pkg/detector/ssh_daemon.go`) with macOS Remote Login guidance

## Phase 3: Subcommands & Operations
- [x] Task 3.1: Implement `rlink status` / `rlink list` command to view configured hosts and tunnels
- [x] Task 3.2: Implement `rlink remove` command to cleanly unbind remote wrapper and revert `~/.ssh/config`
- [x] Task 3.3: Implement shell auto-completion subcommand (`rlink completion`)

## Phase 4: CI/CD Quality Gates & Documentation Polish
- [x] Task 4.1: Add GitHub Actions CI workflow (`.github/workflows/ci.yml`) for automated test & vet
- [x] Task 4.2: Add Architecture ASCII / Mermaid diagram and troubleshooting section in `README.md`
- [x] Task 4.3: End-to-end verification and quality check (`go test -v ./...` & `go vet ./...`)

## Phase 5: Extended Tool Support (System Opener, Clipboard, Custom Commands, Multi-OS Detection)
- [x] Task 5.1: Implement unified Tool model & OS-specific detector for System Opener, Clipboard, and Custom tools (`pkg/detector`)
- [x] Task 5.2: Implement POSIX shell wrapper templates for Editor, Opener, Clipboard (Copy/Paste), and Custom commands (`pkg/template`)
- [x] Task 5.3: Update setup wizard (`cmd/setup.go`) and command flags to support all tool categories and custom inputs
- [x] Task 5.4: Comprehensive test coverage, `go vet` verification, and documentation update
