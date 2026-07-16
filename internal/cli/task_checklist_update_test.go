package cli

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTaskChecklistUpdateHappyAndValidation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPatch && r.URL.Path == "/v2/checklist-item/C-1":
			body, _ := io.ReadAll(r.Body)
			var m map[string]any
			_ = json.Unmarshal(body, &m)
			done, _ := m["done"].(bool)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(sampleCLIChecklistJSON("C-1", "Item", "T-1", done))
		case r.Method == http.MethodPatch && r.URL.Path == "/v2/checklist-item/C-missing":
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`missing`))
		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)

	withTaskConfig(t, srv.URL, "test-token-cl-upd", func() {
		stdout, stderr, err := executeForTest([]string{"task", "checklist", "update", "C-1", "--done", "-o", "json"})
		if err != nil {
			t.Fatalf("err=%v stderr=%q", err, stderr)
		}
		var obj map[string]any
		if err := json.Unmarshal([]byte(stdout), &obj); err != nil {
			t.Fatalf("object: %v\n%s", err, stdout)
		}
		if obj["done"] != true {
			t.Fatalf("obj=%v", obj)
		}

		stdout, _, err = executeForTest([]string{"task", "checklist", "update", "C-1"})
		if err == nil {
			t.Fatal("expected no flags error")
		}
		if ExitCode(err) != 1 {
			t.Fatalf("ExitCode=%d", ExitCode(err))
		}
		if strings.TrimSpace(stdout) != "" {
			t.Fatalf("stdout=%q", stdout)
		}

		_, _, err = executeForTest([]string{"task", "checklist", "update", "C-1", "--done", "--undone"})
		if err == nil {
			t.Fatal("expected mutually exclusive error")
		}
		if ExitCode(err) != 1 {
			t.Fatalf("ExitCode=%d", ExitCode(err))
		}

		_, _, err = executeForTest([]string{"task", "checklist", "update", "C-1", "--title", "  "})
		if err == nil {
			t.Fatal("expected empty title error")
		}
		if ExitCode(err) != 1 {
			t.Fatalf("ExitCode=%d", ExitCode(err))
		}

		stdout, stderr, err = executeForTest([]string{"task", "checklist", "update", "C-missing", "--done"})
		if err == nil {
			t.Fatal("expected 404")
		}
		if ExitCode(err) != 3 {
			t.Fatalf("ExitCode=%d stderr=%q", ExitCode(err), stderr)
		}
		if strings.TrimSpace(stdout) != "" {
			t.Fatalf("stdout=%q", stdout)
		}
	})
}
