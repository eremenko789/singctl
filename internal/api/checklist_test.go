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

func sampleChecklistJSON(id, title, parent string, done bool, extra map[string]any) []byte {
	m := map[string]any{
		"id":              id,
		"title":           title,
		"parent":          parent,
		"parentOrder":     1,
		"done":            done,
		"removed":         false,
		"modificatedDate": "1584530599718",
		"crypted":         "",
		"modificated":     map[string]any{},
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

func TestListChecklistItemsMapsParentOnly(t *testing.T) {
	t.Parallel()
	var gotPath, gotQuery, gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"checklistItems":[` + string(sampleChecklistJSON("C-1", "Buy milk", "T-1", false, nil)) + `]}`))
	}))
	t.Cleanup(srv.Close)

	s, err := api.NewSession(srv.URL, "test-token-cl-list", "5s")
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	items, err := s.ListChecklistItems(context.Background(), api.ChecklistListQuery{Parent: "T-1"})
	if err != nil {
		t.Fatalf("ListChecklistItems: %v", err)
	}
	if gotAuth != "Bearer test-token-cl-list" {
		t.Fatalf("Authorization = %q", gotAuth)
	}
	if gotPath != "/v2/checklist-item" {
		t.Fatalf("path = %q", gotPath)
	}
	if !strings.Contains(gotQuery, "parent=T-1") {
		t.Fatalf("query %q missing parent=T-1", gotQuery)
	}
	for _, bad := range []string{"maxCount=", "offset=", "includeRemoved="} {
		if strings.Contains(gotQuery, bad) {
			t.Fatalf("query %q must not include %q", gotQuery, bad)
		}
	}
	if len(items) != 1 || items[0].ID != "C-1" || items[0].Title != "Buy milk" || items[0].Parent != "T-1" {
		t.Fatalf("items=%#v", items)
	}
}

func TestChecklistGetCreateUpdateDeleteHappyPaths(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token-cl-crud" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v2/checklist-item/C-get":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(sampleChecklistJSON("C-get", "Got", "T-1", false, nil))
		case r.Method == http.MethodPost && r.URL.Path == "/v2/checklist-item":
			body, _ := io.ReadAll(r.Body)
			var m map[string]any
			_ = json.Unmarshal(body, &m)
			if m["parent"] != "T-1" || m["title"] != "New" {
				t.Errorf("create body=%v", m)
			}
			if _, hasOrder := m["parentOrder"]; hasOrder {
				t.Errorf("create must not send parentOrder: %v", m)
			}
			if _, hasCrypted := m["crypted"]; hasCrypted {
				t.Errorf("create must not send crypted: %v", m)
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(sampleChecklistJSON("C-new", "New", "T-1", false, nil))
		case r.Method == http.MethodPatch && r.URL.Path == "/v2/checklist-item/C-upd":
			body, _ := io.ReadAll(r.Body)
			var m map[string]any
			_ = json.Unmarshal(body, &m)
			if m["title"] != "Updated" {
				t.Errorf("update title=%v", m["title"])
			}
			if done, ok := m["done"].(bool); !ok || !done {
				t.Errorf("update done=%v", m["done"])
			}
			if _, hasParent := m["parent"]; hasParent {
				t.Errorf("update must not send parent: %v", m)
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(sampleChecklistJSON("C-upd", "Updated", "T-1", true, nil))
		case r.Method == http.MethodDelete && r.URL.Path == "/v2/checklist-item/C-del":
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)

	s, err := api.NewSession(srv.URL, "test-token-cl-crud", "5s")
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	ctx := context.Background()

	got, err := s.GetChecklistItem(ctx, "C-get")
	if err != nil || got.ID != "C-get" {
		t.Fatalf("GetChecklistItem: %#v %v", got, err)
	}

	parent := "T-1"
	title := "New"
	created, err := s.CreateChecklistItem(ctx, api.ChecklistWriteInput{Parent: &parent, Title: &title})
	if err != nil || created.ID != "C-new" {
		t.Fatalf("CreateChecklistItem: %#v %v", created, err)
	}

	updTitle := "Updated"
	done := true
	updated, err := s.UpdateChecklistItem(ctx, "C-upd", api.ChecklistWriteInput{Title: &updTitle, Done: &done})
	if err != nil || !updated.Done || updated.Title != "Updated" {
		t.Fatalf("UpdateChecklistItem: %#v %v", updated, err)
	}

	if err := s.DeleteChecklistItem(ctx, "C-del"); err != nil {
		t.Fatalf("DeleteChecklistItem: %v", err)
	}
}

func TestGetChecklistItem404KindNotFound(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`missing`))
	}))
	t.Cleanup(srv.Close)

	s, err := api.NewSession(srv.URL, "test-token-cl-404", "5s")
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	_, err = s.GetChecklistItem(context.Background(), "C-missing")
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
	if !strings.Contains(ce.Message, "C-missing") {
		t.Fatalf("message should include entity id: %q", ce.Message)
	}
}

func TestCreateChecklistItemOptionalDone(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var m map[string]any
		_ = json.Unmarshal(body, &m)
		if m["done"] != true {
			t.Errorf("done=%v", m["done"])
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(sampleChecklistJSON("C-done", "Done item", "T-1", true, nil))
	}))
	t.Cleanup(srv.Close)

	s, err := api.NewSession(srv.URL, "test-token-cl-done", "5s")
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	parent := "T-1"
	title := "Done item"
	done := true
	item, err := s.CreateChecklistItem(context.Background(), api.ChecklistWriteInput{
		Parent: &parent, Title: &title, Done: &done,
	})
	if err != nil || !item.Done {
		t.Fatalf("CreateChecklistItem: %#v %v", item, err)
	}
}
