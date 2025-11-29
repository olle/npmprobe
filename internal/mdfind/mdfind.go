package mdfind

import (
	"bytes"
	"os/exec"
)

// FindFiles runs mdfind with the given query and returns matching file paths.
func FindFiles(query string) ([]string, error) {
	cmd := exec.Command("mdfind", query)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &bytes.Buffer{}

	if err := cmd.Run(); err != nil {
		// Return empty list on error, don't fail
		return []string{}, nil
	}

	// Parse null-separated output or newline-separated if not using -0
	result := bytes.Split(bytes.TrimSpace(out.Bytes()), []byte("\n"))
	var files []string
	for _, b := range result {
		if len(b) > 0 {
			files = append(files, string(b))
		}
	}
	return files, nil
}
