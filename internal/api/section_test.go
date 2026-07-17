package api_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/eremenko789/singctl/internal/api"
)

func sampleSectionJSON(id, title string, extra map[string]any) []byte {
	m := map[string]any{
		"id":              id,
		"title":           title,
		"fake":            false,
		"modificatedDate": "2026-07-01",
		"parentOrder":     float32(0),
		"removed":         false,
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

func TestListSectionsMapsFiltersAndReturnsSections(t *testing.T) {
	t.Parallel()
	var gotPath, gotQuery, gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"taskGroups":[` + string(sampleSectionJSON("Q-1", "Inbox", map[string]any{"parent": "P-1"})) + `]}`))
	}))
	t.Cleanup(srv.Close)

	s, err := api.NewSession(srv.URL, "test-token-slist", "5s")
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	limit := 10
	offset := 5
	removed := true
	sections, err := s.ListSections(context.Background(), api.SectionListQuery{
		Parent:         "P-1",
		MaxCount:       &limit,
		Offset:         &offset,
		IncludeRemoved: &removed,
	})
	if err != nil {
		t.Fatalf("ListSections: %v", err)
	}
	if gotAuth != "Bearer test-token-slist" {
		t.Fatalf("Authorization = %q", gotAuth)
	}
	if gotPath != "/v2/task-group" {
		t.Fatalf("path = %q", gotPath)
	}
	for _, want := range []string{
		"parent=P-1",
		"maxCount=10",
		"offset=5",
		"includeRemoved=true",
	} {
		if !strings.Contains(gotQuery, want) {
			t.Fatalf("query %q missing %q", gotQuery, want)
		}
	}
	if len(sections) != 1 || sections[0].ID != "Q-1" || sections[0].Title != "Inbox" || sections[0].Parent != "P-1" {
		t.Fatalf("sections=%#v", sections)
	}
}

func TestListSectionsEmptyParentRejected(t *testing.T) {
	t.Parallel()
	s, err := api.NewSession("http://127.0.0.1:9", "test-token-sempty", "5s")
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	_, err = s.ListSections(context.Background(), api.SectionListQuery{Parent: "  "})
	if err == nil {
		t.Fatal("expected error for empty parent")
	}
}

func TestGetCreateUpdateDeleteSectionHappyPaths(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token-scrud" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v2/task-group/Q-get":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(sampleSectionJSON("Q-get", "Got", map[string]any{"parent": "P-1"}))
		case r.Method == http.MethodPost && r.URL.Path == "/v2/task-group":
			body, _ := io.ReadAll(r.Body)
			var m map[string]any
			_ = json.Unmarshal(body, &m)
			if m["title"] != "New" || m["parent"] != "P-new" {
				t.Errorf("create body=%v", m)
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(sampleSectionJSON("Q-new", "New", map[string]any{"parent": "P-new"}))
		case r.Method == http.MethodPatch && r.URL.Path == "/v2/task-group/Q-upd":
			body, _ := io.ReadAll(r.Body)
			var m map[string]any
			_ = json.Unmarshal(body, &m)
			if m["title"] != "Updated" {
				t.Errorf("update title=%v", m["title"])
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(sampleSectionJSON("Q-upd", "Updated", map[string]any{"parent": "P-1"}))
		case r.Method == http.MethodDelete && r.URL.Path == "/v2/task-group/Q-del":
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)

	s, err := api.NewSession(srv.URL, "test-token-scrud", "5s")
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	ctx := context.Background()

	got, err := s.GetSection(ctx, "Q-get")
	if err != nil || got.ID != "Q-get" {
		t.Fatalf("GetSection: %#v %v", got, err)
	}

	title := "New"
	parent := "P-new"
	created, err := s.CreateSection(ctx, api.SectionWriteInput{Title: &title, Parent: &parent})
	if err != nil || created.ID != "Q-new" {
		t.Fatalf("CreateSection: %#v %v", created, err)
	}

	updTitle := "Updated"
	updated, err := s.UpdateSection(ctx, "Q-upd", api.SectionWriteInput{Title: &updTitle})
	if err != nil || updated.Title != "Updated" {
		t.Fatalf("UpdateSection: %#v %v", updated, err)
	}

	if err := s.DeleteSection(ctx, "Q-del"); err != nil {
		t.Fatalf("DeleteSection: %v", err)
	}
}

func TestGetSection404KindNotFound(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`missing`))
	}))
	t.Cleanup(srv.Close)

	s, err := api.NewSession(srv.URL, "test-token-s404", "5s")
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	_, err = s.GetSection(context.Background(), "Q-missing")
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
	if !strings.Contains(ce.Message, "Q-missing") {
		t.Fatalf("message should include entity id: %q", ce.Message)
	}
}

func TestUpdateSectionPartialBody(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		input    api.SectionWriteInput
		wantKeys []string
		wantVals map[string]any
	}{
		{
			name:     "title only",
			input:    api.SectionWriteInput{Title: ptr("T1")},
			wantKeys: []string{"title"},
			wantVals: map[string]any{"title": "T1"},
		},
		{
			name:     "parent only",
			input:    api.SectionWriteInput{Parent: ptr("P-2")},
			wantKeys: []string{"parent"},
			wantVals: map[string]any{"parent": "P-2"},
		},
		{
			name:     "both",
			input:    api.SectionWriteInput{Title: ptr("T2"), Parent: ptr("P-3")},
			wantKeys: []string{"title", "parent"},
			wantVals: map[string]any{"title": "T2", "parent": "P-3"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, _ := io.ReadAll(r.Body)
				var m map[string]any
				_ = json.Unmarshal(body, &m)
				if len(m) != len(tc.wantKeys) {
					t.Errorf("body keys=%v want %v", m, tc.wantKeys)
				}
				for k, v := range tc.wantVals {
					if m[k] != v {
						t.Errorf("%s=%v want %v", k, m[k], v)
					}
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(sampleSectionJSON("Q-partial", "X", nil))
			}))
			t.Cleanup(srv.Close)

			s, err := api.NewSession(srv.URL, "test-token-spartial", "5s")
			if err != nil {
				t.Fatalf("NewSession: %v", err)
			}
			if _, err := s.UpdateSection(context.Background(), "Q-partial", tc.input); err != nil {
				t.Fatalf("UpdateSection: %v", err)
			}
		})
	}
}

func ptr(s string) *string { return &s }
