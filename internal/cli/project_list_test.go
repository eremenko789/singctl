package cli

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	cfgpkg "github.com/eremenko789/singctl/internal/config"
)

func sampleCLIProjectJSON(id, title string, extra map[string]any) []byte {
	m := map[string]any{
		"id":    id,
		"title": title,
	}
	for k, v := range extra {
		m[k] = v
	}
	b, _ := json.Marshal(m)
	return b
}

func withProjectConfig(t *testing.T, baseURL, token string, fn func()) {
	t.Helper()
	withCLIPaths(t, func(paths testPaths) {
		configPath := filepath.Join(paths.XDG, "singctl", "config.yaml")
		mustSaveConfig(t, configPath, cfgpkg.Document{
			API: cfgpkg.APIConfig{
				BaseURL: baseURL,
				Token:   token,
				Timeout: "5s",
			},
		})
		fn()
	})
}

func TestProjectListHappyPathEmptyAndFilters(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		if r.Method != http.MethodGet || r.URL.Path != "/v2/project" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		q := r.URL.Query()
		if q.Get("includeArchived") != "true" {
			t.Errorf("includeArchived=%q", q.Get("includeArchived"))
		}
		if q.Get("includeRemoved") != "false" {
			t.Errorf("includeRemoved=%q", q.Get("includeRemoved"))
		}
		if q.Get("maxCount") != "20" {
			t.Errorf("maxCount=%q", q.Get("maxCount"))
		}
		if q.Get("offset") != "5" {
			t.Errorf("offset=%q", q.Get("offset"))
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"projects":[` + string(sampleCLIProjectJSON("P-1", "One", nil)) + `]}`))
	}))
	t.Cleanup(srv.Close)

	withProjectConfig(t, srv.URL, "test-token-plist", func() {
		stdout, stderr, err := executeForTest([]string{
			"project", "list", "--archived", "--removed=false", "--limit", "20", "--offset", "5", "-o", "json",
		})
		if err != nil {
			t.Fatalf("err=%v stderr=%q", err, stderr)
		}
		if strings.TrimSpace(stderr) != "" {
			t.Fatalf("stderr=%q", stderr)
		}
		var arr []map[string]any
		if err := json.Unmarshal([]byte(stdout), &arr); err != nil {
			t.Fatalf("json array: %v\n%s", err, stdout)
		}
		if len(arr) != 1 || arr[0]["id"] != "P-1" {
			t.Fatalf("arr=%v", arr)
		}
	})

	emptySrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"projects":[]}`))
	}))
	t.Cleanup(emptySrv.Close)
	withProjectConfig(t, emptySrv.URL, "test-token-plist-empty", func() {
		stdout, stderr, err := executeForTest([]string{"project", "list", "-o", "json"})
		if err != nil {
			t.Fatalf("err=%v stderr=%q", err, stderr)
		}
		var arr []any
		if err := json.Unmarshal([]byte(stdout), &arr); err != nil || len(arr) != 0 {
			t.Fatalf("want [], got %q", stdout)
		}
	})
}

func TestProjectListLimitOffsetValidationNoNetwork(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"projects":[]}`))
	}))
	t.Cleanup(srv.Close)

	withProjectConfig(t, srv.URL, "test-token-plimit", func() {
		for _, args := range [][]string{
			{"project", "list", "--limit", "0"},
			{"project", "list", "--limit", "1001"},
			{"project", "list", "--offset", "-1"},
		} {
			stdout, stderr, err := executeForTest(args)
			if err == nil {
				t.Fatalf("args %v: expected error", args)
			}
			if ExitCode(err) != 1 {
				t.Fatalf("args %v: ExitCode=%d", args, ExitCode(err))
			}
			if strings.TrimSpace(stdout) != "" {
				t.Fatalf("args %v: stdout=%q", args, stdout)
			}
			if strings.TrimSpace(stderr) == "" {
				t.Fatalf("args %v: empty stderr", args)
			}
		}
	})
	if hits.Load() != 0 {
		t.Fatalf("network hits=%d want 0", hits.Load())
	}
}

func TestProjectAuthMissingTokenExit2(t *testing.T) {
	withCLIPaths(t, func(paths testPaths) {
		configPath := filepath.Join(paths.XDG, "singctl", "config.yaml")
		mustSaveConfig(t, configPath, cfgpkg.Document{
			API: cfgpkg.APIConfig{
				BaseURL: "https://example.invalid",
				Timeout: "5s",
			},
		})
		for _, args := range [][]string{
			{"project", "list"},
			{"project", "get", "P-1"},
		} {
			stdout, stderr, err := executeForTest(args)
			if err == nil {
				t.Fatalf("%v: expected error", args)
			}
			if ExitCode(err) != 2 {
				t.Fatalf("%v: ExitCode=%d stderr=%q", args, ExitCode(err), stderr)
			}
			if strings.TrimSpace(stdout) != "" {
				t.Fatalf("%v: stdout=%q", args, stdout)
			}
		}
	})
}
