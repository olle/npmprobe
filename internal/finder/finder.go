package finder

import (
	"github.com/olle/npmprobe/internal/parser"
	"github.com/olle/npmprobe/internal/spinner"
	"github.com/olle/npmprobe/internal/store"
)

// LoadPackageStore loads all package.json and package-lock.json files using the default store.
// It automatically selects the appropriate file finder based on the current platform.
// Displays a spinner animation while loading.
// This function panics on any underlying file system errors; unreadable files are skipped.
func LoadPackageStore() store.PackageStore {
	spin := spinner.NewSpinner()
	spin.Start("Scanning filesystem for package files...")

	finder := store.NewAutoFileFinder()
	packageStore := store.NewDefaultStoreWithFinder(finder)

	spin.Stop()
	return packageStore
}

// FindPackageInStore searches for a package (represented by a parser.Matcher)
// in the package store and returns paths where the package appears.
func FindPackageInStore(s store.PackageStore, matcher parser.Matcher) []string {

	// Quick check: if the store does not contain the package name at all, return empty result
	if s.DoesNotContainPackage(matcher.Name()) {
		return []string{}
	}

	// Look for a full match in the store
	return s.Find(matcher)
}
