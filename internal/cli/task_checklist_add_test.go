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

func TestTaskChecklistAddHappyPath(t *testing.T) {
	var createHits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v2/task/T-1":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(sampleCLITaskJSON("T-1", "Task", nil))
		case r.Method == http.MethodPost && r.URL.Path == "/v2/checklist-item":
			createHits.Add(1)
			body, _ := io.ReadAll(r.Body)
			var m map[string]any
			_ = json.Unmarshal(body, &m)
			if m["parent"] != "T-1" || m["title"] != "Buy milk" {
				t.Errorf("body=%v", m)
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(sampleCLIChecklistJSON("C-new", "Buy milk", "T-1", false))
		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)

	withTaskConfig(t, srv.URL, "test-token-cl-add", func() {
		stdout, stderr, err := executeForTest([]string{
			"task", "checklist", "add", "T-1", "--title", "Buy milk", "-o", "json",
		})
		if err != nil {
			t.Fatalf("err=%v stderr=%q", err, stderr)
		}
		if createHits.Load() != 1 {
			t.Fatalf("createHits=%d", createHits.Load())
		}
		var obj map[string]any
		if err := json.Unmarshal([]byte(stdout), &obj); err != nil {
			t.Fatalf("object: %v\n%s", err, stdout)
		}
		if obj["id"] != "C-new" {
			t.Fatalf("obj=%v", obj)
		}
	})
}

func TestTaskChecklistAddValidationAndTask404(t *testing.T) {
	var createHits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/v2/checklist-item" {
			createHits.Add(1)
		}
		if r.Method == http.MethodGet && r.URL.Path == "/v2/task/T-missing" {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`missing`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(sampleCLITaskJSON("T-1", "Task", nil))
	}))
	t.Cleanup(srv.Close)

	withTaskConfig(t, srv.URL, "test-token-cl-add-val", func() {
		stdout, _, err := executeForTest([]string{"task", "checklist", "add", "T-1"})
		if err == nil {
			t.Fatal("expected missing title error")
		}
		if ExitCode(err) != 1 {
			t.Fatalf("ExitCode=%d", ExitCode(err))
		}
		if strings.TrimSpace(stdout) != "" {
			t.Fatalf("stdout=%q", stdout)
		}

		stdout, _, err = executeForTest([]string{"task", "checklist", "add", "T-1", "--title", "   "})
		if err == nil {
			t.Fatal("expected empty title error")
		}
		if ExitCode(err) != 1 {
			t.Fatalf("ExitCode=%d", ExitCode(err))
		}

		stdout, stderr, err := executeForTest([]string{"task", "checklist", "add", "T-missing", "--title", "X"})
		if err == nil {
			t.Fatal("expected not found")
		}
		if ExitCode(err) != 3 {
			t.Fatalf("ExitCode=%d stderr=%q", ExitCode(err), stderr)
		}
		if strings.TrimSpace(stdout) != "" {
			t.Fatalf("stdout=%q", stdout)
		}
		if createHits.Load() != 0 {
			t.Fatalf("create must not be called, hits=%d", createHits.Load())
		}
	})
}
