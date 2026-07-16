package output

import "strings"

// Format is a record-set output representation.
type Format string

// Supported Format values.
const (
	FormatTable Format = "table"
	FormatJSON  Format = "json"
	FormatYAML  Format = "yaml"
	FormatCSV   Format = "csv"
)

var validFormats = map[string]Format{
	"table": FormatTable,
	"json":  FormatJSON,
	"yaml":  FormatYAML,
	"csv":   FormatCSV,
}

// ParseFormat maps a string to Format. ok is false when the value is not supported.
func ParseFormat(s string) (Format, bool) {
	f, ok := validFormats[strings.ToLower(strings.TrimSpace(s))]
	return f, ok
}

// ResolveFormat picks the effective format: explicit flag > config > table.
// When flagSet is true and flagValue is invalid, falls back to FormatTable
// (F01 normally rejects invalid flags before this is called).
func ResolveFormat(flagSet bool, flagValue, configFormat string) Format {
	if flagSet {
		if f, ok := ParseFormat(flagValue); ok {
			return f
		}
		return FormatTable
	}
	if f, ok := ParseFormat(configFormat); ok {
		return f
	}
	return FormatTable
}
