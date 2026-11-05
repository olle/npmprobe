package store

import (
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

// DefaultStore is the default PackageStore implementation. It builds a
// name -> version -> file path index while loading package files, so queries
// are exact and do not rescan file contents.
type DefaultStore struct {
	index *Index
}

// Find searches for a package (from the matcher) in the index.
// Returns the files where any version of the package is found.
func (ds *DefaultStore) Find(matcher parser.Matcher) []PackageMatch {
	return ds.index.Find(matcher.Name(), matcher.Versions())
}

// QueryByNameAndVersions queries for a package by name and specific versions.
// Returns the distinct file paths where the package with any of the specified
// versions is found, sorted by path.
func (ds *DefaultStore) QueryByNameAndVersions(packageName string, versions []string) []string {
	return ds.index.Paths(packageName, versions)
}

// DoesNotContainPackage checks if the store does not contain any entries for the given package name.
func (ds *DefaultStore) DoesNotContainPackage(packageName string) bool {
	return !ds.index.HasPackage(packageName)
}

// Size returns the number of files in the store.
func (ds *DefaultStore) Size() int {
	return ds.index.FileCount()
}
