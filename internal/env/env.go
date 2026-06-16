// Package env resolves OPENAI_API_KEY from process env, .env, and ~/.env
// without overriding existing environment variables.
package env

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// LoadChain loads .env files in order: ./.env then ~/..env.
// Existing process environment variables are never overwritten.
func LoadChain() error {
	home, err := os.UserHomeDir()
	if err != nil {
		home = ""
	}
	for _, path := range []string{
		".env",
		filepath.Join(home, ".env"),
	} {
		if path == "" {
			continue
		}
		if err := loadFile(path); err != nil {
			// Ignore missing or unreadable files silently.
			if !os.IsNotExist(err) {
				// Best-effort: we don't want a permissions issue to crash the CLI.
				_ = err
			}
		}
	}
	return nil
}

func loadFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := splitEnvLine(line)
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		// Do not override an already-set environment variable.
		if _, set := os.LookupEnv(key); set {
			continue
		}
		if err := os.Setenv(key, value); err != nil {
			return fmt.Errorf("setenv %s: %w", key, err)
		}
	}
	return scanner.Err()
}

func splitEnvLine(line string) (string, string, bool) {
	// Supports KEY=VALUE and KEY="VALUE" with simple unescaping.
	idx := strings.Index(line, "=")
	if idx <= 0 {
		return "", "", false
	}
	key := line[:idx]
	value := strings.TrimSpace(line[idx+1:])
	if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
		value = strings.Trim(value, "\"")
		value = strings.ReplaceAll(value, "\\n", "\n")
		value = strings.ReplaceAll(value, "\\t", "\t")
		value = strings.ReplaceAll(value, "\\\"", "\"")
		value = strings.ReplaceAll(value, "\\\\", "\\")
	}
	return key, value, true
}
