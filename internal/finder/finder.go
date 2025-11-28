package finder

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/olle/npmprobe/internal/mdfind"
)

// FileMap holds file paths and their contents indexed by filepath
type FileMap map[string]string

// LoadPackageFiles loads all package.json and package-lock.json files into memory
func LoadPackageFiles() (FileMap, error) {
	fm := make(FileMap)

	// Query for package.json and package-lock.json files
	query := `kMDItemFSName == "package.json" || kMDItemFSName == "package-lock.json"`
	files, err := mdfind.FindFiles(query)
	if err != nil {
		return nil, fmt.Errorf("mdfind failed: %w", err)
	}

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

	return fm, nil
}

// FindPackageInFiles searches for a package@version in the loaded file map
func FindPackageInFiles(fm FileMap, pkgName, version string) []string {
	var matches []string
	searchStr := fmt.Sprintf(`"%s":`, pkgName)
	versionStr := fmt.Sprintf(`%s"`, version)

	for path, content := range fm {
		// Only search in package.json or package-lock.json files
		base := filepath.Base(path)
		if base != "package.json" && base != "package-lock.json" {
			continue
		}

		// Check if both package name and version appear in the file
		if strings.Contains(content, searchStr) && strings.Contains(content, versionStr) {
			matches = append(matches, path)
		}
	}

	return matches
}
