package api

import (
	"context"
	"fmt"
	"strings"

	"github.com/eremenko789/singctl/internal/apiclient"
)

// Task is the facade view of a SingularityApp task (subset for CLI/TUI).
type Task struct {
	ID          string
	Title       string
	ProjectID   string
	Parent      string
	Priority    *float32
	Start       string
	JournalDate string
	DeleteDate  string
	IsNote      bool
}

// TaskListQuery holds filters for ListTasks.
type TaskListQuery struct {
	ProjectID            *string
	Parent               *string
	StartFrom            *string // YYYY-MM-DD
	StartTo              *string // YYYY-MM-DD
	IncludeArchived      *bool
	IncludeRemoved       *bool
	MaxCount             *int
	Offset               *int
	IncludeAllRecurrence *bool
}

// TaskWriteInput holds create/update fields. Pointers mark fields that are set
// (partial update sends only non-nil fields).
type TaskWriteInput struct {
	Title       *string
	ProjectID   *string
	Parent      *string
	Start       *string // YYYY-MM-DD
	Note        *string
	Priority    *int // 0, 1, or 2
	IsNote      *bool
	JournalDate *string // YYYY-MM-DD
	DeleteDate  *string // YYYY-MM-DD; on create → follow-up PATCH
}

// ListTasks lists tasks matching query via TaskController_list.
func (s *Session) ListTasks(ctx context.Context, query TaskListQuery) ([]Task, error) {
	if s == nil || s.client == nil {
		return nil, Classify(fmt.Errorf("сеанс API не инициализирован"))
	}
	params := listParamsFromQuery(query)
	resp, err := s.client.TaskControllerListWithResponse(ctx, params)
	if err != nil {
		return nil, Classify(err)
	}
	if err := EnsureSuccess(resp.StatusCode(), resp.Body); err != nil {
		return nil, Classify(err)
	}
	if resp.JSON200 == nil {
		return nil, Classify(fmt.Errorf("пустой ответ list tasks"))
	}
	out := make([]Task, 0, len(resp.JSON200.Tasks))
	for i := range resp.JSON200.Tasks {
		out = append(out, taskFromDTO(resp.JSON200.Tasks[i]))
	}
	return out, nil
}

// GetTask fetches a task by id via TaskController_getById.
func (s *Session) GetTask(ctx context.Context, id string) (Task, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return Task{}, Classify(fmt.Errorf("id задачи не задан"))
	}
	if s == nil || s.client == nil {
		return Task{}, Classify(fmt.Errorf("сеанс API не инициализирован"))
	}
	resp, err := s.client.TaskControllerGetByIdWithResponse(ctx, id)
	if err != nil {
		return Task{}, Classify(err, WithEntityID(id))
	}
	if err := EnsureSuccess(resp.StatusCode(), resp.Body); err != nil {
		return Task{}, Classify(err, WithEntityID(id))
	}
	if resp.JSON200 == nil {
		return Task{}, Classify(fmt.Errorf("пустой ответ get task"), WithEntityID(id))
	}
	return taskFromDTO(*resp.JSON200), nil
}

// CreateTask creates a task via TaskController_create.
// When DeleteDate is set, issues a follow-up TaskController_update (create DTO has no deleteDate).
func (s *Session) CreateTask(ctx context.Context, in TaskWriteInput) (Task, error) {
	if s == nil || s.client == nil {
		return Task{}, Classify(fmt.Errorf("сеанс API не инициализирован"))
	}
	body := createDTOFromInput(in)
	resp, err := s.client.TaskControllerCreateWithResponse(ctx, body)
	if err != nil {
		return Task{}, Classify(err)
	}
	if err := EnsureSuccess(resp.StatusCode(), resp.Body); err != nil {
		return Task{}, Classify(err)
	}
	if resp.JSON200 == nil {
		return Task{}, Classify(fmt.Errorf("пустой ответ create task"))
	}
	created := taskFromDTO(*resp.JSON200)
	if in.DeleteDate == nil {
		return created, nil
	}
	return s.UpdateTask(ctx, created.ID, TaskWriteInput{DeleteDate: in.DeleteDate})
}

// UpdateTask partially updates a task via TaskController_update.
func (s *Session) UpdateTask(ctx context.Context, id string, in TaskWriteInput) (Task, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return Task{}, Classify(fmt.Errorf("id задачи не задан"))
	}
	if s == nil || s.client == nil {
		return Task{}, Classify(fmt.Errorf("сеанс API не инициализирован"))
	}
	body := updateDTOFromInput(in)
	resp, err := s.client.TaskControllerUpdateWithResponse(ctx, id, body)
	if err != nil {
		return Task{}, Classify(err, WithEntityID(id))
	}
	if err := EnsureSuccess(resp.StatusCode(), resp.Body); err != nil {
		return Task{}, Classify(err, WithEntityID(id))
	}
	if resp.JSON200 == nil {
		return Task{}, Classify(fmt.Errorf("пустой ответ update task"), WithEntityID(id))
	}
	return taskFromDTO(*resp.JSON200), nil
}

