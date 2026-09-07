package auth

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// KeyPair holds paths and contents for an SSH keypair.
type KeyPair struct {
	PrivateKeyPath    string
	PublicKeyPath     string
	PrivateKeyContent string
	PublicKeyContent  string
}

// getUserSSHDir returns the local user's ~/.ssh directory.
func getUserSSHDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}
	return filepath.Join(home, ".ssh"), nil
}

// DefaultKeyPaths returns the standard location for rlink's dedicated keypair.
func DefaultKeyPaths() (priv string, pub string, err error) {
	sshDir, err := getUserSSHDir()
	if err != nil {
		return "", "", err
	}
	return filepath.Join(sshDir, "rlink_ed25519"), filepath.Join(sshDir, "rlink_ed25519.pub"), nil
}

// EnsureLocalSSHKeyPair checks for an existing rlink keypair or creates a new dedicated Ed25519 keypair.
func EnsureLocalSSHKeyPair() (*KeyPair, error) {
	privPath, pubPath, err := DefaultKeyPaths()
	if err != nil {
		return nil, err
	}

	sshDir := filepath.Dir(privPath)
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create local .ssh directory: %w", err)
	}

	// If key does not exist, generate it
	if _, err := os.Stat(privPath); os.IsNotExist(err) {
		cmd := exec.Command("ssh-keygen", "-t", "ed25519", "-N", "", "-C", "rlink-remote-trigger", "-f", privPath)
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			return nil, fmt.Errorf("failed to generate ed25519 keypair: %s: %w", strings.TrimSpace(stderr.String()), err)
		}
		_ = os.Chmod(privPath, 0600)
		_ = os.Chmod(pubPath, 0644)
	}

	privBytes, err := os.ReadFile(privPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key: %w", err)
	}

	pubBytes, err := os.ReadFile(pubPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read public key: %w", err)
	}

	return &KeyPair{
		PrivateKeyPath:    privPath,
		PublicKeyPath:     pubPath,
		PrivateKeyContent: string(privBytes),
		PublicKeyContent:  strings.TrimSpace(string(pubBytes)),
	}, nil
}

// AuthorizeLocalKey ensures the given public key is listed in ~/.ssh/authorized_keys.
func AuthorizeLocalKey(pubKeyContent string) error {
	sshDir, err := getUserSSHDir()
	if err != nil {
		return err
	}
	authKeysPath := filepath.Join(sshDir, "authorized_keys")

	if err := os.MkdirAll(sshDir, 0700); err != nil {
		return fmt.Errorf("failed to create .ssh directory: %w", err)
	}

	cleanKey := strings.TrimSpace(pubKeyContent)
	if cleanKey == "" {
		return fmt.Errorf("public key content is empty")
	}

	existingBytes, err := os.ReadFile(authKeysPath)
	if err == nil {
		existing := string(existingBytes)
		if strings.Contains(existing, cleanKey) {
			// Already authorized
			return nil
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to read authorized_keys: %w", err)
	}

	// Append to authorized_keys
	f, err := os.OpenFile(authKeysPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return fmt.Errorf("failed to open authorized_keys for writing: %w", err)
	}
	defer f.Close()

	entry := fmt.Sprintf("\n# rlink remote trigger key\n%s\n", cleanKey)
	if _, err := f.WriteString(entry); err != nil {
		return fmt.Errorf("failed to append key to authorized_keys: %w", err)
	}

	return nil
}
