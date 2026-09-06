# rlink 🔗

Bridge your remote terminal sessions (Ghostty, Alacritty, iTerm2, tmux) to your local GUI editors (**Zed**, **VS Code**, **Cursor**) with a single command (`zr .`, `cr .`, `zed .`, etc.).

## 💡 The Problem & Solution

When developing inside a remote server over SSH, opening the current directory in a local GUI editor usually requires complex manual steps: opening a new window locally, selecting Remote-SSH, navigating folders, or setting up reverse tunnels by hand.

`rlink` is an automated CLI wizard that runs on your **local machine**:
1. 🔍 **Detects your installed GUI editors** (Zed, VS Code, Cursor).
2. 📖 **Parses `~/.ssh/config`** to let you pick the remote server.
3. 🔀 **Configures reverse connectivity** via **Reverse SSH Tunnel** (`RemoteForward`) or **Direct/Tailscale IP**.
4. 🚀 **Provisions a zero-dependency POSIX shell wrapper script** onto your remote server (`~/.local/bin/`).

### 🎯 Multi-Editor Support & Custom Command Naming

You can choose any custom command name for your wrapper (`--name` / `-n`), enabling you to wrap multiple editors on the exact same server!

```bash
# Example: Wrap Zed as 'zr' or 'zed'
rlink setup --editor zed --name zr --host dev-server

# Example: Wrap VS Code as 'cr' or 'code' on the same server (automatically reuses tunnel!)
rlink setup --editor code --name cr --host dev-server

# Example: Wrap Cursor as 'cur' or 'cursor'
rlink setup --editor cursor --name cur --host dev-server
```

Once provisioned, simply type on your remote terminal:
```bash
# On your remote server:
zr .              # Opens current directory in Zed locally
zr src/main.rs:42 # Opens file at line 42 in Zed locally
cr .              # Opens current directory in VS Code locally
cur .             # Opens current directory in Cursor locally
```

---

## 🏗️ Architecture & Project Structure

Aligned with the standard Go project structure and `agys`:

```
rlink/
├── .github/
│   └── workflows/
│       └── release.yml         # GitHub Actions GoReleaser automation
├── .goreleaser.yaml            # Multi-arch GoReleaser v2 configuration
├── AGENTS.md                   # Agent development rules and safety standards
├── GEMINI.md                   # AI consistency guidelines
├── LICENSE                     # MIT License
├── README.md                   # Project documentation
├── cmd/                        # Cobra CLI commands
│   ├── root.go                 # Root command & version integration
│   ├── setup.go                # Interactive TUI Wizard (Charm huh) & CLI flags
│   └── version.go              # Version command
├── install.sh                  # One-line curl installer for macOS & Linux
├── main.go                     # Application entrypoint calling cmd.Execute()
├── pkg/                        # Core reusable libraries
│   ├── config/                 # OpenSSH config parser & RemoteForward injector
│   │   ├── model.go
│   │   ├── ssh_config.go
│   │   └── ssh_config_test.go
│   ├── detector/               # Local editor & network discovery
│   │   ├── editor.go
│   │   ├── editor_test.go
│   │   └── network.go
│   ├── remote/                 # Remote SSH client & wrapper deployment
│   │   └── ssh.go
│   ├── template/               # Zero-dependency POSIX script generator
│   │   ├── wrapper.go
│   │   ├── wrapper.sh.tmpl
│   │   └── wrapper_test.go
│   └── version/                # Build-time version metadata
│       └── version.go
├── go.mod
└── go.sum
```

---

## ⚙️ How the Remote Wrapper Script Works

The generated remote wrapper script has **Zero External Dependencies** (`#!/bin/sh` POSIX compliant):

1. **Path Resolution**:
   - Resolves target relative paths (`.`, `src/foo.rs`) to absolute canonical paths.
   - Falls back gracefully across `realpath`, `readlink -f`, and POSIX `cd && pwd`.
   - Preserves line numbers (e.g., `main.go:42:5`).

2. **Triggering Local Editor**:
   - **Reverse SSH Tunnel Mode**: Connects to `127.0.0.1:<PORT>` (forwarded to local port 22).
   - **Direct / Tailscale Mode**: Connects directly to local Tailscale IP or LAN IP.
   - Executes local CLI command:
     - Zed: `zed "ssh://<HostAlias>/canonical/path"`
     - VS Code: `code --remote ssh-remote+<HostAlias> /canonical/path`
     - Cursor: `cursor --remote ssh-remote+<HostAlias> /canonical/path`

3. **Status Feedback & Troubleshooting**:
   - Includes `-c` / `--check` connectivity test.
   - Outputs clear error tips if reverse tunnel is not active.

---

## 🚀 Quick Start

### One-line Install:
```bash
curl -fsSL https://raw.githubusercontent.com/quaywin/rlink/main/install.sh | bash
```

### Local Build & Development:
```bash
go build -o rlink .
./rlink setup
```

### CLI Flags:
```bash
rlink setup --help
  -e, --editor string   Target GUI editor: zed, code, or cursor
  -n, --name string     Custom remote wrapper command name (e.g. zr, zed, cr, code, cur, cursor)
  -H, --host string     Remote SSH host alias from ~/.ssh/config or user@hostname
  -p, --port int        Forward/connect port (default: 22222 for tunnel, 22 for direct)
  -m, --mode string     Connection mode: 'tunnel' (reverse SSH) or 'direct' (Tailscale/LAN)
  -y, --yes             Automatically deploy without interactive confirmation prompt
```
