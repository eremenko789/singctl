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

func TestTaskCreateHappyPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v2/task" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		body, _ := io.ReadAll(r.Body)
		var m map[string]any
		_ = json.Unmarshal(body, &m)
		if m["title"] != "New task" {
			t.Errorf("title=%v", m["title"])
		}
		if m["projectId"] != "P-1" {
			t.Errorf("projectId=%v", m["projectId"])
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(sampleCLITaskJSON("T-new", "New task", map[string]any{"projectId": "P-1"}))
	}))
	t.Cleanup(srv.Close)

	withTaskConfig(t, srv.URL, "test-token-create", func() {
		stdout, stderr, err := executeForTest([]string{
			"task", "create", "--title", "New task", "--project", "P-1", "-o", "json",
		})
		if err != nil {
			t.Fatalf("err=%v stderr=%q", err, stderr)
		}
		if strings.TrimSpace(stderr) != "" {
			t.Fatalf("stderr=%q", stderr)
		}
		var obj map[string]any
		if err := json.Unmarshal([]byte(stdout), &obj); err != nil {
			t.Fatalf("object: %v\n%s", err, stdout)
		}
		if obj["id"] != "T-new" {
			t.Fatalf("obj=%v", obj)
		}
	})
}

func TestTaskCreateValidationAndDeleteDate(t *testing.T) {
	var hits atomic.Int32
	var methods []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		methods = append(methods, r.Method)
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPost:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(sampleCLITaskJSON("T-deldate", "X", nil))
		case r.Method == http.MethodPatch:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(sampleCLITaskJSON("T-deldate", "X", map[string]any{"deleteDate": "2026-07-16"}))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)

	withTaskConfig(t, srv.URL, "test-token-create-val", func() {
		stdout, _, err := executeForTest([]string{"task", "create"})
		if err == nil {
			t.Fatal("expected missing title error")
		}
		if ExitCode(err) != 1 {
			t.Fatalf("ExitCode=%d", ExitCode(err))
		}
		if strings.TrimSpace(stdout) != "" {
			t.Fatalf("stdout=%q", stdout)
		}
		if hits.Load() != 0 {
			t.Fatalf("network on missing title: %d", hits.Load())
		}

		stdout, _, err = executeForTest([]string{"task", "create", "--title", "x", "--priority", "9"})
		if err == nil {
			t.Fatal("expected bad priority")
		}
		if ExitCode(err) != 1 {
			t.Fatalf("ExitCode=%d", ExitCode(err))
		}
		if hits.Load() != 0 {
			t.Fatalf("network on bad priority: %d", hits.Load())
		}

		stdout, stderr, err := executeForTest([]string{
			"task", "create", "--title", "X", "--delete-date", "2026-07-16", "-o", "json",
		})
		if err != nil {
			t.Fatalf("create+delete-date: %v stderr=%q", err, stderr)
		}
		if len(methods) != 2 || methods[0] != http.MethodPost || methods[1] != http.MethodPatch {
			t.Fatalf("methods=%v", methods)
		}
		var obj map[string]any
		if err := json.Unmarshal([]byte(stdout), &obj); err != nil {
			t.Fatal(err)
		}
		if obj["deleteDate"] != "2026-07-16" {
			t.Fatalf("obj=%v", obj)
		}
	})
}

func TestTaskUpdatePartialAndEmptyAnd404(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPatch && r.URL.Path == "/v2/task/T-upd":
			body, _ := io.ReadAll(r.Body)
			var m map[string]any
			_ = json.Unmarshal(body, &m)
			if m["title"] != "Renamed" || len(m) != 1 {
				t.Errorf("body=%v", m)
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(sampleCLITaskJSON("T-upd", "Renamed", nil))
		case r.Method == http.MethodPatch && r.URL.Path == "/v2/task/T-missing":
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`no`))
		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)

	withTaskConfig(t, srv.URL, "test-token-update", func() {
		stdout, stderr, err := executeForTest([]string{"task", "update", "T-upd"})
		if err == nil {
			t.Fatal("expected no-flags error")
		}
		if ExitCode(err) != 1 {
			t.Fatalf("ExitCode=%d", ExitCode(err))
		}
		if strings.TrimSpace(stdout) != "" {
			t.Fatalf("stdout=%q", stdout)
		}

		stdout, stderr, err = executeForTest([]string{"task", "update", "T-upd", "--title", "Renamed", "-o", "json"})
		if err != nil {
			t.Fatalf("update: %v stderr=%q", err, stderr)
		}
		var obj map[string]any
		if err := json.Unmarshal([]byte(stdout), &obj); err != nil {
			t.Fatal(err)
		}
		if obj["title"] != "Renamed" {
			t.Fatalf("obj=%v", obj)
		}

		stdout, _, err = executeForTest([]string{"task", "update", "T-missing", "--title", "x"})
		if err == nil {
			t.Fatal("expected 404")
		}
		if ExitCode(err) != 3 {
			t.Fatalf("ExitCode=%d", ExitCode(err))
		}
		if strings.TrimSpace(stdout) != "" {
			t.Fatalf("stdout=%q", stdout)
		}
	})
}
