package store

import "strings"

// yarnLockIndexer indexes installed versions from Yarn lockfiles. It supports
// both Yarn classic (v1) and Yarn Berry (v2+), which share a file name but use
// different entry syntax.
type yarnLockIndexer struct{}

// yarnEntry holds the fields of one lockfile entry that matter for indexing.
type yarnEntry struct {
	selector   string
	version    string
	resolution string
}

func (yarnLockIndexer) Index(index *Index, path string, content []byte) {
	var current *yarnEntry

	flush := func() {
		if current == nil {
			return
		}
		indexYarnEntry(index, path, *current)
		current = nil
	}

	for _, raw := range strings.Split(string(content), "\n") {
		line := strings.TrimRight(raw, "\r")
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		// Entry headers start at column zero; fields are indented.
		if line[0] != ' ' && line[0] != '\t' {
			flush()
			current = &yarnEntry{selector: strings.TrimSuffix(trimmed, ":")}
			continue
		}

		if current == nil {
			continue
		}

		field, value, ok := parseYarnField(trimmed)
		if !ok {
			continue
		}

		switch field {
		case "version":
			current.version = value
		case "resolution":
			current.resolution = value
		}
	}

	flush()
}

// parseYarnField parses the two field syntaxes used by Yarn lockfiles:
// Berry's "version: 1.2.3" and classic's `version "1.2.3"`.
func parseYarnField(line string) (field, value string, ok bool) {
	if colon := strings.Index(line, ":"); colon != -1 {
		field = strings.TrimSpace(line[:colon])
		if field != "version" && field != "resolution" {
			return "", "", false
		}
		return field, trimLockQuotes(line[colon+1:]), true
	}

	parts := strings.Fields(line)
	if len(parts) < 2 || (parts[0] != "version" && parts[0] != "resolution") {
		return "", "", false
	}
	return parts[0], trimLockQuotes(parts[1]), true
}

func indexYarnEntry(index *Index, path string, entry yarnEntry) {
	selector := trimLockQuotes(entry.selector)
	if entry.version == "" || selector == "" || selector == "__metadata" {
		return
	}

	// Workspace, link, and file entries describe local projects, not packages
	// installed from a registry.
	for _, local := range []string{"@workspace:", "@link:", "@file:", "@portal:"} {
		if strings.Contains(entry.selector, local) || strings.Contains(entry.resolution, local) {
			return
		}
	}

	name := ""
	if entry.resolution != "" {
		name = yarnPackageName(entry.resolution)
	}
	if name == "" {
		name = yarnPackageName(firstYarnSelector(entry.selector))
	}

	index.Add(name, entry.version, path)
}

// firstYarnSelector returns the first selector from a classic lockfile key.
// Keys may combine several ranges for the same package, separated by commas.
func firstYarnSelector(selector string) string {
	if comma := strings.Index(selector, ","); comma != -1 {
		selector = selector[:comma]
	}
	return trimLockQuotes(selector)
}

// yarnPackageName extracts the package name from a Yarn selector or resolution.
// Resolutions look like "lodash@npm:4.17.21", "@scope/pkg@npm:1.0.0", or
// "pkg@patch:pkg@npm%3A1.0.0#...".
func yarnPackageName(value string) string {
	value = trimLockQuotes(value)

	// An npm alias selector looks like "alias@npm:real-package@^1.0.0";
	// a resolution looks like "real-package@npm:1.0.0". Index the real package
	// in the alias case.
	if at := strings.Index(value, "@npm:"); at > 0 {
		rest := value[at+len("@npm:"):]
		if strings.LastIndex(rest, "@") > 0 {
			return yarnPackageName(rest)
		}
		return value[:at]
	}

	for _, protocol := range []string{
		"@patch:", "@workspace:", "@link:", "@file:", "@portal:",
		"@exec:", "@git:", "@github:", "@http:", "@https:",
	} {
		if at := strings.Index(value, protocol); at > 0 {
			return value[:at]
		}
	}

	// Classic selectors use "name@range"; scoped names start with "@".
	if at := strings.LastIndex(value, "@"); at > 0 {
		return value[:at]
	}

	return value
}
