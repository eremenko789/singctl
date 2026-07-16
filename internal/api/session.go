package api

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/eremenko789/singctl/internal/apiclient"
	"github.com/eremenko789/singctl/internal/config"
)

// Session is a configured API access session with auth and timeout applied.
type Session struct {
	client *apiclient.ClientWithResponses
}

// Client returns the underlying OpenAPI client with responses.
func (s *Session) Client() *apiclient.ClientWithResponses {
	if s == nil {
		return nil
	}
	return s.client
}

// NewSession builds a Session from base URL, bare token, and timeout duration string.
// Fail-fast: empty token/URL or invalid timeout returns an error without creating a session.
func NewSession(baseURL, token, timeout string) (*Session, error) {
	baseURL = strings.TrimSpace(baseURL)
	token = strings.TrimSpace(token)
	timeout = strings.TrimSpace(timeout)

	if token == "" {
		return nil, fmt.Errorf("токен API не задан")
	}
	if baseURL == "" {
		return nil, fmt.Errorf("базовый URL API не задан")
	}
	if timeout == "" {
		return nil, fmt.Errorf("таймаут API не задан")
	}
	d, err := time.ParseDuration(timeout)
	if err != nil {
		return nil, fmt.Errorf("некорректный таймаут API: %w", err)
	}

	authValue := config.AuthorizationHeader(token)
	client, err := apiclient.NewClientWithResponses(
		baseURL,
		apiclient.WithHTTPClient(&http.Client{Timeout: d}),
		apiclient.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
			req.Header.Set("Authorization", authValue)
			return nil
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("не удалось создать API-клиент: %w", err)
	}
	return &Session{client: client}, nil
}

// NewFromSettings builds a Session from effective configuration settings.
func NewFromSettings(settings config.EffectiveSettings) (*Session, error) {
	return NewSession(settings.Config.API.BaseURL, settings.Config.API.Token, settings.Config.API.Timeout)
}
