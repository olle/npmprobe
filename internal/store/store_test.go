package store

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"github.com/olle/npmprobe/internal/parser"
)

type staticFinder struct {
	files []string
}

func (f staticFinder) FindFiles() []string {
	return f.files
}

// newTestStore writes the supplied files (keyed by slash-separated relative
// path) to a temporary directory and loads them into a DefaultStore.
func newTestStore(t *testing.T, files map[string]string) (PackageStore, string) {
	t.Helper()

	root := t.TempDir()
	paths := make([]string, 0, len(files))
	for name, content := range files {
		path := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("create directory for %s: %v", name, err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
		paths = append(paths, path)
	}

	return NewDefaultStoreWithFinder(staticFinder{files: paths}), root
}

// findings returns the matches for a package line as "relative/path@version",
// sorted for comparison.
func findings(t *testing.T, store PackageStore, root, line string) []string {
	t.Helper()

	matcher, err := parser.ParseLine(line)
	if err != nil {
		t.Fatalf("parse %q: %v", line, err)
	}

	matches := store.Find(matcher)
	got := make([]string, 0, len(matches))
	for _, match := range matches {
		relative, err := filepath.Rel(root, match.Path)
		if err != nil {
			t.Fatalf("relative path for %s: %v", match.Path, err)
		}
		got = append(got, filepath.ToSlash(relative)+"@"+match.Version)
	}
	sort.Strings(got)
	return got
}

func assertFindings(t *testing.T, store PackageStore, root, line string, want ...string) {
	t.Helper()

	got := findings(t, store, root, line)
	if want == nil {
		want = []string{}
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Find(%q) = %v, want %v", line, got, want)
	}
}

func TestPackageJSONIndexesOnlyItsOwnVersion(t *testing.T) {
	store, root := newTestStore(t, map[string]string{
		"app/node_modules/yaml-eslint-parser/package.json": `{
			"name": "yaml-eslint-parser",
			"version": "1.2.3",
			"dependencies": {"lodash": "^4.17.20"}
		}`,
		"app/node_modules/lodash/package.json": `{
			"name": "lodash",
			"version": "4.17.21"
		}`,
	})

	// A dependency range is not an installed version.
	assertFindings(t, store, root, "lodash 4.17.20")

	// Exact versions still match the package's own manifest.
	assertFindings(t, store, root, "lodash 4.17.21",
		"app/node_modules/lodash/package.json@4.17.21")
	assertFindings(t, store, root, "yaml-eslint-parser 1.2.3",
		"app/node_modules/yaml-eslint-parser/package.json@1.2.3")

	// A version prefix must not match a longer version.
	assertFindings(t, store, root, "lodash 4.17.2")
}

func TestLockfilePackagesV3(t *testing.T) {
	store, root := newTestStore(t, map[string]string{
		"app/package-lock.json": `{
			"name": "app",
			"version": "1.0.0",
			"lockfileVersion": 3,
			"packages": {
				"": {"name": "app", "version": "1.0.0"},
				"node_modules/lodash": {"version": "4.17.20"},
				"node_modules/@scope/pkg": {"version": "2.0.0"},
				"node_modules/a/node_modules/b": {"version": "3.0.0"},
				"packages/workspace": {"name": "workspace", "version": "9.9.9"}
			}
		}`,
	})

	assertFindings(t, store, root, "lodash 4.17.20", "app/package-lock.json@4.17.20")
	assertFindings(t, store, root, "@scope/pkg 2.0.0", "app/package-lock.json@2.0.0")
	assertFindings(t, store, root, "b 3.0.0", "app/package-lock.json@3.0.0")

	// The root project and workspace source entries are not installed packages.
	assertFindings(t, store, root, "app 1.0.0")
	assertFindings(t, store, root, "workspace 9.9.9")

	// Lockfile versions are matched exactly too.
	assertFindings(t, store, root, "lodash 4.17.2")
}

func TestLockfileDependenciesV1(t *testing.T) {
	store, root := newTestStore(t, map[string]string{
		"app/package-lock.json": `{
			"name": "app",
			"version": "1.0.0",
			"lockfileVersion": 1,
			"dependencies": {
				"lodash": {"version": "4.17.20"},
				"outer": {
					"version": "1.0.0",
					"dependencies": {
						"inner": {"version": "2.0.0"}
					}
				}
			}
		}`,
	})

	assertFindings(t, store, root, "lodash 4.17.20", "app/package-lock.json@4.17.20")
	assertFindings(t, store, root, "outer 1.0.0", "app/package-lock.json@1.0.0")
	assertFindings(t, store, root, "inner 2.0.0", "app/package-lock.json@2.0.0")
}

func TestLockfileV2DoesNotDuplicatePackagesAndDependencies(t *testing.T) {
	store, root := newTestStore(t, map[string]string{
		"app/package-lock.json": `{
			"name": "app",
			"version": "1.0.0",
			"lockfileVersion": 2,
			"packages": {
				"node_modules/lodash": {"version": "4.17.20"}
			},
			"dependencies": {
				"lodash": {"version": "4.17.20"}
			}
		}`,
	})

	// One file, one package version: the index deduplicates by file path.
	assertFindings(t, store, root, "lodash 4.17.20", "app/package-lock.json@4.17.20")
}

func TestLockfileOnlyProjectIsDetected(t *testing.T) {
	store, root := newTestStore(t, map[string]string{
		"app/package-lock.json": `{
			"name": "app",
			"version": "1.0.0",
			"lockfileVersion": 3,
			"packages": {
				"node_modules/lodash": {"version": "4.17.20"}
			}
		}`,
	})

	assertFindings(t, store, root, "lodash 4.17.20", "app/package-lock.json@4.17.20")
}

func TestFindAllVersions(t *testing.T) {
	store, root := newTestStore(t, map[string]string{
		"one/package.json":   `{"name": "lodash", "version": "4.17.20"}`,
		"two/package.json":   `{"name": "lodash", "version": "4.17.21"}`,
		"three/package.json": `{"name": "other", "version": "1.0.0"}`,
	})

	assertFindings(t, store, root, "lodash",
		"one/package.json@4.17.20",
		"two/package.json@4.17.21")
	assertFindings(t, store, root, "not-installed")
}

func TestQueryByNameAndVersions(t *testing.T) {
	store, root := newTestStore(t, map[string]string{
		"one/package.json":   `{"name": "lodash", "version": "4.17.20"}`,
		"two/package.json":   `{"name": "lodash", "version": "4.17.20"}`,
		"three/package.json": `{"name": "lodash", "version": "4.17.21"}`,
	})

	got := store.QueryByNameAndVersions("lodash", []string{"4.17.20", "4.17.21"})
	want := make([]string, 0, len(got))
	for _, relative := range []string{"one/package.json", "three/package.json", "two/package.json"} {
		want = append(want, filepath.Join(root, filepath.FromSlash(relative)))
	}
	sort.Strings(want)

	if !reflect.DeepEqual(got, want) {
		t.Errorf("QueryByNameAndVersions() = %v, want %v", got, want)
	}
}

func TestDoesNotContainPackageUsesIndexedNames(t *testing.T) {
	store, _ := newTestStore(t, map[string]string{
		"app/package-lock.json": `{
			"lockfileVersion": 3,
			"packages": {"node_modules/lodash": {"version": "4.17.20"}}
		}`,
	})

	if store.DoesNotContainPackage("lodash") {
		t.Error("DoesNotContainPackage(lodash) = true, want false")
	}
	if !store.DoesNotContainPackage("not-installed") {
		t.Error("DoesNotContainPackage(not-installed) = false, want true")
	}
}

func TestSizeCountsUniqueFiles(t *testing.T) {
	store, root := newTestStore(t, map[string]string{
		"app/package.json":      `{"name": "app", "version": "1.0.0"}`,
		"app/package-lock.json": `{"lockfileVersion": 3, "packages": {}}`,
	})

	if got := store.Size(); got != 2 {
		t.Errorf("Size() = %d, want 2", got)
	}

	store = NewDefaultStoreWithFinder(staticFinder{files: []string{
		filepath.Join(root, "app", "package.json"),
		filepath.Join(root, "app", "package.json"),
	}})
	if got := store.Size(); got != 1 {
		t.Errorf("Size() with duplicate finder results = %d, want 1", got)
	}
}

func TestMalformedFilesAreSkipped(t *testing.T) {
	store, root := newTestStore(t, map[string]string{
		"bad/package.json":      `{ not json`,
		"bad/package-lock.json": `also not json`,
	})

	assertFindings(t, store, root, "lodash 4.17.20")
	if got := store.Size(); got != 2 {
		t.Errorf("Size() = %d, want 2 (readable files are still counted)", got)
	}
}
