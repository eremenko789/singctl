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

func TestProjectSectionUpdatePartialEmptyAnd404(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		switch {
		case r.Method == http.MethodPatch && r.URL.Path == "/v2/task-group/Q-upd":
			body, _ := io.ReadAll(r.Body)
			var m map[string]any
			_ = json.Unmarshal(body, &m)
			if len(m) != 1 {
				t.Errorf("want single field patch, got %v", m)
			}
			if m["parent"] != "P-2" {
				t.Errorf("parent=%v", m["parent"])
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(sampleCLISectionJSON("Q-upd", "Upd", "P-2", nil))
		case r.Method == http.MethodPatch && r.URL.Path == "/v2/task-group/Q-missing":
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`gone`))
		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)

	withProjectConfig(t, srv.URL, "test-token-sec-update", func() {
		stdout, stderr, err := executeForTest([]string{"project", "section", "update", "Q-upd"})
		if err == nil {
			t.Fatal("expected empty update error")
		}
		if ExitCode(err) != 1 {
			t.Fatalf("empty update ExitCode=%d", ExitCode(err))
		}
		if strings.TrimSpace(stdout) != "" {
			t.Fatalf("stdout=%q", stdout)
		}
		_ = stderr

		stdout, stderr, err = executeForTest([]string{"project", "section", "update", "Q-upd", "--parent", "P-2", "-o", "json"})
		if err != nil {
			t.Fatalf("update: %v stderr=%q", err, stderr)
		}
		var obj map[string]any
		if err := json.Unmarshal([]byte(stdout), &obj); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if obj["parent"] != "P-2" {
			t.Fatalf("obj=%v", obj)
		}

		stdout, stderr, err = executeForTest([]string{"project", "section", "update", "Q-missing", "--title", "X"})
		if err == nil {
			t.Fatal("expected 404")
		}
		if ExitCode(err) != 3 {
			t.Fatalf("ExitCode=%d want 3", ExitCode(err))
		}
		if strings.TrimSpace(stdout) != "" {
			t.Fatalf("stdout=%q", stdout)
		}

		for _, args := range [][]string{
			{"project", "section", "update", "Q-upd", "--title", "   "},
			{"project", "section", "update", "Q-upd", "--parent", "   "},
		} {
			stdout, _, err = executeForTest(args)
			if err == nil {
				t.Fatalf("expected error for %v", args)
			}
			if ExitCode(err) != 1 {
				t.Fatalf("ExitCode=%d for %v", ExitCode(err), args)
			}
			if strings.TrimSpace(stdout) != "" {
				t.Fatalf("stdout=%q for %v", stdout, args)
			}
		}
	})
	if hits.Load() != 2 {
		t.Fatalf("hits=%d want 2 (validation must skip extra network)", hits.Load())
	}
}
