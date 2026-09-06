package template

import (
	"bytes"
	_ "embed"
	"fmt"
	"strings"
	"text/template"
)

//go:embed wrapper.sh.tmpl
var wrapperTemplateSource string

// ConnectionMode defines whether connection is via reverse tunnel or direct IP.
type ConnectionMode string

const (
	ModeReverseTunnel ConnectionMode = "tunnel"
	ModeDirectIP      ConnectionMode = "direct"
)

// WrapperConfig contains the template parameters to generate the remote script.
type WrapperConfig struct {
	EditorName     string
	CommandName    string // e.g. "zr", "cr", "cur"
	HostAlias      string // SSH Host alias as recognized by local machine
	LocalUser      string // Username on local machine to SSH into
	ConnectHost    string // "127.0.0.1" for tunnel, or Tailscale/LAN IP
	ConnectPort    int    // Forward port (e.g. 22222) or local SSH port (22)
	SSHKeyPath     string // Path to identity file on remote server, e.g. "~/.ssh/id_rsa"
	ConnectionMode ConnectionMode
	// SyntaxPattern: Command pattern with __RLINK_HOST__ and __RLINK_PATH__ tokens
	SyntaxPattern string
}

// GenerateWrapper renders the POSIX wrapper script for the remote machine.
func GenerateWrapper(cfg WrapperConfig) (string, error) {
	tmpl, err := template.New("wrapper").Parse(wrapperTemplateSource)
	if err != nil {
		return "", fmt.Errorf("failed to parse wrapper template: %w", err)
	}

	// Prepare SyntaxPattern with tokens if using Go template syntax
	// e.g. `zed "ssh://{{.Host}}{{.Path}}"` -> `zed "ssh://__RLINK_HOST____RLINK_PATH__"`
	pattern := cfg.SyntaxPattern
	pattern = strings.ReplaceAll(pattern, "{{.Host}}", "__RLINK_HOST__")
	pattern = strings.ReplaceAll(pattern, "{{.Path}}", "__RLINK_PATH__")
	cfg.SyntaxPattern = pattern

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, cfg); err != nil {
		return "", fmt.Errorf("failed to execute wrapper template: %w", err)
	}

	return buf.String(), nil
}
