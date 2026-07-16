package output

import "testing"

func TestResolveFormat(t *testing.T) {
	tests := []struct {
		name         string
		flagSet      bool
		flagValue    string
		configFormat string
		want         Format
	}{
		{"flag_wins", true, "json", "csv", FormatJSON},
		{"flag_yaml", true, "yaml", "table", FormatYAML},
		{"config_when_no_flag", false, "", "csv", FormatCSV},
		{"default_table", false, "", "", FormatTable},
		{"invalid_config", false, "", "xml", FormatTable},
		{"invalid_flag_fallback", true, "xml", "json", FormatTable},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ResolveFormat(tt.flagSet, tt.flagValue, tt.configFormat)
			if got != tt.want {
				t.Fatalf("got %q want %q", got, tt.want)
			}
		})
	}
}
