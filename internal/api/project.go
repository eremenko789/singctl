package api

import (
	"context"
	"fmt"
	"strings"

	"github.com/eremenko789/singctl/internal/apiclient"
)

// Project is the facade view of a SingularityApp project (subset for CLI/TUI).
type Project struct {
	ID          string
	Title       string
	Emoji       string
	Color       string
	IsNotebook  bool
	Parent      string
	JournalDate string
	DeleteDate  string
}

// ProjectListQuery holds filters for ListProjects.
type ProjectListQuery struct {
	IncludeArchived *bool
	IncludeRemoved  *bool
	MaxCount        *int
	Offset          *int
}

// ProjectWriteInput holds create/update fields. Pointers mark fields that are set
// (partial update sends only non-nil fields). JournalDate/DeleteDate are set only
// via ArchiveProject/TrashProject, not via CLI write flags.
type ProjectWriteInput struct {
	Title       *string
	Note        *string
	IsNotebook  *bool
	Emoji       *string
	Color       *string
	Parent      *string
	JournalDate *string // YYYY-MM-DD; archive helper
	DeleteDate  *string // YYYY-MM-DD; trash helper
}

// ListProjects lists projects matching query via ProjectController_list.
func (s *Session) ListProjects(ctx context.Context, query ProjectListQuery) ([]Project, error) {
	if s == nil || s.client == nil {
		return nil, Classify(fmt.Errorf("сеанс API не инициализирован"))
	}
	params := projectListParamsFromQuery(query)
	resp, err := s.client.ProjectControllerListWithResponse(ctx, params)
	if err != nil {
		return nil, Classify(err)
	}
	if err := EnsureSuccess(resp.StatusCode(), resp.Body); err != nil {
		return nil, Classify(err)
	}
	if resp.JSON200 == nil {
		return nil, Classify(fmt.Errorf("пустой ответ list projects"))
	}
	out := make([]Project, 0, len(resp.JSON200.Projects))
	for i := range resp.JSON200.Projects {
		out = append(out, projectFromDTO(resp.JSON200.Projects[i]))
	}
	return out, nil
}

// GetProject fetches a project by id via ProjectController_getById.
func (s *Session) GetProject(ctx context.Context, id string) (Project, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return Project{}, Classify(fmt.Errorf("id проекта не задан"))
	}
	if s == nil || s.client == nil {
		return Project{}, Classify(fmt.Errorf("сеанс API не инициализирован"))
	}
	resp, err := s.client.ProjectControllerGetByIdWithResponse(ctx, id)
	if err != nil {
		return Project{}, Classify(err, WithEntityID(id))
	}
	if err := EnsureSuccess(resp.StatusCode(), resp.Body); err != nil {
		return Project{}, Classify(err, WithEntityID(id))
	}
	if resp.JSON200 == nil {
		return Project{}, Classify(fmt.Errorf("пустой ответ get project"), WithEntityID(id))
	}
	return projectFromDTO(*resp.JSON200), nil
}

// CreateProject creates a project via ProjectController_create.
// Returns only the nested project from ProjectCreateResponseDto (taskGroup ignored).
func (s *Session) CreateProject(ctx context.Context, in ProjectWriteInput) (Project, error) {
	if s == nil || s.client == nil {
		return Project{}, Classify(fmt.Errorf("сеанс API не инициализирован"))
	}
	body := projectCreateDTOFromInput(in)
	resp, err := s.client.ProjectControllerCreateWithResponse(ctx, body)
	if err != nil {
		return Project{}, Classify(err)
	}
	if err := EnsureSuccess(resp.StatusCode(), resp.Body); err != nil {
		return Project{}, Classify(err)
	}
	if resp.JSON200 == nil {
		return Project{}, Classify(fmt.Errorf("пустой ответ create project"))
	}
	return projectFromDTO(resp.JSON200.Project), nil
}

