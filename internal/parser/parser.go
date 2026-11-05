package parser

import (
	"fmt"
	"strings"
)

// Matcher is an abstraction for matching packages by name and version(s).
// It supports matching against single versions, multiple versions, or every
// version of a package.
type Matcher interface {
	Name() string
	Version() string
	Versions() []string

	// AllVersions reports whether every version of the package matches.
	AllVersions() bool
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

// AllVersions returns false: this matcher matches one specific version.
func (sm *SimpleMatcher) AllVersions() bool {
	return false
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

// AllVersions returns false: this matcher matches specific versions.
func (mvm *MultiVersionMatcher) AllVersions() bool {
	return false
}

// AllVersionsMatcher matches every version of a package.
type AllVersionsMatcher struct {
	name string
}

// Name returns the package name
func (m *AllVersionsMatcher) Name() string {
	return m.name
}

// Version returns an empty string, since no specific version is selected.
func (m *AllVersionsMatcher) Version() string {
	return ""
}

// Versions returns nil, since every version matches.
func (m *AllVersionsMatcher) Versions() []string {
	return nil
}

// AllVersions returns true: this matcher matches any version.
func (m *AllVersionsMatcher) AllVersions() bool {
	return true
}

// ParseLine parses a package line in multiple formats:
//   - "package" (all versions)
//   - "package version" (simple format)
//   - "package version1,version2,version3" (comma-separated versions)
//   - "package version1, version2, version3" (comma-separated with spaces)
//
// Fields may be separated by any whitespace (spaces, tabs, etc.).
// Extra fields beyond the version spec are ignored.
// Returns a Matcher implementation and an error if the line cannot be parsed.
func ParseLine(line string) (Matcher, error) {

	// Fields splits on any whitespace and discards surrounding whitespace.
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return nil, fmt.Errorf("empty line")
	}

	name := parts[0]
	if len(parts) == 1 {
		// A bare package name means every version is compromised.
		return &AllVersionsMatcher{name: name}, nil
	}

	// The version spec starts at parts[1]. If it contains commas, collect
	// subsequent fields that belong to the same "v1, v2, v3" list.
	versionSpec := parts[1]
	if strings.Contains(versionSpec, ",") {
		versionParts := []string{versionSpec}
		for i := 2; i < len(parts); i++ {
			last := versionParts[len(versionParts)-1]
			if strings.HasSuffix(last, ",") || strings.Contains(parts[i], ",") {
				versionParts = append(versionParts, parts[i])
				continue
			}
			break
		}
		versionSpec = strings.Join(versionParts, " ")
	}

	// Check if versions are comma-separated.
	if strings.Contains(versionSpec, ",") {
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
			return &SimpleMatcher{
				name:    name,
				version: trimmedVersions[0],
			}, nil
		}

		return &MultiVersionMatcher{
			name:     name,
			versions: trimmedVersions,
		}, nil
	}

	return &SimpleMatcher{
		name:    name,
		version: versionSpec,
	}, nil
}

// LineError reports a line that could not be parsed.
type LineError struct {
	Line int
	Text string
	Err  error
}

// Error formats the line error with its one-based line number.
func (e *LineError) Error() string {
	return fmt.Sprintf("line %d: %v: %q", e.Line, e.Err, e.Text)
}

// Unwrap exposes the underlying parse error.
func (e *LineError) Unwrap() error {
	return e.Err
}

// ParseLines parses a list of package lines, skipping blank lines. It returns
// every successfully parsed matcher and every line error, rather than stopping
// at the first invalid line. Line numbers in errors are one-based.
func ParseLines(lines []string) ([]Matcher, []error) {
	matchers := make([]Matcher, 0, len(lines))
	var failures []error

	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		matcher, err := ParseLine(line)
		if err != nil {
			failures = append(failures, &LineError{
				Line: i + 1,
				Text: strings.TrimSpace(line),
				Err:  err,
			})
			continue
		}

		matchers = append(matchers, matcher)
	}

	return matchers, failures
}
