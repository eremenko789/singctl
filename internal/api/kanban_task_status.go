package api

import (
	"context"
	"fmt"
	"strings"

	"github.com/eremenko789/singctl/internal/apiclient"
)

// KanbanLink is the facade view of a task↔kanban-status link.
type KanbanLink struct {
	ID          string
	TaskID      string
	StatusID    string
	KanbanOrder float32
}

// KanbanLinkListQuery holds optional filters for ListKanbanLinks (F10).
type KanbanLinkListQuery struct {
	TaskID   string
	StatusID string
}

// KanbanLinkWriteInput holds create/update fields. Pointers mark fields that are set
// (partial update sends only non-nil fields).
type KanbanLinkWriteInput struct {
	TaskID      *string
	StatusID    *string
	KanbanOrder *float32
}

// ListKanbanLinks lists task-kanban links via KanbanTaskStatusController_list.
// Only TaskId/StatusId are set when non-empty; pagination/includeRemoved are omitted.
func (s *Session) ListKanbanLinks(ctx context.Context, query KanbanLinkListQuery) ([]KanbanLink, error) {
	if s == nil || s.client == nil {
		return nil, Classify(fmt.Errorf("сеанс API не инициализирован"))
	}
	params := &apiclient.KanbanTaskStatusControllerListParams{}
	if tid := strings.TrimSpace(query.TaskID); tid != "" {
		params.TaskId = &tid
	}
	if sid := strings.TrimSpace(query.StatusID); sid != "" {
		params.StatusId = &sid
	}
	resp, err := s.client.KanbanTaskStatusControllerListWithResponse(ctx, params)
	if err != nil {
		return nil, Classify(err)
	}
	if err := EnsureSuccess(resp.StatusCode(), resp.Body); err != nil {
		return nil, Classify(err)
	}
	if resp.JSON200 == nil {
		return nil, Classify(fmt.Errorf("пустой ответ list kanban-task-status"))
	}
	out := make([]KanbanLink, 0, len(resp.JSON200.KanbanTaskStatuses))
	for i := range resp.JSON200.KanbanTaskStatuses {
		out = append(out, kanbanLinkFromDTO(resp.JSON200.KanbanTaskStatuses[i]))
	}
	return out, nil
}

// GetKanbanLink fetches a link by id via KanbanTaskStatusController_getById.
func (s *Session) GetKanbanLink(ctx context.Context, id string) (KanbanLink, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return KanbanLink{}, Classify(fmt.Errorf("id канбан-связи не задан"))
	}
	if s == nil || s.client == nil {
		return KanbanLink{}, Classify(fmt.Errorf("сеанс API не инициализирован"))
	}
	resp, err := s.client.KanbanTaskStatusControllerGetByIdWithResponse(ctx, id)
	if err != nil {
		return KanbanLink{}, Classify(err, WithEntityID(id))
	}
	if err := EnsureSuccess(resp.StatusCode(), resp.Body); err != nil {
		return KanbanLink{}, Classify(err, WithEntityID(id))
	}
	if resp.JSON200 == nil {
		return KanbanLink{}, Classify(fmt.Errorf("пустой ответ get kanban-task-status"), WithEntityID(id))
	}
	return kanbanLinkFromDTO(*resp.JSON200), nil
}

// CreateKanbanLink creates a link via KanbanTaskStatusController_create.
// Does not enforce uniqueness; callers may create multiple links for one task.
func (s *Session) CreateKanbanLink(ctx context.Context, in KanbanLinkWriteInput) (KanbanLink, error) {
	if s == nil || s.client == nil {
		return KanbanLink{}, Classify(fmt.Errorf("сеанс API не инициализирован"))
	}
	body := kanbanCreateDTOFromInput(in)
	resp, err := s.client.KanbanTaskStatusControllerCreateWithResponse(ctx, body)
	if err != nil {
		return KanbanLink{}, Classify(err)
	}
	if err := EnsureSuccess(resp.StatusCode(), resp.Body); err != nil {
		return KanbanLink{}, Classify(err)
	}
	if resp.JSON200 == nil {
		return KanbanLink{}, Classify(fmt.Errorf("пустой ответ create kanban-task-status"))
	}
	return kanbanLinkFromDTO(*resp.JSON200), nil
}

