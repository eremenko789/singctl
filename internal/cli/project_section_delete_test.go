package cli

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func TestProjectSectionDeleteEmptyStdoutAnd404(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodDelete && r.URL.Path == "/v2/task-group/Q-del":
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodDelete && r.URL.Path == "/v2/task-group/Q-missing":
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`gone`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)

	withProjectConfig(t, srv.URL, "test-token-sec-del", func() {
		stdout, stderr, err := executeForTest([]string{"project", "section", "delete", "Q-del"})
		if err != nil {
			t.Fatalf("delete: %v stderr=%q", err, stderr)
		}
		if strings.TrimSpace(stdout) != "" {
			t.Fatalf("stdout must be empty, got %q", stdout)
		}
		if strings.TrimSpace(stderr) != "" {
			t.Fatalf("stderr=%q", stderr)
		}

		stdout, stderr, err = executeForTest([]string{"project", "section", "delete", "Q-missing"})
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
}

func TestProjectSectionDeleteEmptyIDNoNetwork(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	withProjectConfig(t, srv.URL, "test-token-sec-del-id", func() {
		stdout, _, err := executeForTest([]string{"project", "section", "delete", "   "})
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
	if hits.Load() != 0 {
		t.Fatalf("hits=%d want 0", hits.Load())
	}
}
