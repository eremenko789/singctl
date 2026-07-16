package openapipipeline_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	// tests live in internal/openapipipeline → two levels up
	root := filepath.Clean(filepath.Join(dir, "../.."))
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		t.Fatalf("repo root (go.mod) not found from %s: %v", dir, err)
	}
	return root
}

func TestClientGenGoExists(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "internal", "apiclient", "client.gen.go")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected generated client at %s: %v", path, err)
	}
}

func TestApiclientPackageBuilds(t *testing.T) {
	root := repoRoot(t)
	list := exec.Command("go", "list", "./internal/apiclient")
	list.Dir = root
	if out, err := list.CombinedOutput(); err != nil {
		t.Fatalf("go list ./internal/apiclient failed (package must exist with Go sources): %v\n%s", err, out)
	}
	cmd := exec.Command("go", "build", "./internal/apiclient/...")
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go build ./internal/apiclient/... failed: %v\n%s", err, out)
	}
}
