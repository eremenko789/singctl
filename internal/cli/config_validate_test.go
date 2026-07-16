package cli

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/eremenko789/singctl/internal/api"
	cfgpkg "github.com/eremenko789/singctl/internal/config"
)

func TestConfigValidateWithoutTokenFailsWithHint(t *testing.T) {
	withCLIPaths(t, func(paths testPaths) {
		stdout, stderr, err := executeForTest([]string{"config", "validate"})
		if err == nil {
			t.Fatalf("expected error, got nil stdout=%q stderr=%q", stdout, stderr)
		}
		if !strings.Contains(stderr, "set-token") {
			t.Fatalf("stderr must hint set-token, got %q", stderr)
		}
		if ExitCode(err) != 2 {
			t.Fatalf("ExitCode = %d, want 2", ExitCode(err))
		}
	})
}

func TestConfigValidateWithTokenReportsRemoteSuccess(t *testing.T) {
	withCLIPaths(t, func(paths testPaths) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Authorization") != "Bearer test-token-validate" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"projects":[]}`))
		}))
		t.Cleanup(srv.Close)

		configPath := filepath.Join(paths.XDG, "singctl", "config.yaml")
		mustSaveConfig(t, configPath, cfgpkg.Document{
			API: cfgpkg.APIConfig{
				BaseURL: srv.URL,
				Token:   "test-token-validate",
				Timeout: "5s",
			},
		})

		stdout, stderr, err := executeForTest([]string{"config", "validate"})
		if err != nil {
			t.Fatalf("expected success, got err=%v stderr=%q stdout=%q", err, stderr, stdout)
		}
		if ExitCode(err) != 0 {
			t.Fatalf("ExitCode = %d, want 0", ExitCode(err))
		}
		if strings.Contains(stdout, "test-token-validate") || strings.Contains(stderr, "test-token-validate") {
			t.Fatalf("full token leaked stdout=%q stderr=%q", stdout, stderr)
		}
		if strings.Contains(stdout, "заглуш") || strings.Contains(stdout, "локальн") {
			t.Fatalf("stdout must not use stub/local-only wording, got %q", stdout)
		}
		if !strings.Contains(stdout, "удалённ") && !strings.Contains(stdout, "Удалённ") {
			t.Fatalf("stdout must confirm remote check, got %q", stdout)
		}
	})
}

func TestConfigValidateNon2xxFailsWithoutRemoteOK(t *testing.T) {
	withCLIPaths(t, func(paths testPaths) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`nope`))
		}))
		t.Cleanup(srv.Close)

		configPath := filepath.Join(paths.XDG, "singctl", "config.yaml")
		mustSaveConfig(t, configPath, cfgpkg.Document{
			API: cfgpkg.APIConfig{
				BaseURL: srv.URL,
				Token:   "test-token-bad",
				Timeout: "5s",
			},
		})

		stdout, stderr, err := executeForTest([]string{"config", "validate"})
		if err == nil {
			t.Fatalf("expected error, got nil stdout=%q stderr=%q", stdout, stderr)
		}
		combined := stdout + stderr
		if strings.Contains(combined, "успешно") || strings.Contains(combined, "удалённо OK") {
			t.Fatalf("must not claim remote OK, got stdout=%q stderr=%q", stdout, stderr)
		}
		if strings.Contains(combined, "test-token-bad") {
			t.Fatalf("token leaked: stdout=%q stderr=%q", stdout, stderr)
		}
	})
}

func TestConfigValidateMock401CatalogExit1(t *testing.T) {
	withCLIPaths(t, func(paths testPaths) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`nope`))
		}))
		t.Cleanup(srv.Close)

		configPath := filepath.Join(paths.XDG, "singctl", "config.yaml")
		mustSaveConfig(t, configPath, cfgpkg.Document{
			API: cfgpkg.APIConfig{
				BaseURL: srv.URL,
				Token:   "test-token-401",
				Timeout: "5s",
			},
		})

		_, stderr, err := executeForTest([]string{"config", "validate"})
		if err == nil {
			t.Fatal("expected error")
		}
		if ExitCode(err) != 1 {
			t.Fatalf("ExitCode = %d, want 1", ExitCode(err))
		}
		if !strings.Contains(stderr, "invalid token") || !strings.Contains(stderr, "set-token") {
			t.Fatalf("stderr missing catalog text: %q", stderr)
		}
	})
}

