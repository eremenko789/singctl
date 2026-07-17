package cli

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func sampleCLISectionJSON(id, title, parent string, extra map[string]any) []byte {
	m := map[string]any{
		"id":              id,
		"title":           title,
		"parent":          parent,
		"parentOrder":     float32(1),
		"removed":         false,
		"fake":            false,
		"modificatedDate": "2026-07-01",
	}
	for k, v := range extra {
		m[k] = v
	}
	b, _ := json.Marshal(m)
	return b
}

func TestProjectSectionListHappyPathEmptyAndFilters(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		if r.Method != http.MethodGet || r.URL.Path != "/v2/task-group" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		q := r.URL.Query()
		if q.Get("parent") != "P-1" {
			t.Errorf("parent=%q", q.Get("parent"))
		}
		if q.Get("includeRemoved") != "true" {
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
		_, _ = w.Write([]byte(`{"taskGroups":[` + string(sampleCLISectionJSON("Q-1", "One", "P-1", nil)) + `]}`))
	}))
	t.Cleanup(srv.Close)

	withProjectConfig(t, srv.URL, "test-token-sec-list", func() {
		stdout, stderr, err := executeForTest([]string{
			"project", "section", "list", "P-1", "--removed", "--limit", "20", "--offset", "5", "-o", "json",
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
		if len(arr) != 1 || arr[0]["id"] != "Q-1" {
			t.Fatalf("arr=%v", arr)
		}
	})

	emptySrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"taskGroups":[]}`))
	}))
	t.Cleanup(emptySrv.Close)
	withProjectConfig(t, emptySrv.URL, "test-token-sec-empty", func() {
		stdout, stderr, err := executeForTest([]string{"project", "section", "list", "P-empty", "-o", "json"})
		if err != nil {
			t.Fatalf("err=%v stderr=%q", err, stderr)
		}
		var arr []any
		if err := json.Unmarshal([]byte(stdout), &arr); err != nil || len(arr) != 0 {
			t.Fatalf("want [], got %q", stdout)
		}
	})
}

func TestProjectSectionListMissingProjectIDNoNetwork(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	withProjectConfig(t, srv.URL, "test-token-sec-noid", func() {
		stdout, _, err := executeForTest([]string{"project", "section", "list", "   "})
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

func TestProjectSectionListLimitOffsetValidationNoNetwork(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	withProjectConfig(t, srv.URL, "test-token-sec-limit", func() {
		for _, args := range [][]string{
			{"project", "section", "list", "P-1", "--limit", "0"},
			{"project", "section", "list", "P-1", "--limit", "1001"},
			{"project", "section", "list", "P-1", "--offset", "-1"},
		} {
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

func TestProjectSectionListNoTokenExit2(t *testing.T) {
	withCLIPaths(t, func(paths testPaths) {
		stdout, stderr, err := executeForTest([]string{"project", "section", "list", "P-1", "-o", "json"})
		if err == nil {
			t.Fatal("expected error")
		}
		if ExitCode(err) != 2 {
			t.Fatalf("ExitCode=%d stderr=%q stdout=%q", ExitCode(err), stderr, stdout)
		}
	})
}
