package auth

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAuthorizeLocalKey(t *testing.T) {
	tempHome := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", origHome)

	testPubKey := "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIGfakekey1234567890 test@local"

	// Test adding key
	err := AuthorizeLocalKey(testPubKey)
	if err != nil {
		t.Fatalf("failed to authorize local key: %v", err)
	}

	authKeysPath := filepath.Join(tempHome, ".ssh", "authorized_keys")
	content, err := os.ReadFile(authKeysPath)
	if err != nil {
		t.Fatalf("failed to read authorized_keys: %v", err)
	}

	if !strings.Contains(string(content), testPubKey) {
		t.Errorf("authorized_keys does not contain test public key")
	}

	// Idempotency: adding again should not duplicate
	err = AuthorizeLocalKey(testPubKey)
	if err != nil {
		t.Fatalf("failed on second authorization: %v", err)
	}

	contentSecond, _ := os.ReadFile(authKeysPath)
	if strings.Count(string(contentSecond), testPubKey) != 1 {
		t.Errorf("expected key to appear exactly once, got %d times", strings.Count(string(contentSecond), testPubKey))
	}
}

func TestEnsureLocalSSHKeyPair(t *testing.T) {
	tempHome := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", origHome)

	keypair, err := EnsureLocalSSHKeyPair()
	if err != nil {
		t.Fatalf("unexpected error ensuring keypair: %v", err)
	}

	if keypair.PrivateKeyContent == "" {
		t.Errorf("private key content is empty")
	}
	if !strings.HasPrefix(keypair.PublicKeyContent, "ssh-ed25519 ") {
		t.Errorf("expected ed25519 public key, got %s", keypair.PublicKeyContent)
	}

	// Calling again should load existing key
	keypair2, err := EnsureLocalSSHKeyPair()
	if err != nil {
		t.Fatalf("second call failed: %v", err)
	}
	if keypair2.PublicKeyContent != keypair.PublicKeyContent {
		t.Errorf("expected same public key on subsequent call")
	}
}
