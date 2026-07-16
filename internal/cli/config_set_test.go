package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	cfgpkg "github.com/eremenko789/singctl/internal/config"
)

func TestConfigSetPersistsValidKeyValues(t *testing.T) {
	withCLIPaths(t, func(paths testPaths) {
		stdout, stderr, err := executeForTest([]string{"config", "set", "api.base_url", "https://example.invalid"})
		if err != nil {
			t.Fatalf("set api.base_url error = %v stderr=%q stdout=%q", err, stderr, stdout)
		}

		stdout, stderr, err = executeForTest([]string{"config", "set", "output.format", "json"})
		if err != nil {
			t.Fatalf("set output.format error = %v stderr=%q stdout=%q", err, stderr, stdout)
		}

		configPath := filepath.Join(paths.XDG, "singctl", "config.yaml")
		cfg, err := cfgpkg.LoadConfig(configPath)
		if err != nil {
			t.Fatalf("LoadConfig() error = %v", err)
		}
		if cfg.API.BaseURL != "https://example.invalid" {
			t.Fatalf("API.BaseURL = %q, want updated value", cfg.API.BaseURL)
		}
		if cfg.Output.Format != "json" {
			t.Fatalf("Output.Format = %q, want updated value", cfg.Output.Format)
		}
	})
}

func TestConfigSetRejectsUnknownKeyWithoutChangingFile(t *testing.T) {
	withCLIPaths(t, func(paths testPaths) {
		configPath := filepath.Join(paths.XDG, "singctl", "config.yaml")
		mustSaveConfig(t, configPath, cfgpkg.Document{
			API: cfgpkg.APIConfig{
				BaseURL: "https://keep.invalid",
				Token:   "test-token-keep",
			},
		})

		stdout, stderr, err := executeForTest([]string{"config", "set", "api.no_such_key", "value"})
		if err == nil {
			t.Fatalf("expected error, got nil stdout=%q stderr=%q", stdout, stderr)
		}
		if !strings.Contains(stderr, "недопуст") && !strings.Contains(stderr, "неизвест") {
			t.Fatalf("stderr must explain invalid key, got %q", stderr)
		}

		cfg, err := cfgpkg.LoadConfig(configPath)
		if err != nil {
			t.Fatalf("LoadConfig() error = %v", err)
		}
		if cfg.API.BaseURL != "https://keep.invalid" || cfg.API.Token != "test-token-keep" {
			t.Fatalf("config file changed on invalid key: %#v", cfg)
		}
	})
}

func TestConfigSetRejectsInvalidValueWithoutChangingFile(t *testing.T) {
	withCLIPaths(t, func(paths testPaths) {
		configPath := filepath.Join(paths.XDG, "singctl", "config.yaml")
		mustSaveConfig(t, configPath, cfgpkg.Document{
			Output: cfgpkg.OutputConfig{
				Format: "yaml",
				Color:  true,
			},
		})

		stdout, stderr, err := executeForTest([]string{"config", "set", "output.format", "xml"})
		if err == nil {
			t.Fatalf("expected error, got nil stdout=%q stderr=%q", stdout, stderr)
		}

		cfg, err := cfgpkg.LoadConfig(configPath)
		if err != nil {
			t.Fatalf("LoadConfig() error = %v", err)
		}
		if cfg.Output.Format != "yaml" {
			t.Fatalf("Output.Format = %q, want original yaml", cfg.Output.Format)
		}
	})
}

func TestConfigSetConfigFlagWritesOnlyExplicitPath(t *testing.T) {
	withCLIPaths(t, func(paths testPaths) {
		explicitPath := filepath.Join(paths.Root, "custom", "config.yaml")
		stdout, stderr, err := executeForTest([]string{"--config", explicitPath, "config", "set", "tui.theme", "light"})
		if err != nil {
			t.Fatalf("executeForTest() error = %v stderr=%q stdout=%q", err, stderr, stdout)
		}

		explicitCfg, err := cfgpkg.LoadConfig(explicitPath)
		if err != nil {
			t.Fatalf("LoadConfig(explicit) error = %v", err)
		}
		if explicitCfg.TUI.Theme != "light" {
			t.Fatalf("TUI.Theme = %q, want light", explicitCfg.TUI.Theme)
		}

		defaultPath := filepath.Join(paths.XDG, "singctl", "config.yaml")
		if _, err := os.Stat(defaultPath); !os.IsNotExist(err) {
			t.Fatalf("default config path %q must not be created, stat err=%v", defaultPath, err)
		}
	})
}
