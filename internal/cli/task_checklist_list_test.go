package cli

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func sampleCLIChecklistJSON(id, title, parent string, done bool) []byte {
	m := map[string]any{
		"id": id, "title": title, "parent": parent, "parentOrder": 1,
		"done": done, "removed": false, "modificatedDate": "1584530599718",
		"crypted": "", "modificated": map[string]any{},
	}
	b, _ := json.Marshal(m)
	return b
}

func TestTaskChecklistListHappyPathAndEmpty(t *testing.T) {
	var taskHits, listHits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v2/task/T-1":
			taskHits.Add(1)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(sampleCLITaskJSON("T-1", "Task", nil))
		case r.Method == http.MethodGet && r.URL.Path == "/v2/checklist-item":
			listHits.Add(1)
			if r.URL.Query().Get("parent") != "T-1" {
				t.Errorf("parent=%q", r.URL.Query().Get("parent"))
			}
			if r.URL.Query().Get("maxCount") != "" || r.URL.Query().Get("offset") != "" {
				t.Errorf("unexpected pagination query %q", r.URL.RawQuery)
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"checklistItems":[` + string(sampleCLIChecklistJSON("C-1", "One", "T-1", false)) + `]}`))
		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)

	withTaskConfig(t, srv.URL, "test-token-cl-list", func() {
		stdout, stderr, err := executeForTest([]string{"task", "checklist", "list", "T-1", "-o", "json"})
		if err != nil {
			t.Fatalf("err=%v stderr=%q", err, stderr)
		}
		if strings.TrimSpace(stderr) != "" {
			t.Fatalf("stderr=%q", stderr)
		}
		if taskHits.Load() != 1 || listHits.Load() != 1 {
			t.Fatalf("hits task=%d list=%d", taskHits.Load(), listHits.Load())
		}
		var arr []map[string]any
		if err := json.Unmarshal([]byte(stdout), &arr); err != nil {
			t.Fatalf("json: %v\n%s", err, stdout)
		}
		if len(arr) != 1 || arr[0]["id"] != "C-1" {
			t.Fatalf("arr=%v", arr)
		}
		// task body must not appear as root object
		if _, ok := arr[0]["projectId"]; ok {
			t.Fatalf("stdout looks like task object leaked: %v", arr[0])
		}
	})

	emptySrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/v2/task/"):
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(sampleCLITaskJSON("T-empty", "Empty", nil))
		case r.Method == http.MethodGet && r.URL.Path == "/v2/checklist-item":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"checklistItems":[]}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(emptySrv.Close)
	withTaskConfig(t, emptySrv.URL, "test-token-cl-empty", func() {
		stdout, _, err := executeForTest([]string{"task", "checklist", "list", "T-empty", "-o", "json"})
		if err != nil {
			t.Fatalf("err=%v", err)
		}
		var arr []any
		if err := json.Unmarshal([]byte(stdout), &arr); err != nil || len(arr) != 0 {
			t.Fatalf("want [], got %q", stdout)
		}
	})
}

func TestTaskChecklistListTaskNotFoundNoChecklistCall(t *testing.T) {
	var listHits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v2/checklist-item" {
			listHits.Add(1)
		}
		if r.Method == http.MethodGet && r.URL.Path == "/v2/task/T-missing" {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`missing`))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	t.Cleanup(srv.Close)

	withTaskConfig(t, srv.URL, "test-token-cl-404-task", func() {
		stdout, stderr, err := executeForTest([]string{"task", "checklist", "list", "T-missing", "-o", "json"})
		if err == nil {
			t.Fatal("expected error")
		}
		if ExitCode(err) != 3 {
			t.Fatalf("ExitCode=%d stderr=%q", ExitCode(err), stderr)
		}
		if strings.TrimSpace(stdout) != "" {
			t.Fatalf("stdout=%q", stdout)
		}
		if listHits.Load() != 0 {
			t.Fatalf("checklist list must not be called, hits=%d", listHits.Load())
		}
	})
}

func TestTaskChecklistListAuthAndEmptyID(t *testing.T) {
	withCLIPaths(t, func(paths testPaths) {
		stdout, stderr, err := executeForTest([]string{"task", "checklist", "list", "T-1", "-o", "json"})
		if err == nil {
			t.Fatal("expected error")
		}
		if ExitCode(err) != 2 {
			t.Fatalf("ExitCode=%d stderr=%q stdout=%q", ExitCode(err), stderr, stdout)
		}
	})

	withTaskConfig(t, "http://127.0.0.1:9", "test-token-cl-empty-id", func() {
		stdout, _, err := executeForTest([]string{"task", "checklist", "list", "   "})
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
