package api

import (
	"context"
	"fmt"
)

// ValidateConnectivity probes the API with ProjectController_list (GET /v2/project).
// Returns nil on 2xx, or a ClassifiedError on failure (HTTP taxonomy or transport).
// One logical call may issue up to 3 HTTP requests when the server responds with 429.
func (s *Session) ValidateConnectivity(ctx context.Context) error {
	if s == nil || s.client == nil {
		return Classify(fmt.Errorf("сеанс API не инициализирован"))
	}
	resp, err := s.client.ProjectControllerListWithResponse(ctx, nil)
	if err != nil {
		return Classify(err)
	}
	return Classify(EnsureSuccess(resp.StatusCode(), resp.Body))
}