// UpdateKanbanLink partially updates a link via KanbanTaskStatusController_update.
func (s *Session) UpdateKanbanLink(ctx context.Context, id string, in KanbanLinkWriteInput) (KanbanLink, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return KanbanLink{}, Classify(fmt.Errorf("id канбан-связи не задан"))
	}
	if s == nil || s.client == nil {
		return KanbanLink{}, Classify(fmt.Errorf("сеанс API не инициализирован"))
	}
	body := kanbanUpdateDTOFromInput(in)
	resp, err := s.client.KanbanTaskStatusControllerUpdateWithResponse(ctx, id, body)
	if err != nil {
		return KanbanLink{}, Classify(err, WithEntityID(id))
	}
	if err := EnsureSuccess(resp.StatusCode(), resp.Body); err != nil {
		return KanbanLink{}, Classify(err, WithEntityID(id))
	}
	if resp.JSON200 == nil {
		return KanbanLink{}, Classify(fmt.Errorf("пустой ответ update kanban-task-status"), WithEntityID(id))
	}
	return kanbanLinkFromDTO(*resp.JSON200), nil
}

// DeleteKanbanLink permanently deletes a link via KanbanTaskStatusController_delete.
func (s *Session) DeleteKanbanLink(ctx context.Context, id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return Classify(fmt.Errorf("id канбан-связи не задан"))
	}
	if s == nil || s.client == nil {
		return Classify(fmt.Errorf("сеанс API не инициализирован"))
	}
	resp, err := s.client.KanbanTaskStatusControllerDeleteWithResponse(ctx, id)
	if err != nil {
		return Classify(err, WithEntityID(id))
	}
	if err := EnsureSuccess(resp.StatusCode(), resp.Body); err != nil {
		return Classify(err, WithEntityID(id))
	}
	return nil
}

// MoveTaskToKanban moves a task into a column: create if no links, update statusId
// if exactly one link (even when already equal), or KindValidation if multiple.
// Does not call GetTask and never sets kanbanOrder.
func (s *Session) MoveTaskToKanban(ctx context.Context, taskID, columnID string) (KanbanLink, error) {
	taskID = strings.TrimSpace(taskID)
	columnID = strings.TrimSpace(columnID)
	if taskID == "" {
		return KanbanLink{}, Classify(fmt.Errorf("id задачи не задан"))
	}
	if columnID == "" {
		return KanbanLink{}, Classify(fmt.Errorf("id колонки не задан"))
	}
	links, err := s.ListKanbanLinks(ctx, KanbanLinkListQuery{TaskID: taskID})
	if err != nil {
		return KanbanLink{}, err
	}
	switch len(links) {
	case 0:
		return s.CreateKanbanLink(ctx, KanbanLinkWriteInput{
			TaskID:   &taskID,
			StatusID: &columnID,
		})
	case 1:
		return s.UpdateKanbanLink(ctx, links[0].ID, KanbanLinkWriteInput{
			StatusID: &columnID,
		})
	default:
		return KanbanLink{}, &ClassifiedError{
			Kind: KindValidation,
			Message: fmt.Sprintf(
				"%s: несколько канбан-связей у задачи %s; используйте task kanban list / update",
				msgValidation, taskID,
			),
		}
	}
}

func kanbanCreateDTOFromInput(in KanbanLinkWriteInput) apiclient.KanbanTaskStatusCreateDto {
	dto := apiclient.KanbanTaskStatusCreateDto{}
	if in.TaskID != nil {
		dto.TaskId = *in.TaskID
	}
	if in.StatusID != nil {
		dto.StatusId = *in.StatusID
	}
	if in.KanbanOrder != nil {
		dto.KanbanOrder = in.KanbanOrder
	}
	return dto
}

func kanbanUpdateDTOFromInput(in KanbanLinkWriteInput) apiclient.KanbanTaskStatusUpdateDto {
	dto := apiclient.KanbanTaskStatusUpdateDto{}
	if in.TaskID != nil {
		dto.TaskId = in.TaskID
	}
	if in.StatusID != nil {
		dto.StatusId = in.StatusID
	}
	if in.KanbanOrder != nil {
		dto.KanbanOrder = in.KanbanOrder
	}
	return dto
}

func kanbanLinkFromDTO(d apiclient.KanbanTaskStatusResponseDto) KanbanLink {
	return KanbanLink{
		ID:          d.Id,
		TaskID:      d.TaskId,
		StatusID:    d.StatusId,
		KanbanOrder: d.KanbanOrder,
	}
}
