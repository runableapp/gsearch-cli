package version

import (
	_ "embed"
	"strings"
)

//go:embed VERSION.txt
var versionFile string

// Get returns the version string from the embedded VERSION.txt file
func Get() string {
	version := strings.TrimSpace(versionFile)
	if version == "" {
		return "unknown"
	}
	return version
}
