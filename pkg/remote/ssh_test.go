package remote

import "testing"

func TestSanitizeRemotePath(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"~/.local/bin/zr", "$HOME/.local/bin/zr"},
		{"~", "$HOME"},
		{"/usr/local/bin/zr", "/usr/local/bin/zr"},
		{"bin/zr", "bin/zr"},
	}

	for _, tc := range tests {
		result := sanitizeRemotePath(tc.input)
		if result != tc.expected {
			t.Errorf("sanitizeRemotePath(%q) = %q; expected %q", tc.input, result, tc.expected)
		}
	}
}
