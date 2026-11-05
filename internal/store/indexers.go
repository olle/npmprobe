package store

import (
	"encoding/json"
	"path/filepath"
	"strings"
)

// fileIndexer extracts installed package versions from one kind of package
// file. Adding support for another file format (yarn.lock, pnpm-lock.yaml,
// npm-shrinkwrap.json, ...) is a matter of implementing this interface and
// registering it in indexerFor.
type fileIndexer interface {
	Index(index *Index, path string, content []byte)
}

// indexerFor selects an indexing strategy by file name. Returning nil means
// the file is unknown to the indexer.
func indexerFor(path string) fileIndexer {
	switch filepath.Base(path) {
	case "package.json":
		return packageJSONIndexer{}
	case "package-lock.json":
		return packageLockIndexer{}
	default:
		return nil
	}
}

// packageJSONIndexer indexes a package.json as the package it describes.
// Dependency ranges in the manifest are deliberately ignored: they are not
// installed versions and must not be treated as such.
type packageJSONIndexer struct{}

func (packageJSONIndexer) Index(index *Index, path string, content []byte) {
	var manifest struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	}
	if err := json.Unmarshal(content, &manifest); err != nil {
		return
	}

	index.Add(manifest.Name, manifest.Version, path)
}

type lockPackage struct {
	Version string `json:"version"`
}

type lockDependency struct {
	Version      string                    `json:"version"`
	Dependencies map[string]lockDependency `json:"dependencies"`
}

// packageLockIndexer indexes installed versions from npm lockfiles. It
// supports the v2/v3 "packages" map and the v1 recursive "dependencies" tree.
type packageLockIndexer struct{}

func (packageLockIndexer) Index(index *Index, path string, content []byte) {
	var lockfile struct {
		Packages     map[string]lockPackage    `json:"packages"`
		Dependencies map[string]lockDependency `json:"dependencies"`
	}
	if err := json.Unmarshal(content, &lockfile); err != nil {
		return
	}

	// Lockfile v2 and v3 carry installed packages in "packages". The legacy
	// "dependencies" tree is present in v2 as well, but "packages" is the
	// authoritative source there.
	if len(lockfile.Packages) > 0 {
		for key, pkg := range lockfile.Packages {
			if name := installedNameFromLockKey(key); name != "" {
				index.Add(name, pkg.Version, path)
			}
		}
		return
	}

	addLockDependencies(index, path, lockfile.Dependencies)
}

// installedNameFromLockKey extracts the package name from a lockfile v2/v3
// package key. Keys look like "node_modules/lodash",
// "node_modules/@scope/pkg", or "packages/a/node_modules/b"; the last
// node_modules segment identifies the installed package. The root entry ("")
// and workspace source entries are not installed packages and are skipped.
func installedNameFromLockKey(key string) string {
	const nodeModules = "node_modules/"

	at := strings.LastIndex(key, nodeModules)
	if at == -1 {
		return ""
	}
	return key[at+len(nodeModules):]
}

// addLockDependencies walks the recursive dependency tree used by lockfile
// v1 (and the legacy section of v2).
func addLockDependencies(index *Index, path string, dependencies map[string]lockDependency) {
	for name, dependency := range dependencies {
		index.Add(name, dependency.Version, path)
		addLockDependencies(index, path, dependency.Dependencies)
	}
}
