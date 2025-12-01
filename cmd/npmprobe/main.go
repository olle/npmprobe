package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/olle/npmprobe/internal/finder"
)

func main() {
	// Verbose: show packages that are not found in any files
	verbose := flag.Bool("v", false, "show packages not found in any files")
	flag.Parse()

	inputFile := "-"
	if len(flag.Args()) > 0 {
		inputFile = flag.Args()[0]
	}

	result := 0

	// Initialize the file map from mdfind results (errors will panic)
	fmt.Fprintf(os.Stderr, "Initializing file map from mdfind results...\n")
	fileMap := finder.LoadPackageFiles()
	fmt.Fprintf(os.Stderr, "Loaded %d package files\n", len(fileMap))

	// Read package list from input
	input := os.Stdin
	if inputFile != "-" {
		f, err := os.Open(inputFile)
		if err != nil {
			log.Fatalf("Error opening input file: %v\n", err)
		}
		defer f.Close()
		input = f
	}

	// Read all input packages first so we can show a percentage progress bar.
	scanner := bufio.NewScanner(input)
	packages := make([]string, 0)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		// Keep only lines that look like 'package version' (space-separated)
		parts := strings.Split(line, " ")
		if len(parts) < 2 {
			continue
		}
		packages = append(packages, line)
	}
	if err := scanner.Err(); err != nil {
		log.Fatalf("Error reading input: %v\n", err)
	}

	packageCount := 0
	foundCount := 0

	total := len(packages)
	fmt.Fprintf(os.Stderr, "Scanning packages...\n")
	if total == 0 {
		fmt.Fprintf(os.Stderr, "No packages to scan\n")
	}

	// Progress bar settings
	barWidth := 30
	lastPercent := -1

	for i, pkg := range packages {
		parts := strings.Split(pkg, " ")
		if len(parts) < 2 {
			continue
		}

		packageCount++
		pkgName := parts[0]
		version := parts[1]

		found := false

		// Search in loaded files
		matches := finder.FindPackageInFiles(fileMap, pkgName, version)
		if len(matches) > 0 {
			fmt.Printf("[FOUND] %s@%s in the following package files:\n", pkgName, version)
			for _, path := range matches {
				fmt.Printf("\t%s\n", path)
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
