package detector

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// ToolCategory categorizes supported remote-bridgeable tools.
type ToolCategory string

const (
	CategoryEditor    ToolCategory = "editor"
	CategoryOpener    ToolCategory = "opener"
	CategoryClipboard ToolCategory = "clipboard"
	CategoryCustom    ToolCategory = "custom"
)

// ToolType identifies specific tools.
type ToolType string

const (
	// GUI Code Editors
	ToolZed          ToolType = "zed"
	ToolVSCode       ToolType = "code"
	ToolCursor       ToolType = "cursor"
	ToolWindsurf     ToolType = "windsurf"
	ToolCodeInsiders ToolType = "code-insiders"
	ToolSublime      ToolType = "sublime"

	// System Opener
	ToolOpener ToolType = "open"

	// System Clipboard
	ToolClipCopy  ToolType = "clip"
	ToolClipPaste ToolType = "paste"

	// Custom
	ToolCustom ToolType = "custom"
)

// Backward compatibility aliases for existing editor references.
type EditorType = ToolType

const (
	EditorZed          = ToolZed
	EditorVSCode       = ToolVSCode
	EditorCursor       = ToolCursor
	EditorWindsurf     = ToolWindsurf
	EditorCodeInsiders = ToolCodeInsiders
	EditorSublime      = ToolSublime
)

// DetectedTool represents the discovery result of a tool or editor on the local machine.
type DetectedTool struct {
	Type               ToolType     `json:"type"`
	Category           ToolCategory `json:"category"`
	Name               string       `json:"name"`
	Description        string       `json:"description"`
	BinaryPath         string       `json:"binary_path"`
	IsInstalled        bool         `json:"is_installed"`
	DefaultWrapperName string       `json:"default_wrapper_name"`
	// SyntaxTemplate is the command executed on local machine.
	// Placeholders: {{.Host}}, {{.Path}}, {{.Args}}
	SyntaxTemplate string `json:"syntax_template"`
}

// Backward compatibility alias for DetectedEditor.
type DetectedEditor = DetectedTool

// DisplayLabel formats a friendly label for interactive CLI selection.
func (t *DetectedTool) DisplayLabel() string {
	status := "Not Found"
	if t.IsInstalled {
		status = fmt.Sprintf("Found: %s", t.BinaryPath)
	}

	categoryIcon := "📝"
	switch t.Category {
	case CategoryOpener:
		categoryIcon = "🌐"
	case CategoryClipboard:
		categoryIcon = "📋"
	case CategoryCustom:
		categoryIcon = "⚙️ "
	}

	return fmt.Sprintf("%s %-22s [%s] -> default: '%s'", categoryIcon, t.Name, status, t.DefaultWrapperName)
}

// ToolSpec defines how to search and configure a tool.
type ToolSpec struct {
	Type               ToolType
	Category           ToolCategory
	Name               string
	Description        string
	Binaries           []string
	MacCandidates      []string
	LinuxCandidates    []string
	DefaultWrapperName string
	SyntaxTemplate     string
}

// Backward compatibility alias for EditorSpec.
type EditorSpec = ToolSpec

