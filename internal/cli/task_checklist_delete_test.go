package cli

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTaskChecklistDeleteSuccessAnd404(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodDelete && r.URL.Path == "/v2/checklist-item/C-1":
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodDelete && r.URL.Path == "/v2/checklist-item/C-missing":
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`missing`))
		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)

	withTaskConfig(t, srv.URL, "test-token-cl-del", func() {
		stdout, stderr, err := executeForTest([]string{"task", "checklist", "delete", "C-1"})
		if err != nil {
			t.Fatalf("err=%v stderr=%q", err, stderr)
		}
		if strings.TrimSpace(stdout) != "" {
			t.Fatalf("stdout=%q", stdout)
		}

		stdout, stderr, err = executeForTest([]string{"task", "checklist", "delete", "C-missing"})
		if err == nil {
			t.Fatal("expected error")
		}
		if ExitCode(err) != 3 {
			t.Fatalf("ExitCode=%d stderr=%q", ExitCode(err), stderr)
		}
		if strings.TrimSpace(stdout) != "" {
			t.Fatalf("stdout=%q", stdout)
		}
	})
}

func TestTaskChecklistDeleteEmptyID(t *testing.T) {
	withTaskConfig(t, "http://127.0.0.1:9", "test-token-cl-del-id", func() {
		stdout, _, err := executeForTest([]string{"task", "checklist", "delete", " \t "})
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
}
