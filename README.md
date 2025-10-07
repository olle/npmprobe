NPM Probe
=========

A simple script that does an exhaustive check for for compromised installed
`npm` packages - from an authored list.

## Quickstart

Simply probes for all packages and versions in the list of `compromised.txt`
packages.

```sh
> ./npmprobe compromised.txt ## do check
```

Probing may take a while and the results are output to the console, so that the
user can take further actions.

Happy hacking!