package detector

import (
	"runtime"
	"strings"
	"testing"
)

func TestDetectSystemTools(t *testing.T) {
	tools := DetectSystemTools()

	if len(tools) != 3 {
		t.Fatalf("expected 3 system tools, got %d", len(tools))
	}

	foundOpener := false
	foundClipCopy := false
	foundClipPaste := false

	for _, tool := range tools {
		switch tool.Type {
		case ToolOpener:
			foundOpener = true
			if tool.DefaultWrapperName != "ropen" {
				t.Errorf("expected default wrapper 'ropen', got %s", tool.DefaultWrapperName)
			}
			if tool.Category != CategoryOpener {
				t.Errorf("expected CategoryOpener, got %s", tool.Category)
			}
			if runtime.GOOS == "darwin" && !tool.IsInstalled {
				t.Errorf("expected 'open' to be installed on macOS")
			}
		case ToolClipCopy:
			foundClipCopy = true
			if tool.DefaultWrapperName != "rclip" {
				t.Errorf("expected default wrapper 'rclip', got %s", tool.DefaultWrapperName)
			}
			if tool.Category != CategoryClipboard {
				t.Errorf("expected CategoryClipboard, got %s", tool.Category)
			}
			if runtime.GOOS == "darwin" && !tool.IsInstalled {
				t.Errorf("expected 'pbcopy' to be installed on macOS")
			}
		case ToolClipPaste:
			foundClipPaste = true
			if tool.DefaultWrapperName != "rpaste" {
				t.Errorf("expected default wrapper 'rpaste', got %s", tool.DefaultWrapperName)
			}
			if tool.Category != CategoryClipboard {
				t.Errorf("expected CategoryClipboard, got %s", tool.Category)
			}
			if runtime.GOOS == "darwin" && !tool.IsInstalled {
				t.Errorf("expected 'pbpaste' to be installed on macOS")
			}
		}
	}

	if !foundOpener || !foundClipCopy || !foundClipPaste {
		t.Errorf("missing system tools: opener=%v, clipCopy=%v, clipPaste=%v", foundOpener, foundClipCopy, foundClipPaste)
	}
}

func TestDetectAllTools(t *testing.T) {
	allTools := DetectAllTools()
	expectedCount := len(SupportedEditors) + len(GetSystemToolSpecs())

	if len(allTools) != expectedCount {
		t.Fatalf("expected %d total tools, got %d", expectedCount, len(allTools))
	}
}

func TestValidateCustomCommand(t *testing.T) {
	// 1. Existing command (sh is present on all Unix/macOS)
	custom, err := ValidateCustomCommand("sh")
	if err != nil {
		t.Fatalf("unexpected error validating 'sh': %v", err)
	}
	if custom.Type != ToolCustom {
		t.Errorf("expected ToolCustom, got %s", custom.Type)
	}
	if custom.Category != CategoryCustom {
		t.Errorf("expected CategoryCustom, got %s", custom.Category)
	}
	if custom.DefaultWrapperName != "rsh" {
		t.Errorf("expected wrapper name 'rsh', got %s", custom.DefaultWrapperName)
	}
	if !custom.IsInstalled {
		t.Errorf("expected 'sh' to be marked as installed")
	}

	// 2. Non-existent command
	_, err = ValidateCustomCommand("non_existent_binary_xyz_98765")
	if err == nil {
		t.Fatalf("expected error for non-existent command, got nil")
	}

	// 3. Empty command
	_, err = ValidateCustomCommand("   ")
	if err == nil {
		t.Fatalf("expected error for empty command, got nil")
	}
}

func TestDisplayLabel(t *testing.T) {
	tool := &DetectedTool{
		Type:               ToolOpener,
		Category:           CategoryOpener,
		Name:               "Web & File Opener",
		IsInstalled:        true,
		BinaryPath:         "/usr/bin/open",
		DefaultWrapperName: "ropen",
	}

	label := tool.DisplayLabel()
	if !strings.Contains(label, "🌐") {
		t.Errorf("expected globe icon for opener, got: %s", label)
	}
	if !strings.Contains(label, "Found: /usr/bin/open") {
		t.Errorf("expected found path in label, got: %s", label)
	}
	if !strings.Contains(label, "'ropen'") {
		t.Errorf("expected default wrapper name in label, got: %s", label)
	}
}
