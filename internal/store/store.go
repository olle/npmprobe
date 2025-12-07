package store

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/olle/npmprobe/internal/mdfind"
	"github.com/olle/npmprobe/internal/parser"
)

// PackageStore is an abstraction for storing and searching package files.
// It provides an interface for finding packages by name and version(s).
type PackageStore interface {

	// DoesNotContainPackage checks if the store does not contain any entries for the given package name.
	DoesNotContainPackage(packageName string) bool

	// Find searches for a package matcher in the store and returns file paths where matches are found.
	Find(matcher parser.Matcher) []string

	// Size returns the number of files in the store.
	Size() int
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
// This function panics on any underlying mdfind error; unreadable files are skipped.
func NewDefaultStore() PackageStore {
	ds := &DefaultStore{
		files: make(map[string]string),
	}

	// Query for package.json and package-lock.json files
	query := `kMDItemFSName == "package.json" || kMDItemFSName == "package-lock.json"`
	files := mdfind.FindFiles(query)

	// Read each file into the store
	for _, path := range files {
		if _, err := os.Stat(path); err != nil {
			// Skip files that don't exist or can't be stat'd
			continue
		}

		content, err := ioutil.ReadFile(path)
		if err != nil {
			// Skip unreadable files
			continue
		}

		ds.files[path] = string(content)
	}

	// Build the package index for quick existence checks
	ds.packageIndex = make(map[string]struct{})
	for _, content := range ds.files {
		lines := strings.Split(content, "\n")
		for _, line := range lines {
			parts := strings.Fields(line)
			for _, part := range parts {

				// Trim and sanitize the part, skipping if empty
				part = strings.TrimSpace(part)
				part = strings.Trim(part, `",:{}$%()`)
				if len(part) == 0 {
					continue
				}

				// Skip common non-package name parts
				prefixes := []string{"sha512-", "https://", "http://", "git+", "file://"}
				skip := false
				for _, prefix := range prefixes {
					if strings.HasPrefix(part, prefix) {
						skip = true
						break
					}
				}

				// Skip known keywords
				for _, skipWord := range []string{"version", "dependencies", "devDependencies", "optionalDependencies", "peerDependencies"} {
					if part == skipWord {
						skip = true
						break
					}
				}

				// This is not a package name, skip it!
				if skip {
					continue
				}

				// Add to the package index
				ds.packageIndex[part] = struct{}{}
			}
		}
	}

	return ds
}

// Find searches for a package (from the matcher) in all stored files.
// Returns a list of file paths where any version of the package is found.
func (ds *DefaultStore) Find(matcher parser.Matcher) []string {
	var matches []string
	packageSearchStr := fmt.Sprintf(`"%s":`, matcher.Name())
	versions := []string{matcher.Version()}

	for path, content := range ds.files {

		found := false

		// Check if package name appears in the file at all
		if !strings.Contains(content, packageSearchStr) {
			continue
		}

		// Check each line for package name and the version
		for _, line := range strings.Split(content, "\n") {

			if !strings.Contains(line, packageSearchStr) {
				continue
			}

			// Check if any of the versions appear in this line
			for _, version := range versions {
				versionStr := fmt.Sprintf(`%s`, version)
				if strings.Contains(line, versionStr) {
					found = true
					break
				}
			}

			if found {
				break
			}
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
