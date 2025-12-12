package gomod

import (
	"fmt"
	"strings"
)

type ModuleSpec struct {
	Path    string
	Version string
}

type Module struct {
	Path     string
	Version  string
	Dir      string
	InfoFile string // Path to .info file in cache
	ModFile  string // Path to .mod file in cache
	ZipFile  string // Path to .zip file in cache
}

func ParseModulesList(content string) ([]ModuleSpec, error) {
	var specs []ModuleSpec
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse line as "path@version" or "path"
		parts := strings.Split(line, "@")
		if len(parts) > 2 {
			return nil, fmt.Errorf("invalid module spec: %s", line)
		}

		path := strings.TrimSpace(parts[0])
		version := ""
		if len(parts) == 2 {
			version = strings.TrimSpace(parts[1])
		}

		specs = append(specs, ModuleSpec{
			Path:    path,
			Version: version,
		})
	}

	return specs, nil
}

func IsValidSemver(version string) bool {
	// Simple check: must start with 'v' and have at least v + digit
	if len(version) < 2 {
		return false
	}
	return version[0] == 'v' && version[1] >= '0' && version[1] <= '9'
}
