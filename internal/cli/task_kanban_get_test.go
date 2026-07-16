package cli

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTaskKanbanGetJSONAnd404(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v2/kanban-task-status/KTS-1":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(sampleCLIKanbanLinkJSON("KTS-1", "T-1", "KS-1", 2))
		case r.Method == http.MethodGet && r.URL.Path == "/v2/kanban-task-status/KTS-missing":
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`missing`))
		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)

	withTaskConfig(t, srv.URL, "test-token-kts-get", func() {
		stdout, stderr, err := executeForTest([]string{"task", "kanban", "get", "KTS-1", "-o", "json"})
		if err != nil {
			t.Fatalf("err=%v stderr=%q", err, stderr)
		}
		var obj map[string]any
		if err := json.Unmarshal([]byte(stdout), &obj); err != nil {
			t.Fatalf("want object: %v\n%s", err, stdout)
		}
		if obj["id"] != "KTS-1" {
			t.Fatalf("obj=%v", obj)
		}
	})

	withTaskConfig(t, srv.URL, "test-token-kts-get", func() {
		stdout, stderr, err := executeForTest([]string{"task", "kanban", "get", "KTS-missing", "-o", "json"})
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

func TestTaskKanbanGetEmptyID(t *testing.T) {
	withTaskConfig(t, "http://127.0.0.1:9", "test-token-kts-get-id", func() {
		stdout, _, err := executeForTest([]string{"task", "kanban", "get", "   "})
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
