package parser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseLine_ValidInput(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		expName string
		expVer  string
		expVers []string
		expAll  bool
	}{
		{
			name:    "simple package and version",
			input:   "lodash 4.17.21",
			expName: "lodash",
			expVer:  "4.17.21",
			expVers: []string{"4.17.21"},
		},
		{
			name:    "scoped package",
			input:   "@babel/core 7.20.0",
			expName: "@babel/core",
			expVer:  "7.20.0",
			expVers: []string{"7.20.0"},
		},
		{
			name:    "version with pre-release",
			input:   "react 18.0.0-alpha.1",
			expName: "react",
			expVer:  "18.0.0-alpha.1",
			expVers: []string{"18.0.0-alpha.1"},
		},
		{
			name:    "leading/trailing whitespace",
			input:   "  express 4.18.2  ",
			expName: "express",
			expVer:  "4.18.2",
			expVers: []string{"4.18.2"},
		},
		{
			name:    "extra fields ignored",
			input:   "typescript 4.9.4 extra stuff here",
			expName: "typescript",
			expVer:  "4.9.4",
			expVers: []string{"4.9.4"},
		},
		{
			name:    "comma-separated versions",
			input:   "lodash 4.17.20,4.17.21,4.17.22",
			expName: "lodash",
			expVer:  "4.17.20",
			expVers: []string{"4.17.20", "4.17.21", "4.17.22"},
		},
		{
			name:    "comma-separated with spaces",
			input:   "react 18.0.0, 18.1.0 , 18.2.0",
			expName: "react",
			expVer:  "18.0.0",
			expVers: []string{"18.0.0", "18.1.0", "18.2.0"},
		},
		{
			name:    "multiple spaces between fields",
			input:   "vue    3.2.31",
			expName: "vue",
			expVer:  "3.2.31",
			expVers: []string{"3.2.31"},
		},
		{
			name:    "tab characters as whitespace",
			input:   "angular\t12.2.16",
			expName: "angular",
			expVer:  "12.2.16",
			expVers: []string{"12.2.16"},
		},
		{
			name:    "mixed whitespace characters",
			input:   "jquery \t 3.6.0",
			expName: "jquery",
			expVer:  "3.6.0",
			expVers: []string{"3.6.0"},
		},
		{
			name:    "single version with trailing comma",
			input:   "ember 3.28.0,",
			expName: "ember",
			expVer:  "3.28.0",
			expVers: []string{"3.28.0"},
		},
		{
			name:    "versions with spaces after commas",
			input:   "backbone 1.4.0, 1.4.1 , 1.4.2 ",
			expName: "backbone",
			expVer:  "1.4.0",
			expVers: []string{"1.4.0", "1.4.1", "1.4.2"},
		},
		{
			name:    "scoped package with multiple versions",
			input:   "@angular/core 12.0.0,12.1.0,12.2.0",
			expName: "@angular/core",
			expVer:  "12.0.0",
			expVers: []string{"12.0.0", "12.1.0", "12.2.0"},
		},
		{
			name:    "package with hyphen in name",
			input:   "my-package 1.0.0,1.1.0",
			expName: "my-package",
			expVer:  "1.0.0",
			expVers: []string{"1.0.0", "1.1.0"},
		},
		{
			name:    "bare package means all versions",
			input:   "get-shit-done-cc",
			expName: "get-shit-done-cc",
			expVer:  "",
			expVers: nil,
			expAll:  true,
		},
		{
			name:    "bare scoped package with surrounding whitespace",
			input:   " \t@gsd-build/sdk  ",
			expName: "@gsd-build/sdk",
			expVer:  "",
			expVers: nil,
			expAll:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matcher, err := ParseLine(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if matcher == nil {
				t.Fatal("expected non-nil matcher")
			}

			if got := matcher.Name(); got != tt.expName {
				t.Errorf("Name() = %q, want %q", got, tt.expName)
			}

			if got := matcher.Version(); got != tt.expVer {
				t.Errorf("Version() = %q, want %q", got, tt.expVer)
			}

			if got := matcher.Versions(); len(got) != len(tt.expVers) {
				t.Errorf("Versions() length = %d, want %d", len(got), len(tt.expVers))
			} else {
				for i, v := range got {
					if v != tt.expVers[i] {
						t.Errorf("Versions()[%d] = %q, want %q", i, v, tt.expVers[i])
					}
				}
			}

			if got := matcher.AllVersions(); got != tt.expAll {
				t.Errorf("AllVersions() = %v, want %v", got, tt.expAll)
			}
		})
	}
}

