package cli

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTaskChecklistGetJSONAnd404(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v2/checklist-item/C-1":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(sampleCLIChecklistJSON("C-1", "Item", "T-1", false))
		case r.Method == http.MethodGet && r.URL.Path == "/v2/checklist-item/C-missing":
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`missing`))
		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)

	withTaskConfig(t, srv.URL, "test-token-cl-get", func() {
		stdout, stderr, err := executeForTest([]string{"task", "checklist", "get", "C-1", "-o", "json"})
		if err != nil {
			t.Fatalf("err=%v stderr=%q", err, stderr)
		}
		var obj map[string]any
		if err := json.Unmarshal([]byte(stdout), &obj); err != nil {
			t.Fatalf("want object: %v\n%s", err, stdout)
		}
		if obj["id"] != "C-1" {
			t.Fatalf("obj=%v", obj)
		}
		if _, isArr := obj["0"]; isArr {
			t.Fatal("must be object not array")
		}
	})

	withTaskConfig(t, srv.URL, "test-token-cl-get", func() {
		stdout, stderr, err := executeForTest([]string{"task", "checklist", "get", "C-missing"})
		if err == nil {
			t.Fatal("expected error")
		}
		if ExitCode(err) != 3 {
			t.Fatalf("ExitCode=%d stderr=%q", ExitCode(err), stderr)
		}
		if strings.TrimSpace(stdout) != "" {
			t.Fatalf("stdout=%q", stdout)
		}
	})
}

func TestTaskChecklistGetEmptyID(t *testing.T) {
	withTaskConfig(t, "http://127.0.0.1:9", "test-token-cl-get-id", func() {
		stdout, _, err := executeForTest([]string{"task", "checklist", "get", "  "})
		if err == nil {
			t.Fatal("expected error")
		}
		if ExitCode(err) != 1 {
			t.Fatalf("ExitCode=%d", ExitCode(err))
		}
		if strings.TrimSpace(stdout) != "" {
			t.Fatalf("stdout=%q", stdout)
		}
	})
}