// UpdateProject partially updates a project via ProjectController_update.
func (s *Session) UpdateProject(ctx context.Context, id string, in ProjectWriteInput) (Project, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return Project{}, Classify(fmt.Errorf("id проекта не задан"))
	}
	if s == nil || s.client == nil {
		return Project{}, Classify(fmt.Errorf("сеанс API не инициализирован"))
	}
	body := projectUpdateDTOFromInput(in)
	resp, err := s.client.ProjectControllerUpdateWithResponse(ctx, id, body)
	if err != nil {
		return Project{}, Classify(err, WithEntityID(id))
	}
	if err := EnsureSuccess(resp.StatusCode(), resp.Body); err != nil {
		return Project{}, Classify(err, WithEntityID(id))
	}
	if resp.JSON200 == nil {
		return Project{}, Classify(fmt.Errorf("пустой ответ update project"), WithEntityID(id))
	}
	return projectFromDTO(*resp.JSON200), nil
}

// DeleteProject permanently deletes a project via ProjectController_delete.
func (s *Session) DeleteProject(ctx context.Context, id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return Classify(fmt.Errorf("id проекта не задан"))
	}
	if s == nil || s.client == nil {
		return Classify(fmt.Errorf("сеанс API не инициализирован"))
	}
	resp, err := s.client.ProjectControllerDeleteWithResponse(ctx, id)
	if err != nil {
		return Classify(err, WithEntityID(id))
	}
	if err := EnsureSuccess(resp.StatusCode(), resp.Body); err != nil {
		return Classify(err, WithEntityID(id))
	}
	return nil
}

// ArchiveProject sets journalDate via Update.
func (s *Session) ArchiveProject(ctx context.Context, id string, dateYYYYMMDD string) (Project, error) {
	d := dateYYYYMMDD
	return s.UpdateProject(ctx, id, ProjectWriteInput{JournalDate: &d})
}

// TrashProject sets deleteDate via Update.
func (s *Session) TrashProject(ctx context.Context, id string, dateYYYYMMDD string) (Project, error) {
	d := dateYYYYMMDD
	return s.UpdateProject(ctx, id, ProjectWriteInput{DeleteDate: &d})
}

func projectListParamsFromQuery(q ProjectListQuery) *apiclient.ProjectControllerListParams {
	p := &apiclient.ProjectControllerListParams{}
	empty := true
	if q.IncludeArchived != nil {
		p.IncludeArchived = q.IncludeArchived
		empty = false
	}
	if q.IncludeRemoved != nil {
		p.IncludeRemoved = q.IncludeRemoved
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

func projectCreateDTOFromInput(in ProjectWriteInput) apiclient.ProjectCreateDto {
	dto := apiclient.ProjectCreateDto{}
	if in.Title != nil {
		dto.Title = *in.Title
	}
	if in.Note != nil {
		dto.Note = in.Note
	}
	if in.IsNotebook != nil {
		dto.IsNotebook = in.IsNotebook
	}
	if in.Emoji != nil {
		dto.Emoji = in.Emoji
	}
	if in.Color != nil {
		dto.Color = in.Color
	}
	if in.Parent != nil {
		dto.Parent = in.Parent
	}
	return dto
}

func projectUpdateDTOFromInput(in ProjectWriteInput) apiclient.ProjectUpdateDto {
	dto := apiclient.ProjectUpdateDto{}
	if in.Title != nil {
		dto.Title = in.Title
	}
	if in.Note != nil {
		dto.Note = in.Note
	}
	if in.IsNotebook != nil {
		dto.IsNotebook = in.IsNotebook
	}
	if in.Emoji != nil {
		dto.Emoji = in.Emoji
	}
	if in.Color != nil {
		dto.Color = in.Color
	}
	if in.Parent != nil {
		dto.Parent = in.Parent
	}
	if in.JournalDate != nil {
		dto.JournalDate = in.JournalDate
	}
	if in.DeleteDate != nil {
		dto.DeleteDate = in.DeleteDate
	}
	return dto
}

func projectFromDTO(d apiclient.ProjectResponseDto) Project {
	p := Project{
		ID:    d.Id,
		Title: d.Title,
	}
	if d.Emoji != nil {
		p.Emoji = *d.Emoji
	}
	if d.Color != nil {
		p.Color = *d.Color
	}
	if d.IsNotebook != nil {
		p.IsNotebook = *d.IsNotebook
	}
	if d.Parent != nil {
		p.Parent = *d.Parent
	}
	if d.JournalDate != nil {
		p.JournalDate = *d.JournalDate
	}
	// ProjectResponseDto has no deleteDate field in OpenAPI; trash is write-only via Update.
	return p
}
