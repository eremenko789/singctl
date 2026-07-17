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

func TestProjectArchiveAndTrashWithDateAndToday(t *testing.T) {
	today := api.TodayCalendarDate()
	var archiveBodies, trashBodies []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("want PATCH, got %s", r.Method)
		}
		body, _ := io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/v2/project/P-arch":
			archiveBodies = append(archiveBodies, string(body))
			var m map[string]any
			_ = json.Unmarshal(body, &m)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(sampleCLIProjectJSON("P-arch", "A", map[string]any{"journalDate": m["journalDate"]}))
		case "/v2/project/P-trash":
			trashBodies = append(trashBodies, string(body))
			var m map[string]any
			_ = json.Unmarshal(body, &m)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(sampleCLIProjectJSON("P-trash", "B", nil))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)

	withProjectConfig(t, srv.URL, "test-token-parch", func() {
		stdout, stderr, err := executeForTest([]string{"project", "archive", "P-arch", "--date", "2026-07-10", "-o", "json"})
		if err != nil {
			t.Fatalf("archive dated: %v stderr=%q", err, stderr)
		}
		var obj map[string]any
		if err := json.Unmarshal([]byte(stdout), &obj); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if obj["journalDate"] != "2026-07-10" {
			t.Fatalf("obj=%v", obj)
		}

		stdout, stderr, err = executeForTest([]string{"project", "archive", "P-arch", "-o", "json"})
		if err != nil {
			t.Fatalf("archive today: %v stderr=%q", err, stderr)
		}
		if !strings.Contains(archiveBodies[1], today) {
			t.Fatalf("archive body without --date should use today %s: %s", today, archiveBodies[1])
		}

		stdout, stderr, err = executeForTest([]string{"project", "trash", "P-trash", "--date", "2026-07-11", "-o", "json"})
		if err != nil {
			t.Fatalf("trash dated: %v stderr=%q", err, stderr)
		}
		if !strings.Contains(trashBodies[0], "2026-07-11") {
			t.Fatalf("trash body=%s", trashBodies[0])
		}

		_, stderr, err = executeForTest([]string{"project", "trash", "P-trash", "-o", "json"})
		if err != nil {
			t.Fatalf("trash today: %v stderr=%q", err, stderr)
		}
		if !strings.Contains(trashBodies[1], today) {
			t.Fatalf("trash body without --date should use today %s: %s", today, trashBodies[1])
		}
	})
}

func TestProjectArchiveTrashInvalidDateNoNetwork(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	withProjectConfig(t, srv.URL, "test-token-pdate", func() {
		for _, args := range [][]string{
			{"project", "archive", "P-1", "--date", "not-a-date"},
			{"project", "trash", "P-1", "--date", "not-a-date"},
		} {
			stdout, stderr, err := executeForTest(args)
			if err == nil {
				t.Fatalf("%v: expected error", args)
			}
			if ExitCode(err) != 1 {
				t.Fatalf("%v: ExitCode=%d", args, ExitCode(err))
			}
			if strings.TrimSpace(stdout) != "" {
				t.Fatalf("%v: stdout=%q", args, stdout)
			}
			if strings.TrimSpace(stderr) == "" {
				t.Fatalf("%v: empty stderr", args)
			}
		}
	})
	if hits.Load() != 0 {
		t.Fatalf("hits=%d want 0", hits.Load())
	}
}
