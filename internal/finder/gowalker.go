package finder

import (
	"fmt"
	"os"
	"path/filepath"
)

// GoWalkerFinder is a pure Go implementation of the PackageFileFinder.
// It walks the filesystem to find package.json and package-lock.json files
// without relying on external tools like mdfind or locate.
// This implementation is slower than system-specific tools but provides
// better cross-platform support, including Windows.
type GoWalkerFinder struct {
	filenames map[string]struct{}
}

// NewGoWalkerFinder creates a new GoWalkerFinder that looks for the given filenames.
func NewGoWalkerFinder(filenames ...string) *GoWalkerFinder {
	f := &GoWalkerFinder{
		filenames: make(map[string]struct{}),
	}
	for _, fn := range filenames {
		f.filenames[fn] = struct{}{}
	}
	return f
}

// FindFiles walks the filesystem starting from the root directory (/)
// and returns all files matching the configured filenames.
// It returns errors on stderr but continues searching.
func (gwf *GoWalkerFinder) FindFiles() []string {
	var results []string

	// Determine root path based on OS
	root := "/"

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// Skip inaccessible directories and continue
			return nil
		}

		// Check if file matches one of our target filenames
		if !info.IsDir() {
			if _, exists := gwf.filenames[info.Name()]; exists {
				results = append(results, path)
			}
		}

		return nil
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: filesystem walk encountered error: %v\n", err)
	}

	return results
}
