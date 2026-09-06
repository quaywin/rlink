# Implementation Plan: Extended Tool Support in `rlink`

## 🎯 Objective & Overview
Expand `rlink` beyond GUI code editors into a comprehensive **Remote Action Runner** that connects remote server terminal sessions to local machine capabilities via Reverse SSH Tunnel.

Support:
1. **GUI Code Editors**: Zed (`rzed`), VS Code (`rcode`), Cursor (`rcursor`), Windsurf (`rwindsurf`), Sublime (`rsubl`), etc.
2. **System Opener (`ropen`)**: Opens URLs in local default browser (macOS `open`, Linux `xdg-open`) and opens remote files locally via automated temporary streaming.
3. **Clipboard Sync (`rclip` & `rpaste`)**:
   - `rclip`: Pipe remote command output/files to local clipboard (`pbcopy`, `wl-copy`, `xclip`).
   - `rpaste`: Stream local clipboard to remote terminal output (`pbpaste`, `wl-paste`, `xclip -o`).
4. **Custom Command Runner (`r<custom>`)**: Map any local binary to an `r<name>` remote wrapper with argument and stdin support.
5. **Multi-OS Intelligence**: Auto-detect macOS vs Linux local environments and select appropriate native commands.

---

## 🏗️ Architecture & Component Design

### 1. `pkg/detector`
- Introduce `ToolKind` enum: `ToolKindEditor`, `ToolKindOpener`, `ToolKindClipboardCopy`, `ToolKindClipboardPaste`, `ToolKindCustom`.
- Detect system tools based on `runtime.GOOS`:
  - macOS: `open` (opener), `pbcopy` (clip copy), `pbpaste` (clip paste).
  - Linux: `xdg-open` (opener), `wl-copy` / `xclip` / `xsel` (clip copy), `wl-paste` / `xclip -o` / `xsel -o` (clip paste).
- Extend `DetectedTool` model to encompass editors, system tools, and custom executables.
- Validate custom commands using `exec.LookPath`.

### 2. `pkg/template`
- Modularize or support templates for each `ToolKind`:
  - **Editor Template**: Resolves remote absolute path, passes formatted URI (`zed ssh://...`, `code --remote ssh-remote+...`) to local machine.
  - **Opener Template (`ropen`)**:
    - If argument is URL (`http://`, `https://`): runs local `open "$URL"` or `xdg-open "$URL"`.
    - If argument is file/dir: streams file to `/tmp/rlink_downloads/...` on local machine and triggers `open` / `xdg-open`.
  - **Clipboard Copy Template (`rclip`)**:
    - Reads stdin or argument string/file, pipes to local `pbcopy` or `wl-copy || xclip -selection clipboard`.
  - **Clipboard Paste Template (`rpaste`)**:
    - Queries local `pbpaste` or `wl-paste || xclip -selection clipboard -o`, outputs directly to remote stdout.
  - **Custom Template**:
    - Executes specified local binary with remote arguments and/or stdin.

### 3. `cmd/setup.go`
- Update Step 1 to present a categorized menu of available tools:
  - 📝 GUI Editors (Zed, VS Code, Cursor, Windsurf, Sublime...)
  - 🌐 Web / File Opener (`ropen`)
  - 📋 System Clipboard Copy (`rclip`)
  - 📋 System Clipboard Paste (`rpaste`)
  - ⚙️  `[+] Custom Local Command...`
- Handle `--tool` flag with alias `--editor` for backwards compatibility.
- Seamlessly reuse existing SSH config, reverse tunnel, and Ed25519 authentication mechanism.

---

## 📋 Task Breakdown

### Task 5.1: Define Unified Tool Model & System Tool Detection (`pkg/detector`)
- Define `ToolKind` and `DetectedTool`.
- Add OS-specific detection for `open` / `xdg-open`, `pbcopy` / `wl-copy`, `pbpaste` / `wl-paste`.
- Add helper for validating custom commands via `exec.LookPath`.
- Write unit tests in `pkg/detector/tool_test.go`.

### Task 5.2: Implement Shell Wrapper Templates for All Tool Kinds (`pkg/template`)
- Update `wrapper.go` and add templates for Opener, Clipboard, and Custom tools.
- Ensure all scripts are 100% POSIX `/bin/sh` with zero remote dependencies.
- Write unit tests in `pkg/template/wrapper_test.go`.

### Task 5.3: Update Setup Wizard & CLI Flags (`cmd/setup.go`)
- Update interactive wizard menu in `setup.go`.
- Support `--tool` / `--editor` flags.
- Support interactive input for custom local commands.

### Task 5.4: End-to-End Verification & Documentation Update
- Run `go test -v ./...` and `go vet ./...`.
- Build `./rlink` locally.
- Update `README.md` with examples of `ropen`, `rclip`, `rpaste`, and custom commands.
