package parser

import (
	"fmt"
	"strings"
)

// Matcher is an abstraction for matching packages by name and version(s).
// It supports matching against single or multiple versions.
type Matcher interface {
	Name() string
	Version() string
	Versions() []string
}

// SimpleMatcher is a basic implementation of Matcher for single version
type SimpleMatcher struct {
	name    string
	version string
}

// Name returns the package name
func (sm *SimpleMatcher) Name() string {
	return sm.name
}

// Version returns the package version
func (sm *SimpleMatcher) Version() string {
	return sm.version
}

// Versions returns a slice containing the single version
func (sm *SimpleMatcher) Versions() []string {
	return []string{sm.version}
}

// MultiVersionMatcher is an implementation of Matcher for multiple versions
type MultiVersionMatcher struct {
	name     string
	versions []string
}

// Name returns the package name
func (mvm *MultiVersionMatcher) Name() string {
	return mvm.name
}

// Version returns the first version (for compatibility)
func (mvm *MultiVersionMatcher) Version() string {
	if len(mvm.versions) > 0 {
		return mvm.versions[0]
	}
	return ""
}

// Versions returns all versions
func (mvm *MultiVersionMatcher) Versions() []string {
	return mvm.versions
}

// ParseLine parses a package line in multiple formats:
//   - "package version" (simple format)
//   - "package version1,version2,version3" (comma-separated versions)
//   - "package version1, version2, version3" (comma-separated with spaces)
//
// Extra fields beyond the version spec are ignored.
// Returns a Matcher implementation and an error if the line cannot be parsed.
func ParseLine(line string) (Matcher, error) {

	line = strings.TrimSpace(line)
	line = strings.ReplaceAll(line, "\t", " ")
	line = strings.ReplaceAll(line, " ", " ")

	// Skip empty lines
	if line == "" {
		return nil, fmt.Errorf("empty line")
	}

	// Split by whitespace to extract package name and version(s)
	parts := strings.Fields(line)
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid package line format: expected 'package version[,...]', got '%s'", line)
	}

	name := parts[0]

	// The version spec starts at parts[1]
	// If it contains commas, we might need to collect more parts (for "v1, v2, v3" format)
	// Otherwise, just use the first part
	versionSpec := parts[1]

	// If the first version part contains a comma, check if we need to collect more parts
	if strings.Contains(versionSpec, ",") {
		versionParts := []string{versionSpec}

		// Collect additional parts until we have a complete version list
		// Continue collecting if the part contains a comma or the previous part ended with a comma
		for i := 2; i < len(parts); i++ {
			// If previous part ended with comma, this is part of the version list
			if strings.HasSuffix(versionParts[len(versionParts)-1], ",") {
				versionParts = append(versionParts, parts[i])
			} else if strings.Contains(parts[i], ",") {
				// This part contains a comma, so it's part of the version list
				versionParts = append(versionParts, parts[i])
			} else {
				// No comma in this or previous part, stop collecting
				break
			}
		}
		versionSpec = strings.Join(versionParts, " ")
	}

	// Check if versions are comma-separated
	if strings.Contains(versionSpec, ",") {
		// Parse comma-separated versions
		versions := strings.Split(versionSpec, ",")
		trimmedVersions := make([]string, 0, len(versions))
		for _, v := range versions {
			v = strings.TrimSpace(v)
			if v != "" {
				trimmedVersions = append(trimmedVersions, v)
			}
		}

		if len(trimmedVersions) == 0 {
			return nil, fmt.Errorf("no valid versions found in '%s'", versionSpec)
		}

		if len(trimmedVersions) == 1 {
			// Single version, use SimpleMatcher
			return &SimpleMatcher{
				name:    name,
				version: trimmedVersions[0],
			}, nil
		}

		// Multiple versions, use MultiVersionMatcher
		return &MultiVersionMatcher{
			name:     name,
			versions: trimmedVersions,
		}, nil
	}

	// Single version format
	return &SimpleMatcher{
		name:    name,
		version: versionSpec,
	}, nil
}
