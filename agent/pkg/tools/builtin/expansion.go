package builtin

import (
	"os"
	"path/filepath"
	"strings"
)

// ExpandArguments expands tilde and glob patterns in command arguments
func ExpandArguments(args []string) []string {
	expanded := make([]string, 0, len(args))

	for _, arg := range args {
		// Expand tilde first
		arg = expandTilde(arg)

		// Then expand globs
		matches, err := filepath.Glob(arg)
		if err == nil && len(matches) > 0 {
			expanded = append(expanded, matches...)
		} else {
			expanded = append(expanded, arg)
		}
	}

	return expanded
}

// expandTilde expands the tilde in a path (internal helper)
func expandTilde(path string) string {
	if path == "~" {
		home, err := os.UserHomeDir()
		if err == nil {
			return home
		}
	} else if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			return strings.Replace(path, "~", home, 1)
		}
	}
	return path
}
