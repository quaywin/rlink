package template

import (
	"bytes"
	_ "embed"
	"fmt"
	"strings"
	"text/template"
)

//go:embed wrapper.sh.tmpl
var editorTemplateSource string

//go:embed wrapper_opener.sh.tmpl
var openerTemplateSource string

//go:embed wrapper_clip.sh.tmpl
var clipTemplateSource string

//go:embed wrapper_paste.sh.tmpl
var pasteTemplateSource string

//go:embed wrapper_custom.sh.tmpl
var customTemplateSource string

// WrapperConfig contains the template parameters to generate the remote script.
type WrapperConfig struct {
	ToolType      string // "zed", "code", "open", "clip", "paste", "custom"
	ToolCategory  string // "editor", "opener", "clipboard", "custom"
	ToolName      string // Display name, e.g. "Zed", "Web & File Opener"
	EditorName    string // Backwards compatibility alias for ToolName
	CommandName   string // Remote wrapper command name, e.g. "rzed", "ropen", "rclip"
	HostAlias     string // SSH Host alias as recognized by local machine
	LocalUser     string // Username on local machine to SSH into
	ConnectPort   int    // Reverse forward port (e.g. 22222)
	SSHKeyPath    string // Path to identity file on remote server, e.g. "$HOME/.ssh/rlink_id_ed25519"
	SyntaxPattern string // Command pattern with __RLINK_HOST__ and __RLINK_PATH__ tokens
	LocalBinary   string // Local executable name or path for custom commands
}

// GenerateWrapper renders the POSIX wrapper script for the remote machine.
func GenerateWrapper(cfg WrapperConfig) (string, error) {
	// Reconcile ToolName and EditorName for backwards compatibility
	if cfg.ToolName == "" && cfg.EditorName != "" {
		cfg.ToolName = cfg.EditorName
	}
	if cfg.EditorName == "" && cfg.ToolName != "" {
		cfg.EditorName = cfg.ToolName
	}
	if cfg.LocalBinary == "" {
		cfg.LocalBinary = cfg.CommandName
	}

	var rawTemplate string

	// Select appropriate template based on category and tool type
	switch {
	case cfg.ToolCategory == "opener" || cfg.ToolType == "open":
		rawTemplate = openerTemplateSource

	case cfg.ToolCategory == "clipboard" || cfg.ToolType == "clip" || cfg.ToolType == "paste":
		if cfg.ToolType == "paste" {
			rawTemplate = pasteTemplateSource
		} else {
			rawTemplate = clipTemplateSource
		}

	case cfg.ToolCategory == "custom" || cfg.ToolType == "custom":
		rawTemplate = customTemplateSource

	default: // Category "editor" or legacy editor calls
		rawTemplate = editorTemplateSource

		// Prepare SyntaxPattern with tokens if using Go template syntax
		// e.g. `zed "ssh://{{.Host}}{{.Path}}"` -> `zed "ssh://__RLINK_HOST____RLINK_PATH__"`
		pattern := cfg.SyntaxPattern
		pattern = strings.ReplaceAll(pattern, "{{.Host}}", "__RLINK_HOST__")
		pattern = strings.ReplaceAll(pattern, "{{.Path}}", "__RLINK_PATH__")
		cfg.SyntaxPattern = pattern
	}

	tmpl, err := template.New("wrapper").Parse(rawTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse wrapper template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, cfg); err != nil {
		return "", fmt.Errorf("failed to execute wrapper template: %w", err)
	}

	return buf.String(), nil
}