// SupportedEditors lists all supported GUI editors in priority order.
var SupportedEditors = []ToolSpec{
	{
		Type:               ToolZed,
		Category:           CategoryEditor,
		Name:               "Zed",
		Description:        "High-performance, multiplayer code editor",
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
		Type:               ToolVSCode,
		Category:           CategoryEditor,
		Name:               "Visual Studio Code",
		Description:        "Code editing redefined",
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
		Type:               ToolCursor,
		Category:           CategoryEditor,
		Name:               "Cursor",
		Description:        "The AI-first Code Editor",
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
		Type:               ToolWindsurf,
		Category:           CategoryEditor,
		Name:               "Windsurf",
		Description:        "Agentic AI IDE by Codeium",
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
		Type:               ToolCodeInsiders,
		Category:           CategoryEditor,
		Name:               "VS Code Insiders",
		Description:        "Early preview build of VS Code",
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
		Type:               ToolSublime,
		Category:           CategoryEditor,
		Name:               "Sublime Text",
		Description:        "Sophisticated text editor for code, markup and prose",
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

// GetSystemToolSpecs returns specs for OS-level utility tools (open, clip, paste).
func GetSystemToolSpecs() []ToolSpec {
	var openerBinaries, openerMac, openerLinux []string
	var openerSyntax string

	var clipCopyBinaries, clipCopyMac, clipCopyLinux []string
	var clipCopySyntax string

	var clipPasteBinaries, clipPasteMac, clipPasteLinux []string
	var clipPasteSyntax string

	if runtime.GOOS == "darwin" {
		openerBinaries = []string{"open"}
		openerMac = []string{"/usr/bin/open"}
		openerSyntax = `open "{{.Path}}"`

		clipCopyBinaries = []string{"pbcopy"}
		clipCopyMac = []string{"/usr/bin/pbcopy"}
		clipCopySyntax = `pbcopy`

		clipPasteBinaries = []string{"pbpaste"}
		clipPasteMac = []string{"/usr/bin/pbpaste"}
		clipPasteSyntax = `pbpaste`
	} else {
		// Linux & BSD defaults
		openerBinaries = []string{"xdg-open", "gio"}
		openerLinux = []string{"/usr/bin/xdg-open", "/usr/local/bin/xdg-open"}
		openerSyntax = `xdg-open "{{.Path}}"`

		clipCopyBinaries = []string{"wl-copy", "xclip", "xsel"}
		clipCopyLinux = []string{"/usr/bin/wl-copy", "/usr/bin/xclip", "/usr/bin/xsel"}
		clipCopySyntax = `wl-copy 2>/dev/null || xclip -selection clipboard 2>/dev/null || xsel --clipboard 2>/dev/null`

		clipPasteBinaries = []string{"wl-paste", "xclip", "xsel"}
		clipPasteLinux = []string{"/usr/bin/wl-paste", "/usr/bin/xclip", "/usr/bin/xsel"}
		clipPasteSyntax = `wl-paste 2>/dev/null || xclip -selection clipboard -o 2>/dev/null || xsel --clipboard -o 2>/dev/null`
	}

	return []ToolSpec{
		{
			Type:               ToolOpener,
			Category:           CategoryOpener,
			Name:               "Web & File Opener",
			Description:        "Open URLs in local browser or stream remote files to view locally",
			Binaries:           openerBinaries,
			DefaultWrapperName: "ropen",
			SyntaxTemplate:     openerSyntax,
			MacCandidates:      openerMac,
			LinuxCandidates:    openerLinux,
		},
		{
			Type:               ToolClipCopy,
			Category:           CategoryClipboard,
			Name:               "Clipboard Copy",
			Description:        "Pipe remote stdout / files directly to local system clipboard",
			Binaries:           clipCopyBinaries,
			DefaultWrapperName: "rclip",
			SyntaxTemplate:     clipCopySyntax,
			MacCandidates:      clipCopyMac,
			LinuxCandidates:    clipCopyLinux,
		},
		{
			Type:               ToolClipPaste,
			Category:           CategoryClipboard,
			Name:               "Clipboard Paste",
			Description:        "Stream local system clipboard contents into remote terminal",
			Binaries:           clipPasteBinaries,
			DefaultWrapperName: "rpaste",
			SyntaxTemplate:     clipPasteSyntax,
			MacCandidates:      clipPasteMac,
			LinuxCandidates:    clipPasteLinux,
		},
	}
}

// detectSpec discovers if a given ToolSpec is installed locally.
func detectSpec(spec ToolSpec) DetectedTool {
	detected := DetectedTool{
		Type:               spec.Type,
		Category:           spec.Category,
		Name:               spec.Name,
		Description:        spec.Description,
		DefaultWrapperName: spec.DefaultWrapperName,
		SyntaxTemplate:     spec.SyntaxTemplate,
	}

	// 1. Search in system PATH
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

	return detected
}

// DetectInstalledEditors scans the local machine for all supported GUI editors.
func DetectInstalledEditors() []DetectedTool {
	results := make([]DetectedTool, 0, len(SupportedEditors))
	for _, spec := range SupportedEditors {
		results = append(results, detectSpec(spec))
	}
	return results
}

// DetectSystemTools scans the local machine for system opener and clipboard tools.
func DetectSystemTools() []DetectedTool {
	specs := GetSystemToolSpecs()
	results := make([]DetectedTool, 0, len(specs))
	for _, spec := range specs {
		results = append(results, detectSpec(spec))
	}
	return results
}

// DetectAllTools returns both GUI editors and system tools.
func DetectAllTools() []DetectedTool {
	editors := DetectInstalledEditors()
	systemTools := DetectSystemTools()
	return append(editors, systemTools...)
}

// ValidateCustomCommand checks if a custom binary exists locally and generates a DetectedTool specification.
func ValidateCustomCommand(binOrPath string) (*DetectedTool, error) {
	cleanInput := strings.TrimSpace(binOrPath)
	if cleanInput == "" {
		return nil, fmt.Errorf("command name cannot be empty")
	}

	binaryPath, err := exec.LookPath(cleanInput)
	if err != nil {
		// Try absolute/relative file path
		if fi, statErr := os.Stat(cleanInput); statErr == nil && !fi.IsDir() {
			binaryPath = cleanInput
		} else {
			return nil, fmt.Errorf("command %q not found in local PATH or filesystem: %w", cleanInput, err)
		}
	}

	baseName := filepath.Base(cleanInput)
	wrapperName := "r" + baseName

	return &DetectedTool{
		Type:               ToolCustom,
		Category:           CategoryCustom,
		Name:               baseName,
		Description:        fmt.Sprintf("Custom local command: %s", binaryPath),
		BinaryPath:         binaryPath,
		IsInstalled:        true,
		DefaultWrapperName: wrapperName,
		SyntaxTemplate:     fmt.Sprintf("%s {{.Args}}", cleanInput),
	}, nil
}
