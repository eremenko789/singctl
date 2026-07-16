package cli

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTaskKanbanUpdateHappyAndValidation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPatch && r.URL.Path == "/v2/kanban-task-status/KTS-1":
			body, _ := io.ReadAll(r.Body)
			var m map[string]any
			_ = json.Unmarshal(body, &m)
			if m["statusId"] != "KS-2" {
				t.Errorf("body=%v", m)
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(sampleCLIKanbanLinkJSON("KTS-1", "T-1", "KS-2", 1))
		case r.Method == http.MethodPatch && r.URL.Path == "/v2/kanban-task-status/KTS-missing":
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`missing`))
		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)

	withTaskConfig(t, srv.URL, "test-token-kts-upd", func() {
		stdout, _, err := executeForTest([]string{"task", "kanban", "update", "KTS-1"})
		if err == nil || ExitCode(err) != 1 {
			t.Fatalf("no flags: err=%v", err)
		}

		stdout, stderr, err := executeForTest([]string{"task", "kanban", "update", "KTS-1", "--column", "KS-2", "-o", "json"})
		if err != nil {
			t.Fatalf("err=%v stderr=%q", err, stderr)
		}
		var obj map[string]any
		if err := json.Unmarshal([]byte(stdout), &obj); err != nil || obj["statusId"] != "KS-2" {
			t.Fatalf("stdout=%q", stdout)
		}

		stdout, stderr, err = executeForTest([]string{"task", "kanban", "update", "KTS-missing", "--column", "KS-2"})
		if err == nil || ExitCode(err) != 3 {
			t.Fatalf("ExitCode=%d stderr=%q", ExitCode(err), stderr)
		}
		if strings.TrimSpace(stdout) != "" {
			t.Fatalf("stdout=%q", stdout)
		}
	})
}
