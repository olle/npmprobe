package parser

import (
	"testing"
)

func TestParseLine_ValidInput(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		expName string
		expVer  string
		expVers []string
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
			name:  "single field",
			input: "lodash",
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
}
