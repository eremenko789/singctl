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

func sampleTaskJSON(id, title string, extra map[string]any) []byte {
	m := map[string]any{
		"id":                    id,
		"title":                 title,
		"projectId":             "P-1",
		"parent":                "",
		"priority":              1,
		"start":                 "2026-07-01",
		"journalDate":           "",
		"deleteDate":            "",
		"isNote":                false,
		"alarmNotify":           false,
		"checked":               0,
		"complete":              0,
		"completeLast":          "",
		"createdDate":           "2026-07-01",
		"crypted":               "",
		"deadline":              "",
		"deadlineNotifyReaded":  false,
		"deferred":              false,
		"externalId":            "",
		"group":                 "",
		"integrationItemId":     "",
		"modificated":           map[string]any{},
		"modificatedDate":       "2026-07-01",
		"note":                  "",
		"notifies":              []any{},
		"notify":                0,
		"parentOrder":           0,
		"pomodoroCount":         0,
		"pomodoroTotalTime":     0,
		"recurrenceGeneratorId": "",
		"removed":               false,
		"scheduleOrder":         0,
		"seenToday":             "",
		"showInBasket":          false,
		"startNotifiesReaded":   []any{},
		"startNotifyReaded":     false,
		"state":                 1,
		"tags":                  []any{},
		"timeLength":            0,
		"useTime":               false,
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

func TestListTasksMapsFiltersAndReturnsTasks(t *testing.T) {
	t.Parallel()
	var gotPath, gotQuery, gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"tasks":[` + string(sampleTaskJSON("T-1", "Buy milk", nil)) + `]}`))
	}))
	t.Cleanup(srv.Close)

	s, err := api.NewSession(srv.URL, "test-token-list", "5s")
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	project := "P-abc"
	parent := "T-parent"
	from := "2026-07-01"
	to := "2026-07-31"
	limit := 10
	offset := 5
	archived := true
	removed := false
	allRec := true
	tasks, err := s.ListTasks(context.Background(), api.TaskListQuery{
		ProjectID:            &project,
		Parent:               &parent,
		StartFrom:            &from,
		StartTo:              &to,
		MaxCount:             &limit,
		Offset:               &offset,
		IncludeArchived:      &archived,
		IncludeRemoved:       &removed,
		IncludeAllRecurrence: &allRec,
	})
	if err != nil {
		t.Fatalf("ListTasks: %v", err)
	}
	if gotAuth != "Bearer test-token-list" {
		t.Fatalf("Authorization = %q", gotAuth)
	}
	if gotPath != "/v2/task" {
		t.Fatalf("path = %q", gotPath)
	}
	for _, want := range []string{
		"projectId=P-abc",
		"parent=T-parent",
		"startDateFrom=2026-07-01",
		"startDateTo=2026-07-31",
		"maxCount=10",
		"offset=5",
		"includeArchived=true",
		"includeRemoved=false",
		"includeAllRecurrenceInstances=true",
	} {
		if !strings.Contains(gotQuery, want) {
			t.Fatalf("query %q missing %q", gotQuery, want)
		}
	}
	if len(tasks) != 1 || tasks[0].ID != "T-1" || tasks[0].Title != "Buy milk" {
		t.Fatalf("tasks=%#v", tasks)
	}
}

func TestGetCreateUpdateDeleteHappyPaths(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token-crud" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v2/task/T-get":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(sampleTaskJSON("T-get", "Got", nil))
		case r.Method == http.MethodPost && r.URL.Path == "/v2/task":
			body, _ := io.ReadAll(r.Body)
			var m map[string]any
			_ = json.Unmarshal(body, &m)
			if m["title"] != "New" {
				t.Errorf("create title=%v", m["title"])
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(sampleTaskJSON("T-new", "New", nil))
		case r.Method == http.MethodPatch && r.URL.Path == "/v2/task/T-upd":
			body, _ := io.ReadAll(r.Body)
			var m map[string]any
			_ = json.Unmarshal(body, &m)
			if m["title"] != "Updated" {
				t.Errorf("update title=%v", m["title"])
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(sampleTaskJSON("T-upd", "Updated", nil))
		case r.Method == http.MethodDelete && r.URL.Path == "/v2/task/T-del":
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)

	s, err := api.NewSession(srv.URL, "test-token-crud", "5s")
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	ctx := context.Background()

	got, err := s.GetTask(ctx, "T-get")
	if err != nil || got.ID != "T-get" {
		t.Fatalf("GetTask: %#v %v", got, err)
	}

	title := "New"
	created, err := s.CreateTask(ctx, api.TaskWriteInput{Title: &title})
	if err != nil || created.ID != "T-new" {
		t.Fatalf("CreateTask: %#v %v", created, err)
	}

	updTitle := "Updated"
	updated, err := s.UpdateTask(ctx, "T-upd", api.TaskWriteInput{Title: &updTitle})
	if err != nil || updated.Title != "Updated" {
		t.Fatalf("UpdateTask: %#v %v", updated, err)
	}

	if err := s.DeleteTask(ctx, "T-del"); err != nil {
		t.Fatalf("DeleteTask: %v", err)
	}
}

func TestGetTask404KindNotFound(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`missing`))
	}))
	t.Cleanup(srv.Close)

	s, err := api.NewSession(srv.URL, "test-token-404", "5s")
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	_, err = s.GetTask(context.Background(), "T-missing")
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
	if !strings.Contains(ce.Message, "T-missing") {
		t.Fatalf("message should include entity id: %q", ce.Message)
	}
}

func TestCreateTaskWithDeleteDatePostsThenPatches(t *testing.T) {
	t.Parallel()
	var methods []string
	var paths []string
	var bodies []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		methods = append(methods, r.Method)
		paths = append(paths, r.URL.Path)
		b, _ := io.ReadAll(r.Body)
		bodies = append(bodies, string(b))
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v2/task":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(sampleTaskJSON("T-created", "Trash me", nil))
		case r.Method == http.MethodPatch && r.URL.Path == "/v2/task/T-created":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(sampleTaskJSON("T-created", "Trash me", map[string]any{"deleteDate": "2026-07-16"}))
		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)

	s, err := api.NewSession(srv.URL, "test-token-delete-date", "5s")
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	title := "Trash me"
	del := "2026-07-16"
	task, err := s.CreateTask(context.Background(), api.TaskWriteInput{Title: &title, DeleteDate: &del})
	if err != nil {
		t.Fatalf("CreateTask: %v", err)
	}
	if task.DeleteDate != "2026-07-16" {
		t.Fatalf("DeleteDate=%q", task.DeleteDate)
	}
	if len(methods) != 2 || methods[0] != http.MethodPost || methods[1] != http.MethodPatch {
		t.Fatalf("methods=%v", methods)
	}
	if paths[0] != "/v2/task" || paths[1] != "/v2/task/T-created" {
		t.Fatalf("paths=%v", paths)
	}
	if strings.Contains(bodies[0], "deleteDate") {
		t.Fatalf("create body must not include deleteDate: %s", bodies[0])
	}
	if !strings.Contains(bodies[1], `"deleteDate":"2026-07-16"`) && !strings.Contains(bodies[1], `"deleteDate": "2026-07-16"`) {
		t.Fatalf("patch body missing deleteDate: %s", bodies[1])
	}
}

func TestArchiveAndTrashPatchOnlyDates(t *testing.T) {
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
		case "/v2/task/T-arch":
			n.Add(1)
			if _, ok := m["journalDate"]; !ok || len(m) != 1 {
				t.Errorf("archive body=%v want only journalDate", m)
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(sampleTaskJSON("T-arch", "A", map[string]any{"journalDate": m["journalDate"]}))
		case "/v2/task/T-trash":
			n.Add(1)
			if _, ok := m["deleteDate"]; !ok || len(m) != 1 {
				t.Errorf("trash body=%v want only deleteDate", m)
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(sampleTaskJSON("T-trash", "B", map[string]any{"deleteDate": m["deleteDate"]}))
		default:
			t.Errorf("path=%s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)

	s, err := api.NewSession(srv.URL, "test-token-arch-trash", "5s")
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	ctx := context.Background()
	a, err := s.ArchiveTask(ctx, "T-arch", "2026-07-10")
	if err != nil || a.JournalDate != "2026-07-10" {
		t.Fatalf("ArchiveTask: %#v %v", a, err)
	}
	tr, err := s.TrashTask(ctx, "T-trash", "2026-07-11")
	if err != nil || tr.DeleteDate != "2026-07-11" {
		t.Fatalf("TrashTask: %#v %v", tr, err)
	}
	if n.Load() != 2 {
		t.Fatalf("hits=%d", n.Load())
	}
}
