package output

import (
	"fmt"
	"io"
)

// Render writes set to w using opts.Format.
func Render(w io.Writer, set RecordSet, opts RenderOptions) error {
	if err := validateRecordSet(set); err != nil {
		return err
	}
	switch opts.Format {
	case FormatJSON:
		return renderJSON(w, set, opts)
	case FormatYAML:
		return renderYAML(w, set, opts)
	case FormatCSV:
		return renderCSV(w, set, opts)
	case FormatTable:
		return renderTable(w, set, opts)
	default:
		return fmt.Errorf("output: unknown format %q", opts.Format)
	}
}
