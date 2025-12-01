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

// FindPackageInFiles searches for a package@version in the loaded file map
// FindPackageInFiles searches for a package (represented by a parser.Matcher)
// in the loaded file map and returns paths where the package@version appears.
func FindPackageInFiles(fm FileMap, m parser.Matcher) []string {
	var matches []string
	searchStr := fmt.Sprintf(`"%s":`, m.Name())
	versionStr := fmt.Sprintf(`%s"`, m.Version())

	for path, content := range fm {

		found := false
		// Check if both package name and version appear in the file
		if strings.Contains(content, searchStr) && strings.Contains(content, versionStr) {

			// Check each line for both package name and version
			for _, line := range strings.Split(content, "\n") {
				if strings.Contains(line, searchStr) && strings.Contains(line, versionStr) {
					found = true
					break
				}
			}

		}

		if found {
			matches = append(matches, path)
		}
	}

	return matches
}
