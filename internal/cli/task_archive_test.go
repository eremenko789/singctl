package cli

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/eremenko789/singctl/internal/api"
)

func TestTaskArchiveAndTrash(t *testing.T) {
	today := api.TodayCalendarDate()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("method=%s", r.Method)
		}
		body, _ := io.ReadAll(r.Body)
		var m map[string]any
		_ = json.Unmarshal(body, &m)
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/v2/task/T-arch":
			if m["journalDate"] != "2026-07-10" {
				t.Errorf("journalDate=%v", m["journalDate"])
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(sampleCLITaskJSON("T-arch", "A", map[string]any{"journalDate": "2026-07-10"}))
		case "/v2/task/T-arch-today":
			if m["journalDate"] != today {
				t.Errorf("journalDate=%v want today %s", m["journalDate"], today)
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(sampleCLITaskJSON("T-arch-today", "A", map[string]any{"journalDate": today}))
		case "/v2/task/T-trash":
			if m["deleteDate"] != "2026-07-11" {
				t.Errorf("deleteDate=%v", m["deleteDate"])
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(sampleCLITaskJSON("T-trash", "B", map[string]any{"deleteDate": "2026-07-11"}))
		case "/v2/task/T-trash-today":
			if m["deleteDate"] != today {
				t.Errorf("deleteDate=%v want %s", m["deleteDate"], today)
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(sampleCLITaskJSON("T-trash-today", "B", map[string]any{"deleteDate": today}))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)

	withTaskConfig(t, srv.URL, "test-token-arch", func() {
		stdout, stderr, err := executeForTest([]string{"task", "archive", "T-arch", "--date", "2026-07-10", "-o", "json"})
		if err != nil {
			t.Fatalf("archive: %v stderr=%q", err, stderr)
		}
		var obj map[string]any
		_ = json.Unmarshal([]byte(stdout), &obj)
		if obj["journalDate"] != "2026-07-10" {
			t.Fatalf("obj=%v", obj)
		}

		stdout, stderr, err = executeForTest([]string{"task", "archive", "T-arch-today", "-o", "json"})
		if err != nil {
			t.Fatalf("archive today: %v stderr=%q", err, stderr)
		}
		_ = json.Unmarshal([]byte(stdout), &obj)
		if obj["journalDate"] != today {
			t.Fatalf("obj=%v", obj)
		}

		stdout, stderr, err = executeForTest([]string{"task", "trash", "T-trash", "--date", "2026-07-11", "-o", "json"})
		if err != nil {
			t.Fatalf("trash: %v stderr=%q", err, stderr)
		}
		_ = json.Unmarshal([]byte(stdout), &obj)
		if obj["deleteDate"] != "2026-07-11" {
			t.Fatalf("obj=%v", obj)
		}

		stdout, stderr, err = executeForTest([]string{"task", "trash", "T-trash-today", "-o", "json"})
		if err != nil {
			t.Fatalf("trash today: %v stderr=%q", err, stderr)
		}
		_ = json.Unmarshal([]byte(stdout), &obj)
		if obj["deleteDate"] != today {
			t.Fatalf("obj=%v", obj)
		}
	})
}

func TestTaskArchiveTrashInvalidDateNoNetwork(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	withTaskConfig(t, srv.URL, "test-token-date-bad", func() {
		for _, args := range [][]string{
			{"task", "archive", "T-1", "--date", "not-a-date"},
			{"task", "trash", "T-1", "--date", "28.07.2026"},
		} {
			stdout, _, err := executeForTest(args)
			if err == nil {
				t.Fatalf("args %v expected error", args)
			}
			if ExitCode(err) != 1 {
				t.Fatalf("ExitCode=%d", ExitCode(err))
			}
			if strings.TrimSpace(stdout) != "" {
				t.Fatalf("stdout=%q", stdout)
			}
		}
	})
	if hits.Load() != 0 {
		t.Fatalf("hits=%d", hits.Load())
	}
}

func TestTaskDeleteEmptyStdoutAnd404(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodDelete && r.URL.Path == "/v2/task/T-del":
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodDelete && r.URL.Path == "/v2/task/T-missing":
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`no`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)

	withTaskConfig(t, srv.URL, "test-token-delete", func() {
		stdout, stderr, err := executeForTest([]string{"task", "delete", "T-del"})
		if err != nil {
			t.Fatalf("delete: %v stderr=%q", err, stderr)
		}
		if strings.TrimSpace(stdout) != "" {
			t.Fatalf("stdout must be empty, got %q", stdout)
		}
		if strings.TrimSpace(stderr) != "" {
			t.Fatalf("stderr=%q", stderr)
		}

		stdout, _, err = executeForTest([]string{"task", "delete", "T-missing"})
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
