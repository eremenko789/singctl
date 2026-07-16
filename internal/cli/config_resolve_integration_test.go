package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	cfgpkg "github.com/eremenko789/singctl/internal/config"
)

func TestConfigShowReadsHighestPriorityExistingFile(t *testing.T) {
	withCLIPaths(t, func(paths testPaths) {
		homePath := filepath.Join(paths.Home, ".config", "singctl", "config.yaml")
		cwdPath := filepath.Join(paths.CWD, ".singctl.yaml")
		explicitPath := filepath.Join(paths.Root, "explicit.yaml")

		mustSaveConfig(t, homePath, cfgpkg.Document{
			API: cfgpkg.APIConfig{BaseURL: "https://home.invalid", Token: "test-token-home"},
		})
		mustSaveConfig(t, cwdPath, cfgpkg.Document{
			API: cfgpkg.APIConfig{BaseURL: "https://cwd.invalid", Token: "test-token-cwd"},
		})

		stdout, stderr, err := executeForTest([]string{"config", "show"})
		if err != nil {
			t.Fatalf("show without explicit config error = %v stderr=%q stdout=%q", err, stderr, stdout)
		}
		if !strings.Contains(stdout, "https://cwd.invalid") {
			t.Fatalf("expected cwd config to win, got %q", stdout)
		}
		if strings.Contains(stdout, "https://home.invalid") {
			t.Fatalf("home config must not override cwd config, got %q", stdout)
		}

		mustSaveConfig(t, explicitPath, cfgpkg.Document{
			API: cfgpkg.APIConfig{BaseURL: "https://explicit.invalid", Token: "test-token-explicit"},
		})

		stdout, stderr, err = executeForTest([]string{"--config", explicitPath, "config", "show"})
		if err != nil {
			t.Fatalf("show with explicit config error = %v stderr=%q stdout=%q", err, stderr, stdout)
		}
		if !strings.Contains(stdout, "https://explicit.invalid") {
			t.Fatalf("expected explicit config to win, got %q", stdout)
		}
	})
}

func TestConfigSetTokenUpdatesOnlyEffectiveFile(t *testing.T) {
	withCLIPaths(t, func(paths testPaths) {
		homePath := filepath.Join(paths.Home, ".config", "singctl", "config.yaml")
		cwdPath := filepath.Join(paths.CWD, ".singctl.yaml")

		mustSaveConfig(t, homePath, cfgpkg.Document{
			API: cfgpkg.APIConfig{Token: "test-token-home"},
		})
		mustSaveConfig(t, cwdPath, cfgpkg.Document{
			API: cfgpkg.APIConfig{Token: "test-token-cwd"},
		})

		stdout, stderr, err := executeForTest([]string{"config", "set-token", "test-token-new"})
		if err != nil {
			t.Fatalf("set-token error = %v stderr=%q stdout=%q", err, stderr, stdout)
		}

		homeCfg, err := cfgpkg.LoadConfig(homePath)
		if err != nil {
			t.Fatalf("LoadConfig(home) error = %v", err)
		}
		if homeCfg.API.Token != "test-token-home" {
			t.Fatalf("home token = %q, want original token preserved", homeCfg.API.Token)
		}

		cwdCfg, err := cfgpkg.LoadConfig(cwdPath)
		if err != nil {
			t.Fatalf("LoadConfig(cwd) error = %v", err)
		}
		if cwdCfg.API.Token != "test-token-new" {
			t.Fatalf("cwd token = %q, want updated token", cwdCfg.API.Token)
		}
	})
}

func mustSaveConfig(t *testing.T, path string, cfg cfgpkg.Document) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir parent for %s: %v", path, err)
	}
	if err := cfgpkg.SaveConfig(path, cfg); err != nil {
		t.Fatalf("SaveConfig(%q) error = %v", path, err)
	}
}
