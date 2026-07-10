// Package env resolves credentials from the user configuration file or the
// legacy environment-variable chain.
package env

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const configDirName = ".gpt-image2-cli"
const configFileName = "config.json"

// Config is the user-scoped CLI configuration stored in
// ~/.gpt-image2-cli/config.json.
type Config struct {
	APIKey  string `json:"api_key"`
	BaseURL string `json:"base_url"`
}

// Credentials contains the API connection settings used by the CLI.
type Credentials struct {
	APIKey  string
	BaseURL string
}

// Resolve returns credentials from ~/.gpt-image2-cli/config.json when that
// file exists. The configuration file is authoritative: it must contain an
// api_key, and environment variables are not consulted. When the file does
// not exist, Resolve preserves the legacy env, .env, and ~/.env behavior.
func Resolve() (Credentials, error) {
	config, found, err := LoadConfig()
	if err != nil {
		return Credentials{}, err
	}
	if found {
		if strings.TrimSpace(config.APIKey) == "" {
			return Credentials{}, fmt.Errorf("configuration file %s must contain api_key", ConfigPath())
		}
		return Credentials{APIKey: config.APIKey, BaseURL: config.BaseURL}, nil
	}

	_ = LoadChain()
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("API_KEY")
	}
	return Credentials{APIKey: apiKey, BaseURL: os.Getenv("BASE_URL")}, nil
}

// ConfigPath returns the conventional path for the user-scoped configuration
// file. An empty string means the home directory could not be determined.
func ConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return ""
	}
	return filepath.Join(home, configDirName, configFileName)
}

// LoadConfig reads the user-scoped configuration file. found is false only
// when the file does not exist (or the home directory is unavailable).
func LoadConfig() (config Config, found bool, err error) {
	path := ConfigPath()
	if path == "" {
		return Config{}, false, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Config{}, false, nil
		}
		return Config{}, true, fmt.Errorf("read configuration file %s: %w", path, err)
	}

	config, err = parseConfig(data)
	if err != nil {
		return Config{}, true, fmt.Errorf("parse configuration file %s: %w", path, err)
	}
	return config, true, nil
}

func parseConfig(data []byte) (Config, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()

	var config Config
	if err := decoder.Decode(&config); err != nil {
		return Config{}, err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return Config{}, fmt.Errorf("expected a single JSON object")
	}
	return config, nil
}

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
