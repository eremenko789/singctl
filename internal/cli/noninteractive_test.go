package cli

import (
	"io"
	"strings"
	"testing"
	"time"
)

func TestNonInteractiveClosedStdinVersion(t *testing.T) {
	done := make(chan struct{})
	var stdout, stderr string
	var err error

	go func() {
		defer close(done)
		stdout, stderr, err = executeForTestWithIn([]string{"version"}, io.NopCloser(strings.NewReader("")))
	}()

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("version with closed stdin blocked (timeout)")
	}

	if err != nil {
		t.Fatalf("version error = %v stderr=%q stdout=%q", err, stderr, stdout)
	}
	if strings.TrimSpace(stdout) == "" {
		t.Fatal("expected version on stdout")
	}
	if strings.TrimSpace(stderr) != "" {
		t.Fatalf("stderr should be empty on success, got %q", stderr)
	}
}

func TestNonInteractivePipedStdinUnknownCommand(t *testing.T) {
	done := make(chan struct{})
	var stdout, stderr string
	var err error

	go func() {
		defer close(done)
		stdout, stderr, err = executeForTestWithIn(
			[]string{"nosuch"},
			strings.NewReader("line1\nline2\n"),
		)
	}()

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("unknown command with piped stdin blocked (timeout)")
	}

	if err == nil {
		t.Fatal("expected error")
	}
	if ExitCode(err) == 0 {
		t.Fatal("ExitCode must be non-zero")
	}
	if strings.TrimSpace(stdout) != "" {
		t.Errorf("stdout should be empty, got %q", stdout)
	}
	if strings.TrimSpace(stderr) == "" {
		t.Error("expected stderr message")
	}
}
