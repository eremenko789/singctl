package openapipipeline_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestGenerateFailsWhenConfigAbsent(t *testing.T) {
	root := repoRoot(t)
	cfg := filepath.Join(root, "api", "oapi-codegen.yaml")
	backup := cfg + ".bak-test"

	existed := true
	if _, err := os.Stat(cfg); err != nil {
		if !os.IsNotExist(err) {
			t.Fatalf("stat config: %v", err)
		}
		existed = false
	}

	if existed {
		if err := os.Rename(cfg, backup); err != nil {
			t.Fatalf("rename config aside: %v", err)
		}
		defer func() {
			if err := os.Rename(backup, cfg); err != nil {
				t.Errorf("restore config: %v", err)
			}
		}()
	}

	cmd := exec.Command("make", "generate")
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected make generate to fail without api/oapi-codegen.yaml; output:\n%s", out)
	}
	if len(out) == 0 {
		t.Fatal("expected non-empty error message from make generate")
	}
}
