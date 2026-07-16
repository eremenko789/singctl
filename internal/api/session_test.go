package api_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/eremenko789/singctl/internal/api"
	"github.com/eremenko789/singctl/internal/config"
)

func TestNewSessionEmptyToken(t *testing.T) {
	t.Parallel()
	for _, token := range []string{"", "   ", "\t"} {
		s, err := api.NewSession("https://example.invalid", token, "30s")
		if err == nil {
			t.Fatalf("token %q: expected error, got nil", token)
		}
		if s != nil {
			t.Fatalf("token %q: expected nil session, got %#v", token, s)
		}
		if strings.Contains(err.Error(), token) && strings.TrimSpace(token) != "" {
			t.Fatalf("error must not leak token: %v", err)
		}
	}
}

func TestNewSessionEmptyBaseURL(t *testing.T) {
	t.Parallel()
	for _, base := range []string{"", "   ", "\t"} {
		s, err := api.NewSession(base, "test-token-url", "30s")
		if err == nil {
			t.Fatalf("baseURL %q: expected error, got nil", base)
		}
		if s != nil {
			t.Fatalf("baseURL %q: expected nil session", base)
		}
	}
}

func TestNewSessionInvalidTimeout(t *testing.T) {
	t.Parallel()
	s, err := api.NewSession("https://example.invalid", "test-token-timeout", "not-a-duration")
	if err == nil {
		t.Fatal("expected error for invalid timeout")
	}
	if s != nil {
		t.Fatal("expected nil session")
	}
}

func TestNewSessionAndNewFromSettingsSuccess(t *testing.T) {
	t.Parallel()
	s, err := api.NewSession("https://example.invalid", "test-token-ok", "15s")
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	if s == nil || s.Client() == nil {
		t.Fatal("expected non-nil session with client")
	}

	settings := config.EffectiveSettings{
		Config: config.Document{
			API: config.APIConfig{
				BaseURL: "https://example.invalid",
				Token:   "test-token-from-settings",
				Timeout: "20s",
			},
		},
	}
	s2, err := api.NewFromSettings(settings)
	if err != nil {
		t.Fatalf("NewFromSettings: %v", err)
	}
	if s2 == nil || s2.Client() == nil {
		t.Fatal("expected non-nil session from settings")
	}
}

func TestHappyPathBearerHeader(t *testing.T) {
	t.Parallel()
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/v2/project" {
			t.Errorf("path = %q, want /v2/project", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"projects":[]}`))
	}))
	t.Cleanup(srv.Close)

	s, err := api.NewSession(srv.URL, "test-token-happy", "5s")
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	if err := s.ValidateConnectivity(context.Background()); err != nil {
		t.Fatalf("ValidateConnectivity: %v", err)
	}
	if gotAuth != "Bearer test-token-happy" {
		t.Fatalf("Authorization = %q, want %q", gotAuth, "Bearer test-token-happy")
	}
}

func TestBearerPrefixNotDoubled(t *testing.T) {
	t.Parallel()
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"projects":[]}`))
	}))
	t.Cleanup(srv.Close)

	s, err := api.NewSession(srv.URL, "Bearer test-token-prefixed", "5s")
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	if err := s.ValidateConnectivity(context.Background()); err != nil {
		t.Fatalf("ValidateConnectivity: %v", err)
	}
	if gotAuth != "Bearer test-token-prefixed" {
		t.Fatalf("Authorization = %q, want no double Bearer", gotAuth)
	}
	if strings.HasPrefix(gotAuth, "Bearer Bearer") {
		t.Fatalf("double Bearer prefix: %q", gotAuth)
	}
}

func TestSessionTimeoutAgainstHang(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	t.Cleanup(srv.Close)

	s, err := api.NewSession(srv.URL, "test-token-timeout-hang", "50ms")
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	start := time.Now()
	err = s.ValidateConnectivity(context.Background())
	elapsed := time.Since(start)
	if err == nil {
		t.Fatal("expected timeout error")
	}
	var httpErr *api.HTTPError
	if errors.As(err, &httpErr) {
		t.Fatalf("timeout should not be HTTPError, got %#v", httpErr)
	}
	if elapsed > 500*time.Millisecond {
		t.Fatalf("hung too long: %v", elapsed)
	}
}
