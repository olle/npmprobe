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
	}{
		{
			name:    "simple package and version",
			input:   "lodash 4.17.21",
			expName: "lodash",
			expVer:  "4.17.21",
		},
		{
			name:    "scoped package",
			input:   "@babel/core 7.20.0",
			expName: "@babel/core",
			expVer:  "7.20.0",
		},
		{
			name:    "version with pre-release",
			input:   "react 18.0.0-alpha.1",
			expName: "react",
			expVer:  "18.0.0-alpha.1",
		},
		{
			name:    "leading/trailing whitespace",
			input:   "  express 4.18.2  ",
			expName: "express",
			expVer:  "4.18.2",
		},
		{
			name:    "extra fields ignored",
			input:   "typescript 4.9.4 extra stuff here",
			expName: "typescript",
			expVer:  "4.9.4",
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
