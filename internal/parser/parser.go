package parser

import (
	"fmt"
	"strings"
)

// Matcher is an abstraction for matching packages by name and version.
type Matcher interface {
	Name() string
	Version() string
}

// SimpleMatcher is a basic implementation of Matcher
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

// ParseLine parses a package line in the format "package version" and returns a Matcher.
// Returns an error if the line cannot be parsed.
func ParseLine(line string) (Matcher, error) {
	line = strings.TrimSpace(line)

	// Skip empty lines
	if line == "" {
		return nil, fmt.Errorf("empty line")
	}

	// Split by whitespace to extract package name and version
	parts := strings.Fields(line)
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid package line format: expected 'package version', got '%s'", line)
	}

	// Take only the first two parts (package name and version)
	// Additional fields are ignored
	name := parts[0]
	version := parts[1]

	return &SimpleMatcher{
		name:    name,
		version: version,
	}, nil
}
