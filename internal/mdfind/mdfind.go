package mdfind

import (
	"bytes"
	"fmt"
	"os/exec"
)

// FindFiles runs mdfind with the given query and returns matching file paths.
// On any error it will panic so callers don't need to handle the error case.
func FindFiles(query string) []string {
	cmd := exec.Command("mdfind", query)
	var out bytes.Buffer
	var errb bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errb

	if err := cmd.Run(); err != nil {
		panic(fmt.Errorf("mdfind failed: %w; stderr=%s", err, errb.String()))
	}

	// Parse newline-separated output
	result := bytes.Split(bytes.TrimSpace(out.Bytes()), []byte("\n"))
	var files []string
	for _, b := range result {
		if len(b) > 0 {
			files = append(files, string(b))
		}
	}
	return files
}
