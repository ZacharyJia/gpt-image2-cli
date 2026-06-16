package env

import "testing"

func TestSplitEnvLine(t *testing.T) {
	cases := []struct {
		line    string
		wantKey string
		wantVal string
		wantOk  bool
	}{
		{"KEY=value", "KEY", "value", true},
		{"KEY=\"quoted value\"", "KEY", "quoted value", true},
		{"KEY=", "KEY", "", true},
		{"KEY", "", "", false},
		{"# comment", "", "", false},
	}
	for _, tc := range cases {
		k, v, ok := splitEnvLine(tc.line)
		if ok != tc.wantOk || k != tc.wantKey || v != tc.wantVal {
			t.Errorf("splitEnvLine(%q) = (%q, %q, %v), want (%q, %q, %v)",
				tc.line, k, v, ok, tc.wantKey, tc.wantVal, tc.wantOk)
		}
	}
}
