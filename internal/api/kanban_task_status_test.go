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

func sampleKanbanLinkJSON(id, taskID, statusID string, order float32) []byte {
	m := map[string]any{
		"id":              id,
		"taskId":          taskID,
		"statusId":        statusID,
		"kanbanOrder":     order,
		"removed":         false,
		"modificatedDate": "1584530599718",
		"modificated":     map[string]any{},
	}
	b, err := json.Marshal(m)
	if err != nil {
		panic(err)
	}
	return b
}

func TestListKanbanLinksMapsFiltersOnly(t *testing.T) {
	t.Parallel()
	var gotPath, gotQuery, gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"kanbanTaskStatuses":[` + string(sampleKanbanLinkJSON("KTS-1", "T-1", "KS-1", 1)) + `]}`))
	}))
	t.Cleanup(srv.Close)

	s, err := api.NewSession(srv.URL, "test-token-kts-list", "5s")
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	links, err := s.ListKanbanLinks(context.Background(), api.KanbanLinkListQuery{
		TaskID:   "T-1",
		StatusID: "KS-1",
	})
	if err != nil {
		t.Fatalf("ListKanbanLinks: %v", err)
	}
	if gotAuth != "Bearer test-token-kts-list" {
		t.Fatalf("Authorization = %q", gotAuth)
	}
	if gotPath != "/v2/kanban-task-status" {
		t.Fatalf("path = %q", gotPath)
	}
	if !strings.Contains(gotQuery, "taskId=T-1") || !strings.Contains(gotQuery, "statusId=KS-1") {
		t.Fatalf("query %q missing filters", gotQuery)
	}
	for _, bad := range []string{"maxCount=", "offset=", "includeRemoved="} {
		if strings.Contains(gotQuery, bad) {
			t.Fatalf("query %q must not include %q", gotQuery, bad)
		}
	}
	if len(links) != 1 || links[0].ID != "KTS-1" || links[0].TaskID != "T-1" || links[0].StatusID != "KS-1" {
		t.Fatalf("links=%#v", links)
	}
}

func TestKanbanGetCreateUpdateDeleteHappyPaths(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token-kts-crud" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v2/kanban-task-status/KTS-get":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(sampleKanbanLinkJSON("KTS-get", "T-1", "KS-1", 2))
		case r.Method == http.MethodPost && r.URL.Path == "/v2/kanban-task-status":
			body, _ := io.ReadAll(r.Body)
			var m map[string]any
			_ = json.Unmarshal(body, &m)
			if m["taskId"] != "T-1" || m["statusId"] != "KS-2" {
				t.Errorf("create body=%v", m)
			}
			if _, hasExt := m["externalId"]; hasExt {
				t.Errorf("create must not send externalId: %v", m)
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(sampleKanbanLinkJSON("KTS-new", "T-1", "KS-2", 0))
		case r.Method == http.MethodPatch && r.URL.Path == "/v2/kanban-task-status/KTS-upd":
			body, _ := io.ReadAll(r.Body)
			var m map[string]any
			_ = json.Unmarshal(body, &m)
			if m["statusId"] != "KS-3" {
				t.Errorf("update statusId=%v", m["statusId"])
			}
			if order, ok := m["kanbanOrder"].(float64); !ok || order != 5 {
				t.Errorf("update kanbanOrder=%v", m["kanbanOrder"])
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(sampleKanbanLinkJSON("KTS-upd", "T-1", "KS-3", 5))
		case r.Method == http.MethodDelete && r.URL.Path == "/v2/kanban-task-status/KTS-del":
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)

	s, err := api.NewSession(srv.URL, "test-token-kts-crud", "5s")
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	ctx := context.Background()

	got, err := s.GetKanbanLink(ctx, "KTS-get")
	if err != nil || got.ID != "KTS-get" {
		t.Fatalf("GetKanbanLink: %#v %v", got, err)
	}

	taskID := "T-1"
	statusID := "KS-2"
	created, err := s.CreateKanbanLink(ctx, api.KanbanLinkWriteInput{TaskID: &taskID, StatusID: &statusID})
	if err != nil || created.ID != "KTS-new" {
		t.Fatalf("CreateKanbanLink: %#v %v", created, err)
	}

	col := "KS-3"
	order := float32(5)
	updated, err := s.UpdateKanbanLink(ctx, "KTS-upd", api.KanbanLinkWriteInput{StatusID: &col, KanbanOrder: &order})
	if err != nil || updated.StatusID != "KS-3" || updated.KanbanOrder != 5 {
		t.Fatalf("UpdateKanbanLink: %#v %v", updated, err)
	}

	if err := s.DeleteKanbanLink(ctx, "KTS-del"); err != nil {
		t.Fatalf("DeleteKanbanLink: %v", err)
	}
}

func TestGetKanbanLink404KindNotFound(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`missing`))
	}))
	t.Cleanup(srv.Close)

	s, err := api.NewSession(srv.URL, "test-token-kts-404", "5s")
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	_, err = s.GetKanbanLink(context.Background(), "KTS-missing")
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
	if !strings.Contains(ce.Message, "KTS-missing") {
		t.Fatalf("message should include entity id: %q", ce.Message)
	}
}

func TestCreateKanbanLinkOptionalOrderNoUniquenessList(t *testing.T) {
	t.Parallel()
	var methods []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		methods = append(methods, r.Method+" "+r.URL.Path)
		body, _ := io.ReadAll(r.Body)
		var m map[string]any
		_ = json.Unmarshal(body, &m)
		if m["kanbanOrder"] != float64(3) {
			t.Errorf("kanbanOrder=%v", m["kanbanOrder"])
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(sampleKanbanLinkJSON("KTS-ord", "T-1", "KS-1", 3))
	}))
	t.Cleanup(srv.Close)

	s, err := api.NewSession(srv.URL, "test-token-kts-ord", "5s")
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	taskID, statusID := "T-1", "KS-1"
	order := float32(3)
	_, err = s.CreateKanbanLink(context.Background(), api.KanbanLinkWriteInput{
		TaskID: &taskID, StatusID: &statusID, KanbanOrder: &order,
	})
	if err != nil {
		t.Fatalf("CreateKanbanLink: %v", err)
	}
	if len(methods) != 1 || methods[0] != "POST /v2/kanban-task-status" {
		t.Fatalf("create must POST only, got %v", methods)
	}
}

func TestMoveTaskToKanbanBranches(t *testing.T) {
	t.Parallel()

	t.Run("zero_links_create", func(t *testing.T) {
		var sawCreateBody map[string]any
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			switch {
			case r.Method == http.MethodGet && r.URL.Path == "/v2/kanban-task-status":
				if r.URL.Query().Get("taskId") != "T-0" {
					t.Errorf("taskId=%q", r.URL.Query().Get("taskId"))
				}
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"kanbanTaskStatuses":[]}`))
			case r.Method == http.MethodPost && r.URL.Path == "/v2/kanban-task-status":
				body, _ := io.ReadAll(r.Body)
				_ = json.Unmarshal(body, &sawCreateBody)
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(sampleKanbanLinkJSON("KTS-c", "T-0", "KS-new", 0))
			default:
				t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
				w.WriteHeader(http.StatusNotFound)
			}
		}))
		t.Cleanup(srv.Close)
		s, err := api.NewSession(srv.URL, "test-token-kts-move0", "5s")
		if err != nil {
			t.Fatalf("NewSession: %v", err)
		}
		link, err := s.MoveTaskToKanban(context.Background(), "T-0", "KS-new")
		if err != nil || link.ID != "KTS-c" {
			t.Fatalf("Move: %#v %v", link, err)
		}
		if sawCreateBody["taskId"] != "T-0" || sawCreateBody["statusId"] != "KS-new" {
			t.Fatalf("create body=%v", sawCreateBody)
		}
		if _, hasOrder := sawCreateBody["kanbanOrder"]; hasOrder {
			t.Fatalf("move create must omit kanbanOrder: %v", sawCreateBody)
		}
	})

	t.Run("one_link_update_same_column", func(t *testing.T) {
		var patchBody map[string]any
		var patched bool
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			switch {
			case r.Method == http.MethodGet && r.URL.Path == "/v2/kanban-task-status":
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"kanbanTaskStatuses":[` + string(sampleKanbanLinkJSON("KTS-1", "T-1", "KS-same", 1)) + `]}`))
			case r.Method == http.MethodPatch && r.URL.Path == "/v2/kanban-task-status/KTS-1":
				patched = true
				body, _ := io.ReadAll(r.Body)
				_ = json.Unmarshal(body, &patchBody)
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(sampleKanbanLinkJSON("KTS-1", "T-1", "KS-same", 1))
			default:
				t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
				w.WriteHeader(http.StatusNotFound)
			}
		}))
		t.Cleanup(srv.Close)
		s, err := api.NewSession(srv.URL, "test-token-kts-move1", "5s")
		if err != nil {
			t.Fatalf("NewSession: %v", err)
		}
		link, err := s.MoveTaskToKanban(context.Background(), "T-1", "KS-same")
		if err != nil || link.ID != "KTS-1" {
			t.Fatalf("Move: %#v %v", link, err)
		}
		if !patched {
			t.Fatal("expected PATCH even for same column")
		}
		if patchBody["statusId"] != "KS-same" {
			t.Fatalf("patch body=%v", patchBody)
		}
		if _, hasOrder := patchBody["kanbanOrder"]; hasOrder {
			t.Fatalf("move update must omit kanbanOrder: %v", patchBody)
		}
		if _, hasTask := patchBody["taskId"]; hasTask {
			t.Fatalf("move update must omit taskId: %v", patchBody)
		}
	})

	t.Run("many_links_error", func(t *testing.T) {
		var wrote bool
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodPost || r.Method == http.MethodPatch {
				wrote = true
			}
			w.Header().Set("Content-Type", "application/json")
			if r.Method == http.MethodGet && r.URL.Path == "/v2/kanban-task-status" {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"kanbanTaskStatuses":[` +
					string(sampleKanbanLinkJSON("KTS-a", "T-m", "KS-1", 1)) + `,` +
					string(sampleKanbanLinkJSON("KTS-b", "T-m", "KS-2", 2)) + `]}`))
				return
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
		}))
		t.Cleanup(srv.Close)
		s, err := api.NewSession(srv.URL, "test-token-kts-moveN", "5s")
		if err != nil {
			t.Fatalf("NewSession: %v", err)
		}
		_, err = s.MoveTaskToKanban(context.Background(), "T-m", "KS-3")
		if err == nil {
			t.Fatal("expected error")
		}
		var ce *api.ClassifiedError
		if !errors.As(err, &ce) || ce.Kind != api.KindValidation {
			t.Fatalf("want KindValidation, got %v", err)
		}
		if !strings.Contains(ce.Message, "kanban list") && !strings.Contains(ce.Message, "task kanban") {
			t.Fatalf("message should hint kanban list/update: %q", ce.Message)
		}
		if wrote {
			t.Fatal("must not write when ambiguous")
		}
	})
}
