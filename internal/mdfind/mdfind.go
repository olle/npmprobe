package mdfind

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// FindFiles runs the appropriate system search command (mdfind on macOS, locate on Linux)
// to find files matching the given query. On Windows, it falls back to pure Go filesystem walking.
// On any error it will panic so callers don't need to handle the error case.
func FindFiles(query string) []string {
	var files []string

	switch runtime.GOOS {
	case "darwin":
		files = findFilesMacOS(query)
	case "linux":
		files = findFilesLinux(query)
	case "windows":
		files = findFilesWindows(query)
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

// findFilesWindows uses pure Go filesystem walking to find files on Windows
// This is a fallback for Windows where system-specific search tools are not available
func findFilesWindows(query string) []string {
	// Extract filenames from the query
	filenames := extractFilenamesFromQuery(query)
	if len(filenames) == 0 {
		return []string{}
	}

	// Build a map of filenames to search for
	targetFiles := make(map[string]struct{})
	for _, fn := range filenames {
		targetFiles[fn] = struct{}{}
	}

	// Search from common root drives on Windows
	var allFiles []string
	drives := []string{"C:\\", "D:\\", "E:\\", "F:\\"}

	for _, drive := range drives {
		if _, err := os.Stat(drive); err != nil {
			// Skip if drive doesn't exist
			continue
		}

		files := findFilesInDirectory(drive, targetFiles)
		allFiles = append(allFiles, files...)
	}

	return allFiles
}

// findFilesInDirectory recursively searches for target files starting from a root directory
func findFilesInDirectory(root string, targetFiles map[string]struct{}) []string {
	var results []string

	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// Skip inaccessible directories and continue
			return nil
		}

		// Check if this file is one of our targets
		if !info.IsDir() {
			if _, exists := targetFiles[info.Name()]; exists {
				results = append(results, path)
			}
		}

		return nil
	})

	return results
}
