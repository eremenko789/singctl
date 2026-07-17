package cli

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestProjectSectionGetJSONObjectAnd404(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v2/task-group/Q-ok":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(sampleCLISectionJSON("Q-ok", "Hello", "P-1", nil))
		case r.Method == http.MethodGet && r.URL.Path == "/v2/task-group/Q-missing":
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`gone`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)

	withProjectConfig(t, srv.URL, "test-token-sec-get", func() {
		stdout, stderr, err := executeForTest([]string{"project", "section", "get", "Q-ok", "-o", "json"})
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
		if obj["id"] != "Q-ok" {
			t.Fatalf("obj=%v", obj)
		}

		stdout, stderr, err = executeForTest([]string{"project", "section", "get", "Q-missing", "-o", "json"})
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

func TestProjectSectionGetNoTokenExit2(t *testing.T) {
	withCLIPaths(t, func(paths testPaths) {
		stdout, stderr, err := executeForTest([]string{"project", "section", "get", "Q-1", "-o", "json"})
		if err == nil {
			t.Fatal("expected error")
		}
		if ExitCode(err) != 2 {
			t.Fatalf("ExitCode=%d stderr=%q stdout=%q", ExitCode(err), stderr, stdout)
		}
	})
}
