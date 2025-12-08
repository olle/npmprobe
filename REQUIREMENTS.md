Requirements
------------

This file lists the requirements for the npmprobe project.

- A command line tool written in Go, compiled into a single binary `npmprobe`.
- Cross-compiles for multiple platforms (darwin/arm64, linux/amd64)

## Features

- Exhaustive search for npm package files in whole filesystem.
- Finds any package.json and package-lock.json files, regardless of location.
- Easy to use, scanning for compromised packages returning an exit code.
- Compatible with CI/CD pipelines.
- Prints progress information to the user, including number of files scanned.

## Technical Details

- The main command is easy to read and understand, making use of 3 critical
abstractions:
  - The package store - an in-memory database of found package files, and their
    contents. The store supports direct queries for package names and their
    existence, per file. It provides methods to scan for distinct packages and
    version per line by line in each file. It also provides for ways to extend
    the store with additional methods to query the data.
  - Package file finder. A file system walker that searches for package.json
    and package-lock.json files. It uses efficient file system traversal
    techniques to minimize the time taken to find files. It returns a list
    of found files to be loaded by the package store. Initial implementation
    uses native tools, such as mdfind on macOS, and mlocate or plocate on
    Linux. Future implementations include pure Go implementations that do not
    rely on external tools, allowing for better cross-platform support - even
    on Windows.
  - Compromised list parser. A parser that reads a list of compromised
    packages from a text file. It has the ability to parse different line-
    formats, initially package and version pairs, or packages and a list of
    comma-separated versions. The parser returns a data structure that can be
    used to query if a package and version is compromised. It is designed to be
    extensible to support additional formats in the future.

## Pseudo code

1. parse command line arguments, resolving flags for:
   - compromised list file path
   - verbose output
2. Initialize package store, loading all package files found by the finder,
reporting progress to the user.
3. Parse compromised list file, reporting progress to the user and preparing
data structure for queries against the store.
4. For each compromised package and version, query the store for existence.
   - If found, print details to the user or aggregate results for CI/CD or
   a final report with exit code.
5. Exit with code 0 if no compromised packages found, otherwise exit with code 1.