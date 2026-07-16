package api_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/eremenko789/singctl/internal/api"
)

func instantSleep(_ context.Context, _ time.Duration) error { return nil }

func TestRetry429ThenSuccessThreeHits(t *testing.T) {
	t.Parallel()
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := hits.Add(1)
		if n < 3 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"projects":[]}`))
	}))
	t.Cleanup(srv.Close)

	s, err := api.NewSession(srv.URL, "test-token-retry-ok", "5s", api.WithSleep(instantSleep))
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	if err := s.ValidateConnectivity(context.Background()); err != nil {
		t.Fatalf("ValidateConnectivity: %v", err)
	}
	if hits.Load() != 3 {
		t.Fatalf("hits = %d, want 3", hits.Load())
	}
}

func TestRetry429ExhaustedRateLimited(t *testing.T) {
	t.Parallel()
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`slow down`))
	}))
	t.Cleanup(srv.Close)

	s, err := api.NewSession(srv.URL, "test-token-retry-exhaust", "5s", api.WithSleep(instantSleep))
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	err = s.ValidateConnectivity(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
	if hits.Load() != 3 {
		t.Fatalf("hits = %d, want 3", hits.Load())
	}
	var ce *api.ClassifiedError
	if !errors.As(err, &ce) {
		t.Fatalf("want *ClassifiedError, got %T %v", err, err)
	}
	if ce.Message != "Error: rate limited. Retry later" {
		t.Fatalf("Message = %q", ce.Message)
	}
}

func TestRetryNon429SingleHit(t *testing.T) {
	t.Parallel()
	for _, code := range []int{401, 404, 422, 500} {
		code := code
		t.Run(http.StatusText(code), func(t *testing.T) {
			t.Parallel()
			var hits atomic.Int32
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				hits.Add(1)
				w.WriteHeader(code)
				_, _ = w.Write([]byte(`err`))
			}))
			t.Cleanup(srv.Close)

			s, err := api.NewSession(srv.URL, "test-token-no-retry", "5s", api.WithSleep(instantSleep))
			if err != nil {
				t.Fatalf("NewSession: %v", err)
			}
			err = s.ValidateConnectivity(context.Background())
			if err == nil {
				t.Fatal("expected error")
			}
			if hits.Load() != 1 {
				t.Fatalf("hits = %d, want 1 for status %d", hits.Load(), code)
			}
		})
	}
}

func TestRetryExponentialBackoffDelays(t *testing.T) {
	t.Parallel()
	var hits atomic.Int32
	var sleeps []time.Duration
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	t.Cleanup(srv.Close)

	sleeper := func(_ context.Context, d time.Duration) error {
		sleeps = append(sleeps, d)
		return nil
	}
	s, err := api.NewSession(srv.URL, "test-token-backoff", "5s", api.WithSleep(sleeper))
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	_ = s.ValidateConnectivity(context.Background())
	if hits.Load() != 3 {
		t.Fatalf("hits = %d, want 3", hits.Load())
	}
	if len(sleeps) != 2 {
		t.Fatalf("sleeps = %v, want 2 entries", sleeps)
	}
	if sleeps[0] != time.Second {
		t.Fatalf("first sleep = %v, want 1s", sleeps[0])
	}
	if sleeps[1] != 2*time.Second {
		t.Fatalf("second sleep = %v, want 2s", sleeps[1])
	}
}

func TestRetryAfterHeaderRespectedAndCapped(t *testing.T) {
	t.Parallel()
	t.Run("Retry-After 1", func(t *testing.T) {
		t.Parallel()
		var hits atomic.Int32
		var sleeps []time.Duration
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			n := hits.Add(1)
			if n < 3 {
				w.Header().Set("Retry-After", "1")
				w.WriteHeader(http.StatusTooManyRequests)
				return
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
		}))
		t.Cleanup(srv.Close)

		sleeper := func(_ context.Context, d time.Duration) error {
			sleeps = append(sleeps, d)
			return nil
		}
		s, err := api.NewSession(srv.URL, "test-token-ra1", "5s", api.WithSleep(sleeper))
		if err != nil {
			t.Fatalf("NewSession: %v", err)
		}
		if err := s.ValidateConnectivity(context.Background()); err != nil {
			t.Fatalf("ValidateConnectivity: %v", err)
		}
		if len(sleeps) != 2 {
			t.Fatalf("sleeps = %v", sleeps)
		}
		for i, d := range sleeps {
			if d != time.Second {
				t.Fatalf("sleep[%d] = %v, want 1s", i, d)
			}
		}
	})

	t.Run("Retry-After capped at 30s", func(t *testing.T) {
		t.Parallel()
		var hits atomic.Int32
		var sleeps []time.Duration
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			hits.Add(1)
			w.Header().Set("Retry-After", "999")
			w.WriteHeader(http.StatusTooManyRequests)
		}))
		t.Cleanup(srv.Close)

		sleeper := func(_ context.Context, d time.Duration) error {
			sleeps = append(sleeps, d)
			return nil
		}
		s, err := api.NewSession(srv.URL, "test-token-ra-cap", "5s", api.WithSleep(sleeper))
		if err != nil {
			t.Fatalf("NewSession: %v", err)
		}
		_ = s.ValidateConnectivity(context.Background())
		if len(sleeps) != 2 {
			t.Fatalf("sleeps = %v", sleeps)
		}
		for i, d := range sleeps {
			if d != 30*time.Second {
				t.Fatalf("sleep[%d] = %v, want 30s", i, d)
			}
		}
	})
}

func TestRetryContextCancelDuringBackoff(t *testing.T) {
	t.Parallel()
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	t.Cleanup(srv.Close)

	ctx, cancel := context.WithCancel(context.Background())
	sleeper := func(c context.Context, _ time.Duration) error {
		cancel()
		return c.Err()
	}
	s, err := api.NewSession(srv.URL, "test-token-ctx", "5s", api.WithSleep(sleeper))
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	err = s.ValidateConnectivity(ctx)
	if err == nil {
		t.Fatal("expected cancel error")
	}
	if hits.Load() != 1 {
		t.Fatalf("hits = %d, want 1 (stopped during backoff)", hits.Load())
	}
}
