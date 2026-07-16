package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found walking from cwd")
		}
		dir = parent
	}
}

func TestScriptabilityDocsExitTable(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "docs", "scriptability.md")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	body := string(data)
	for _, code := range []string{"0", "1", "2", "3"} {
		if !strings.Contains(body, "`"+code+"`") && !strings.Contains(body, "| `"+code+"`") {
			// table uses `0` style
			if !strings.Contains(body, code) {
				t.Errorf("docs/scriptability.md must mention exit code %s", code)
			}
		}
	}
	for _, meaning := range []string{"Успех", "конфигурац", "не найден"} {
		if !strings.Contains(body, meaning) {
			t.Errorf("docs/scriptability.md missing meaning fragment %q", meaning)
		}
	}
	if !strings.Contains(body, "API") && !strings.Contains(body, "использования") {
		t.Error("docs/scriptability.md must describe code 1 (API/usage)")
	}
}

func TestScriptabilityDocsPipeScenarioIDs(t *testing.T) {
	root := repoRoot(t)
	docsPath := filepath.Join(root, "docs", "scriptability.md")
	docs, err := os.ReadFile(docsPath)
	if err != nil {
		t.Fatalf("read docs: %v", err)
	}
	docsBody := string(docs)

	contractPath := filepath.Join(root, "specs", "007-scriptability-exits", "contracts", "pipe-scenarios.md")
	contract, err := os.ReadFile(contractPath)
	if err != nil {
		t.Fatalf("read contract: %v", err)
	}
	contractBody := string(contract)

	ids := []string{"json-redirect", "list-jq-xargs", "csv-awk", "xargs-create"}
	for _, id := range ids {
		if !strings.Contains(docsBody, id) {
			t.Errorf("docs/scriptability.md missing pipe scenario id %q", id)
		}
		if !strings.Contains(contractBody, id) {
			t.Errorf("pipe-scenarios.md missing id %q", id)
		}
		idx := strings.Index(contractBody, "`"+id+"`")
		if idx < 0 {
			continue
		}
		window := contractBody[idx:]
		if len(window) > 800 {
			window = window[:800]
		}
		if !strings.Contains(window, "Status") {
			t.Errorf("pipe-scenarios.md scenario %q missing Status nearby", id)
		}
	}
}
