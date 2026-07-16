package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	cfgpkg "github.com/eremenko789/singctl/internal/config"
)

func withCLIPaths(t *testing.T, fn func(paths testPaths)) {
	t.Helper()

	root := t.TempDir()
	home := filepath.Join(root, "home")
	xdg := filepath.Join(root, "xdg")
	cwd := filepath.Join(root, "cwd")

	for _, dir := range []string{home, xdg, cwd} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", dir, err)
		}
	}

	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", xdg)

	prevWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(cwd); err != nil {
		t.Fatalf("chdir %s: %v", cwd, err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(prevWD)
	})

	fn(testPaths{
		Root: root,
		Home: home,
		XDG:  xdg,
		CWD:  cwd,
	})
}

type testPaths struct {
	Root string
	Home string
	XDG  string
	CWD  string
}

func TestConfigSetTokenWritesBareTokenToResolvedPath(t *testing.T) {
	withCLIPaths(t, func(paths testPaths) {
		stdout, stderr, err := executeForTest([]string{"config", "set-token", "test-token-aaaa"})
		if err != nil {
			t.Fatalf("executeForTest() error = %v stderr=%q stdout=%q", err, stderr, stdout)
		}
		if strings.Contains(stdout, "test-token-aaaa") || strings.Contains(stderr, "test-token-aaaa") {
			t.Fatal("full token leaked to output")
		}

		configPath := filepath.Join(paths.XDG, "singctl", "config.yaml")
		cfg, err := cfgpkg.LoadConfig(configPath)
		if err != nil {
			t.Fatalf("LoadConfig(%q) error = %v", configPath, err)
		}
		if cfg.API.Token != "test-token-aaaa" {
			t.Fatalf("cfg.API.Token = %q, want bare token", cfg.API.Token)
		}
	})
}

func TestConfigSetTokenRejectsBearerInputWithoutChangingFile(t *testing.T) {
	withCLIPaths(t, func(paths testPaths) {
		configPath := filepath.Join(paths.XDG, "singctl", "config.yaml")
		if err := cfgpkg.SaveConfig(configPath, cfgpkg.Document{
			API: cfgpkg.APIConfig{
				Token: "test-token-keep",
			},
		}); err != nil {
			t.Fatalf("SaveConfig() error = %v", err)
		}

		stdout, stderr, err := executeForTest([]string{"config", "set-token", "Bearer should-not-pass"})
		if err == nil {
			t.Fatalf("expected error, got nil stdout=%q stderr=%q", stdout, stderr)
		}
		if !strings.Contains(stderr, "Bearer") {
			t.Fatalf("stderr must mention Bearer rule, got %q", stderr)
		}

		cfg, err := cfgpkg.LoadConfig(configPath)
		if err != nil {
			t.Fatalf("LoadConfig() error = %v", err)
		}
		if cfg.API.Token != "test-token-keep" {
			t.Fatalf("cfg.API.Token = %q, want original token preserved", cfg.API.Token)
		}
	})
}

func TestConfigSetTokenRejectsBlankInput(t *testing.T) {
	withCLIPaths(t, func(paths testPaths) {
		stdout, stderr, err := executeForTest([]string{"config", "set-token", "   "})
		if err == nil {
			t.Fatalf("expected error, got nil stdout=%q stderr=%q", stdout, stderr)
		}

		defaultPath := filepath.Join(paths.XDG, "singctl", "config.yaml")
		if _, statErr := os.Stat(defaultPath); !os.IsNotExist(statErr) {
			t.Fatalf("config file must not be created for blank token, stat err=%v", statErr)
		}
	})
}

func TestConfigSetTokenConfigFlagWritesOnlyExplicitPath(t *testing.T) {
	withCLIPaths(t, func(paths testPaths) {
		explicitPath := filepath.Join(paths.Root, "custom", "config.yaml")
		stdout, stderr, err := executeForTest([]string{"--config", explicitPath, "config", "set-token", "test-token-flag"})
		if err != nil {
			t.Fatalf("executeForTest() error = %v stderr=%q stdout=%q", err, stderr, stdout)
		}

		explicitCfg, err := cfgpkg.LoadConfig(explicitPath)
		if err != nil {
			t.Fatalf("LoadConfig(explicit) error = %v", err)
		}
		if explicitCfg.API.Token != "test-token-flag" {
			t.Fatalf("explicitCfg.API.Token = %q, want %q", explicitCfg.API.Token, "test-token-flag")
		}

		defaultPath := filepath.Join(paths.XDG, "singctl", "config.yaml")
		if _, err := os.Stat(defaultPath); !os.IsNotExist(err) {
			t.Fatalf("default config path %q must not be created, stat err=%v", defaultPath, err)
		}
	})
}