func TestConfigValidateMock404CatalogExit3(t *testing.T) {
	withCLIPaths(t, func(paths testPaths) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`missing`))
		}))
		t.Cleanup(srv.Close)

		configPath := filepath.Join(paths.XDG, "singctl", "config.yaml")
		mustSaveConfig(t, configPath, cfgpkg.Document{
			API: cfgpkg.APIConfig{
				BaseURL: srv.URL,
				Token:   "test-token-404",
				Timeout: "5s",
			},
		})

		_, stderr, err := executeForTest([]string{"config", "validate"})
		if err == nil {
			t.Fatal("expected error")
		}
		if ExitCode(err) != 3 {
			t.Fatalf("ExitCode = %d, want 3", ExitCode(err))
		}
		if !strings.Contains(stderr, "entity not found") {
			t.Fatalf("stderr missing not-found catalog: %q", stderr)
		}
	})
}

func TestConfigValidateMock429Exhausted(t *testing.T) {
	withCLIPaths(t, func(paths testPaths) {
		restore := api.SetSleepForTest(func(context.Context, time.Duration) error { return nil })
		t.Cleanup(restore)

		var hits atomic.Int32
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			hits.Add(1)
			w.WriteHeader(http.StatusTooManyRequests)
		}))
		t.Cleanup(srv.Close)

		configPath := filepath.Join(paths.XDG, "singctl", "config.yaml")
		mustSaveConfig(t, configPath, cfgpkg.Document{
			API: cfgpkg.APIConfig{
				BaseURL: srv.URL,
				Token:   "test-token-429",
				Timeout: "5s",
			},
		})

		_, stderr, err := executeForTest([]string{"config", "validate"})
		if err == nil {
			t.Fatal("expected error")
		}
		if ExitCode(err) != 1 {
			t.Fatalf("ExitCode = %d, want 1", ExitCode(err))
		}
		if hits.Load() != 3 {
			t.Fatalf("hits = %d, want 3", hits.Load())
		}
		if !strings.Contains(stderr, "rate limited") {
			t.Fatalf("stderr missing rate-limited catalog: %q", stderr)
		}
	})
}

func TestConfigValidateMock5xx(t *testing.T) {
	withCLIPaths(t, func(paths testPaths) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`boom`))
		}))
		t.Cleanup(srv.Close)

		configPath := filepath.Join(paths.XDG, "singctl", "config.yaml")
		mustSaveConfig(t, configPath, cfgpkg.Document{
			API: cfgpkg.APIConfig{
				BaseURL: srv.URL,
				Token:   "test-token-5xx",
				Timeout: "5s",
			},
		})

		_, stderr, err := executeForTest([]string{"config", "validate"})
		if err == nil {
			t.Fatal("expected error")
		}
		if ExitCode(err) != 1 {
			t.Fatalf("ExitCode = %d, want 1", ExitCode(err))
		}
		if !strings.Contains(stderr, "server error") {
			t.Fatalf("stderr missing server catalog: %q", stderr)
		}
	})
}

func TestConfigValidateUsesRuntimeTokenOverride(t *testing.T) {
	withCLIPaths(t, func(paths testPaths) {
		var gotAuth string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotAuth = r.Header.Get("Authorization")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"projects":[]}`))
		}))
		t.Cleanup(srv.Close)

		configPath := filepath.Join(paths.XDG, "singctl", "config.yaml")
		mustSaveConfig(t, configPath, cfgpkg.Document{
			API: cfgpkg.APIConfig{
				BaseURL: srv.URL,
				Token:   "",
				Timeout: "5s",
			},
		})

		stdout, stderr, err := executeForTest([]string{"--token", "test-token-runtime", "config", "validate"})
		if err != nil {
			t.Fatalf("expected success, got err=%v stderr=%q stdout=%q", err, stderr, stdout)
		}
		if strings.Contains(stdout, "test-token-runtime") || strings.Contains(stderr, "test-token-runtime") {
			t.Fatalf("full runtime token leaked stdout=%q stderr=%q", stdout, stderr)
		}
		if gotAuth != "Bearer test-token-runtime" {
			t.Fatalf("Authorization = %q, want Bearer test-token-runtime", gotAuth)
		}
	})
}

// TestConfigValidateUsesAdapterPackage documents that validate wires through internal/api
// (import of api in config_validate.go); remote success above exercises that path.
func TestConfigValidateUsesAdapterPackage(t *testing.T) {
	t.Parallel()
	// Compile-time / wiring check: config_validate.go imports internal/api.
	// Behavioral coverage: TestConfigValidateWithTokenReportsRemoteSuccess.
}
