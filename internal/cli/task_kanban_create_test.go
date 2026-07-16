package cli

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func TestTaskKanbanCreateHappyPath(t *testing.T) {
	var taskHits, createHits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v2/task/T-1":
			taskHits.Add(1)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(sampleCLITaskJSON("T-1", "Task", nil))
		case r.Method == http.MethodPost && r.URL.Path == "/v2/kanban-task-status":
			createHits.Add(1)
			body, _ := io.ReadAll(r.Body)
			var m map[string]any
			_ = json.Unmarshal(body, &m)
			if m["taskId"] != "T-1" || m["statusId"] != "KS-1" {
				t.Errorf("body=%v", m)
			}
			if m["kanbanOrder"] != float64(2) {
				t.Errorf("order=%v", m["kanbanOrder"])
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(sampleCLIKanbanLinkJSON("KTS-new", "T-1", "KS-1", 2))
		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)

	withTaskConfig(t, srv.URL, "test-token-kts-create", func() {
		stdout, stderr, err := executeForTest([]string{
			"task", "kanban", "create", "--task", "T-1", "--column", "KS-1", "--order", "2", "-o", "json",
		})
		if err != nil {
			t.Fatalf("err=%v stderr=%q", err, stderr)
		}
		if taskHits.Load() != 1 || createHits.Load() != 1 {
			t.Fatalf("hits task=%d create=%d", taskHits.Load(), createHits.Load())
		}
		var obj map[string]any
		if err := json.Unmarshal([]byte(stdout), &obj); err != nil || obj["id"] != "KTS-new" {
			t.Fatalf("stdout=%q err=%v", stdout, err)
		}
	})
}

func TestTaskKanbanCreateValidationAndTask404(t *testing.T) {
	var createHits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			createHits.Add(1)
		}
		if r.Method == http.MethodGet && r.URL.Path == "/v2/task/T-missing" {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`missing`))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(sampleCLITaskJSON("T-1", "Task", nil))
	}))
	t.Cleanup(srv.Close)

	withTaskConfig(t, srv.URL, "test-token-kts-create-val", func() {
		stdout, _, err := executeForTest([]string{"task", "kanban", "create", "--task", "T-1"})
		if err == nil || ExitCode(err) != 1 {
			t.Fatalf("missing column: err=%v", err)
		}
		if strings.TrimSpace(stdout) != "" {
			t.Fatalf("stdout=%q", stdout)
		}

		stdout, _, err = executeForTest([]string{"task", "kanban", "create", "--task", "T-1", "--column", "KS-1", "--order", "-1"})
		if err == nil || ExitCode(err) != 1 {
			t.Fatalf("neg order: err=%v", err)
		}

		stdout, stderr, err := executeForTest([]string{"task", "kanban", "create", "--task", "T-missing", "--column", "KS-1"})
		if err == nil || ExitCode(err) != 3 {
			t.Fatalf("ExitCode=%d stderr=%q", ExitCode(err), stderr)
		}
		if strings.TrimSpace(stdout) != "" {
			t.Fatalf("stdout=%q", stdout)
		}
		if createHits.Load() != 0 {
			t.Fatalf("create must not run, hits=%d", createHits.Load())
		}
	})
}
