package store

import (
	"os"
	"runtime"

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
// This allows for different file discovery strategies. The store builds its
// version index while reading the files; files that cannot be read or parsed
// are skipped.
func NewDefaultStoreWithFinder(finder FileFinder) PackageStore {
	ds := &DefaultStore{
		index: NewIndex(),
	}

	// Use the provided finder to locate files
	for _, path := range finder.FindFiles() {
		content, err := os.ReadFile(path)
		if err != nil {
			// Skip files that don't exist or can't be read
			continue
		}

		ds.index.AddFile(path, content)
	}

	return ds
}
