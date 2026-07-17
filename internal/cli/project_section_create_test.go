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

func TestProjectSectionCreateHappyPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v2/task-group" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		body, _ := io.ReadAll(r.Body)
		var m map[string]any
		_ = json.Unmarshal(body, &m)
		if m["title"] != "New Section" || m["parent"] != "P-parent" {
			t.Errorf("body=%v", m)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(sampleCLISectionJSON("Q-new", "New Section", "P-parent", nil))
	}))
	t.Cleanup(srv.Close)

	withProjectConfig(t, srv.URL, "test-token-sec-create", func() {
		stdout, stderr, err := executeForTest([]string{
			"project", "section", "create", "P-parent", "--title", "New Section", "-o", "json",
		})
		if err != nil {
			t.Fatalf("err=%v stderr=%q", err, stderr)
		}
		if strings.TrimSpace(stderr) != "" {
			t.Fatalf("stderr=%q", stderr)
		}
		trimmed := strings.TrimSpace(stdout)
		if strings.HasPrefix(trimmed, "[") {
			t.Fatalf("create json must be object: %s", trimmed)
		}
		var obj map[string]any
		if err := json.Unmarshal([]byte(stdout), &obj); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if obj["id"] != "Q-new" {
			t.Fatalf("obj=%v", obj)
		}
	})
}

func TestProjectSectionCreateValidationNoNetwork(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	withProjectConfig(t, srv.URL, "test-token-sec-cval", func() {
		cases := [][]string{
			{"project", "section", "create", "P-1"},
			{"project", "section", "create", "P-1", "--title", "   "},
			{"project", "section", "create", "   ", "--title", "X"},
		}
		for _, args := range cases {
			stdout, _, err := executeForTest(args)
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
	if hits.Load() != 0 {
		t.Fatalf("hits=%d want 0", hits.Load())
	}
}
