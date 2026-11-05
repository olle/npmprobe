package store

import (
	"encoding/json"
	"strings"
)

// bunLockIndexer indexes installed versions from Bun's text lockfile
// ("bun.lock"). Bun's binary lockfile ("bun.lockb") is not indexed.
type bunLockIndexer struct{}

func (bunLockIndexer) Index(index *Index, path string, content []byte) {
	var lockfile struct {
		Packages map[string][]any `json:"packages"`
	}
	if err := json.Unmarshal(stripJSONC(content), &lockfile); err != nil {
		return
	}

	for _, entry := range lockfile.Packages {
		if len(entry) == 0 {
			continue
		}
		resolution, ok := entry[0].(string)
		if !ok {
			continue
		}

		name, version := bunPackageNameVersion(resolution)
		index.Add(name, version, path)
	}
}

// bunPackageNameVersion splits a Bun resolution string such as
// "lodash@4.17.21" or "@scope/pkg@2.0.0", ignoring peer suffixes.
func bunPackageNameVersion(resolution string) (name, version string) {
	resolution = strings.TrimSpace(resolution)
	if paren := strings.Index(resolution, "("); paren != -1 {
		resolution = resolution[:paren]
	}

	at := strings.LastIndex(resolution, "@")
	if at <= 0 {
		return "", ""
	}

	name = resolution[:at]
	version = resolution[at+1:]
	if !isPackageNameToken(name) || !isVersionToken(version) {
		return "", ""
	}
	return name, version
}

// stripJSONC removes comments and trailing commas from Bun's JSONC lockfile so
// it can be decoded with encoding/json. String contents are preserved.
func stripJSONC(content []byte) []byte {
	withoutComments := make([]byte, 0, len(content))
	inString := false
	escaped := false

	for i := 0; i < len(content); i++ {
		c := content[i]

		if inString {
			withoutComments = append(withoutComments, c)
			switch {
			case escaped:
				escaped = false
			case c == '\\':
				escaped = true
			case c == '"':
				inString = false
			}
			continue
		}

		if c == '"' {
			inString = true
			withoutComments = append(withoutComments, c)
			continue
		}

		if c == '/' && i+1 < len(content) {
			if content[i+1] == '/' {
				i += 2
				for i < len(content) && content[i] != '\n' {
					i++
				}
				if i < len(content) {
					withoutComments = append(withoutComments, '\n')
				}
				continue
			}
			if content[i+1] == '*' {
				i += 2
				for i+1 < len(content) && !(content[i] == '*' && content[i+1] == '/') {
					i++
				}
				i++ // skip the closing '/'
				continue
			}
		}

		withoutComments = append(withoutComments, c)
	}

	return stripTrailingCommas(withoutComments)
}

// stripTrailingCommas removes commas that appear immediately before a closing
// object or array, as allowed by JSONC.
func stripTrailingCommas(content []byte) []byte {
	result := make([]byte, 0, len(content))
	inString := false
	escaped := false

	for i := 0; i < len(content); i++ {
		c := content[i]

		if inString {
			result = append(result, c)
			switch {
			case escaped:
				escaped = false
			case c == '\\':
				escaped = true
			case c == '"':
				inString = false
			}
			continue
		}

		if c == '"' {
			inString = true
			result = append(result, c)
			continue
		}

		if c == ',' {
			next := i + 1
			for next < len(content) && strings.ContainsRune(" \t\r\n", rune(content[next])) {
				next++
			}
			if next < len(content) && (content[next] == '}' || content[next] == ']') {
				continue
			}
		}

		result = append(result, c)
	}

	return result
}
