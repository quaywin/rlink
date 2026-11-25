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
	if !strings.Contains(script, `sh -c `) {
		t.Errorf("script must wrap remote command in sh -c for non-POSIX local login shells like fish")
	}

	verifyShellSyntax(t, script)
}

func TestGenerateWrapperWu(t *testing.T) {
	cfg := WrapperConfig{
		ToolType:      "wu",
		ToolCategory:  "editor",
		ToolName:      "Wu",
		CommandName:   "rwu",
		HostAlias:     "dev-box",
		LocalUser:     "myuser",
		ConnectPort:   22222,
		SSHKeyPath:    "$HOME/.ssh/rlink_id_ed25519",
		SyntaxPattern: `wu "ssh://{{.Host}}{{.Path}}"`,
	}

	script, err := GenerateWrapper(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(script, `wu "ssh://__RLINK_HOST____RLINK_PATH__"`) {
		t.Errorf("script missing wu ssh syntax pattern")
	}
	if !strings.Contains(script, `sh -c `) {
		t.Errorf("script must wrap remote command in sh -c for non-POSIX local login shells like fish")
	}

	verifyShellSyntax(t, script)
}

func TestCustomWrapperFishCompatibility(t *testing.T) {
	fishPath, err := exec.LookPath("fish")
	if err != nil {
		t.Skip("fish shell not installed, skipping live fish parsing test")
	}

	// Generate custom wrapper and execute it with a mocked ssh binary that runs the remote command through fish -c
	tmpDir := t.TempDir()
	mockSSH := tmpDir + "/ssh"
	// Mock ssh extracts the last argument (the remote command string) and executes it via fish -c
	mockSSHScript := "#!/bin/sh\nfor last; do :; done\nexec " + fishPath + " -c \"$last\"\n"
	if err := os.WriteFile(mockSSH, []byte(mockSSHScript), 0755); err != nil {
		t.Fatalf("failed to write mock ssh: %v", err)
	}

	cfg := WrapperConfig{
		ToolType:     "custom",
		ToolCategory: "custom",
		ToolName:     "printf",
		CommandName:  "rprintf",
		HostAlias:    "dev-box",
		LocalUser:    "myuser",
		ConnectPort:  22222,
		LocalBinary:  "/usr/bin/printf",
	}

	script, err := GenerateWrapper(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wrapperPath := tmpDir + "/rprintf"
	if err := os.WriteFile(wrapperPath, []byte(script), 0755); err != nil {
		t.Fatalf("failed to write wrapper: %v", err)
	}

	cmd := exec.Command(wrapperPath, "<%s>|<%s>|", "hello world", "a'b$c\"d\\e")
	cmd.Env = append(os.Environ(), "PATH="+tmpDir+":"+os.Getenv("PATH"))
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("wrapper execution via fish failed: %v\nOutput:\n%s", err, string(out))
	}

	want := "<hello world>|<a'b$c\"d\\e>|"
	if string(out) != want {
		t.Errorf("unexpected output: got %q, want %q", string(out), want)
	}

	// Test zero-argument invocation + piped stdin through fish -c
	catCfg := cfg
	catCfg.CommandName = "rcat"
	catCfg.LocalBinary = "/bin/cat"
	catScript, err := GenerateWrapper(catCfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	catWrapperPath := tmpDir + "/rcat"
	if err := os.WriteFile(catWrapperPath, []byte(catScript), 0755); err != nil {
		t.Fatalf("failed to write cat wrapper: %v", err)
	}

	catCmd := exec.Command(catWrapperPath)
	catCmd.Env = append(os.Environ(), "PATH="+tmpDir+":"+os.Getenv("PATH"))
	catCmd.Stdin = strings.NewReader("piped via fish\n")
	catOut, err := catCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("piped custom wrapper via fish failed: %v\nOutput:\n%s", err, string(catOut))
	}
	if string(catOut) != "piped via fish\n" {
		t.Errorf("unexpected piped output: got %q, want %q", string(catOut), "piped via fish\n")
	}
}

func TestEditorAndOpenerWrappersFishCompatibility(t *testing.T) {
	fishPath, err := exec.LookPath("fish")
	if err != nil {
		t.Skip("fish shell not installed, skipping live fish parsing test")
	}

	tmpDir := t.TempDir()
	// Mock ssh: runs the remote command string under fish -c
	mockSSH := tmpDir + "/ssh"
	mockSSHScript := "#!/bin/sh\nfor last; do :; done\nexec " + fishPath + " -c \"$last\"\n"
	if err := os.WriteFile(mockSSH, []byte(mockSSHScript), 0755); err != nil {
		t.Fatalf("failed to write mock ssh: %v", err)
	}

	// Mock local 'wu' and 'open' commands to verify arguments received after fish -> sh -c
	logFile := tmpDir + "/invocations.log"
	for _, tool := range []string{"wu", "open", "pbcopy", "pbpaste"} {
		mockTool := tmpDir + "/" + tool
		body := "#!/bin/sh\nprintf '" + tool + ":%s\\n' \"$*\" >> " + logFile + "\n"
		if tool == "pbcopy" {
			body = "#!/bin/sh\ncat > " + tmpDir + "/clip.txt\n"
		} else if tool == "pbpaste" {
			body = "#!/bin/sh\nprintf 'from_clipboard'\n"
		}
		if err := os.WriteFile(mockTool, []byte(body), 0755); err != nil {
			t.Fatalf("failed to write mock %s: %v", tool, err)
		}
	}

	// Place tmpDir early in PATH so mock open/pbcopy/pbpaste take precedence over /usr/bin,
	// while SyntaxPattern calls mock wu directly by path.
	env := append(os.Environ(), "PATH="+tmpDir+":"+os.Getenv("PATH"))

	// 1. Editor wrapper (wu)
	editorCfg := WrapperConfig{
		ToolType:      "wu",
		ToolCategory:  "editor",
		ToolName:      "Wu",
		CommandName:   "rwu",
		HostAlias:     "dev-box",
		LocalUser:     "myuser",
		ConnectPort:   22222,
		SyntaxPattern: tmpDir + `/wu "ssh://{{.Host}}{{.Path}}"`,
	}
	editorScript, err := GenerateWrapper(editorCfg)
	if err != nil {
		t.Fatalf("failed to generate editor wrapper: %v", err)
	}
	rwuPath := tmpDir + "/rwu"
	if err := os.WriteFile(rwuPath, []byte(editorScript), 0755); err != nil {
		t.Fatalf("failed to write rwu: %v", err)
	}

	sampleFile := tmpDir + "/sample.txt"
	if err := os.WriteFile(sampleFile, []byte("hello from remote"), 0644); err != nil {
		t.Fatalf("failed to write sample file: %v", err)
	}

	cmd := exec.Command(rwuPath, sampleFile)
	cmd.Env = env
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("rwu via fish failed: %v\nOutput:\n%s", err, string(out))
	}

	// 2. Opener wrapper (URL, dir, and file streaming)
	t.Cleanup(func() { os.RemoveAll("/tmp/rlink_downloads/dev-box") })
	openerCfg := WrapperConfig{
		ToolType:     "open",
		ToolCategory: "opener",
		ToolName:     "Web & File Opener",
		CommandName:  "ropen",
		HostAlias:    "dev-box",
		LocalUser:    "myuser",
		ConnectPort:  22222,
	}
	openerScript, err := GenerateWrapper(openerCfg)
	if err != nil {
		t.Fatalf("failed to generate opener wrapper: %v", err)
	}
	ropenPath := tmpDir + "/ropen"
	if err := os.WriteFile(ropenPath, []byte(openerScript), 0755); err != nil {
		t.Fatalf("failed to write ropen: %v", err)
	}

	for _, arg := range []string{"https://example.com?a=1&b=2", tmpDir, sampleFile} {
		ocmd := exec.Command(ropenPath, arg)
		ocmd.Env = env
		if out, err := ocmd.CombinedOutput(); err != nil {
			t.Fatalf("ropen %q via fish failed: %v\nOutput:\n%s", arg, err, string(out))
		}
	}

	// 3. Clipboard copy & paste wrappers via fish -c
	clipCfg := WrapperConfig{
		ToolType:     "clip",
		ToolCategory: "clipboard",
		ToolName:     "Clipboard Copy",
		CommandName:  "rclip",
		HostAlias:    "dev-box",
		LocalUser:    "myuser",
		ConnectPort:  22222,
	}
	clipScript, err := GenerateWrapper(clipCfg)
	if err != nil {
		t.Fatalf("failed to generate clip wrapper: %v", err)
	}
	rclipPath := tmpDir + "/rclip"
	if err := os.WriteFile(rclipPath, []byte(clipScript), 0755); err != nil {
		t.Fatalf("failed to write rclip: %v", err)
	}
	ccmd := exec.Command(rclipPath, "fish-clip-test")
	ccmd.Env = env
	if out, err := ccmd.CombinedOutput(); err != nil {
		t.Fatalf("rclip via fish failed: %v\nOutput:\n%s", err, string(out))
	}

	pasteCfg := WrapperConfig{
		ToolType:     "paste",
		ToolCategory: "clipboard",
		ToolName:     "Clipboard Paste",
		CommandName:  "rpaste",
		HostAlias:    "dev-box",
		LocalUser:    "myuser",
		ConnectPort:  22222,
	}
	pasteScript, err := GenerateWrapper(pasteCfg)
	if err != nil {
		t.Fatalf("failed to generate paste wrapper: %v", err)
	}
	rpastePath := tmpDir + "/rpaste"
	if err := os.WriteFile(rpastePath, []byte(pasteScript), 0755); err != nil {
		t.Fatalf("failed to write rpaste: %v", err)
	}
	pcmd := exec.Command(rpastePath)
	pcmd.Env = env
	pout, err := pcmd.CombinedOutput()
	if err != nil {
		t.Fatalf("rpaste via fish failed: %v\nOutput:\n%s", err, string(pout))
	}
	if string(pout) != "from_clipboard" {
		t.Errorf("unexpected rpaste output: got %q, want %q", string(pout), "from_clipboard")
	}

	logBytes, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("expected invocation log: %v", err)
	}
	logStr := string(logBytes)
	for _, want := range []string{
		"wu:ssh://dev-box",
		"open:https://example.com?a=1&b=2",
		"open:sftp://dev-box" + tmpDir,
		"open:/tmp/rlink_downloads/dev-box/sample.txt",
	} {
		if !strings.Contains(logStr, want) {
			t.Errorf("expected %q in log, got:\n%s", want, logStr)
		}
	}
}
