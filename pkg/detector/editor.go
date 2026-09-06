package detector

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// EditorType identifies supported code editors.
type EditorType string

const (
	EditorZed          EditorType = "zed"
	EditorVSCode       EditorType = "code"
	EditorCursor       EditorType = "cursor"
	EditorWindsurf     EditorType = "windsurf"
	EditorCodeInsiders EditorType = "code-insiders"
	EditorSublime      EditorType = "sublime"
)

// DetectedEditor represents the discovery result of a GUI editor on the local machine.
type DetectedEditor struct {
	Type               EditorType `json:"type"`
	Name               string     `json:"name"`
	BinaryPath         string     `json:"binary_path"`
	IsInstalled        bool       `json:"is_installed"`
	DefaultWrapperName string     `json:"default_wrapper_name"`
	// SyntaxTemplate is the command executed on local machine to open remote folder.
	// Placeholders: {{.Host}} (SSH host alias) and {{.Path}} (absolute remote path)
	SyntaxTemplate string `json:"syntax_template"`
}

// DisplayLabel formats a friendly label for CLI selection.
func (e *DetectedEditor) DisplayLabel() string {
	status := "Not Found"
	if e.IsInstalled {
		status = fmt.Sprintf("Found: %s", e.BinaryPath)
	}
	return fmt.Sprintf("%-24s [%s] -> default: '%s'", e.Name, status, e.DefaultWrapperName)
}

// EditorSpec defines how to search and configure an editor.
type EditorSpec struct {
	Type               EditorType
	Name               string
	Binaries           []string
	MacCandidates      []string
	LinuxCandidates    []string
	DefaultWrapperName string
	SyntaxTemplate     string
}

// SupportedEditors lists all supported GUI editors in priority order.
// Wrapper names follow the 'r<editor>' prefix convention (rzed, rcode, rcursor, rwindsurf, etc.).
var SupportedEditors = []EditorSpec{
	{
		Type:               EditorZed,
		Name:               "Zed",
		Binaries:           []string{"zed", "zed-editor"},
		DefaultWrapperName: "rzed",
		SyntaxTemplate:     `zed "ssh://{{.Host}}{{.Path}}"`,
		MacCandidates: []string{
			"/usr/local/bin/zed",
			"/Applications/Zed.app/Contents/MacOS/cli",
			"/Applications/Zed Preview.app/Contents/MacOS/cli",
		},
		LinuxCandidates: []string{
			"~/.local/bin/zed",
			"/usr/bin/zed",
			"/usr/local/bin/zed",
		},
	},
	{
		Type:               EditorVSCode,
		Name:               "Visual Studio Code",
		Binaries:           []string{"code"},
		DefaultWrapperName: "rcode",
		SyntaxTemplate:     `code --remote ssh-remote+{{.Host}} {{.Path}}`,
		MacCandidates: []string{
			"/usr/local/bin/code",
			"/Applications/Visual Studio Code.app/Contents/Resources/app/bin/code",
		},
		LinuxCandidates: []string{
			"/usr/bin/code",
			"/usr/local/bin/code",
			"/snap/bin/code",
		},
	},
	{
		Type:               EditorCursor,
		Name:               "Cursor",
		Binaries:           []string{"cursor"},
		DefaultWrapperName: "rcursor",
		SyntaxTemplate:     `cursor --remote ssh-remote+{{.Host}} {{.Path}}`,
		MacCandidates: []string{
			"/usr/local/bin/cursor",
			"/Applications/Cursor.app/Contents/Resources/app/bin/cursor",
		},
		LinuxCandidates: []string{
			"/usr/bin/cursor",
			"/usr/local/bin/cursor",
			"~/.local/bin/cursor",
		},
	},
	{
		Type:               EditorWindsurf,
		Name:               "Windsurf",
		Binaries:           []string{"windsurf"},
		DefaultWrapperName: "rwindsurf",
		SyntaxTemplate:     `windsurf --remote ssh-remote+{{.Host}} {{.Path}}`,
		MacCandidates: []string{
			"/usr/local/bin/windsurf",
			"/Applications/Windsurf.app/Contents/Resources/app/bin/windsurf",
		},
		LinuxCandidates: []string{
			"/usr/bin/windsurf",
			"/usr/local/bin/windsurf",
			"~/.local/bin/windsurf",
		},
	},
	{
		Type:               EditorCodeInsiders,
		Name:               "VS Code Insiders",
		Binaries:           []string{"code-insiders"},
		DefaultWrapperName: "rcode-insiders",
		SyntaxTemplate:     `code-insiders --remote ssh-remote+{{.Host}} {{.Path}}`,
		MacCandidates: []string{
			"/usr/local/bin/code-insiders",
			"/Applications/Visual Studio Code - Insiders.app/Contents/Resources/app/bin/code-insiders",
		},
		LinuxCandidates: []string{
			"/usr/bin/code-insiders",
			"/snap/bin/code-insiders",
		},
	},
	{
		Type:               EditorSublime,
		Name:               "Sublime Text",
		Binaries:           []string{"subl"},
		DefaultWrapperName: "rsubl",
		SyntaxTemplate:     `subl {{.Path}}`,
		MacCandidates: []string{
			"/usr/local/bin/subl",
			"/Applications/Sublime Text.app/Contents/SharedSupport/bin/subl",
		},
		LinuxCandidates: []string{
			"/usr/bin/subl",
			"~/.local/bin/subl",
		},
	},
}

// DetectInstalledEditors scans the local machine for all supported GUI editors.
func DetectInstalledEditors() []DetectedEditor {
	results := make([]DetectedEditor, 0, len(SupportedEditors))

	for _, spec := range SupportedEditors {
		detected := DetectedEditor{
			Type:               spec.Type,
			Name:               spec.Name,
			DefaultWrapperName: spec.DefaultWrapperName,
			SyntaxTemplate:     spec.SyntaxTemplate,
		}

		// 1. Look in system PATH
		for _, bin := range spec.Binaries {
			if path, err := exec.LookPath(bin); err == nil {
				detected.IsInstalled = true
				detected.BinaryPath = path
				break
			}
		}

		// 2. If not found in PATH, check OS-specific candidate locations
		if !detected.IsInstalled {
			var candidates []string
			if runtime.GOOS == "darwin" {
				candidates = spec.MacCandidates
			} else if runtime.GOOS == "linux" {
				candidates = spec.LinuxCandidates
			}

			home, _ := os.UserHomeDir()
			for _, cand := range candidates {
				expanded := cand
				if strings.HasPrefix(cand, "~/") && home != "" {
					expanded = filepath.Join(home, cand[2:])
				}

				if fi, err := os.Stat(expanded); err == nil && !fi.IsDir() {
					detected.IsInstalled = true
					detected.BinaryPath = expanded
					break
				}
			}
		}

		results = append(results, detected)
	}

	return results
}
