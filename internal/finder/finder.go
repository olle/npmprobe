package finder

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/olle/npmprobe/internal/mdfind"
	"github.com/olle/npmprobe/internal/parser"
)

// FileMap holds file paths and their contents indexed by filepath
type FileMap map[string]string

// LoadPackageFiles loads all package.json and package-lock.json files into memory
// This function panics on any underlying mdfind error; unreadable files are skipped.
func LoadPackageFiles() FileMap {
	fm := make(FileMap)

	// Query for package.json and package-lock.json files
	query := `kMDItemFSName == "package.json" || kMDItemFSName == "package-lock.json"`
	files := mdfind.FindFiles(query)

	// Read each file into the map
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

		fm[path] = string(content)
	}

	return fm
}

// FindPackageInFiles searches for a package (represented by a parser.Matcher)
// in the loaded file map and returns paths where the package and any of its versions appear.
func FindPackageInFiles(fm FileMap, m parser.Matcher) []string {
	var matches []string
	searchStr := fmt.Sprintf(`"%s":`, m.Name())

	// Get all versions to check
	versions := m.Versions()
	if len(versions) == 0 {
		panic("Matcher has no versions")
	}

	for path, content := range fm {
		found := false

		// Check if package name appears in the file
		if !strings.Contains(content, searchStr) {
			continue
		}

		// Check each line for package name and any of the versions
		for _, line := range strings.Split(content, "\n") {
			if !strings.Contains(line, searchStr) {
				continue
			}

			// Check if any of the versions appear in this line
			for _, version := range versions {
				versionStr := fmt.Sprintf(`%s"`, version)
				if strings.Contains(line, versionStr) {
					found = true
					break
				}
			}

			if found {
				break
			}
		}

		if found {
			matches = append(matches, path)
		}
	}

	return matches
}
