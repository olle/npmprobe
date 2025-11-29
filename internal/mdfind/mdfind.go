package mdfind

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// FindFiles runs the appropriate system search command (mdfind on macOS, locate on Linux)
// to find files matching the given query. On any error it will panic so callers don't need
// to handle the error case.
func FindFiles(query string) []string {
	var files []string

	switch runtime.GOOS {
	case "darwin":
		files = findFilesMacOS(query)
	case "linux":
		files = findFilesLinux(query)
	default:
		panic(fmt.Errorf("unsupported OS: %s", runtime.GOOS))
	}

	return files
}

// findFilesMacOS uses mdfind to find files matching the query
func findFilesMacOS(query string) []string {
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

// findFilesLinux uses locate to find files matching package.json or package-lock.json
func findFilesLinux(query string) []string {
	// Extract filenames from the Spotlight query
	// Expected format: kMDItemFSName == "package.json" || kMDItemFSName == "package-lock.json"
	// We'll search for both filenames using locate
	filenames := extractFilenamesFromQuery(query)

	var files []string
	for _, filename := range filenames {
		cmd := exec.Command("locate", filename)
		var out bytes.Buffer
		var errb bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &errb

		if err := cmd.Run(); err != nil {
			// If locate fails (e.g., database not updated), skip but warn
			fmt.Fprintf(os.Stderr, "Warning: locate failed for %s: %v\n", filename, err)
			continue
		}

		// Parse newline-separated output
		result := bytes.Split(bytes.TrimSpace(out.Bytes()), []byte("\n"))
		for _, b := range result {
			if len(b) > 0 {
				files = append(files, string(b))
			}
		}
	}

	return files
}

// extractFilenamesFromQuery parses the Spotlight query to extract filenames
// Handles format like: kMDItemFSName == "file1" || kMDItemFSName == "file2"
func extractFilenamesFromQuery(query string) []string {
	var filenames []string

	// Simple parsing: look for quoted strings after kMDItemFSName ==
	parts := strings.Split(query, "||")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		// Extract filename between quotes
		if start := strings.Index(part, `"`); start != -1 {
			if end := strings.Index(part[start+1:], `"`); end != -1 {
				filename := part[start+1 : start+1+end]
				filenames = append(filenames, filename)
			}
		}
	}

	return filenames
}
