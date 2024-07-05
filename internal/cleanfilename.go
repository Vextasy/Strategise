package internal

import (
	"path/filepath"
	"strings"
)

// cleanFilename removes troublesome characters from a potential filename.
func CleanFilename(filename string) string {
	filename = strings.ReplaceAll(filename, " ", "_")
	filename = strings.ReplaceAll(filename, "&", "And")
	filename = strings.ReplaceAll(filename, "/", "-")
	return filepath.Clean(filename)
}
