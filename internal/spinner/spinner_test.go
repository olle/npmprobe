package spinner

import (
	"io"
	"os"
	"strings"
	"testing"
	"time"
)

func TestStopClearsSpinnerLine(t *testing.T) {
	oldStderr := os.Stderr
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("create pipe: %v", err)
	}
	os.Stderr = writer
	defer func() { os.Stderr = oldStderr }()

	s := NewSpinner()
	s.delay = time.Millisecond
	s.Start("loading")
	s.Stop()

	if err := writer.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}
	out, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("read spinner output: %v", err)
	}
	os.Stderr = oldStderr

	got := string(out)
	if strings.Contains(got, "\n") {
		t.Errorf("spinner output contains a newline: %q", got)
	}
	if !strings.HasSuffix(got, "\r") {
		t.Errorf("spinner output does not leave the cursor at the start: %q", got)
	}
}
