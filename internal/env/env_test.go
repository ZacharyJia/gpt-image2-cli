package env

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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

func TestResolveUsesUserConfigBeforeEnvironment(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("OPENAI_API_KEY", "environment-key")
	t.Setenv("BASE_URL", "https://environment.example/v1")

	writeConfig(t, home, `{"api_key":"config-key","base_url":"https://config.example/v1"}`)

	credentials, err := Resolve()
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if credentials.APIKey != "config-key" {
		t.Errorf("APIKey = %q, want config-key", credentials.APIKey)
	}
	if credentials.BaseURL != "https://config.example/v1" {
		t.Errorf("BaseURL = %q, want config URL", credentials.BaseURL)
	}
}

func TestResolveFallsBackToEnvironmentWithoutUserConfig(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("OPENAI_API_KEY", "environment-key")
	t.Setenv("BASE_URL", "https://environment.example/v1")

	credentials, err := Resolve()
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if credentials.APIKey != "environment-key" {
		t.Errorf("APIKey = %q, want environment-key", credentials.APIKey)
	}
	if credentials.BaseURL != "https://environment.example/v1" {
		t.Errorf("BaseURL = %q, want environment URL", credentials.BaseURL)
	}
}

func TestResolveRejectsConfigWithoutAPIKey(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("OPENAI_API_KEY", "environment-key")
	writeConfig(t, home, `{"base_url":"https://config.example/v1"}`)

	_, err := Resolve()
	if err == nil || !strings.Contains(err.Error(), "must contain api_key") {
		t.Fatalf("Resolve() error = %v, want missing api_key error", err)
	}
}

func TestParseConfigRejectsUnknownFieldsAndTrailingData(t *testing.T) {
	if _, err := parseConfig([]byte(`{"api_key":"key","unexpected":true}`)); err == nil {
		t.Fatal("parseConfig() accepted an unknown field")
	}
	if _, err := parseConfig([]byte(`{"api_key":"key"} {}`)); err == nil {
		t.Fatal("parseConfig() accepted trailing JSON")
	}
}

func writeConfig(t *testing.T, home, contents string) {
	t.Helper()
	dir := filepath.Join(home, configDirName)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatalf("MkdirAll(%s): %v", dir, err)
	}
	if err := os.WriteFile(filepath.Join(dir, configFileName), []byte(contents), 0o600); err != nil {
		t.Fatalf("WriteFile(config): %v", err)
	}
}
