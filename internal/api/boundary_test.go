package api_test

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAPIPackageDoesNotImportCLI(t *testing.T) {
	t.Parallel()
	dir := "."
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	fset := token.NewFileSet()
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") {
			continue
		}
		path := filepath.Join(dir, name)
		f, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatalf("parse %s: %v", path, err)
		}
		for _, imp := range f.Imports {
			pathLit := strings.Trim(imp.Path.Value, `"`)
			if strings.Contains(pathLit, "/internal/cli") {
				t.Errorf("%s imports %s — adapter must not depend on CLI", name, pathLit)
			}
		}
	}
}
