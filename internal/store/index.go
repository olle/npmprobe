package store

import "sort"

// Index is an in-memory index of installed package versions and the files
// they were found in. It is built once while loading package files and is
// safe for concurrent reads after construction.
//
// The index is deliberately independent of parser.Matcher and of file
// discovery; those remain separate concerns in their own packages.
type Index struct {
	// byName maps a package name to its versions, and each version to the
	// set of files that contain it.
	byName map[string]map[string]map[string]struct{}

	// files tracks the files that have been indexed, so duplicate finder
	// results are not counted or processed twice.
	files map[string]struct{}
}

// NewIndex creates an empty package index.
func NewIndex() *Index {
	return &Index{
		byName: make(map[string]map[string]map[string]struct{}),
		files:  make(map[string]struct{}),
	}
}

// AddFile indexes the contents of a package file. The file type determines
// which indexing strategy is used. Unknown file types are counted but add no
// package entries.
func (i *Index) AddFile(path string, content []byte) {
	if _, seen := i.files[path]; seen {
		return
	}
	i.files[path] = struct{}{}

	if indexer := indexerFor(path); indexer != nil {
		indexer.Index(i, path, content)
	}
}

// Add records that name@version was found in path.
func (i *Index) Add(name, version, path string) {
	if name == "" || version == "" || path == "" {
		return
	}

	versions, ok := i.byName[name]
	if !ok {
		versions = make(map[string]map[string]struct{})
		i.byName[name] = versions
	}

	paths, ok := versions[version]
	if !ok {
		paths = make(map[string]struct{})
		versions[version] = paths
	}

	paths[path] = struct{}{}
}

// HasPackage reports whether any version of the package is indexed.
func (i *Index) HasPackage(name string) bool {
	_, ok := i.byName[name]
	return ok
}

// FileCount returns the number of files indexed.
func (i *Index) FileCount() int {
	return len(i.files)
}

// Find returns the indexed locations of the given package versions. Versions
// are matched exactly, in the order supplied.
func (i *Index) Find(name string, versions []string) []PackageMatch {
	versionsByName, ok := i.byName[name]
	if !ok {
		return nil
	}

	var matches []PackageMatch
	for _, version := range versions {
		for _, path := range sortedSet(versionsByName[version]) {
			matches = append(matches, PackageMatch{
				Path:    path,
				Version: version,
			})
		}
	}
	return matches
}

// FindAll returns every indexed version of the named package and the files
// containing them. Versions and paths are sorted for deterministic output.
func (i *Index) FindAll(name string) []PackageMatch {
	versionsByName, ok := i.byName[name]
	if !ok {
		return nil
	}

	versions := make([]string, 0, len(versionsByName))
	for version := range versionsByName {
		versions = append(versions, version)
	}
	sort.Strings(versions)

	var matches []PackageMatch
	for _, version := range versions {
		for _, path := range sortedSet(versionsByName[version]) {
			matches = append(matches, PackageMatch{
				Path:    path,
				Version: version,
			})
		}
	}
	return matches
}

// Paths returns the distinct files containing any of the given package
// versions, sorted by path.
func (i *Index) Paths(name string, versions []string) []string {
	matches := i.Find(name, versions)
	if len(matches) == 0 {
		return nil
	}

	seen := make(map[string]struct{}, len(matches))
	paths := make([]string, 0, len(matches))
	for _, match := range matches {
		if _, ok := seen[match.Path]; ok {
			continue
		}
		seen[match.Path] = struct{}{}
		paths = append(paths, match.Path)
	}
	sort.Strings(paths)
	return paths
}

func sortedSet(set map[string]struct{}) []string {
	if len(set) == 0 {
		return nil
	}

	values := make([]string, 0, len(set))
	for value := range set {
		values = append(values, value)
	}
	sort.Strings(values)
	return values
}
