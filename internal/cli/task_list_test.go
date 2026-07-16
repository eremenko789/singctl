package cli

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	cfgpkg "github.com/eremenko789/singctl/internal/config"
)

func sampleCLITaskJSON(id, title string, extra map[string]any) []byte {
	m := map[string]any{
		"id": id, "title": title, "projectId": "P-1", "parent": "",
		"priority": 1, "start": "2026-07-01", "journalDate": "", "deleteDate": "",
		"isNote": false, "alarmNotify": false, "checked": 0, "complete": 0,
		"completeLast": "", "createdDate": "2026-07-01", "crypted": "",
		"deadline": "", "deadlineNotifyReaded": false, "deferred": false,
		"externalId": "", "group": "", "integrationItemId": "",
		"modificated": map[string]any{}, "modificatedDate": "2026-07-01",
		"note": "", "notifies": []any{}, "notify": 0, "parentOrder": 0,
		"pomodoroCount": 0, "pomodoroTotalTime": 0, "recurrenceGeneratorId": "",
		"removed": false, "scheduleOrder": 0, "seenToday": "", "showInBasket": false,
		"startNotifiesReaded": []any{}, "startNotifyReaded": false, "state": 1,
		"tags": []any{}, "timeLength": 0, "useTime": false,
	}
	for k, v := range extra {
		m[k] = v
	}
	b, _ := json.Marshal(m)
	return b
}

func withTaskConfig(t *testing.T, baseURL, token string, fn func()) {
	t.Helper()
	withCLIPaths(t, func(paths testPaths) {
		configPath := filepath.Join(paths.XDG, "singctl", "config.yaml")
		mustSaveConfig(t, configPath, cfgpkg.Document{
			API: cfgpkg.APIConfig{
				BaseURL: baseURL,
				Token:   token,
				Timeout: "5s",
			},
		})
		fn()
	})
}

func TestTaskListHappyPathAndEmptyJSON(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		if r.Method != http.MethodGet || r.URL.Path != "/v2/task" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		q := r.URL.Query()
		if q.Get("projectId") != "P-abc" {
			t.Errorf("projectId=%q", q.Get("projectId"))
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"tasks":[` + string(sampleCLITaskJSON("T-1", "One", nil)) + `]}`))
	}))
	t.Cleanup(srv.Close)

	withTaskConfig(t, srv.URL, "test-token-list", func() {
		stdout, stderr, err := executeForTest([]string{"task", "list", "--project", "P-abc", "-o", "json"})
		if err != nil {
			t.Fatalf("err=%v stderr=%q", err, stderr)
		}
		if strings.TrimSpace(stderr) != "" {
			t.Fatalf("stderr=%q", stderr)
		}
		var arr []map[string]any
		if err := json.Unmarshal([]byte(stdout), &arr); err != nil {
			t.Fatalf("json array: %v\n%s", err, stdout)
		}
		if len(arr) != 1 || arr[0]["id"] != "T-1" {
			t.Fatalf("arr=%v", arr)
		}
	})

	emptySrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"tasks":[]}`))
	}))
	t.Cleanup(emptySrv.Close)
	withTaskConfig(t, emptySrv.URL, "test-token-list-empty", func() {
		stdout, stderr, err := executeForTest([]string{"task", "list", "-o", "json"})
		if err != nil {
			t.Fatalf("err=%v stderr=%q", err, stderr)
		}
		if strings.TrimSpace(stdout) != "[]" && !strings.HasPrefix(strings.TrimSpace(stdout), "[]") {
			// indented empty array
			var arr []any
			if err := json.Unmarshal([]byte(stdout), &arr); err != nil || len(arr) != 0 {
				t.Fatalf("want [], got %q", stdout)
			}
		}
	})
}

func TestTaskListLimitOffsetValidationNoNetwork(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"tasks":[]}`))
	}))
	t.Cleanup(srv.Close)

	withTaskConfig(t, srv.URL, "test-token-limit", func() {
		for _, args := range [][]string{
			{"task", "list", "--limit", "0"},
			{"task", "list", "--limit", "1001"},
			{"task", "list", "--offset", "-1"},
		} {
			stdout, stderr, err := executeForTest(args)
			if err == nil {
				t.Fatalf("args %v: expected error", args)
			}
			if ExitCode(err) != 1 {
				t.Fatalf("args %v: ExitCode=%d", args, ExitCode(err))
			}
			if strings.TrimSpace(stdout) != "" {
				t.Fatalf("args %v: stdout=%q", args, stdout)
			}
			if strings.TrimSpace(stderr) == "" {
				t.Fatalf("args %v: empty stderr", args)
			}
		}
	})
	if hits.Load() != 0 {
		t.Fatalf("network hits=%d want 0", hits.Load())
	}
}

func TestTaskGetJSONObjectAnd404(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v2/task/T-ok":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(sampleCLITaskJSON("T-ok", "Hello", nil))
		case r.Method == http.MethodGet && r.URL.Path == "/v2/task/T-missing":
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`gone`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)

	withTaskConfig(t, srv.URL, "test-token-get", func() {
		stdout, stderr, err := executeForTest([]string{"task", "get", "T-ok", "-o", "json"})
		if err != nil {
			t.Fatalf("get: %v stderr=%q", err, stderr)
		}
		if strings.TrimSpace(stderr) != "" {
			t.Fatalf("stderr=%q", stderr)
		}
		trimmed := strings.TrimSpace(stdout)
		if strings.HasPrefix(trimmed, "[") {
			t.Fatalf("get json must be object: %s", trimmed)
		}
		var obj map[string]any
		if err := json.Unmarshal([]byte(stdout), &obj); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if obj["id"] != "T-ok" {
			t.Fatalf("obj=%v", obj)
		}

		stdout, stderr, err = executeForTest([]string{"task", "get", "T-missing", "-o", "json"})
		if err == nil {
			t.Fatal("expected 404")
		}
		if ExitCode(err) != 3 {
			t.Fatalf("ExitCode=%d want 3", ExitCode(err))
		}
		if strings.TrimSpace(stdout) != "" {
			t.Fatalf("stdout must be empty on error, got %q", stdout)
		}
	})
}

func TestTaskAuthMissingTokenExit2(t *testing.T) {
	withCLIPaths(t, func(paths testPaths) {
		configPath := filepath.Join(paths.XDG, "singctl", "config.yaml")
		mustSaveConfig(t, configPath, cfgpkg.Document{
			API: cfgpkg.APIConfig{
				BaseURL: "https://example.invalid",
				Timeout: "5s",
			},
		})
		for _, args := range [][]string{
			{"task", "list"},
			{"task", "get", "T-1"},
		} {
			stdout, stderr, err := executeForTest(args)
			if err == nil {
				t.Fatalf("args %v: expected error", args)
			}
			if ExitCode(err) != 2 {
				t.Fatalf("args %v: ExitCode=%d want 2 stderr=%q", args, ExitCode(err), stderr)
			}
			if strings.TrimSpace(stdout) != "" {
				t.Fatalf("stdout=%q", stdout)
			}
		}
	})
}
