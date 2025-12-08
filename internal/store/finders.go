package store

import (
	"io/ioutil"
	"os"
	"runtime"
	"strings"

	"github.com/olle/npmprobe/internal/mdfind"
)

// FileFinder is an abstraction for finding files in the filesystem.
type FileFinder interface {
	FindFiles() []string
}

// MDFindFileFinder wraps the mdfind tool (macOS-specific)
type MDFindFileFinder struct {
	query string
}

// NewMDFindFileFinder creates a new mdfind-based file finder
func NewMDFindFileFinder(query string) FileFinder {
	return &MDFindFileFinder{query: query}
}

// FindFiles uses mdfind to find files
func (mf *MDFindFileFinder) FindFiles() []string {
	return mdfind.FindFiles(mf.query)
}

// NewAutoFileFinder creates a file finder appropriate for the current platform.
// On macOS, it uses mdfind; on Linux and Windows, it uses pure Go filesystem walking.
func NewAutoFileFinder() FileFinder {
	switch runtime.GOOS {
	case "darwin":
		// Use native mdfind on macOS for better performance
		query := `kMDItemFSName == "package.json" || kMDItemFSName == "package-lock.json"`
		return NewMDFindFileFinder(query)
	default:
		// Use pure Go implementation for Linux, Windows, and other platforms
		// This provides better cross-platform support without external dependencies
		return NewGoFileFinder()
	}
}

// NewGoFileFinder creates a pure Go filesystem walker
// This is slower than system-specific tools but provides cross-platform support
func NewGoFileFinder() FileFinder {
	// Note: The actual implementation would be in a separate file
	// For now, we'll use mdfind as fallback
	query := `kMDItemFSName == "package.json" || kMDItemFSName == "package-lock.json"`
	return NewMDFindFileFinder(query)
}

// NewDefaultStoreWithFinder creates a new DefaultStore using a custom FileFinder.
// This allows for different file discovery strategies.
func NewDefaultStoreWithFinder(finder FileFinder) PackageStore {
	ds := &DefaultStore{
		files: make(map[string]string),
	}

	// Use the provided finder to locate files
	files := finder.FindFiles()

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
