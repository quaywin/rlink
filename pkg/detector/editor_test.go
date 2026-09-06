package detector

import (
	"testing"
)

func TestDetectInstalledEditors(t *testing.T) {
	editors := DetectInstalledEditors()

	if len(editors) != 6 {
		t.Fatalf("expected 6 editor specifications, got %d", len(editors))
	}

	hasZed := false
	for _, ed := range editors {
		if ed.Type == EditorZed {
			hasZed = true
			if ed.DefaultWrapperName != "rzed" {
				t.Errorf("expected default wrapper 'rzed', got %s", ed.DefaultWrapperName)
			}
			t.Logf("Zed detection: installed=%v, path=%s", ed.IsInstalled, ed.BinaryPath)
		}
		if ed.Type == EditorVSCode && ed.DefaultWrapperName != "rcode" {
			t.Errorf("expected default wrapper 'rcode', got %s", ed.DefaultWrapperName)
		}
		if ed.Type == EditorCursor && ed.DefaultWrapperName != "rcursor" {
			t.Errorf("expected default wrapper 'rcursor', got %s", ed.DefaultWrapperName)
		}
		if ed.Type == EditorWindsurf && ed.DefaultWrapperName != "rwindsurf" {
			t.Errorf("expected default wrapper 'rwindsurf', got %s", ed.DefaultWrapperName)
		}
	}

	if !hasZed {
		t.Errorf("Zed editor was not in the detection list")
	}
}
