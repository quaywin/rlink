package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseSSHConfigFile(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config")

	sampleConfig := `
# Global default
Host *
    ServerAliveInterval 60

Host dev-server
    HostName 192.168.1.100
    User ubuntu
    Port 2222
    IdentityFile ~/.ssh/id_rsa
    RemoteForward 22222 localhost:22

Host prod-api
    HostName api.example.com
    User deploy
`
	if err := os.WriteFile(configPath, []byte(sampleConfig), 0600); err != nil {
		t.Fatalf("failed to write test ssh config: %v", err)
	}

	hosts, err := ParseSSHConfigFile(configPath)
	if err != nil {
		t.Fatalf("unexpected error parsing ssh config: %v", err)
	}

	// Host * should be filtered out
	if len(hosts) != 2 {
		t.Fatalf("expected 2 hosts, got %d", len(hosts))
	}

	h1 := hosts[0]
	if h1.Alias != "dev-server" {
		t.Errorf("expected alias 'dev-server', got %s", h1.Alias)
	}
	if h1.HostName != "192.168.1.100" {
		t.Errorf("expected hostname '192.168.1.100', got %s", h1.HostName)
	}
	if h1.User != "ubuntu" {
		t.Errorf("expected user 'ubuntu', got %s", h1.User)
	}
	if h1.Port != 2222 {
		t.Errorf("expected port 2222, got %d", h1.Port)
	}
	if len(h1.RemoteForwards) != 1 || h1.RemoteForwards[0] != "22222 localhost:22" {
		t.Errorf("expected RemoteForward '22222 localhost:22', got %v", h1.RemoteForwards)
	}

	h2 := hosts[1]
	if h2.Alias != "prod-api" {
		t.Errorf("expected alias 'prod-api', got %s", h2.Alias)
	}
	if h2.Port != 22 {
		t.Errorf("expected default port 22, got %d", h2.Port)
	}
}

func TestInjectRemoteForwardExistingHost(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config")

	sampleConfig := `Host dev-server
    HostName 192.168.1.100
    User ubuntu

Host prod-api
    HostName api.example.com
`
	if err := os.WriteFile(configPath, []byte(sampleConfig), 0600); err != nil {
		t.Fatalf("failed to write test ssh config: %v", err)
	}

	err := InjectRemoteForward(configPath, "dev-server", 22222, 22)
	if err != nil {
		t.Fatalf("unexpected error injecting remoteforward: %v", err)
	}

	bytes, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read back config: %v", err)
	}
	content := string(bytes)

	if !strings.Contains(content, "RemoteForward 22222 localhost:22") {
		t.Errorf("config does not contain injected RemoteForward directive:\n%s", content)
	}

	// Idempotency check: injecting again should not duplicate
	err = InjectRemoteForward(configPath, "dev-server", 22222, 22)
	if err != nil {
		t.Fatalf("idempotent injection failed: %v", err)
	}

	bytesAfter, _ := os.ReadFile(configPath)
	if strings.Count(string(bytesAfter), "RemoteForward 22222 localhost:22") != 1 {
		t.Errorf("expected exactly 1 RemoteForward line, got:\n%s", string(bytesAfter))
	}
}

func TestInjectRemoteForwardNewHost(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config")

	err := InjectRemoteForward(configPath, "new-vps", 22222, 22)
	if err != nil {
		t.Fatalf("failed to inject into non-existent config: %v", err)
	}

	bytes, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read back config: %v", err)
	}
	content := string(bytes)

	if !strings.Contains(content, "Host new-vps") || !strings.Contains(content, "RemoteForward 22222 localhost:22") {
		t.Errorf("unexpected content in new config:\n%s", content)
	}
}

func TestGetExistingTunnelPort(t *testing.T) {
	h1 := &HostConfig{
		Alias:          "srv1",
		RemoteForwards: []string{"22222 localhost:22"},
	}
	if port := h1.GetExistingTunnelPort(); port != 22222 {
		t.Errorf("expected 22222, got %d", port)
	}

	h2 := &HostConfig{
		Alias:          "srv2",
		RemoteForwards: []string{"33333 127.0.0.1:22"},
	}
	if port := h2.GetExistingTunnelPort(); port != 33333 {
		t.Errorf("expected 33333, got %d", port)
	}

	h3 := &HostConfig{
		Alias:          "srv3",
		RemoteForwards: []string{"8080 localhost:80"},
	}
	if port := h3.GetExistingTunnelPort(); port != 0 {
		t.Errorf("expected 0 for non-22 forward, got %d", port)
	}
}

func TestRemoveRemoteForward(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config")

	sampleConfig := `Host dev-server
    HostName 192.168.1.100
    RemoteForward 22222 localhost:22
    User ubuntu

Host prod-api
    HostName api.example.com
`
	if err := os.WriteFile(configPath, []byte(sampleConfig), 0600); err != nil {
		t.Fatalf("failed to write test ssh config: %v", err)
	}

	err := RemoveRemoteForward(configPath, "dev-server")
	if err != nil {
		t.Fatalf("unexpected error removing forward: %v", err)
	}

	bytes, _ := os.ReadFile(configPath)
	content := string(bytes)
	if strings.Contains(content, "RemoteForward") {
		t.Errorf("expected RemoteForward to be removed, but still present:\n%s", content)
	}
	if !strings.Contains(content, "Host dev-server") || !strings.Contains(content, "User ubuntu") {
		t.Errorf("other directives were lost:\n%s", content)
	}
}


