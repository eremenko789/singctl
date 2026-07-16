package config

import (
	"path/filepath"
	"testing"
)

func TestLoadSaveConfigRoundTrip(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "nested", "singctl", "config.yaml")
	want := Document{
		API: APIConfig{
			BaseURL: "https://example.invalid",
			Token:   "test-token-aaaa",
			Timeout: "45s",
		},
		Output: OutputConfig{
			Format:     "json",
			Color:      false,
			DateFormat: "02.01.2006",
		},
		TUI: TUIConfig{
			Theme:           "light",
			ViKeys:          false,
			RefreshInterval: 10,
		},
	}

	if err := SaveConfig(path, want); err != nil {
		t.Fatalf("SaveConfig() error = %v", err)
	}

	got, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	if got.API.BaseURL != want.API.BaseURL {
		t.Fatalf("API.BaseURL = %q, want %q", got.API.BaseURL, want.API.BaseURL)
	}
	if got.API.Token != want.API.Token {
		t.Fatalf("API.Token = %q, want %q", got.API.Token, want.API.Token)
	}
	if got.API.Timeout != want.API.Timeout {
		t.Fatalf("API.Timeout = %q, want %q", got.API.Timeout, want.API.Timeout)
	}
	if got.Output.Format != want.Output.Format {
		t.Fatalf("Output.Format = %q, want %q", got.Output.Format, want.Output.Format)
	}
	if got.Output.Color != want.Output.Color {
		t.Fatalf("Output.Color = %v, want %v", got.Output.Color, want.Output.Color)
	}
	if got.Output.DateFormat != want.Output.DateFormat {
		t.Fatalf("Output.DateFormat = %q, want %q", got.Output.DateFormat, want.Output.DateFormat)
	}
	if got.TUI.Theme != want.TUI.Theme {
		t.Fatalf("TUI.Theme = %q, want %q", got.TUI.Theme, want.TUI.Theme)
	}
	if got.TUI.ViKeys != want.TUI.ViKeys {
		t.Fatalf("TUI.ViKeys = %v, want %v", got.TUI.ViKeys, want.TUI.ViKeys)
	}
	if got.TUI.RefreshInterval != want.TUI.RefreshInterval {
		t.Fatalf("TUI.RefreshInterval = %d, want %d", got.TUI.RefreshInterval, want.TUI.RefreshInterval)
	}
}
