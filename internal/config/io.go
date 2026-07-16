package config

import (
	"fmt"
	"os"
	"path/filepath"

	"go.yaml.in/yaml/v3"
)

// LoadConfig reads and parses a YAML configuration file at path.
func LoadConfig(path string) (Document, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Document{}, fmt.Errorf("прочитать конфиг: %w", err)
	}

	cfg := DefaultConfig()
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Document{}, fmt.Errorf("разобрать YAML конфигурации: %w", err)
	}
	applyDefaults(&cfg)
	return cfg, nil
}

// SaveConfig writes cfg as YAML to path, creating parent directories as needed.
func SaveConfig(path string, cfg Document) error {
	applyDefaults(&cfg)

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("подготовить YAML конфигурации: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("создать каталог конфигурации: %w", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("записать конфиг: %w", err)
	}
	return nil
}

func applyDefaults(cfg *Document) {
	defaults := DefaultConfig()
	if cfg.API.BaseURL == "" {
		cfg.API.BaseURL = defaults.API.BaseURL
	}
	if cfg.API.Timeout == "" {
		cfg.API.Timeout = defaults.API.Timeout
	}
	if cfg.Output.Format == "" {
		cfg.Output.Format = defaults.Output.Format
	}
	if cfg.Output.DateFormat == "" {
		cfg.Output.DateFormat = defaults.Output.DateFormat
	}
	if !cfg.Output.Color && cfg.Output == (OutputConfig{}) {
		cfg.Output.Color = defaults.Output.Color
	}
	if cfg.TUI.Theme == "" {
		cfg.TUI.Theme = defaults.TUI.Theme
	}
	if !cfg.TUI.ViKeys && cfg.TUI == (TUIConfig{}) {
		cfg.TUI.ViKeys = defaults.TUI.ViKeys
	}
}
