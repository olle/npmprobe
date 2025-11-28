NPM Probe
=========

A Go tool that does an exhaustive check for compromised installed `npm` packages
from an authored list. Uses `mdfind` to efficiently locate and search `package.json`
and `package-lock.json` files across the system.

## Building

Build cross-platform binaries for macOS, Linux, and Windows:

```sh
make build
```

This produces binaries in the `bin/` directory:
- `npmprobe-darwin-aarch64` — macOS (Apple Silicon)
- `npmprobe-linux-x86_64` — Linux (x86_64)
- `npmprobe-windows-x86_64.exe` — Windows (x86_64)

## Quickstart

Probe for all packages and versions in the `compromised.txt` file:

```sh
./bin/npmprobe-darwin-aarch64 < compromised.txt
[OK]   @operato/help@9.0.36 not present in any files
[OK]   @operato/help@9.0.37 not present in any files
[OK]   @operato/help@9.0.38 not present in any files
...
```

Probing may take a while on first run (loading files into memory), and results
are output to the console for further action. When a compromised package is found,
it displays the matching files:

```sh
./bin/npmprobe-darwin-aarch64 < compromised.txt
[FOUND] etag@1.8.1 in the following package files:
	/Users/me/Development/project/node_modules/express/package.json
	/Users/me/Development/project/node_modules/vite/package.json
	/Users/me/Development/project/node_modules/send/package.json
```

## Adding new packages

Edit `compromised.txt` and normalize the list using:

```sh
make prepare-list
```

This sorts and removes duplicates. Then commit and push your changes.

## Development

The project is written in Go and organized as follows:
- `cmd/npmprobe/` — main entry point
- `internal/mdfind/` — mdfind query wrapper
- `internal/finder/` — file loading and searching logic

Happy hacking!