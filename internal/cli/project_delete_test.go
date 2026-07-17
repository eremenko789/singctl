package cli

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestProjectDeleteEmptyStdoutAnd404(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodDelete && r.URL.Path == "/v2/project/P-del":
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodDelete && r.URL.Path == "/v2/project/P-missing":
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`gone`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)

	withProjectConfig(t, srv.URL, "test-token-pdel", func() {
		stdout, stderr, err := executeForTest([]string{"project", "delete", "P-del"})
		if err != nil {
			t.Fatalf("delete: %v stderr=%q", err, stderr)
		}
		if strings.TrimSpace(stdout) != "" {
			t.Fatalf("stdout must be empty, got %q", stdout)
		}
		if strings.TrimSpace(stderr) != "" {
			t.Fatalf("stderr=%q", stderr)
		}

		stdout, stderr, err = executeForTest([]string{"project", "delete", "P-missing"})
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
