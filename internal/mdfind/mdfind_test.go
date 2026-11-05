package mdfind

import (
	"context"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestExtractFilenamesFromQuery(t *testing.T) {
	query := `kMDItemFSName == "package.json" || kMDItemFSName == "yarn.lock" || kMDItemFSName == "bun.lock"`

	got := extractFilenamesFromQuery(query)
	want := []string{"package.json", "yarn.lock", "bun.lock"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("extractFilenamesFromQuery() = %v, want %v", got, want)
	}
}

func TestExtractFilenamesFromQueryWithoutNames(t *testing.T) {
	if got := extractFilenamesFromQuery("kMDItemFSName == null"); len(got) != 0 {
		t.Errorf("extractFilenamesFromQuery() = %v, want no filenames", got)
	}
}

func TestRunSearchCommandReportsFailure(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err := runSearchCommand(ctx, "npmprobe-command-that-does-not-exist")
	if err == nil {
		t.Fatal("error = nil, want a command failure")
	}
	if !strings.Contains(err.Error(), "failed") {
		t.Errorf("error = %v, want it to mention the command failure", err)
	}
}

func TestRunSearchCommandTimesOut(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("sleep is not available on Windows")
	}

	oldTimeout := searchTimeout
	searchTimeout = 20 * time.Millisecond
	defer func() { searchTimeout = oldTimeout }()

	ctx, cancel := context.WithTimeout(context.Background(), searchTimeout)
	defer cancel()

	_, err := runSearchCommand(ctx, "sleep", "5")
	if err == nil {
		t.Fatal("error = nil, want a timeout")
	}
	if !strings.Contains(err.Error(), "timed out") {
		t.Errorf("error = %v, want it to mention the timeout", err)
	}
}
