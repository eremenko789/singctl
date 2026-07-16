package cli

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTaskKanbanDeleteSuccessAnd404(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodDelete && r.URL.Path == "/v2/kanban-task-status/KTS-del":
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodDelete && r.URL.Path == "/v2/kanban-task-status/KTS-missing":
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`missing`))
		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)

	withTaskConfig(t, srv.URL, "test-token-kts-del", func() {
		stdout, stderr, err := executeForTest([]string{"task", "kanban", "delete", "KTS-del"})
		if err != nil {
			t.Fatalf("err=%v stderr=%q", err, stderr)
		}
		if strings.TrimSpace(stdout) != "" {
			t.Fatalf("stdout=%q", stdout)
		}

		stdout, stderr, err = executeForTest([]string{"task", "kanban", "delete", "KTS-missing"})
		if err == nil || ExitCode(err) != 3 {
			t.Fatalf("ExitCode=%d stderr=%q", ExitCode(err), stderr)
		}
		if strings.TrimSpace(stdout) != "" {
			t.Fatalf("stdout=%q", stdout)
		}
	})
}

func TestTaskKanbanDeleteEmptyID(t *testing.T) {
	withTaskConfig(t, "http://127.0.0.1:9", "test-token-kts-del-id", func() {
		stdout, _, err := executeForTest([]string{"task", "kanban", "delete", "  "})
		if err == nil || ExitCode(err) != 1 {
			t.Fatalf("err=%v", err)
		}
		if strings.TrimSpace(stdout) != "" {
			t.Fatalf("stdout=%q", stdout)
		}
	})
}
