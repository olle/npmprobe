NPM Probe
=========

A Go tool that does an exhaustive check for compromised installed `npm` packages
from an authored list. Uses `mdfind` or `mlocate` to efficiently locate and search `package.json` and `package-lock.json` files across the system.


## Building

Build cross-platform binaries for macOS and Linux:

```sh
make build
```

This produces binaries in the `bin/` directory:
- `npmprobe` — macOS (Apple Silicon)
- `npmprobe-linux-x86_64` — Linux (x86_64)

### Testing on Linux with Docker

Build and test the Linux binary in a Docker container:

```sh
docker build -t npmprobe:test .
docker run npmprobe:test
```

The Dockerfile builds the npmprobe binary and runs it against the `compromised.txt` file
on a Linux Alpine system with the `locate` database pre-populated.

## Makefile Targets

- `make all` — builds binaries, normalizes the `compromised.txt` list, and runs a
  verbose probe against the current system.
- `make build` — cross-compiles binaries for macOS and Linux.
- `make fmt` — formats all Go source files using `go fmt`.
- `make prepare-list` — normalizes `compromised.txt` (sorts and removes duplicates).
- `make docker-test` — builds and runs the Docker test image on Linux.

## Quickstart

Probe for all packages and versions in the `compromised.txt` file:

```sh
./bin/npmprobe -v compromised.txt
[OK]   @operato/help@9.0.36 not present in any files
[OK]   @operato/help@9.0.37 not present in any files
[OK]   @operato/help@9.0.38 not present in any files
...
```

Probing may take a while on first run (loading files into memory), and results
are output to the console for further action. When a compromised package is found,
it displays the matching files:

```sh
./bin/npmprobe compromised.txt
...
[FOUND] etag@1.8.1 in the following package files:
	/Users/me/Development/project/node_modules/express/package.json
	/Users/me/Development/project/node_modules/vite/package.json
	/Users/me/Development/project/node_modules/send/package.json
```

### Flags

- `-v`: verbose. When set, the tool prints packages that were not found in any
  scanned `package.json`/`package-lock.json` files (lines prefixed with
  `[OK]`). By default (without `-v`) only found matches (`[FOUND]`) are
  printed.


## Adding new packages

Edit `compromised.txt` and normalize the list using:

```sh
make prepare-list
```

This sorts and removes duplicates. Then commit and push your changes.

## Development

The project is written in Go and organized as follows:

- `cmd/npmprobe/` — main entry point
- `internal/mdfind/` — macOS `mdfind` query wrapper
- `internal/mlocate/` — Linux `mlocate` query wrapper
- `internal/finder/` — file loading and searching logic
- `compromised.txt` — list of known compromised packages

Happy hacking!
