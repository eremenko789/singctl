package api_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/eremenko789/singctl/internal/api"
)

func TestValidateConnectivityNon2xxSingleRequest(t *testing.T) {
	t.Parallel()
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`unauthorized`))
	}))
	t.Cleanup(srv.Close)

	s, err := api.NewSession(srv.URL, "test-token-401", "5s")
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	err = s.ValidateConnectivity(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
	var httpErr *api.HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("want *HTTPError, got %T %v", err, err)
	}
	if httpErr.StatusCode != http.StatusUnauthorized {
		t.Fatalf("StatusCode = %d, want 401", httpErr.StatusCode)
	}
	if hits.Load() != 1 {
		t.Fatalf("request count = %d, want 1 (no retry)", hits.Load())
	}
}

func TestValidateConnectivityTransportErrorNotHTTPError(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	baseURL := srv.URL
	srv.Close()

	s, err := api.NewSession(baseURL, "test-token-transport", "2s")
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	err = s.ValidateConnectivity(context.Background())
	if err == nil {
		t.Fatal("expected transport error")
	}
	var httpErr *api.HTTPError
	if errors.As(err, &httpErr) {
		t.Fatalf("transport error must not be *HTTPError, got %#v", httpErr)
	}
}
