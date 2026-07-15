package cli

import (
	"fmt"
	"strings"
)

// OutputFormat is a validated --output / -o enum: table|json|yaml|csv.
type OutputFormat string

// Supported --output / -o values.
const (
	OutputTable OutputFormat = "table"
	OutputJSON  OutputFormat = "json"
	OutputYAML  OutputFormat = "yaml"
	OutputCSV   OutputFormat = "csv"
)

var validOutputs = map[string]OutputFormat{
	"table": OutputTable,
	"json":  OutputJSON,
	"yaml":  OutputYAML,
	"csv":   OutputCSV,
}

// String implements pflag.Value / fmt.Stringer.
func (o *OutputFormat) String() string {
	if o == nil || *o == "" {
		return string(OutputTable)
	}
	return string(*o)
}

// Set implements pflag.Value. Invalid values fail during flag parse (before help/version).
func (o *OutputFormat) Set(s string) error {
	v, ok := validOutputs[strings.ToLower(strings.TrimSpace(s))]
	if !ok {
		return fmt.Errorf("недопустимый формат вывода %q: допустимы table, json, yaml, csv", s)
	}
	*o = v
	return nil
}

// Type implements pflag.Value.
func (o *OutputFormat) Type() string {
	return "format"
}
