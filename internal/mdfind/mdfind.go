package mdfind

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// searchTimeout bounds the external search commands so a hung mdfind or
// locate cannot stall the whole scan.
var searchTimeout = 2 * time.Minute

// FindFiles runs the appropriate system search command (mdfind on macOS,
// locate on Linux) to find files matching the given query. On Windows it falls
// back to pure Go filesystem walking.
//
// It returns an error when discovery cannot be completed reliably, so callers
// can fail closed instead of reporting a clean scan for an empty result set.
func FindFiles(query string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), searchTimeout)
	defer cancel()

	switch runtime.GOOS {
	case "darwin":
		return findFilesMacOS(ctx, query)
	case "linux":
		return findFilesLinux(ctx, query)
	case "windows":
		return findFilesWindows(query)
	default:
		return nil, fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
}

// findFilesMacOS uses mdfind to find files matching the query.
func findFilesMacOS(ctx context.Context, query string) ([]string, error) {
	out, err := runSearchCommand(ctx, "mdfind", query)
	if err != nil {
		return nil, err
	}
	return splitLines(out), nil
}

// findFilesLinux uses locate to find files matching the requested package and
// lockfile names. A locate exit status of 1 means "no matches"; any other
// failure aborts discovery.
func findFilesLinux(ctx context.Context, query string) ([]string, error) {
	// Extract filenames from the Spotlight query. Expected format:
	// kMDItemFSName == "package.json" || kMDItemFSName == "yarn.lock" || ...
	// Each filename is searched separately using locate.
	filenames := extractFilenamesFromQuery(query)
	if len(filenames) == 0 {
		return nil, fmt.Errorf("discovery query contains no file names")
	}

	var files []string
	for _, filename := range filenames {
		out, err := runSearchCommand(ctx, "locate", filename)
		if err != nil {
			var exitErr *exec.ExitError
			if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
				// GNU locate uses exit status 1 when nothing matched.
				continue
			}
			return nil, err
		}
		files = append(files, splitLines(out)...)
	}

	return files, nil
}

// runSearchCommand runs an external search command with the supplied context
// and returns its standard output. Failures include the command's stderr and
// timeouts are reported distinctly.
func runSearchCommand(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	var out bytes.Buffer
	var errb bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errb

	err := cmd.Run()
	if ctx.Err() == context.DeadlineExceeded {
		return nil, fmt.Errorf("%s timed out after %s", name, searchTimeout)
	}
	if err != nil {
		return nil, fmt.Errorf("%s failed: %w; stderr=%s", name, err, strings.TrimSpace(errb.String()))
	}
	return out.Bytes(), nil
}

// splitLines parses newline-separated command output into non-empty paths.
func splitLines(out []byte) []string {
	result := bytes.Split(bytes.TrimSpace(out), []byte("\n"))
	files := make([]string, 0, len(result))
	for _, b := range result {
		if len(b) > 0 {
			files = append(files, string(b))
		}
	}
	return files
}

// extractFilenamesFromQuery parses the Spotlight query to extract filenames.
// It handles a format like:
// kMDItemFSName == "file1" || kMDItemFSName == "file2"
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

// findFilesWindows uses pure Go filesystem walking to find files on Windows.
// This is a fallback for Windows where system-specific search tools are not
// available.
func findFilesWindows(query string) ([]string, error) {
	// Extract filenames from the query
	filenames := extractFilenamesFromQuery(query)
	if len(filenames) == 0 {
		return nil, fmt.Errorf("discovery query contains no file names")
	}

	// Build a map of filenames to search for
	targetFiles := make(map[string]struct{})
	for _, fn := range filenames {
		targetFiles[fn] = struct{}{}
	}

	// Search from common root drives on Windows
	var allFiles []string
	foundDrive := false
	drives := []string{"C:\\", "D:\\", "E:\\", "F:\\"}

	for _, drive := range drives {
		if _, err := os.Stat(drive); err != nil {
			// Skip if drive doesn't exist
			continue
		}
		foundDrive = true

		files, err := findFilesInDirectory(drive, targetFiles)
		if err != nil {
			return nil, fmt.Errorf("walking %s: %w", drive, err)
		}
		allFiles = append(allFiles, files...)
	}

	if !foundDrive {
		return nil, fmt.Errorf("no accessible filesystem drives found")
	}

	return allFiles, nil
}

// findFilesInDirectory recursively searches for target files starting from a
// root directory. Inaccessible subdirectories are skipped.
func findFilesInDirectory(root string, targetFiles map[string]struct{}) ([]string, error) {
	var results []string

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if path == root {
				return err
			}
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
	if err != nil {
		return nil, err
	}

	return results, nil
}
