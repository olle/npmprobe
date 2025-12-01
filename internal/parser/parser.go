package parser

import (
	"fmt"
	"strings"
)

// Matcher is an abstraction for matching packages by name and version(s).
type Matcher interface {
	Name() string
	Version() string    // Returns the first version (for backward compatibility)
	Versions() []string // Returns all versions
}

// SimpleMatcher is a basic implementation of Matcher for a single version
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

// Versions returns all versions as a single-element slice
func (sm *SimpleMatcher) Versions() []string {
	return []string{sm.version}
}

// MultiVersionMatcher is an implementation of Matcher for multiple comma-separated versions
type MultiVersionMatcher struct {
	name     string
	versions []string
}

// Name returns the package name
func (mvm *MultiVersionMatcher) Name() string {
	return mvm.name
}

// Version returns the first version (for backward compatibility)
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

// ParseLine parses a package line in the format "package version" or "package version1,version2,...".
// Returns a Matcher (SimpleMatcher for single version, MultiVersionMatcher for multiple versions),
// or an error if the line cannot be parsed.
func ParseLine(line string) (Matcher, error) {
	line = strings.TrimSpace(line)

	// Skip empty lines
	if line == "" {
		return nil, fmt.Errorf("empty line")
	}

	// Split by whitespace to extract package name and versions
	parts := strings.Fields(line)
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid package line format: expected 'package version[,version2,...]', got '%s'", line)
	}

	name := parts[0]

	// Reconstruct versions part by joining all parts after the name and handling commas
	// This handles cases like "react 16.0.0 , 17.0.0 , 18.0.0"
	versionStr := strings.Join(parts[1:], " ")

	// Check if versions contain commas (multiple versions)
	if strings.Contains(versionStr, ",") {
		versions := strings.Split(versionStr, ",")
		// Trim whitespace from each version
		for i := range versions {
			versions[i] = strings.TrimSpace(versions[i])
		}
		return &MultiVersionMatcher{
			name:     name,
			versions: versions,
		}, nil
	}

	// Single version case
	return &SimpleMatcher{
		name:    name,
		version: versionStr,
	}, nil
}
