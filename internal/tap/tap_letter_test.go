package tap

import "testing"

func TestFormulaFirstLetterDir(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"", "@"},
		{"ripgrep", "r"},
		{"Ripgrep", "r"},
		{"9patch", "0"},
		{"@scope/pkg", "@"},
	}
	for _, tt := range tests {
		if got := formulaFirstLetterDir(tt.name); got != tt.want {
			t.Errorf("formulaFirstLetterDir(%q) = %q, want %q", tt.name, got, tt.want)
		}
	}
}
