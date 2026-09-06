package cmd

import (
	"testing"
)

func TestSetupFlags(t *testing.T) {
	cmd := setupCmd

	toolFlag := cmd.Flags().Lookup("tool")
	if toolFlag == nil {
		t.Fatalf("expected --tool flag to exist")
	}
	if toolFlag.Shorthand != "t" {
		t.Errorf("expected shorthand 't' for --tool, got %s", toolFlag.Shorthand)
	}

	editorFlag := cmd.Flags().Lookup("editor")
	if editorFlag == nil {
		t.Fatalf("expected --editor flag to exist")
	}
	if editorFlag.Shorthand != "e" {
		t.Errorf("expected shorthand 'e' for --editor, got %s", editorFlag.Shorthand)
	}

	nameFlag := cmd.Flags().Lookup("name")
	if nameFlag == nil {
		t.Fatalf("expected --name flag to exist")
	}

	hostFlag := cmd.Flags().Lookup("host")
	if hostFlag == nil {
		t.Fatalf("expected --host flag to exist")
	}

	portFlag := cmd.Flags().Lookup("port")
	if portFlag == nil {
		t.Fatalf("expected --port flag to exist")
	}

	yesFlag := cmd.Flags().Lookup("yes")
	if yesFlag == nil {
		t.Fatalf("expected --yes flag to exist")
	}
}
