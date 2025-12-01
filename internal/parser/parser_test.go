package parser

import (
	"testing"
)

func TestParseLine_ValidInput(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		expName    string
		expVersion string
		expVers    []string
	}{
		{
			name:       "simple package and version",
			input:      "lodash 4.17.21",
			expName:    "lodash",
			expVersion: "4.17.21",
			expVers:    []string{"4.17.21"},
		},
		{
			name:       "scoped package",
			input:      "@babel/core 7.20.0",
			expName:    "@babel/core",
			expVersion: "7.20.0",
			expVers:    []string{"7.20.0"},
		},
		{
			name:       "version with pre-release",
			input:      "react 18.0.0-alpha.1",
			expName:    "react",
			expVersion: "18.0.0-alpha.1",
			expVers:    []string{"18.0.0-alpha.1"},
		},
		{
			name:       "leading/trailing whitespace",
			input:      "  express 4.18.2  ",
			expName:    "express",
			expVersion: "4.18.2",
			expVers:    []string{"4.18.2"},
		},
		{
			name:       "extra fields become part of version",
			input:      "typescript 4.9.4 extra stuff",
			expName:    "typescript",
			expVersion: "4.9.4 extra stuff",
			expVers:    []string{"4.9.4 extra stuff"},
		},
		{
			name:       "multiple comma-separated versions",
			input:      "lodash 4.17.20,4.17.21",
			expName:    "lodash",
			expVersion: "4.17.20",
			expVers:    []string{"4.17.20", "4.17.21"},
		},
		{
			name:       "multiple versions with scoped package",
			input:      "@babel/core 7.18.0,7.19.0,7.20.0",
			expName:    "@babel/core",
			expVersion: "7.18.0",
			expVers:    []string{"7.18.0", "7.19.0", "7.20.0"},
		},
		{
			name:       "versions with spaces around commas",
			input:      "react 16.0.0 , 17.0.0 , 18.0.0",
			expName:    "react",
			expVersion: "16.0.0",
			expVers:    []string{"16.0.0", "17.0.0", "18.0.0"},
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

			if got := matcher.Version(); got != tt.expVersion {
				t.Errorf("Version() = %q, want %q", got, tt.expVersion)
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

	if got := sm.Versions(); len(got) != 1 || got[0] != "1.0.0" {
		t.Errorf("Versions() = %v, want [1.0.0]", got)
	}
}

func TestMultiVersionMatcher_Interface(t *testing.T) {
	mvm := &MultiVersionMatcher{
		name:     "test-package",
		versions: []string{"1.0.0", "2.0.0", "3.0.0"},
	}

	// Verify it implements the Matcher interface
	var _ Matcher = mvm

	if mvm.Name() != "test-package" {
		t.Errorf("Name() = %q, want %q", mvm.Name(), "test-package")
	}

	if mvm.Version() != "1.0.0" {
		t.Errorf("Version() = %q, want %q (first version)", mvm.Version(), "1.0.0")
	}

	if got := mvm.Versions(); len(got) != 3 {
		t.Errorf("Versions() length = %d, want 3", len(got))
	} else {
		expected := []string{"1.0.0", "2.0.0", "3.0.0"}
		for i, v := range got {
			if v != expected[i] {
				t.Errorf("Versions()[%d] = %q, want %q", i, v, expected[i])
			}
		}
	}
}
