package api

import (
	"context"
	"fmt"
	"strings"

	"github.com/eremenko789/singctl/internal/apiclient"
)

// ChecklistItem is the facade view of a SingularityApp checklist item.
type ChecklistItem struct {
	ID          string
	Title       string
	Done        bool
	Parent      string
	ParentOrder float32
}

// ChecklistListQuery holds filters for ListChecklistItems (F09: parent only).
type ChecklistListQuery struct {
	Parent string
}

// ChecklistWriteInput holds create/update fields. Pointers mark fields that are set
// (partial update sends only non-nil fields). Parent is required on create.
type ChecklistWriteInput struct {
	Parent *string
	Title  *string
	Done   *bool
}

// ListChecklistItems lists checklist items matching query via ChecklistItemController_list.
func (s *Session) ListChecklistItems(ctx context.Context, query ChecklistListQuery) ([]ChecklistItem, error) {
	if s == nil || s.client == nil {
		return nil, Classify(fmt.Errorf("сеанс API не инициализирован"))
	}
	parent := strings.TrimSpace(query.Parent)
	if parent == "" {
		return nil, Classify(fmt.Errorf("parent задачи не задан"))
	}
	params := &apiclient.ChecklistItemControllerListParams{Parent: &parent}
	resp, err := s.client.ChecklistItemControllerListWithResponse(ctx, params)
	if err != nil {
		return nil, Classify(err)
	}
	if err := EnsureSuccess(resp.StatusCode(), resp.Body); err != nil {
		return nil, Classify(err)
	}
	if resp.JSON200 == nil {
		return nil, Classify(fmt.Errorf("пустой ответ list checklist items"))
	}
	out := make([]ChecklistItem, 0, len(resp.JSON200.ChecklistItems))
	for i := range resp.JSON200.ChecklistItems {
		out = append(out, checklistItemFromDTO(resp.JSON200.ChecklistItems[i]))
	}
	return out, nil
}

// GetChecklistItem fetches a checklist item by id via ChecklistItemController_getById.
func (s *Session) GetChecklistItem(ctx context.Context, id string) (ChecklistItem, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return ChecklistItem{}, Classify(fmt.Errorf("id пункта чек-листа не задан"))
	}
	if s == nil || s.client == nil {
		return ChecklistItem{}, Classify(fmt.Errorf("сеанс API не инициализирован"))
	}
	resp, err := s.client.ChecklistItemControllerGetByIdWithResponse(ctx, id)
	if err != nil {
		return ChecklistItem{}, Classify(err, WithEntityID(id))
	}
	if err := EnsureSuccess(resp.StatusCode(), resp.Body); err != nil {
		return ChecklistItem{}, Classify(err, WithEntityID(id))
	}
	if resp.JSON200 == nil {
		return ChecklistItem{}, Classify(fmt.Errorf("пустой ответ get checklist item"), WithEntityID(id))
	}
	return checklistItemFromDTO(*resp.JSON200), nil
}

// CreateChecklistItem creates a checklist item via ChecklistItemController_create.
func (s *Session) CreateChecklistItem(ctx context.Context, in ChecklistWriteInput) (ChecklistItem, error) {
	if s == nil || s.client == nil {
		return ChecklistItem{}, Classify(fmt.Errorf("сеанс API не инициализирован"))
	}
	body := checklistCreateDTOFromInput(in)
	resp, err := s.client.ChecklistItemControllerCreateWithResponse(ctx, body)
	if err != nil {
		return ChecklistItem{}, Classify(err)
	}
	if err := EnsureSuccess(resp.StatusCode(), resp.Body); err != nil {
		return ChecklistItem{}, Classify(err)
	}
	if resp.JSON200 == nil {
		return ChecklistItem{}, Classify(fmt.Errorf("пустой ответ create checklist item"))
	}
	return checklistItemFromDTO(*resp.JSON200), nil
}

// UpdateChecklistItem partially updates a checklist item via ChecklistItemController_update.
func (s *Session) UpdateChecklistItem(ctx context.Context, id string, in ChecklistWriteInput) (ChecklistItem, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return ChecklistItem{}, Classify(fmt.Errorf("id пункта чек-листа не задан"))
	}
	if s == nil || s.client == nil {
		return ChecklistItem{}, Classify(fmt.Errorf("сеанс API не инициализирован"))
	}
	body := checklistUpdateDTOFromInput(in)
	resp, err := s.client.ChecklistItemControllerUpdateWithResponse(ctx, id, body)
	if err != nil {
		return ChecklistItem{}, Classify(err, WithEntityID(id))
	}
	if err := EnsureSuccess(resp.StatusCode(), resp.Body); err != nil {
		return ChecklistItem{}, Classify(err, WithEntityID(id))
	}
	if resp.JSON200 == nil {
		return ChecklistItem{}, Classify(fmt.Errorf("пустой ответ update checklist item"), WithEntityID(id))
	}
	return checklistItemFromDTO(*resp.JSON200), nil
}

// DeleteChecklistItem permanently deletes a checklist item via ChecklistItemController_delete.
func (s *Session) DeleteChecklistItem(ctx context.Context, id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return Classify(fmt.Errorf("id пункта чек-листа не задан"))
	}
	if s == nil || s.client == nil {
		return Classify(fmt.Errorf("сеанс API не инициализирован"))
	}
	resp, err := s.client.ChecklistItemControllerDeleteWithResponse(ctx, id)
	if err != nil {
		return Classify(err, WithEntityID(id))
	}
	if err := EnsureSuccess(resp.StatusCode(), resp.Body); err != nil {
		return Classify(err, WithEntityID(id))
	}
	return nil
}

func checklistCreateDTOFromInput(in ChecklistWriteInput) apiclient.ChecklistItemCreateDto {
	dto := apiclient.ChecklistItemCreateDto{}
	if in.Parent != nil {
		dto.Parent = *in.Parent
	}
	if in.Title != nil {
		dto.Title = *in.Title
	}
	if in.Done != nil {
		dto.Done = in.Done
	}
	return dto
}

func checklistUpdateDTOFromInput(in ChecklistWriteInput) apiclient.ChecklistItemUpdateDto {
	dto := apiclient.ChecklistItemUpdateDto{}
	if in.Title != nil {
		dto.Title = in.Title
	}
	if in.Done != nil {
		dto.Done = in.Done
	}
	return dto
}

func checklistItemFromDTO(d apiclient.ChecklistItemResponseDto) ChecklistItem {
	return ChecklistItem{
		ID:          d.Id,
		Title:       d.Title,
		Done:        d.Done,
		Parent:      d.Parent,
		ParentOrder: d.ParentOrder,
	}
}
