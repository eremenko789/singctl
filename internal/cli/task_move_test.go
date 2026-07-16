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

func TestTaskMoveCreateUpdateAmbiguousAndValidation(t *testing.T) {
	t.Run("zero_create", func(t *testing.T) {
		var taskHits, listHits, createHits atomic.Int32
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			switch {
			case r.Method == http.MethodGet && r.URL.Path == "/v2/task/T-0":
				taskHits.Add(1)
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(sampleCLITaskJSON("T-0", "Task", nil))
			case r.Method == http.MethodGet && r.URL.Path == "/v2/kanban-task-status":
				listHits.Add(1)
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"kanbanTaskStatuses":[]}`))
			case r.Method == http.MethodPost && r.URL.Path == "/v2/kanban-task-status":
				createHits.Add(1)
				body, _ := io.ReadAll(r.Body)
				var m map[string]any
				_ = json.Unmarshal(body, &m)
				if _, hasOrder := m["kanbanOrder"]; hasOrder {
					t.Errorf("create must omit order: %v", m)
				}
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(sampleCLIKanbanLinkJSON("KTS-c", "T-0", "KS-1", 0))
			default:
				t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
				w.WriteHeader(http.StatusNotFound)
			}
		}))
		t.Cleanup(srv.Close)
		withTaskConfig(t, srv.URL, "test-token-move0", func() {
			stdout, stderr, err := executeForTest([]string{"task", "move", "T-0", "--column", "KS-1", "-o", "json"})
			if err != nil {
				t.Fatalf("err=%v stderr=%q", err, stderr)
			}
			if taskHits.Load() != 1 || listHits.Load() != 1 || createHits.Load() != 1 {
				t.Fatalf("hits task=%d list=%d create=%d", taskHits.Load(), listHits.Load(), createHits.Load())
			}
			var obj map[string]any
			if err := json.Unmarshal([]byte(stdout), &obj); err != nil || obj["id"] != "KTS-c" {
				t.Fatalf("stdout=%q", stdout)
			}
			// list wrapper must not be root
			if _, ok := obj["kanbanTaskStatuses"]; ok {
				t.Fatalf("list leaked: %v", obj)
			}
		})
	})

	t.Run("one_update_same_column", func(t *testing.T) {
		var patched atomic.Int32
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			switch {
			case r.Method == http.MethodGet && r.URL.Path == "/v2/task/T-1":
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(sampleCLITaskJSON("T-1", "Task", nil))
			case r.Method == http.MethodGet && r.URL.Path == "/v2/kanban-task-status":
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"kanbanTaskStatuses":[` + string(sampleCLIKanbanLinkJSON("KTS-1", "T-1", "KS-same", 1)) + `]}`))
			case r.Method == http.MethodPatch && r.URL.Path == "/v2/kanban-task-status/KTS-1":
				patched.Add(1)
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(sampleCLIKanbanLinkJSON("KTS-1", "T-1", "KS-same", 1))
			default:
				t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
				w.WriteHeader(http.StatusNotFound)
			}
		}))
		t.Cleanup(srv.Close)
		withTaskConfig(t, srv.URL, "test-token-move1", func() {
			stdout, _, err := executeForTest([]string{"task", "move", "T-1", "--column", "KS-same", "-o", "json"})
			if err != nil {
				t.Fatalf("err=%v", err)
			}
			if patched.Load() != 1 {
				t.Fatalf("expected PATCH, hits=%d", patched.Load())
			}
			var obj map[string]any
			if err := json.Unmarshal([]byte(stdout), &obj); err != nil || obj["id"] != "KTS-1" {
				t.Fatalf("stdout=%q", stdout)
			}
		})
	})

	t.Run("ambiguous_and_errors", func(t *testing.T) {
		var wrote atomic.Int32
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodPost || r.Method == http.MethodPatch {
				wrote.Add(1)
			}
			w.Header().Set("Content-Type", "application/json")
			switch {
			case r.Method == http.MethodGet && r.URL.Path == "/v2/task/T-m":
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(sampleCLITaskJSON("T-m", "Task", nil))
			case r.Method == http.MethodGet && r.URL.Path == "/v2/task/T-missing":
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte(`missing`))
			case r.Method == http.MethodGet && r.URL.Path == "/v2/kanban-task-status":
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"kanbanTaskStatuses":[` +
					string(sampleCLIKanbanLinkJSON("KTS-a", "T-m", "KS-1", 1)) + `,` +
					string(sampleCLIKanbanLinkJSON("KTS-b", "T-m", "KS-2", 2)) + `]}`))
			default:
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{}`))
			}
		}))
		t.Cleanup(srv.Close)
		withTaskConfig(t, srv.URL, "test-token-moveN", func() {
			stdout, stderr, err := executeForTest([]string{"task", "move", "T-m", "--column", "KS-3"})
			if err == nil || ExitCode(err) != 1 {
				t.Fatalf("ambiguous ExitCode=%d stderr=%q", ExitCode(err), stderr)
			}
			if strings.TrimSpace(stdout) != "" {
				t.Fatalf("stdout=%q", stdout)
			}
			if !strings.Contains(stderr, "kanban") {
				t.Fatalf("stderr should hint kanban: %q", stderr)
			}
			if wrote.Load() != 0 {
				t.Fatalf("must not write, hits=%d", wrote.Load())
			}

			stdout, stderr, err = executeForTest([]string{"task", "move", "T-missing", "--column", "KS-1"})
			if err == nil || ExitCode(err) != 3 {
				t.Fatalf("task 404 ExitCode=%d stderr=%q", ExitCode(err), stderr)
			}
			if strings.TrimSpace(stdout) != "" {
				t.Fatalf("stdout=%q", stdout)
			}

			stdout, _, err = executeForTest([]string{"task", "move", "T-1"})
			if err == nil || ExitCode(err) != 1 {
				t.Fatalf("missing column: %v", err)
			}

			stdout, _, err = executeForTest([]string{"task", "move", "T-1", "--column", "KS-1", "--order", "1"})
			if err == nil || ExitCode(err) != 1 {
				t.Fatalf("unexpected --order should fail: %v stdout=%q", err, stdout)
			}
		})
	})
}
