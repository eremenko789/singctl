package cli

import (
	"path/filepath"
	"strings"
	"testing"

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
	})
}

func TestConfigValidateWithTokenReportsLocalStubSuccess(t *testing.T) {
	withCLIPaths(t, func(paths testPaths) {
		configPath := filepath.Join(paths.XDG, "singctl", "config.yaml")
		mustSaveConfig(t, configPath, cfgpkg.Document{
			API: cfgpkg.APIConfig{Token: "test-token-file"},
		})

		stdout, stderr, err := executeForTest([]string{"config", "validate"})
		if err != nil {
			t.Fatalf("expected success, got err=%v stderr=%q stdout=%q", err, stderr, stdout)
		}
		if strings.Contains(stdout, "test-token-file") || strings.Contains(stderr, "test-token-file") {
			t.Fatalf("full token leaked stdout=%q stderr=%q", stdout, stderr)
		}
		if !strings.Contains(stdout, "локаль") && !strings.Contains(stdout, "заглуш") {
			t.Fatalf("stdout must explain local/stub validation, got %q", stdout)
		}
	})
}

func TestConfigValidateUsesRuntimeTokenOverride(t *testing.T) {
	withCLIPaths(t, func(paths testPaths) {
		configPath := filepath.Join(paths.XDG, "singctl", "config.yaml")
		mustSaveConfig(t, configPath, cfgpkg.Document{
			API: cfgpkg.APIConfig{Token: ""},
		})

		stdout, stderr, err := executeForTest([]string{"--token", "test-token-runtime", "config", "validate"})
		if err != nil {
			t.Fatalf("expected success, got err=%v stderr=%q stdout=%q", err, stderr, stdout)
		}
		if strings.Contains(stdout, "test-token-runtime") || strings.Contains(stderr, "test-token-runtime") {
			t.Fatalf("full runtime token leaked stdout=%q stderr=%q", stdout, stderr)
		}
	})
}
