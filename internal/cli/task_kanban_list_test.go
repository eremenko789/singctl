package cli

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func sampleCLIKanbanLinkJSON(id, taskID, statusID string, order float32) []byte {
	m := map[string]any{
		"id": id, "taskId": taskID, "statusId": statusID, "kanbanOrder": order,
		"removed": false, "modificatedDate": "1584530599718", "modificated": map[string]any{},
	}
	b, _ := json.Marshal(m)
	return b
}

func TestTaskKanbanListHappyPathEmptyAndNoGetTask(t *testing.T) {
	var taskHits, listHits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.HasPrefix(r.URL.Path, "/v2/task/") {
			taskHits.Add(1)
		}
		if r.Method == http.MethodGet && r.URL.Path == "/v2/kanban-task-status" {
			listHits.Add(1)
			if r.URL.Query().Get("taskId") != "T-1" {
				t.Errorf("taskId=%q", r.URL.Query().Get("taskId"))
			}
			if r.URL.Query().Get("maxCount") != "" {
				t.Errorf("unexpected pagination %q", r.URL.RawQuery)
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"kanbanTaskStatuses":[` + string(sampleCLIKanbanLinkJSON("KTS-1", "T-1", "KS-1", 1)) + `]}`))
			return
		}
		t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)

	withTaskConfig(t, srv.URL, "test-token-kts-list", func() {
		stdout, stderr, err := executeForTest([]string{"task", "kanban", "list", "--task", "T-1", "-o", "json"})
		if err != nil {
			t.Fatalf("err=%v stderr=%q", err, stderr)
		}
		if strings.TrimSpace(stderr) != "" {
			t.Fatalf("stderr=%q", stderr)
		}
		if taskHits.Load() != 0 {
			t.Fatalf("list must not call GetTask, hits=%d", taskHits.Load())
		}
		if listHits.Load() != 1 {
			t.Fatalf("listHits=%d", listHits.Load())
		}
		var arr []map[string]any
		if err := json.Unmarshal([]byte(stdout), &arr); err != nil {
			t.Fatalf("json: %v\n%s", err, stdout)
		}
		if len(arr) != 1 || arr[0]["id"] != "KTS-1" {
			t.Fatalf("arr=%v", arr)
		}
	})

	emptySrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"kanbanTaskStatuses":[]}`))
	}))
	t.Cleanup(emptySrv.Close)
	withTaskConfig(t, emptySrv.URL, "test-token-kts-empty", func() {
		stdout, _, err := executeForTest([]string{"task", "kanban", "list", "-o", "json"})
		if err != nil {
			t.Fatalf("err=%v", err)
		}
		var arr []any
		if err := json.Unmarshal([]byte(stdout), &arr); err != nil || len(arr) != 0 {
			t.Fatalf("want [], got %q", stdout)
		}
	})
}

func TestTaskKanbanListAuth(t *testing.T) {
	withCLIPaths(t, func(paths testPaths) {
		stdout, stderr, err := executeForTest([]string{"task", "kanban", "list", "-o", "json"})
		if err == nil {
			t.Fatal("expected error")
		}
		if ExitCode(err) != 2 {
			t.Fatalf("ExitCode=%d stderr=%q stdout=%q", ExitCode(err), stderr, stdout)
		}
	})
}
