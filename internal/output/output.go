// Package output handles output path resolution and writing image bytes to disk.
package output

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var nonWord = regexp.MustCompile(`[^\w\s-]+`)
var collapse = regexp.MustCompile(`[-\s]+`)

// Slugify turns a prompt into a safe filename stem.
func Slugify(text string, maxLen int) string {
	s := nonWord.ReplaceAllString(strings.ToLower(text), "")
	s = strings.TrimSpace(s)
	s = collapse.ReplaceAllString(s, "-")
	if len(s) > maxLen {
		s = s[:maxLen]
	}
	s = strings.Trim(s, "-")
	if s == "" {
		return "image"
	}
	return s
}

// DefaultPath returns an auto-generated output path.
func DefaultPath(prompt string, ext string) string {
	cwd, _ := os.Getwd()
	target := cwd
	if fi, err := os.Stat(filepath.Join(cwd, "fig")); err == nil && fi.IsDir() {
		target = filepath.Join(cwd, "fig")
	}
	stamp := time.Now().Format("2006-01-02-15-04-05")
	return filepath.Join(target, fmt.Sprintf("%s-%s.%s", stamp, Slugify(prompt, 30), ext))
}

// Write saves image bytes to disk, creating parent directories as needed.
func Write(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
