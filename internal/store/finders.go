package store

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/olle/npmprobe/internal/mdfind"
)

// packageFileNames lists the package and lockfile formats understood by the
// store. Discovery must look for the same names the indexers support.
var packageFileNames = []string{
	"package.json",
	"package-lock.json",
	"npm-shrinkwrap.json",
	"yarn.lock",
	"pnpm-lock.yaml",
	"bun.lock",
}

// PackageFileNames returns a copy of the supported package file names.
func PackageFileNames() []string {
	names := make([]string, len(packageFileNames))
	copy(names, packageFileNames)
	return names
}

// packageFileQuery builds the Spotlight query used by mdfind for all supported
// package file names.
func packageFileQuery() string {
	parts := make([]string, 0, len(packageFileNames))
	for _, name := range packageFileNames {
		parts = append(parts, fmt.Sprintf(`kMDItemFSName == %q`, name))
	}
	return strings.Join(parts, " || ")
}

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
// On macOS it uses mdfind; on Linux it uses locate; on Windows it walks the
// filesystem.
func NewAutoFileFinder() FileFinder {
	switch runtime.GOOS {
	case "darwin":
		// Use native mdfind on macOS for better performance
		return NewMDFindFileFinder(packageFileQuery())
	default:
		// Use the platform-specific system search on Linux and Windows.
		return NewGoFileFinder()
	}
}

// NewGoFileFinder creates a finder for the supported package files. On Linux
// it uses locate; on Windows it walks the filesystem.
func NewGoFileFinder() FileFinder {
	return NewMDFindFileFinder(packageFileQuery())
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
