//go:build linux

package linux

import (
	"os"
	"os/exec"
	"strings"
)

// SystemLocaleName returns the system locale name on Linux.
// It tries multiple methods in order:
// 1. Check LANG environment variable
// 2. Use locale command
// 3. Return empty string as fallback
func SystemLocaleName() string {
	// Try LANG environment variable first
	if lang := os.Getenv("LANG"); lang != "" {
		return lang
	}

	// Try using locale command
	cmd := exec.Command("locale")
	output, err := cmd.Output()
	if err == nil {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "LANG=") {
				lang := strings.TrimPrefix(line, "LANG=")
				lang = strings.Trim(lang, "\"")
				return lang
			}
		}
	}

	return ""
}
