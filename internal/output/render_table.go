package output

import (
	"io"
	"os"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/renderer"
)

func renderTable(w io.Writer, set RecordSet, opts RenderOptions) error {
	layout := layoutOf(opts)
	var table *tablewriter.Table
	if opts.Color {
		// fatih/color.New permanently disables a Color when NO_COLOR is set in the
		// process env (even if color.NoColor is flipped). Colorized Tint.Apply uses
		// New(), so temporarily clear NO_COLOR for this explicit Color=true render.
		prevEnv, hadNOColor := os.LookupEnv("NO_COLOR")
		_ = os.Unsetenv("NO_COLOR")
		prevNoColor := color.NoColor
		color.NoColor = false
		defer func() {
			color.NoColor = prevNoColor
			if hadNOColor {
				_ = os.Setenv("NO_COLOR", prevEnv)
			}
		}()
		table = tablewriter.NewTable(w, tablewriter.WithRenderer(renderer.NewColorized()))
	} else {
		table = tablewriter.NewWriter(w)
	}
	headers := make([]any, len(set.Columns))
	for i, c := range set.Columns {
		headers[i] = c.headerTitle()
	}
	table.Header(headers...)
	for _, row := range set.Rows {
		cells := make([]string, len(set.Columns))
		for i, c := range set.Columns {
			var v any
			if row != nil {
				v = row[c.Key]
			}
			cells[i] = cellString(v, layout)
		}
		if err := table.Append(cells); err != nil {
			return err
		}
	}
	return table.Render()
}
