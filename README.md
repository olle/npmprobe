NPM Probe
=========

A simple script that does an exhaustive check for for compromised installed
`npm` packages - from an authored list.

## Quickstart

Simply probes for all packages and versions in the list of `compromised.txt`
packages.

```sh
> ./npmprobe compromised.txt ## do check
[OK]   @operato/help@operato/help@9.0.36 not present in any files
[OK]   @operato/help@operato/help@9.0.37 not present in any files
[OK]   @operato/help@operato/help@9.0.38 not present in any files
...
```

Probing may take a while and the results are output to the console, so that the
user can take further actions.

## Adding new packages

Simply edit the `compromised.txt` file and prepare it for publishing using the
provided `Makefile`.

```sh
> make ## fixes the package list file
```

Now commit the changes and either create a merge request or simply push it to
the central repository.

Happy hacking!