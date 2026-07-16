package config

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// SetConfigValue updates a dotted-path key on cfg after validating the value.
func SetConfigValue(cfg *Document, key, value string) error {
	switch key {
	case "api.base_url":
		cfg.API.BaseURL = strings.TrimSpace(value)
	case "api.token":
		token, err := NormalizeStoredToken(value)
		if err != nil {
			return err
		}
		cfg.API.Token = token
	case "api.timeout":
		if _, err := time.ParseDuration(strings.TrimSpace(value)); err != nil {
			return fmt.Errorf("недопустимое значение %q для %s: ожидается duration, например 30s", value, key)
		}
		cfg.API.Timeout = strings.TrimSpace(value)
	case "output.format":
		format := strings.ToLower(strings.TrimSpace(value))
		switch format {
		case "table", "json", "yaml", "csv":
			cfg.Output.Format = format
		default:
			return fmt.Errorf("недопустимое значение %q для %s", value, key)
		}
	case "output.color":
		parsed, err := strconv.ParseBool(strings.TrimSpace(value))
		if err != nil {
			return fmt.Errorf("недопустимое значение %q для %s: ожидается true/false", value, key)
		}
		cfg.Output.Color = parsed
	case "output.date_format":
		cfg.Output.DateFormat = strings.TrimSpace(value)
	case "tui.theme":
		theme := strings.ToLower(strings.TrimSpace(value))
		switch theme {
		case "dark", "light":
			cfg.TUI.Theme = theme
		default:
			return fmt.Errorf("недопустимое значение %q для %s", value, key)
		}
	case "tui.vi_keys":
		parsed, err := strconv.ParseBool(strings.TrimSpace(value))
		if err != nil {
			return fmt.Errorf("недопустимое значение %q для %s: ожидается true/false", value, key)
		}
		cfg.TUI.ViKeys = parsed
	case "tui.refresh_interval":
		parsed, err := strconv.Atoi(strings.TrimSpace(value))
		if err != nil {
			return fmt.Errorf("недопустимое значение %q для %s: ожидается целое число", value, key)
		}
		cfg.TUI.RefreshInterval = parsed
	default:
		return fmt.Errorf("неизвестный ключ конфигурации %q", key)
	}

	return nil
}
