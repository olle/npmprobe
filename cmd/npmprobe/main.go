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
	fileMap := finder.LoadPackageFiles()

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

	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		pkg := strings.TrimSpace(scanner.Text())
		if pkg == "" {
			continue
		}

		// Parse 'package@version'
		parts := strings.Split(pkg, " ")
		if len(parts) < 2 {
			continue
		}

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
			result = 1
		}

		if !found && *verbose {

			fmt.Printf("[OK]   %s@%s not present in any files\n", pkgName, version)

		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Error reading input: %v\n", err)
	}

	os.Exit(result)
}
