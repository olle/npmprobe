package store

import "strings"

// pnpmLockIndexer indexes installed versions from pnpm lockfiles. It reads the
// package keys in the "packages" and "snapshots" sections, which cover the
// v5, v6, and v9 lockfile layouts.
type pnpmLockIndexer struct{}

func (pnpmLockIndexer) Index(index *Index, path string, content []byte) {
	section := ""

	for _, raw := range strings.Split(string(content), "\n") {
		line := strings.TrimRight(raw, "\r")
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		// Section headers are at column zero. Nested YAML is not a package key.
		if line[0] != ' ' && line[0] != '\t' {
			switch key := strings.TrimSuffix(trimmed, ":"); key {
			case "packages", "snapshots":
				section = key
			default:
				section = ""
			}
			continue
		}

		if section == "" {
			continue
		}

		// Direct children of a section are indented by two spaces and are
		// resolution keys, either block-style ("key:") or inline ("key: {}").
		if !strings.HasPrefix(line, "  ") || strings.HasPrefix(line, "    ") {
			continue
		}

		colon := strings.Index(trimmed, ":")
		if colon == -1 {
			continue
		}

		name, version := parsePnpmKey(trimLockQuotes(trimmed[:colon]))
		index.Add(name, version, path)
	}
}

// parsePnpmKey extracts a package name and version from a pnpm lockfile key.
// The formats in use are "/name/version" and "/@scope/name/version" (v5) and
// "name@version" or "/name@version" (v6+). Peer suffixes are removed.
func parsePnpmKey(key string) (name, version string) {
	key = strings.TrimPrefix(key, "/")

	// Strip berry-style peer suffixes such as "react-dom@18.2.0(react@18.2.0)".
	if paren := strings.Index(key, "("); paren != -1 {
		key = key[:paren]
	}

	// pnpm v5 uses slashes between name and version.
	for slash := 0; slash < len(key); slash++ {
		if key[slash] != '/' {
			continue
		}
		candidateName := key[:slash]
		candidateVersion := trimPnpmPeerSuffix(key[slash+1:])
		if isPackageNameToken(candidateName) && isVersionToken(candidateVersion) {
			return candidateName, candidateVersion
		}
	}

	// pnpm v6+ uses "@" between name and version.
	if at := strings.LastIndex(key, "@"); at > 0 {
		candidateName := key[:at]
		candidateVersion := trimPnpmPeerSuffix(key[at+1:])
		if isPackageNameToken(candidateName) && isVersionToken(candidateVersion) {
			return candidateName, candidateVersion
		}
	}

	return "", ""
}

// trimPnpmPeerSuffix removes the v5 peer-dependency suffix from a version.
func trimPnpmPeerSuffix(version string) string {
	if underscore := strings.Index(version, "_"); underscore != -1 {
		return version[:underscore]
	}
	return version
}