func TestParseLine_InvalidInput(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "empty string",
			input: "",
		},
		{
			name:  "only whitespace",
			input: "   ",
		},
		{
			name:  "only tabs",
			input: "\t\t",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matcher, err := ParseLine(tt.input)

			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if matcher != nil {
				t.Errorf("expected nil matcher, got %v", matcher)
			}
		})
	}
}

func TestSimpleMatcher_Interface(t *testing.T) {
	sm := &SimpleMatcher{
		name:    "test-package",
		version: "1.0.0",
	}

	// Verify it implements the Matcher interface
	var _ Matcher = sm

	if sm.Name() != "test-package" {
		t.Errorf("Name() = %q, want %q", sm.Name(), "test-package")
	}

	if sm.Version() != "1.0.0" {
		t.Errorf("Version() = %q, want %q", sm.Version(), "1.0.0")
	}

	if sm.AllVersions() {
		t.Error("AllVersions() = true, want false")
	}
}

func TestParseLines(t *testing.T) {
	lines := []string{
		"lodash 4.17.20",
		"",
		"   ",
		"@gsd-build/sdk",
		"@asyncapi/specs\t6.10.1",
		"broken ,",
	}

	matchers, failures := ParseLines(lines)
	if len(failures) != 1 {
		t.Fatalf("ParseLines() failures = %d, want 1: %v", len(failures), failures)
	}
	if len(matchers) != 3 {
		t.Fatalf("ParseLines() matchers = %d, want 3", len(matchers))
	}

	lineErr, ok := failures[0].(*LineError)
	if !ok {
		t.Fatalf("failure type = %T, want *LineError", failures[0])
	}
	if lineErr.Line != 6 {
		t.Errorf("failure line = %d, want 6", lineErr.Line)
	}

	if !matchers[1].AllVersions() || matchers[1].Name() != "@gsd-build/sdk" {
		t.Errorf("matcher 1 = %s (all=%v), want @gsd-build/sdk (all=true)",
			matchers[1].Name(), matchers[1].AllVersions())
	}

	tabSeparated := matchers[2]
	if tabSeparated.Name() != "@asyncapi/specs" || tabSeparated.Version() != "6.10.1" {
		t.Errorf("tab-separated matcher = %s@%s, want @asyncapi/specs@6.10.1",
			tabSeparated.Name(), tabSeparated.Version())
	}
}

// TestParseCompromisedList guards against silently dropping entries from the
// real list, including tab-separated lines and bare package names.
func TestParseCompromisedList(t *testing.T) {
	path := filepath.Join("..", "..", "compromised.txt")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}

	lines := strings.Split(strings.TrimSuffix(string(content), "\n"), "\n")
	matchers, failures := ParseLines(lines)
	if len(failures) > 0 {
		t.Fatalf("ParseLines(%s) returned %d errors, first: %v", path, len(failures), failures[0])
	}

	nonEmpty := 0
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			nonEmpty++
		}
	}
	if len(matchers) != nonEmpty {
		t.Errorf("parsed %d matchers from %d non-empty lines", len(matchers), nonEmpty)
	}

	allVersions := make(map[string]bool)
	for _, matcher := range matchers {
		if matcher.AllVersions() {
			allVersions[matcher.Name()] = true
		}
	}
	for _, name := range []string{"@gsd-build/sdk", "get-shit-done-cc", "get-shit-done-codex"} {
		if !allVersions[name] {
			t.Errorf("%s was not parsed as an all-versions entry", name)
		}
	}
}
