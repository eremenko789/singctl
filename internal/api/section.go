package api

import (
	"context"
	"fmt"
	"strings"

	"github.com/eremenko789/singctl/internal/apiclient"
)

// Section is the facade view of a SingularityApp task group (section).
type Section struct {
	ID          string
	Title       string
	Parent      string
	ParentOrder float32
	Removed     bool
}

// SectionListQuery holds filters for ListSections.
type SectionListQuery struct {
	Parent         string
	IncludeRemoved *bool
	MaxCount       *int
	Offset         *int
}

// SectionWriteInput holds create/update fields. Pointers mark fields that are set.
type SectionWriteInput struct {
	Title  *string
	Parent *string
}

// ListSections lists sections (task groups) for a parent project via TaskGroupController_list.
func (s *Session) ListSections(ctx context.Context, query SectionListQuery) ([]Section, error) {
	parent := strings.TrimSpace(query.Parent)
	if parent == "" {
		return nil, Classify(fmt.Errorf("parent проекта не задан"))
	}
	if s == nil || s.client == nil {
		return nil, Classify(fmt.Errorf("сеанс API не инициализирован"))
	}
	params := sectionListParamsFromQuery(query, parent)
	resp, err := s.client.TaskGroupControllerListWithResponse(ctx, params)
	if err != nil {
		return nil, Classify(err)
	}
	if err := EnsureSuccess(resp.StatusCode(), resp.Body); err != nil {
		return nil, Classify(err)
	}
	if resp.JSON200 == nil {
		return nil, Classify(fmt.Errorf("пустой ответ list sections"))
	}
	out := make([]Section, 0, len(resp.JSON200.TaskGroups))
	for i := range resp.JSON200.TaskGroups {
		out = append(out, sectionFromDTO(resp.JSON200.TaskGroups[i]))
	}
	return out, nil
}

// GetSection fetches a section by id via TaskGroupController_getById.
func (s *Session) GetSection(ctx context.Context, id string) (Section, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return Section{}, Classify(fmt.Errorf("id секции не задан"))
	}
	if s == nil || s.client == nil {
		return Section{}, Classify(fmt.Errorf("сеанс API не инициализирован"))
	}
	resp, err := s.client.TaskGroupControllerGetByIdWithResponse(ctx, id)
	if err != nil {
		return Section{}, Classify(err, WithEntityID(id))
	}
	if err := EnsureSuccess(resp.StatusCode(), resp.Body); err != nil {
		return Section{}, Classify(err, WithEntityID(id))
	}
	if resp.JSON200 == nil {
		return Section{}, Classify(fmt.Errorf("пустой ответ get section"), WithEntityID(id))
	}
	return sectionFromDTO(*resp.JSON200), nil
}

// CreateSection creates a section via TaskGroupController_create.
func (s *Session) CreateSection(ctx context.Context, in SectionWriteInput) (Section, error) {
	if s == nil || s.client == nil {
		return Section{}, Classify(fmt.Errorf("сеанс API не инициализирован"))
	}
	body, err := sectionCreateDTOFromInput(in)
	if err != nil {
		return Section{}, Classify(err)
	}
	resp, err := s.client.TaskGroupControllerCreateWithResponse(ctx, body)
	if err != nil {
		return Section{}, Classify(err)
	}
	if err := EnsureSuccess(resp.StatusCode(), resp.Body); err != nil {
		return Section{}, Classify(err)
	}
	if resp.JSON200 == nil {
		return Section{}, Classify(fmt.Errorf("пустой ответ create section"))
	}
	return sectionFromDTO(*resp.JSON200), nil
}

// UpdateSection partially updates a section via TaskGroupController_update.
func (s *Session) UpdateSection(ctx context.Context, id string, in SectionWriteInput) (Section, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return Section{}, Classify(fmt.Errorf("id секции не задан"))
	}
	if s == nil || s.client == nil {
		return Section{}, Classify(fmt.Errorf("сеанс API не инициализирован"))
	}
	body := sectionUpdateDTOFromInput(in)
	resp, err := s.client.TaskGroupControllerUpdateWithResponse(ctx, id, body)
	if err != nil {
		return Section{}, Classify(err, WithEntityID(id))
	}
	if err := EnsureSuccess(resp.StatusCode(), resp.Body); err != nil {
		return Section{}, Classify(err, WithEntityID(id))
	}
	if resp.JSON200 == nil {
		return Section{}, Classify(fmt.Errorf("пустой ответ update section"), WithEntityID(id))
	}
	return sectionFromDTO(*resp.JSON200), nil
}

// DeleteSection permanently deletes a section via TaskGroupController_delete.
func (s *Session) DeleteSection(ctx context.Context, id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return Classify(fmt.Errorf("id секции не задан"))
	}
	if s == nil || s.client == nil {
		return Classify(fmt.Errorf("сеанс API не инициализирован"))
	}
	resp, err := s.client.TaskGroupControllerDeleteWithResponse(ctx, id)
	if err != nil {
		return Classify(err, WithEntityID(id))
	}
	if err := EnsureSuccess(resp.StatusCode(), resp.Body); err != nil {
		return Classify(err, WithEntityID(id))
	}
	return nil
}

func sectionListParamsFromQuery(q SectionListQuery, parent string) *apiclient.TaskGroupControllerListParams {
	p := &apiclient.TaskGroupControllerListParams{Parent: &parent}
	if q.IncludeRemoved != nil {
		p.IncludeRemoved = q.IncludeRemoved
	}
	if q.MaxCount != nil {
		v := float32(*q.MaxCount)
		p.MaxCount = &v
	}
	if q.Offset != nil {
		v := float32(*q.Offset)
		p.Offset = &v
	}
	return p
}

func sectionCreateDTOFromInput(in SectionWriteInput) (apiclient.TaskGroupCreateDto, error) {
	dto := apiclient.TaskGroupCreateDto{}
	if in.Title == nil || strings.TrimSpace(*in.Title) == "" {
		return dto, fmt.Errorf("title секции не задан")
	}
	dto.Title = strings.TrimSpace(*in.Title)
	if in.Parent == nil || strings.TrimSpace(*in.Parent) == "" {
		return dto, fmt.Errorf("parent проекта не задан")
	}
	dto.Parent = strings.TrimSpace(*in.Parent)
	return dto, nil
}

func sectionUpdateDTOFromInput(in SectionWriteInput) apiclient.TaskGroupUpdateDto {
	dto := apiclient.TaskGroupUpdateDto{}
	if in.Title != nil {
		t := strings.TrimSpace(*in.Title)
		dto.Title = &t
	}
	if in.Parent != nil {
		p := strings.TrimSpace(*in.Parent)
		dto.Parent = &p
	}
	return dto
}

func sectionFromDTO(d apiclient.TaskGroupResponseDto) Section {
	sec := Section{
		ID:          d.Id,
		Title:       d.Title,
		ParentOrder: d.ParentOrder,
		Removed:     d.Removed,
	}
	if d.Parent != nil {
		sec.Parent = *d.Parent
	}
	return sec
}
