package output

// ColorEnabled reports whether ANSI color may be used for human table output.
// Machine formats (json/yaml/csv) never emit ANSI regardless of this result.
//
// Precedence: --no-color or non-empty NO_COLOR → off; non-TTY → off;
// configColor false → off; otherwise on.
func ColorEnabled(isTTY, noColorFlag bool, noColorEnv string, configColor bool) bool {
	if noColorFlag || noColorEnv != "" {
		return false
	}
	if !isTTY {
		return false
	}
	if !configColor {
		return false
	}
	return true
}