// DeleteTask permanently deletes a task via TaskController_delete.
func (s *Session) DeleteTask(ctx context.Context, id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return Classify(fmt.Errorf("id задачи не задан"))
	}
	if s == nil || s.client == nil {
		return Classify(fmt.Errorf("сеанс API не инициализирован"))
	}
	resp, err := s.client.TaskControllerDeleteWithResponse(ctx, id)
	if err != nil {
		return Classify(err, WithEntityID(id))
	}
	if err := EnsureSuccess(resp.StatusCode(), resp.Body); err != nil {
		return Classify(err, WithEntityID(id))
	}
	return nil
}

// ArchiveTask sets journalDate via Update.
func (s *Session) ArchiveTask(ctx context.Context, id string, dateYYYYMMDD string) (Task, error) {
	d := dateYYYYMMDD
	return s.UpdateTask(ctx, id, TaskWriteInput{JournalDate: &d})
}

// TrashTask sets deleteDate via Update.
func (s *Session) TrashTask(ctx context.Context, id string, dateYYYYMMDD string) (Task, error) {
	d := dateYYYYMMDD
	return s.UpdateTask(ctx, id, TaskWriteInput{DeleteDate: &d})
}

func listParamsFromQuery(q TaskListQuery) *apiclient.TaskControllerListParams {
	p := &apiclient.TaskControllerListParams{}
	empty := true
	if q.ProjectID != nil {
		p.ProjectId = q.ProjectID
		empty = false
	}
	if q.Parent != nil {
		p.Parent = q.Parent
		empty = false
	}
	if q.StartFrom != nil {
		p.StartDateFrom = q.StartFrom
		empty = false
	}
	if q.StartTo != nil {
		p.StartDateTo = q.StartTo
		empty = false
	}
	if q.IncludeArchived != nil {
		p.IncludeArchived = q.IncludeArchived
		empty = false
	}
	if q.IncludeRemoved != nil {
		p.IncludeRemoved = q.IncludeRemoved
		empty = false
	}
	if q.IncludeAllRecurrence != nil {
		p.IncludeAllRecurrenceInstances = q.IncludeAllRecurrence
		empty = false
	}
	if q.MaxCount != nil {
		v := float32(*q.MaxCount)
		p.MaxCount = &v
		empty = false
	}
	if q.Offset != nil {
		v := float32(*q.Offset)
		p.Offset = &v
		empty = false
	}
	if empty {
		return nil
	}
	return p
}

func createDTOFromInput(in TaskWriteInput) apiclient.TaskCreateDto {
	dto := apiclient.TaskCreateDto{}
	if in.Title != nil {
		dto.Title = *in.Title
	}
	if in.ProjectID != nil {
		dto.ProjectId = in.ProjectID
	}
	if in.Parent != nil {
		dto.Parent = in.Parent
	}
	if in.Start != nil {
		dto.Start = in.Start
	}
	if in.Note != nil {
		dto.Note = in.Note
	}
	if in.IsNote != nil {
		dto.IsNote = in.IsNote
	}
	if in.JournalDate != nil {
		dto.JournalDate = in.JournalDate
	}
	if in.Priority != nil {
		p := apiclient.TaskPriority(*in.Priority)
		dto.Priority = &p
	}
	return dto
}

func updateDTOFromInput(in TaskWriteInput) apiclient.TaskUpdateDto {
	dto := apiclient.TaskUpdateDto{}
	if in.Title != nil {
		dto.Title = in.Title
	}
	if in.ProjectID != nil {
		dto.ProjectId = in.ProjectID
	}
	if in.Parent != nil {
		dto.Parent = in.Parent
	}
	if in.Start != nil {
		dto.Start = in.Start
	}
	if in.Note != nil {
		dto.Note = in.Note
	}
	if in.IsNote != nil {
		dto.IsNote = in.IsNote
	}
	if in.JournalDate != nil {
		dto.JournalDate = in.JournalDate
	}
	if in.DeleteDate != nil {
		dto.DeleteDate = in.DeleteDate
	}
	if in.Priority != nil {
		p := apiclient.TaskUpdateDtoPriority(*in.Priority)
		dto.Priority = &p
	}
	return dto
}

func taskFromDTO(d apiclient.TaskResponseDto) Task {
	t := Task{
		ID:          d.Id,
		Title:       d.Title,
		ProjectID:   d.ProjectId,
		Parent:      d.Parent,
		Start:       d.Start,
		JournalDate: d.JournalDate,
		DeleteDate:  d.DeleteDate,
		IsNote:      d.IsNote,
	}
	p := d.Priority
	t.Priority = &p
	return t
}
