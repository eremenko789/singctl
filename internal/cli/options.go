package cli

// GlobalOptions holds session-level persistent flags for the root command.
type GlobalOptions struct {
	ConfigPath string
	Token      string
	Output     OutputFormat
	NoColor    bool
	Debug      bool
}
