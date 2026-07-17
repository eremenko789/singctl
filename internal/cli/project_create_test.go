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

func TestProjectCreateHappyPathAndOptionalFlags(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v2/project" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		body, _ := io.ReadAll(r.Body)
		var m map[string]any
		_ = json.Unmarshal(body, &m)
		if m["title"] != "New Proj" {
			t.Errorf("title=%v", m["title"])
		}
		if m["note"] != "hello" {
			t.Errorf("note=%v", m["note"])
		}
		if m["isNotebook"] != true {
			t.Errorf("isNotebook=%v", m["isNotebook"])
		}
		if m["color"] != "#ff0000" {
			t.Errorf("color=%v", m["color"])
		}
		if m["parent"] != "P-parent" {
			t.Errorf("parent=%v", m["parent"])
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		proj := sampleCLIProjectJSON("P-new", "New Proj", map[string]any{
			"color": "#ff0000", "parent": "P-parent", "isNotebook": true,
		})
		tg := `{"id":"Q-1","title":"g","fake":false,"modificatedDate":"2026-07-01","parentOrder":0,"removed":false}`
		_, _ = w.Write([]byte(`{"project":` + string(proj) + `,"taskGroup":` + tg + `}`))
	}))
	t.Cleanup(srv.Close)

	withProjectConfig(t, srv.URL, "test-token-pcreate", func() {
		stdout, stderr, err := executeForTest([]string{
			"project", "create",
			"--title", "New Proj",
			"--note", "hello",
			"--notebook",
			"--color", "#ff0000",
			"--parent", "P-parent",
			"-o", "json",
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
		if obj["id"] != "P-new" {
			t.Fatalf("obj=%v", obj)
		}
	})
}

func TestProjectCreateTitleRequiredAndEmojiNormalize(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		body, _ := io.ReadAll(r.Body)
		var m map[string]any
		_ = json.Unmarshal(body, &m)
		if m["emoji"] != "1f49e" {
			t.Errorf("emoji=%v want 1f49e", m["emoji"])
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		proj := sampleCLIProjectJSON("P-emo", "Emo", map[string]any{"emoji": "1f49e"})
		tg := `{"id":"Q-1","title":"g","fake":false,"modificatedDate":"2026-07-01","parentOrder":0,"removed":false}`
		_, _ = w.Write([]byte(`{"project":` + string(proj) + `,"taskGroup":` + tg + `}`))
	}))
	t.Cleanup(srv.Close)

	withProjectConfig(t, srv.URL, "test-token-pemoji", func() {
		stdout, stderr, err := executeForTest([]string{"project", "create"})
		if err == nil {
			t.Fatal("expected missing title error")
		}
		if ExitCode(err) != 1 {
			t.Fatalf("ExitCode=%d", ExitCode(err))
		}
		if strings.TrimSpace(stdout) != "" {
			t.Fatalf("stdout=%q", stdout)
		}
		_ = stderr

		stdout, stderr, err = executeForTest([]string{"project", "create", "--title", "X", "--emoji", "heart"})
		if err == nil {
			t.Fatal("expected bad emoji")
		}
		if ExitCode(err) != 1 {
			t.Fatalf("bad emoji ExitCode=%d", ExitCode(err))
		}
		if strings.TrimSpace(stdout) != "" {
			t.Fatalf("stdout=%q", stdout)
		}

		stdout, stderr, err = executeForTest([]string{"project", "create", "--title", "Emo", "--emoji", "💞", "-o", "json"})
		if err != nil {
			t.Fatalf("emoji create: %v stderr=%q", err, stderr)
		}
		var obj map[string]any
		if err := json.Unmarshal([]byte(stdout), &obj); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if obj["emoji"] != "1f49e" {
			t.Fatalf("obj emoji=%v", obj["emoji"])
		}
	})
	// title-required and bad-emoji must not hit network; unicode emoji create hits once
	if hits.Load() != 1 {
		t.Fatalf("hits=%d want 1", hits.Load())
	}
}
