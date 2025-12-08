package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/olle/npmprobe/internal/finder"
	"github.com/olle/npmprobe/internal/parser"
)

const helpText = `npmprobe - exhaustive npm package compromise detector

USAGE:
  npmprobe [FLAGS] [FILE]

DESCRIPTION:
  npmprobe performs an exhaustive search of your filesystem to find all npm
  package.json and package-lock.json files, then checks them against a list
  of known compromised packages. It uses efficient filesystem traversal to
  minimize scan time and provides clear output for CI/CD integration.

ARGUMENTS:
  FILE
    Path to a file containing compromised packages to check against.
    Each line should contain a package name and version(s) in one of these formats:
      - package version               (single version)
      - package v1,v2,v3              (comma-separated versions)
      - package v1, v2, v3            (comma-separated with spaces)
    
    If FILE is omitted or "-", npmprobe reads from standard input (stdin).
    This allows piping package lists from other tools.

FLAGS:
  -v
    Verbose output. When enabled, prints all checked packages including those
    not found in the system (prefixed with [OK]). Without this flag, only
    found matches (prefixed with [FOUND]) are displayed.
  
  -h
    Display this help message and exit.

EXAMPLES:
  # Check packages from a file with verbose output:
  $ npmprobe -v compromised.txt

  # Check packages from stdin:
  $ echo "lodash 4.17.20" | npmprobe

  # Pipe results from a vulnerability database:
  $ cat vulnerability-list.txt | npmprobe -v

OUTPUT:
  When a compromised package is found, npmprobe outputs:
    [FOUND] package@version in the following package files:
      /path/to/node_modules/package/package.json
      /path/to/another/node_modules/package/package.json

  With -v flag, non-found packages are shown as:
    [OK]   package@version not present in any files

  Progress information is displayed on stderr while scanning, allowing you to
  monitor the scan progress. Final results show:
    - Total packages scanned
    - Number of compromised packages found
    - Overall completion status

EXIT CODES:
  0 - No compromised packages found (success)
  1 - One or more compromised packages found (failure)

PERFORMANCE:
  npmprobe uses efficient filesystem search techniques:
  - On macOS: Native mdfind Spotlight database for fast searching
  - On Linux: locate command database when available
  - On Windows: Filesystem walking with configurable search paths
  
  First run may take longer as packages are loaded into memory. Subsequent
  queries against the same system snapshot are faster.

CROSS-PLATFORM SUPPORT:
  npmprobe is compiled for:
  - macOS (Apple Silicon and Intel)
  - Linux (x86_64)
  - Windows (with pure Go filesystem walking)

LICENSE:
  See LICENSE file in the repository for licensing information.
`

func printHelp() {
	fmt.Print(helpText)
}

func main() {
	// Handle command-line flags
	// Verbose: show packages that are not found in any files
	verbose := flag.Bool("v", false, "show packages not found in any files")
	help := flag.Bool("h", false, "display help message")

	flag.Parse()

	// Handle help flag
	if *help {
		printHelp()
		os.Exit(0)
	}

	listFile := "-"
	if len(flag.Args()) > 0 {
		listFile = flag.Args()[0]
	}

	// Exit code: 0 = no findings, 1 = findings
	result := 0

	// Initialize the package store with spinner animation
	packageStore := finder.LoadPackageStore()
	fmt.Fprintf(os.Stderr, "Loaded %d package files\n", packageStore.Size())

	// Open reader for input list of compromised packages
	// Default to stdin if listFile is "-", else open the specified file
	input := os.Stdin
	if listFile != "-" {
		f, err := os.Open(listFile)
		if err != nil {
			panic(fmt.Sprintf("Error opening input file: %v\n", err))
		}
		defer f.Close()
		input = f
	}

	// Read all input packages first so we can show a percentage progress bar.
	scanner := bufio.NewScanner(input)
	compromisedPackages := make([]string, 0)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		// Keep only lines that look like 'package version...' (space-separated)
		parts := strings.Split(line, " ")
		if len(parts) < 2 {
			continue
		}
		compromisedPackages = append(compromisedPackages, line)
	}
	if err := scanner.Err(); err != nil {
		log.Fatalf("Error reading input: %v\n", err)
	}

	packageCount := 0
	foundCount := 0

	// Buffer for formatted findings when not in verbose mode
	bufferedFindings := make([]string, 0)

	total := len(compromisedPackages)
	fmt.Fprintf(os.Stderr, "Scanning packages...\n")
	if total == 0 {
		fmt.Fprintf(os.Stderr, "No packages to scan\n")
	}

	// Progress bar settings
	barWidth := 50
	lastPercent := -1

	for i, pkg := range compromisedPackages {

		// Parse package and version into a matcher
		packageMatcher, err := parser.ParseLine(pkg)
		if err != nil {
			panic(fmt.Sprintf("Error parsing package line '%s': %v", pkg, err))
		}

		packageCount++
		pkgName := packageMatcher.Name()
		version := packageMatcher.Version()
		found := false

		// Search for the package in the store using the matcher
		matches := finder.FindPackageInStore(packageStore, packageMatcher)
		if len(matches) > 0 {

			// Format the finding
			var b strings.Builder
			fmt.Fprintf(&b, "[FOUND] %s@%s in the following package files:\n", pkgName, version)
			for _, path := range matches {
				fmt.Fprintf(&b, "\t%s\n", path)
			}

			if *verbose {
				// Print immediately in verbose mode
				fmt.Print(b.String())
			} else {
				// Buffer the formatted output to print at the end
				bufferedFindings = append(bufferedFindings, b.String())
			}

			found = true
			foundCount++
			result = 1
		}

		if !found && *verbose {
			fmt.Printf("[OK]   %s@%s not present in any files\n", pkgName, version)
		}

		// Update progress bar on stderr. Re-draw only when percent changes.
		if total > 0 {
			percent := (i + 1) * 100 / total
			if percent != lastPercent {
				filled := percent * barWidth / 100
				if filled > barWidth {
					filled = barWidth
				}
				if filled < 0 {
					filled = 0
				}
				bar := strings.Repeat("=", filled) + strings.Repeat(" ", barWidth-filled)
				if !*verbose {
					fmt.Fprintf(os.Stderr, "\r[%s] %3d%% (%d/%d)", bar, percent, i+1, total)
				}
				lastPercent = percent
			}
		}
	}

	// Finish progress bar line
	if total > 0 {
		fmt.Fprintf(os.Stderr, "\n")
	}

	// If not verbose, print buffered findings now
	if !*verbose && len(bufferedFindings) > 0 {
		for _, s := range bufferedFindings {
			fmt.Print(s)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Error reading input: %v\n", err)
	}

	fmt.Fprintf(os.Stderr, "Scanned %d packages\n", packageCount)

	if result == 0 {
		fmt.Fprintf(os.Stderr, "No findings, ok\n")
	} else {
		fmt.Fprintf(os.Stderr, "Found %d compromised packages\n", foundCount)
	}

	os.Exit(result)
}
