package finder

import (
	"github.com/olle/npmprobe/internal/parser"
	"github.com/olle/npmprobe/internal/spinner"
	"github.com/olle/npmprobe/internal/store"
)

// LoadPackageStore loads all supported package and lockfiles using the default store.
// It automatically selects the appropriate file finder based on the current platform.
// Displays a spinner animation while loading.
// It returns an error when discovery fails or no package files can be read;
// unreadable individual files are skipped.
func LoadPackageStore() (store.PackageStore, error) {
	spin := spinner.NewSpinner()
	spin.Start("Scanning filesystem for package files...")
	defer spin.Stop()

	return store.NewDefaultStoreWithFinder(store.NewAutoFileFinder())
}

// FindPackageInStore searches for a package (represented by a parser.Matcher)
// in the package store and returns paths where the package appears.
func FindPackageInStore(s store.PackageStore, matcher parser.Matcher) []store.PackageMatch {

	// Quick check: if the store does not contain the package name at all, return empty result
	if s.DoesNotContainPackage(matcher.Name()) {
		return []store.PackageMatch{}
	}

	// Look for a full match in the store
	return s.Find(matcher)
}
