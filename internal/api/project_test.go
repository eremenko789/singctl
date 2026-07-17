package api_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/eremenko789/singctl/internal/api"
)

func sampleProjectJSON(id, title string, extra map[string]any) []byte {
	m := map[string]any{
		"id":    id,
		"title": title,
	}
	for k, v := range extra {
		m[k] = v
	}
	b, err := json.Marshal(m)
	if err != nil {
		panic(err)
	}
	return b
}

func sampleTaskGroupJSON() []byte {
	return []byte(`{"id":"Q-1","title":"Default","fake":false,"modificatedDate":"2026-07-01","parentOrder":0,"removed":false}`)
}

func TestListProjectsMapsFiltersAndReturnsProjects(t *testing.T) {
	t.Parallel()
	var gotPath, gotQuery, gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"projects":[` + string(sampleProjectJSON("P-1", "Inbox", nil)) + `]}`))
	}))
	t.Cleanup(srv.Close)

	s, err := api.NewSession(srv.URL, "test-token-plist", "5s")
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	limit := 10
	offset := 5
	archived := true
	removed := false
	projects, err := s.ListProjects(context.Background(), api.ProjectListQuery{
		MaxCount:        &limit,
		Offset:          &offset,
		IncludeArchived: &archived,
		IncludeRemoved:  &removed,
	})
	if err != nil {
		t.Fatalf("ListProjects: %v", err)
	}
	if gotAuth != "Bearer test-token-plist" {
		t.Fatalf("Authorization = %q", gotAuth)
	}
	if gotPath != "/v2/project" {
		t.Fatalf("path = %q", gotPath)
	}
	for _, want := range []string{
		"maxCount=10",
		"offset=5",
		"includeArchived=true",
		"includeRemoved=false",
	} {
		if !strings.Contains(gotQuery, want) {
			t.Fatalf("query %q missing %q", gotQuery, want)
		}
	}
	if len(projects) != 1 || projects[0].ID != "P-1" || projects[0].Title != "Inbox" {
		t.Fatalf("projects=%#v", projects)
	}
}

func TestGetCreateUpdateDeleteProjectHappyPaths(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token-pcrud" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v2/project/P-get":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(sampleProjectJSON("P-get", "Got", nil))
		case r.Method == http.MethodPost && r.URL.Path == "/v2/project":
			body, _ := io.ReadAll(r.Body)
			var m map[string]any
			_ = json.Unmarshal(body, &m)
			if m["title"] != "New" {
				t.Errorf("create title=%v", m["title"])
			}
			w.WriteHeader(http.StatusOK)
			proj := sampleProjectJSON("P-new", "New", nil)
			_, _ = w.Write([]byte(`{"project":` + string(proj) + `,"taskGroup":` + string(sampleTaskGroupJSON()) + `}`))
		case r.Method == http.MethodPatch && r.URL.Path == "/v2/project/P-upd":
			body, _ := io.ReadAll(r.Body)
			var m map[string]any
			_ = json.Unmarshal(body, &m)
			if m["title"] != "Updated" {
				t.Errorf("update title=%v", m["title"])
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(sampleProjectJSON("P-upd", "Updated", nil))
		case r.Method == http.MethodDelete && r.URL.Path == "/v2/project/P-del":
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)

	s, err := api.NewSession(srv.URL, "test-token-pcrud", "5s")
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	ctx := context.Background()

	got, err := s.GetProject(ctx, "P-get")
	if err != nil || got.ID != "P-get" {
		t.Fatalf("GetProject: %#v %v", got, err)
	}

	title := "New"
	created, err := s.CreateProject(ctx, api.ProjectWriteInput{Title: &title})
	if err != nil || created.ID != "P-new" {
		t.Fatalf("CreateProject: %#v %v", created, err)
	}

	updTitle := "Updated"
	updated, err := s.UpdateProject(ctx, "P-upd", api.ProjectWriteInput{Title: &updTitle})
	if err != nil || updated.Title != "Updated" {
		t.Fatalf("UpdateProject: %#v %v", updated, err)
	}

	if err := s.DeleteProject(ctx, "P-del"); err != nil {
		t.Fatalf("DeleteProject: %v", err)
	}
}

func TestGetProject404KindNotFound(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`missing`))
	}))
	t.Cleanup(srv.Close)

	s, err := api.NewSession(srv.URL, "test-token-p404", "5s")
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	_, err = s.GetProject(context.Background(), "P-missing")
	if err == nil {
		t.Fatal("expected error")
	}
	var ce *api.ClassifiedError
	if !errors.As(err, &ce) {
		t.Fatalf("want ClassifiedError, got %T %v", err, err)
	}
	if ce.Kind != api.KindNotFound {
		t.Fatalf("Kind=%q want not-found", ce.Kind)
	}
	if !strings.Contains(ce.Message, "P-missing") {
		t.Fatalf("message should include entity id: %q", ce.Message)
	}
}

func TestCreateProjectUnwrapsProjectIgnoresTaskGroup(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		proj := sampleProjectJSON("P-created", "With Group", map[string]any{"emoji": "1f49e"})
		// Intentionally large taskGroup that must not leak into Project view.
		tg := `{"id":"Q-secret","title":"SHOULD-NOT-LEAK","fake":false,"modificatedDate":"2026-07-01","parentOrder":0,"removed":false}`
		_, _ = w.Write([]byte(`{"project":` + string(proj) + `,"taskGroup":` + tg + `}`))
	}))
	t.Cleanup(srv.Close)

	s, err := api.NewSession(srv.URL, "test-token-unwrap", "5s")
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	title := "With Group"
	p, err := s.CreateProject(context.Background(), api.ProjectWriteInput{Title: &title})
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	if p.ID != "P-created" || p.Title != "With Group" || p.Emoji != "1f49e" {
		t.Fatalf("project=%#v", p)
	}
	b, _ := json.Marshal(p)
	if strings.Contains(string(b), "SHOULD-NOT-LEAK") || strings.Contains(string(b), "Q-secret") {
		t.Fatalf("taskGroup leaked into Project: %s", b)
	}
}

func TestArchiveAndTrashProjectPatchOnlyDates(t *testing.T) {
	t.Parallel()
	var n atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("want PATCH, got %s", r.Method)
		}
		body, _ := io.ReadAll(r.Body)
		var m map[string]any
		_ = json.Unmarshal(body, &m)
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/v2/project/P-arch":
			n.Add(1)
			if _, ok := m["journalDate"]; !ok || len(m) != 1 {
				t.Errorf("archive body=%v want only journalDate", m)
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(sampleProjectJSON("P-arch", "A", map[string]any{"journalDate": m["journalDate"]}))
		case "/v2/project/P-trash":
			n.Add(1)
			if _, ok := m["deleteDate"]; !ok || len(m) != 1 {
				t.Errorf("trash body=%v want only deleteDate", m)
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(sampleProjectJSON("P-trash", "B", nil))
		default:
			t.Errorf("path=%s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)

	s, err := api.NewSession(srv.URL, "test-token-parch-trash", "5s")
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	ctx := context.Background()
	a, err := s.ArchiveProject(ctx, "P-arch", "2026-07-10")
	if err != nil || a.JournalDate != "2026-07-10" {
		t.Fatalf("ArchiveProject: %#v %v", a, err)
	}
	tr, err := s.TrashProject(ctx, "P-trash", "2026-07-11")
	if err != nil || tr.ID != "P-trash" {
		t.Fatalf("TrashProject: %#v %v", tr, err)
	}
	if n.Load() != 2 {
		t.Fatalf("hits=%d", n.Load())
	}
}
