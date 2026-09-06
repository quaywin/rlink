# Implementation Plan: `rlink`

## 🎯 Objective & Overview
`rlink` is an open-source CLI wizard running on your local machine (macOS/Linux) that automates the bridging of remote terminal sessions (Ghostty, Alacritty, iTerm2, tmux) to local GUI editors (**Zed**, **VS Code**, **Cursor**).

It configures reverse connectivity (Reverse SSH Tunnel or Direct/Tailscale) and provisions a zero-dependency POSIX shell wrapper script (`zr`, `cr`, `cur`) on remote servers.

---

## 🏗️ Architecture & Module Breakdown

1. **`pkg/config`**:
   - Parse OpenSSH client `~/.ssh/config` (extract `Host`, `HostName`, `User`, `Port`, `IdentityFile`, `RemoteForward`).
   - Safely inject `RemoteForward <port> localhost:22` into specified `Host` blocks with automatic backup (`~/.ssh/config.rlink.bak`).
   - Provide removal / rollback capability for uninstallation.

2. **`pkg/detector`**:
   - Local GUI editor discovery (`Zed`, `Visual Studio Code`, `Cursor`) across `$PATH`, macOS `/Applications`, and Linux binary locations.
   - Local network discovery: Tailscale IPv4 (`tailscale ip -4`) and LAN IPv4 interfaces.
   - Local SSH daemon readiness check (verify port 22 is open on local machine; guide macOS Remote Login if closed).

3. **`pkg/template`**:
   - Zero-dependency, pure POSIX `/bin/sh` remote script generator.
   - Dynamic parameter substitution for target editor syntax, connection target, and paths.
   - POSIX path resolution ladder (`realpath` -> `readlink -f` -> `(cd "$(dirname "$1")" && pwd)`).
   - Support file positions (`file.rs:42:5`).
   - Built-in diagnostic flag (`-c` / `--check`).

4. **`pkg/remote` & `pkg/auth`**:
   - Remote server SSH execution and path discovery (prefers `~/.local/bin`, then `/usr/local/bin`).
   - Wrapper script provisioning and permission management (`chmod +x`).
   - Automated isolated SSH keypair generation for reverse connection (`~/.ssh/rlink_ed25519`), authorizing on local machine and deploying private key to remote machine for passwordless operation.

5. **`cmd` & CLI Layer**:
   - `root.go`: Cobra root command with version integration.
   - `setup.go`: Interactive 4-step wizard powered by `charmbracelet/huh`.
   - `status.go` / `list.go`: Inspect configured remote forward tunnels and wrapper statuses.
   - `remove.go`: Clean up wrapper script and revert `~/.ssh/config`.
   - `version.go`: Displays version, OS, Arch, and compiler info.

6. **Release & CI/CD**:
   - `.goreleaser.yaml` (v2) for cross-platform binary compilation (Darwin & Linux, amd64 & arm64).
   - `.github/workflows/ci.yml` for automated testing and linting.
   - `.github/workflows/release.yml` for automated GitHub Releases upon tagging `v*`.
   - `install.sh`: One-line curl installer for end-users.

---

## 📋 Implementation Phases

### Phase 1: Foundation & Core Scaffolding (Completed ✅)
- [x] Initial Go module initialization and Cobra CLI integration.
- [x] SSH config parsing & `RemoteForward` injection logic with unit tests (`pkg/config`).
- [x] Editor detection (`Zed`, `Code`, `Cursor`) & Network suggestion discovery (`pkg/detector`).
- [x] POSIX shell wrapper template and generator (`pkg/template`).
- [x] SSH execution and script deployment engine (`pkg/remote`).
- [x] Interactive setup wizard (`cmd/setup.go`).
- [x] GoReleaser, GitHub Actions release, POSIX installer script, and AGENTS.md guidelines.

### Phase 2: Security & Passwordless Reverse Auth
- [ ] Implement `pkg/auth`:
  - Detect or generate isolated Ed25519 keypair for reverse rlink triggers.
  - Safely add public key to local `~/.ssh/authorized_keys` with restricted options.
  - Deploy private key to remote host at `~/.ssh/rlink_id_ed25519` (`chmod 600`).
- [ ] Add local SSH daemon listening check (`pkg/detector/ssh_daemon.go`) with user tips for macOS Remote Login.

### Phase 3: Additional Management Subcommands
- [ ] Implement `cmd/status.go` / `cmd/list.go`: List hosts configured with `rlink` and verify their tunnel status.
- [ ] Implement `cmd/remove.go`: Wizard to uninstall remote wrapper and clean up `~/.ssh/config`.
- [ ] Implement `cmd/completion.go`: Generate shell completions for zsh, bash, fish.

### Phase 4: CI/CD Quality Gates & Release Verification
- [ ] Add `.github/workflows/ci.yml` (test + vet on all PRs/pushes).
- [ ] Full end-to-end dry run verification.
- [ ] Polish documentation, architectural diagrams, and troubleshooting guides in `README.md`.
