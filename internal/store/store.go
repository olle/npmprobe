package store

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/olle/npmprobe/internal/parser"
)

type PackageMatch struct {
	Path    string
	Version string
}

// PackageStore is an abstraction for storing and searching package files.
// It provides an interface for finding packages by name and version(s).
// The interface is designed to be extensible, allowing additional query methods
// to be added by embedding or implementing additional interfaces.
type PackageStore interface {

	// DoesNotContainPackage checks if the store does not contain any entries for the given package name.
	DoesNotContainPackage(packageName string) bool

	// Find searches for a package matcher in the store and returns file paths where matches are found.
	Find(matcher parser.Matcher) []PackageMatch

	// Size returns the number of files in the store.
	Size() int

	// QueryByNameAndVersions queries for a package by name and specific versions.
	// This method is designed to allow extensible query patterns.
	QueryByNameAndVersions(packageName string, versions []string) []string
}

// DefaultStore is a basic implementation of PackageStore that stores file contents in memory.
type DefaultStore struct {

	// File contents mapped by file path
	files map[string]string // filepath -> content

	// Index of all existing words (package names) for quick existence checks
	// This is built during initialization for fast DoesNotContainPackage checks
	packageIndex map[string]struct{}
}

// NewDefaultStore creates a new DefaultStore by loading all package.json and package-lock.json files.
// This is a convenience function that uses the auto-detecting FileFinder.
// This function panics on any underlying file system errors; unreadable files are skipped.
// Deprecated: Use NewDefaultStoreWithFinder with NewAutoFileFinder() instead for better flexibility.
func NewDefaultStore() PackageStore {
	return NewDefaultStoreWithFinder(NewAutoFileFinder())
}

// Find searches for a package (from the matcher) in all stored files.
// Returns a list of file paths where any version of the package is found.
func (ds *DefaultStore) Find(matcher parser.Matcher) []PackageMatch {

	versions := matcher.Versions()
	packageName := matcher.Name()
	var matches []PackageMatch

	for path, content := range ds.files {

		if !strings.Contains(content, packageName) {
			continue
		}

		packageLineSearchStr := fmt.Sprintf(`"%s":`, packageName)
		for _, version := range versions {
			if findLineByLine(content, packageLineSearchStr, []string{version}) {
				matches = append(matches, PackageMatch{
					Path:    path,
					Version: version,
				})
			} else if findByPackageNameAndVersion(content, packageName, []string{version}) {
				matches = append(matches, PackageMatch{
					Path:    path,
					Version: version,
				})	
			}
		}
	}

	return matches
}



// QueryByNameAndVersions queries for a package by name and specific versions.
// Returns a list of file paths where the package with any of the specified versions is found.
func (ds *DefaultStore) QueryByNameAndVersions(packageName string, versions []string) []string {
	var matches []string

	for path, content := range ds.files {

		found := false

		// Check if package name appears in the file at all
		if !strings.Contains(content, packageName) {
			continue
		}

		// Dependency or line search string assumes `"<packageName>":` as format.
		packageLineSearchStr := fmt.Sprintf(`"%s":`, packageName)
		if findLineByLine(content, packageLineSearchStr, versions) {
			found = true
		} else if findByPackageNameAndVersion(content, packageName, versions) {
			found = true
		}

		if found {
			matches = append(matches, path)
		}
	}

	return matches
}

// DoesNotContainPackage checks if the store does not contain any entries for the given package name.
func (ds *DefaultStore) DoesNotContainPackage(packageName string) bool {
	_, exists := ds.packageIndex[packageName]
	return !exists
}

// Size returns the number of files in the store.
func (ds *DefaultStore) Size() int {
	return len(ds.files)
}

// findLineByLine searches for a package name and versions line by line in content.
// Returns true if any line contains the package name and any of the versions.
func findLineByLine(content string, packageSearchStr string, versions []string) bool {
	for _, line := range strings.Split(content, "\n") {
		if !strings.Contains(line, packageSearchStr) {
			continue
		}

		// Check if any of the versions appear in this line
		for _, version := range versions {
			if strings.Contains(line, version) {
				return true
			}
		}
	}
	return false
}

// Parses the actual JSON content and checks for package name and versions.
func findByPackageNameAndVersion(content string, packageName string, versions []string) bool {

	type PackageJson struct {
		Name    string `json:"name,omitempty"`
		Version string `json:"version,omitempty"`
	}

	var packageJson PackageJson
	err := json.Unmarshal([]byte(content), &packageJson)
	if err != nil {
		panic(fmt.Sprintf("Error unmarshaling package.json content: %v", err))
	}

	for _, version := range versions {
		if packageJson.Name == packageName && packageJson.Version == version {
			return true
		}
	}

	return false
}
