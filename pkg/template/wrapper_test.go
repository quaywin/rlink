package template

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

// verifyShellSyntax uses 'sh -n' to verify that the script has valid POSIX syntax.
func verifyShellSyntax(t *testing.T, scriptContent string) {
	t.Helper()
	tmpFile, err := os.CreateTemp("", "rlink_test_*.sh")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(scriptContent); err != nil {
		t.Fatalf("failed to write temp script: %v", err)
	}
	tmpFile.Close()

	cmd := exec.Command("sh", "-n", tmpFile.Name())
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("POSIX shell syntax check failed: %v\nOutput:\n%s\nScript:\n%s", err, string(out), scriptContent)
	}
}

func TestGenerateWrapperZed(t *testing.T) {
	cfg := WrapperConfig{
		EditorName:    "Zed",
		CommandName:   "rzed",
		HostAlias:     "dev-box",
		LocalUser:     "myuser",
		ConnectPort:   22222,
		SSHKeyPath:    "$HOME/.ssh/rlink_id_ed25519",
		SyntaxPattern: `zed "ssh://{{.Host}}{{.Path}}"`,
	}

	script, err := GenerateWrapper(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.HasPrefix(script, "#!/bin/sh") {
		t.Errorf("script must start with #!/bin/sh")
	}
	if !strings.Contains(script, `HOST_ALIAS="dev-box"`) {
		t.Errorf("script missing HOST_ALIAS")
	}
	if !strings.Contains(script, `CONNECT_PORT="22222"`) {
		t.Errorf("script missing CONNECT_PORT")
	}
	if !strings.Contains(script, `zed "ssh://__RLINK_HOST____RLINK_PATH__"`) {
		t.Errorf("script missing syntax pattern")
	}

	verifyShellSyntax(t, script)
}

func TestGenerateWrapperOpener(t *testing.T) {
	cfg := WrapperConfig{
		ToolType:     "open",
		ToolCategory: "opener",
		ToolName:     "Web & File Opener",
		CommandName:  "ropen",
		HostAlias:    "remote-server",
		LocalUser:    "alice",
		ConnectPort:  22222,
		SSHKeyPath:   "$HOME/.ssh/rlink_id_ed25519",
	}

	script, err := GenerateWrapper(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.HasPrefix(script, "#!/bin/sh") {
		t.Errorf("script must start with #!/bin/sh")
	}
	if !strings.Contains(script, `COMMAND_NAME="ropen"`) {
		t.Errorf("script missing COMMAND_NAME")
	}
	if !strings.Contains(script, `HOST_ALIAS="remote-server"`) {
		t.Errorf("script missing HOST_ALIAS")
	}
	if !strings.Contains(script, `/tmp/rlink_downloads`) {
		t.Errorf("script missing temp downloads folder")
	}

	verifyShellSyntax(t, script)
}

func TestGenerateWrapperClipCopy(t *testing.T) {
	cfg := WrapperConfig{
		ToolType:     "clip",
		ToolCategory: "clipboard",
		ToolName:     "Clipboard Copy",
		CommandName:  "rclip",
		HostAlias:    "dev-server",
		LocalUser:    "bob",
		ConnectPort:  22222,
		SSHKeyPath:   "$HOME/.ssh/rlink_id_ed25519",
	}

	script, err := GenerateWrapper(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.HasPrefix(script, "#!/bin/sh") {
		t.Errorf("script must start with #!/bin/sh")
	}
	if !strings.Contains(script, `pbcopy`) || !strings.Contains(script, `wl-copy`) {
		t.Errorf("script missing clipboard copy commands")
	}

	verifyShellSyntax(t, script)
}

func TestGenerateWrapperClipPaste(t *testing.T) {
	cfg := WrapperConfig{
		ToolType:     "paste",
		ToolCategory: "clipboard",
		ToolName:     "Clipboard Paste",
		CommandName:  "rpaste",
		HostAlias:    "dev-server",
		LocalUser:    "bob",
		ConnectPort:  22222,
		SSHKeyPath:   "$HOME/.ssh/rlink_id_ed25519",
	}

	script, err := GenerateWrapper(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.HasPrefix(script, "#!/bin/sh") {
		t.Errorf("script must start with #!/bin/sh")
	}
	if !strings.Contains(script, `pbpaste`) || !strings.Contains(script, `wl-paste`) {
		t.Errorf("script missing clipboard paste commands")
	}

	verifyShellSyntax(t, script)
}

func TestGenerateWrapperCustom(t *testing.T) {
	cfg := WrapperConfig{
		ToolType:     "custom",
		ToolCategory: "custom",
		ToolName:     "mpv",
		CommandName:  "rmpv",
		HostAlias:    "media-server",
		LocalUser:    "quaywin",
		ConnectPort:  22222,
		SSHKeyPath:   "$HOME/.ssh/rlink_id_ed25519",
		LocalBinary:  "/opt/homebrew/bin/mpv",
	}

	script, err := GenerateWrapper(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.HasPrefix(script, "#!/bin/sh") {
		t.Errorf("script must start with #!/bin/sh")
	}
	if !strings.Contains(script, `LOCAL_BIN="/opt/homebrew/bin/mpv"`) {
		t.Errorf("script missing LOCAL_BIN")
	}

	verifyShellSyntax(t, script)
}
