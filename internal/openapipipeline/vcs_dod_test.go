package openapipipeline_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestDoDPathsNotGitIgnored(t *testing.T) {
	root := repoRoot(t)
	paths := []string{
		"api/oapi-codegen.yaml",
		"internal/apiclient/client.gen.go",
		"docs/api/openapi.json",
		"docs/api/openapi.yaml",
		"docs/api/coverage.md",
	}
	for _, rel := range paths {
		abs := filepath.Join(root, rel)
		if _, err := os.Stat(abs); err != nil {
			t.Errorf("DoD path missing on disk: %s (%v)", rel, err)
			continue
		}
		cmd := exec.Command("git", "check-ignore", "-q", rel)
		cmd.Dir = root
		err := cmd.Run()
		// exit 0 => ignored; exit 1 => not ignored
		if err == nil {
			t.Errorf("%s is ignored by git (must be trackable for F03 DoD)", rel)
			continue
		}
		if ee, ok := err.(*exec.ExitError); ok && ee.ExitCode() == 1 {
			continue // not ignored — good
		}
		t.Errorf("git check-ignore %s: %v", rel, err)
	}
}

func TestGenerateWorksOfflineWithoutFetch(t *testing.T) {
	root := repoRoot(t)
	if _, err := exec.LookPath("oapi-codegen"); err != nil {
		t.Skip("oapi-codegen not in PATH")
	}
	cfg := filepath.Join(root, "api", "oapi-codegen.yaml")
	yamlSnap := filepath.Join(root, "docs", "api", "openapi.yaml")
	for _, p := range []string{cfg, yamlSnap} {
		if _, err := os.Stat(p); err != nil {
			t.Fatalf("required offline input missing: %s: %v", p, err)
		}
	}

	cmd := exec.Command("make", "generate")
	cmd.Dir = root
	// Block accidental network use by the recipe (generate must not fetch).
	cmd.Env = append(os.Environ(),
		"OPENAPI_JSON_URL=http://127.0.0.1:1/blocked",
		"OPENAPI_YAML_URL=http://127.0.0.1:1/blocked",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("make generate should succeed offline from committed snapshot: %v\n%s", err, out)
	}
	if strings.Contains(string(out), "curl") && strings.Contains(string(out), "blocked") {
		t.Fatalf("generate appears to have attempted network fetch:\n%s", out)
	}
}
