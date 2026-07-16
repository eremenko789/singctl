package cli

import (
	"encoding/csv"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	cfgpkg "github.com/eremenko789/singctl/internal/config"
	"go.yaml.in/yaml/v3"
)

func TestConfigShowWithoutConfigFailsSafely(t *testing.T) {
	withCLIPaths(t, func(paths testPaths) {
		stdout, stderr, err := executeForTest([]string{"config", "show"})
		if err == nil {
			t.Fatalf("expected error, got nil stdout=%q stderr=%q", stdout, stderr)
		}
		if strings.Contains(stdout, "test-token") || strings.Contains(stderr, "test-token") {
			t.Fatalf("unexpected secret leak in output stdout=%q stderr=%q", stdout, stderr)
		}
		if !strings.Contains(stderr, "конфиг") {
			t.Fatalf("stderr must explain missing config, got %q", stderr)
		}
	})
}

func TestConfigShowMasksTokenAndPrefersRuntimeOverride(t *testing.T) {
	withCLIPaths(t, func(paths testPaths) {
		configPath := filepath.Join(paths.XDG, "singctl", "config.yaml")
		if err := cfgpkg.SaveConfig(configPath, cfgpkg.Document{
			API: cfgpkg.APIConfig{
				BaseURL: "https://example.invalid",
				Token:   "test-token-file",
				Timeout: "45s",
			},
		}); err != nil {
			t.Fatalf("SaveConfig() error = %v", err)
		}

		stdout, stderr, err := executeForTest([]string{"--token", "test-token-runtime", "config", "show"})
		if err != nil {
			t.Fatalf("executeForTest() error = %v stderr=%q stdout=%q", err, stderr, stdout)
		}
		if strings.Contains(stdout, "test-token-runtime") || strings.Contains(stdout, "test-token-file") {
			t.Fatalf("full token leaked to stdout: %q", stdout)
		}
		if !strings.Contains(stdout, "test****time") {
			t.Fatalf("stdout must contain masked runtime token, got %q", stdout)
		}
	})
}

func TestConfigShowOutputDefaultsToYAMLAndSupportsJSON(t *testing.T) {
	withCLIPaths(t, func(paths testPaths) {
		configPath := filepath.Join(paths.XDG, "singctl", "config.yaml")
		if err := cfgpkg.SaveConfig(configPath, cfgpkg.Document{
			API: cfgpkg.APIConfig{
				BaseURL: "https://example.invalid",
				Token:   "test-token-aaaa",
				Timeout: "45s",
			},
			Output: cfgpkg.OutputConfig{
				Format: "csv",
			},
		}); err != nil {
			t.Fatalf("SaveConfig() error = %v", err)
		}

		yamlOut, yamlErr, err := executeForTest([]string{"config", "show"})
		if err != nil {
			t.Fatalf("yaml show error = %v stderr=%q stdout=%q", err, yamlErr, yamlOut)
		}
		var yamlDoc map[string]any
		if err := yaml.Unmarshal([]byte(yamlOut), &yamlDoc); err != nil {
			t.Fatalf("yaml output is invalid: %v\n%s", err, yamlOut)
		}

		jsonOut, jsonErr, err := executeForTest([]string{"-o", "json", "config", "show"})
		if err != nil {
			t.Fatalf("json show error = %v stderr=%q stdout=%q", err, jsonErr, jsonOut)
		}
		var jsonDoc map[string]any
		if err := json.Unmarshal([]byte(jsonOut), &jsonDoc); err != nil {
			t.Fatalf("json output is invalid: %v\n%s", err, jsonOut)
		}
	})
}

func TestConfigShowSupportsCSVAndTableWithMaskedToken(t *testing.T) {
	withCLIPaths(t, func(paths testPaths) {
		configPath := filepath.Join(paths.XDG, "singctl", "config.yaml")
		if err := cfgpkg.SaveConfig(configPath, cfgpkg.Document{
			API: cfgpkg.APIConfig{
				BaseURL: "https://example.invalid",
				Token:   "test-token-aaaa",
			},
		}); err != nil {
			t.Fatalf("SaveConfig() error = %v", err)
		}

		csvOut, csvErr, err := executeForTest([]string{"-o", "csv", "config", "show"})
		if err != nil {
			t.Fatalf("csv show error = %v stderr=%q stdout=%q", err, csvErr, csvOut)
		}
		if strings.Contains(csvOut, "test-token-aaaa") {
			t.Fatalf("csv leaked full token: %q", csvOut)
		}
		records, err := csv.NewReader(strings.NewReader(csvOut)).ReadAll()
		if err != nil {
			t.Fatalf("invalid csv output: %v\n%s", err, csvOut)
		}
		if len(records) == 0 {
			t.Fatalf("expected csv rows, got none")
		}

		tableOut, tableErr, err := executeForTest([]string{"-o", "table", "config", "show"})
		if err != nil {
			t.Fatalf("table show error = %v stderr=%q stdout=%q", err, tableErr, tableOut)
		}
		if strings.Contains(tableOut, "test-token-aaaa") {
			t.Fatalf("table leaked full token: %q", tableOut)
		}
		if !strings.Contains(tableOut, "test****aaaa") {
			t.Fatalf("table output must contain masked token, got %q", tableOut)
		}
	})
}

func TestConfigShowReportsInvalidYAMLSafely(t *testing.T) {
	withCLIPaths(t, func(paths testPaths) {
		configPath := filepath.Join(paths.XDG, "singctl", "config.yaml")
		if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
			t.Fatalf("mkdir parent: %v", err)
		}
		if err := os.WriteFile(configPath, []byte("api: [broken"), 0o600); err != nil {
			t.Fatalf("write invalid yaml: %v", err)
		}

		stdout, stderr, err := executeForTest([]string{"config", "show"})
		if err == nil {
			t.Fatalf("expected error, got nil stdout=%q stderr=%q", stdout, stderr)
		}
		if !strings.Contains(stderr, "YAML") && !strings.Contains(stderr, "конфиг") {
			t.Fatalf("stderr must explain YAML/config issue, got %q", stderr)
		}
	})
}
