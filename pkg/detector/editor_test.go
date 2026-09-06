package detector

import (
	"testing"
)

func TestDetectInstalledEditors(t *testing.T) {
	editors := DetectInstalledEditors()

	if len(editors) != 3 {
		t.Fatalf("expected 3 editor specifications, got %d", len(editors))
	}

	hasZed := false
	for _, ed := range editors {
		if ed.Type == EditorZed {
			hasZed = true
			t.Logf("Zed detection: installed=%v, path=%s", ed.IsInstalled, ed.BinaryPath)
		}
	}

	if !hasZed {
		t.Errorf("Zed editor was not in the detection list")
	}
}
