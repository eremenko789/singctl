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

func TestProjectUpdatePartialEmptyAnd404(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		switch {
		case r.Method == http.MethodPatch && r.URL.Path == "/v2/project/P-upd":
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
			_, _ = w.Write(sampleCLIProjectJSON("P-upd", "Upd", map[string]any{"parent": "P-2"}))
		case r.Method == http.MethodPatch && r.URL.Path == "/v2/project/P-missing":
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`gone`))
		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)

	withProjectConfig(t, srv.URL, "test-token-pupdate", func() {
		stdout, stderr, err := executeForTest([]string{"project", "update", "P-upd"})
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

		stdout, stderr, err = executeForTest([]string{"project", "update", "P-upd", "--parent", "P-2", "-o", "json"})
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

		stdout, stderr, err = executeForTest([]string{"project", "update", "P-missing", "--title", "X"})
		if err == nil {
			t.Fatal("expected 404")
		}
		if ExitCode(err) != 3 {
			t.Fatalf("ExitCode=%d want 3", ExitCode(err))
		}
		if strings.TrimSpace(stdout) != "" {
			t.Fatalf("stdout=%q", stdout)
		}
	})
	if hits.Load() != 2 {
		t.Fatalf("hits=%d want 2 (empty update must skip network)", hits.Load())
	}
}
