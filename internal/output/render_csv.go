package output

import (
	"encoding/csv"
	"io"
)

func renderCSV(w io.Writer, set RecordSet, opts RenderOptions) error {
	layout := layoutOf(opts)
	cw := csv.NewWriter(w)
	header := make([]string, len(set.Columns))
	for i, c := range set.Columns {
		header[i] = c.Key
	}
	if err := cw.Write(header); err != nil {
		return err
	}
	for _, row := range set.Rows {
		cells := make([]string, len(set.Columns))
		for i, c := range set.Columns {
			var v any
			if row != nil {
				v = row[c.Key]
			}
			cells[i] = cellString(v, layout)
		}
		if err := cw.Write(cells); err != nil {
			return err
		}
	}
	cw.Flush()
	return cw.Error()
}
