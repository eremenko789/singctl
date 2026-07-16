package cli

import (
	"encoding/csv"
	"encoding/json"
	"io"
	"strconv"

	cfgpkg "github.com/eremenko789/singctl/internal/config"
	"github.com/olekukonko/tablewriter"
	"go.yaml.in/yaml/v3"
)

type configRow struct {
	Key   string
	Value string
}

func renderConfigYAML(w io.Writer, cfg cfgpkg.Document) error {
	data, err := yaml.Marshal(maskedConfig(cfg))
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

func renderConfigJSON(w io.Writer, cfg cfgpkg.Document) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(maskedConfig(cfg))
}

func renderConfigCSV(w io.Writer, cfg cfgpkg.Document) error {
	writer := csv.NewWriter(w)
	if err := writer.Write([]string{"key", "value"}); err != nil {
		return err
	}
	for _, row := range configRows(cfg) {
		if err := writer.Write([]string{row.Key, row.Value}); err != nil {
			return err
		}
	}
	writer.Flush()
	return writer.Error()
}

func renderConfigTable(w io.Writer, cfg cfgpkg.Document) error {
	table := tablewriter.NewWriter(w)
	table.Header("Ключ", "Значение")
	for _, row := range configRows(cfg) {
		if err := table.Append([]string{row.Key, row.Value}); err != nil {
			return err
		}
	}
	return table.Render()
}

func maskedConfig(cfg cfgpkg.Document) cfgpkg.Document {
	masked := cfg
	masked.API.Token = cfgpkg.MaskToken(masked.API.Token)
	return masked
}

func configRows(cfg cfgpkg.Document) []configRow {
	cfg = maskedConfig(cfg)
	return []configRow{
		{Key: "api.base_url", Value: cfg.API.BaseURL},
		{Key: "api.token", Value: cfg.API.Token},
		{Key: "api.timeout", Value: cfg.API.Timeout},
		{Key: "output.format", Value: cfg.Output.Format},
		{Key: "output.color", Value: strconv.FormatBool(cfg.Output.Color)},
		{Key: "output.date_format", Value: cfg.Output.DateFormat},
		{Key: "tui.theme", Value: cfg.TUI.Theme},
		{Key: "tui.vi_keys", Value: strconv.FormatBool(cfg.TUI.ViKeys)},
		{Key: "tui.refresh_interval", Value: strconv.Itoa(cfg.TUI.RefreshInterval)},
	}
}
