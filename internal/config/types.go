package config

// Default configuration values used when a field is absent from the YAML file.
const (
	DefaultAPIBaseURL       = "https://api.singularity-app.ru"
	DefaultAPITimeout       = "30s"
	DefaultOutputFormat     = "table"
	DefaultOutputDateFormat = "2006-01-02"
	DefaultTUITheme         = "dark"
)

// Document is the on-disk YAML configuration schema for singctl.
type Document struct {
	API    APIConfig    `yaml:"api" json:"api"`
	Output OutputConfig `yaml:"output" json:"output"`
	TUI    TUIConfig    `yaml:"tui" json:"tui"`
}

// APIConfig holds API endpoint, token, and timeout settings.
type APIConfig struct {
	BaseURL string `yaml:"base_url" json:"base_url"`
	Token   string `yaml:"token,omitempty" json:"token,omitempty"`
	Timeout string `yaml:"timeout" json:"timeout"`
}

// OutputConfig holds CLI output format and presentation settings.
type OutputConfig struct {
	Format     string `yaml:"format" json:"format"`
	Color      bool   `yaml:"color" json:"color"`
	DateFormat string `yaml:"date_format" json:"date_format"`
}

// TUIConfig holds interactive TUI presentation settings.
type TUIConfig struct {
	Theme           string `yaml:"theme" json:"theme"`
	ViKeys          bool   `yaml:"vi_keys" json:"vi_keys"`
	RefreshInterval int    `yaml:"refresh_interval" json:"refresh_interval"`
}

// EffectiveSettings is the resolved config path plus loaded document (with runtime overrides).
type EffectiveSettings struct {
	ConfigPath string
	FromFile   bool
	Config     Document
}

// DefaultConfig returns a Document filled with package defaults.
func DefaultConfig() Document {
	return Document{
		API: APIConfig{
			BaseURL: DefaultAPIBaseURL,
			Timeout: DefaultAPITimeout,
		},
		Output: OutputConfig{
			Format:     DefaultOutputFormat,
			Color:      true,
			DateFormat: DefaultOutputDateFormat,
		},
		TUI: TUIConfig{
			Theme:           DefaultTUITheme,
			ViKeys:          true,
			RefreshInterval: 0,
		},
	}
}
