Requirements
------------

This file lists the requirements for the npmprobe project.

- A command line tool written in Go, compiled into a single binary `npmprobe`.
- Cross-compiles for multiple platforms (darwin/arm64, linux/amd64)

## Features

- Exhaustive search for npm package files in whole filesystem.
- Finds package manifests and lockfiles regardless of location: `package.json`,
  `package-lock.json`, `npm-shrinkwrap.json`, `yarn.lock`, `pnpm-lock.yaml`, and
  `bun.lock`.
- Easy to use, scanning for compromised packages returning an exit code.
- Compatible with CI/CD pipelines.
- Fails closed when no package files can be discovered or read, instead of
  reporting a clean scan.
- Prints progress information to the user, including number of files scanned.

## Technical Details

- The main command is easy to read and understand, making use of 3 critical
abstractions:
  - The package store - an in-memory index of installed package versions,
    mapping package name and version to the files that contain them. The index
    is built once while loading package manifests and lockfiles using
    per-file-format indexers, so queries are exact and do not rescan file
    contents. The indexers cover npm, Yarn classic and Berry, pnpm v5/v6/v9,
    and Bun text lockfiles. The store supports direct queries for package names, versions,
    and their existence, and is extensible with additional query methods.
  - Package file finder. A file system walker that searches for supported
    package manifests and lockfiles. It uses efficient file system traversal
    techniques to minimize the time taken to find files. It returns a list
    of found files to be loaded by the package store. Initial implementation
    uses native tools, such as mdfind on macOS, and mlocate or plocate on
    Linux. Future implementations include pure Go implementations that do not
    rely on external tools, allowing for better cross-platform support - even
    on Windows.
  - Compromised list parser. A parser that reads a list of compromised
    packages from a text file. It parses whitespace-separated package and
    version entries, including package and version pairs, packages with a list
    of comma-separated versions, and bare package names meaning every version
    is compromised. Invalid lines are reported with their line number instead
    of being skipped. The parser returns a data structure that can be used to
    query if a package and version is compromised. It is designed to be
    extensible to support additional formats in the future.

## Pseudo code

1. parse command line arguments, resolving flags for:
   - compromised list file path
   - verbose output
2. Initialize package store, loading all package files found by the finder,
reporting progress to the user. Abort with exit code 2 if discovery fails or
finds no readable package files.
3. Parse compromised list file, reporting progress to the user and preparing
data structure for queries against the store.
4. For each compromised package and version, query the store for existence.
   - If found, print details to the user or aggregate results for CI/CD or
   a final report with exit code.
5. Exit with code 0 if no compromised packages found, otherwise exit with code 1.

## Development

- `cmd/npmprobe/` — Main entry point handling CLI arguments, file I/O, and output
  formatting.
- `internal/finder/` — Loads the package store and provides the public API.
- `internal/parser/` — Parses compromised package lines supporting multiple
  formats (single versions, comma-separated versions).
- `internal/store/` — In-memory version index (package name → version → file
  paths) built from parsed package manifests and lockfiles. Supports pluggable
  file finders and per-file-format indexers for npm, Yarn, pnpm, and Bun files.
- `internal/mdfind/` — Platform-specific file search (macOS mdfind, Linux
  locate, Windows filesystem walking).
- `internal/spinner/` — CLI spinner animation for user feedback during
  initialization.

To build: `make build` | To test: `go test ./...` | To format: `make fmt`