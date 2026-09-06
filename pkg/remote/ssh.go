package remote

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// SSHClient wraps the system SSH and SCP commands for seamless integration
// with ~/.ssh/config, SSH agents, and ProxyJump.
type SSHClient struct {
	HostAlias string
}

// NewSSHClient creates an SSHClient targeting a specific host (either an alias from ssh config or user@host).
func NewSSHClient(hostAlias string) *SSHClient {
	return &SSHClient{HostAlias: hostAlias}
}

// Run executes a command on the remote host via SSH and returns its standard output.
func (c *SSHClient) Run(command string) (string, error) {
	cmd := exec.Command("ssh", "-o", "BatchMode=no", c.HostAlias, command)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("ssh command failed (%s): %s: %w", command, strings.TrimSpace(stderr.String()), err)
	}

	return strings.TrimSpace(stdout.String()), nil
}

// UploadScript writes scriptContent to a remote file path with specified executable permissions.
func (c *SSHClient) UploadScript(remotePath string, scriptContent string) error {
	// We can stream content directly via SSH stdin to avoid writing a local temp file:
	// ssh <host> 'cat > <remotePath> && chmod +x <remotePath>'
	remoteCmd := fmt.Sprintf("mkdir -p $(dirname '%s') && cat > '%s' && chmod +x '%s'", remotePath, remotePath, remotePath)
	cmd := exec.Command("ssh", c.HostAlias, remoteCmd)
	cmd.Stdin = strings.NewReader(scriptContent)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to upload script to %s: %s: %w", remotePath, strings.TrimSpace(stderr.String()), err)
	}

	return nil
}

// DeployPrivateKey writes the private key securely to the remote machine with 0600 permissions.
func (c *SSHClient) DeployPrivateKey(remoteKeyPath string, privateKeyContent string) error {
	remoteCmd := fmt.Sprintf("mkdir -p $(dirname '%s') && chmod 700 $(dirname '%s') && cat > '%s' && chmod 600 '%s'", remoteKeyPath, remoteKeyPath, remoteKeyPath, remoteKeyPath)
	cmd := exec.Command("ssh", c.HostAlias, remoteCmd)
	cmd.Stdin = strings.NewReader(privateKeyContent)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to deploy private key to %s: %s: %w", remoteKeyPath, strings.TrimSpace(stderr.String()), err)
	}

	return nil
}


// DetectRemoteBinDir inspects the remote server's PATH and determines the best directory to install the wrapper.
// Priority:
// 1. ~/.local/bin (user-level, non-root, standard on modern Linux)
// 2. ~/bin
// 3. /usr/local/bin (if writable or in PATH)
func (c *SSHClient) DetectRemoteBinDir() (string, error) {
	// Query remote PATH
	out, err := c.Run("echo $PATH")
	if err != nil {
		return "", fmt.Errorf("failed to check remote $PATH: %w", err)
	}

	remotePath := out
	paths := strings.Split(remotePath, ":")

	// Check candidate directories
	candidates := []string{
		"$HOME/.local/bin",
		"$HOME/bin",
		"/usr/local/bin",
	}

	for _, cand := range candidates {
		// Expand $HOME on remote
		expanded, err := c.Run(fmt.Sprintf("echo %s", cand))
		if err == nil && expanded != "" {
			// Check if expanded path is in PATH
			for _, p := range paths {
				if strings.TrimRight(p, "/") == strings.TrimRight(expanded, "/") {
					return expanded, nil
				}
			}
		}
	}

	// Default fallback to ~/.local/bin and ensure directory exists
	home, err := c.Run("echo $HOME")
	if err != nil || home == "" {
		home = "~"
	}
	defaultDir := home + "/.local/bin"
	return defaultDir, nil
}

// IsDirInRemotePath checks if a given directory is in the remote server's PATH.
func (c *SSHClient) IsDirInRemotePath(dir string) bool {
	out, err := c.Run("echo $PATH")
	if err != nil {
		return false
	}
	paths := strings.Split(out, ":")
	for _, p := range paths {
		if strings.TrimRight(p, "/") == strings.TrimRight(dir, "/") {
			return true
		}
	}
	return false
}
