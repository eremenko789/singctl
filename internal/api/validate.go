package api

import (
	"context"
	"fmt"
)

// ValidateConnectivity probes the API with ProjectController_list (GET /v2/project).
// Returns nil on 2xx, *HTTPError on non-2xx, or a transport/context error otherwise.
// Performs exactly one HTTP attempt (no retry).
func (s *Session) ValidateConnectivity(ctx context.Context) error {
	if s == nil || s.client == nil {
		return fmt.Errorf("сеанс API не инициализирован")
	}
	resp, err := s.client.ProjectControllerListWithResponse(ctx, nil)
	if err != nil {
		return err
	}
	return EnsureSuccess(resp.StatusCode(), resp.Body)
}
