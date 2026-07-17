package cli

import (
	"github.com/eremenko789/singctl/internal/api"
	cfgpkg "github.com/eremenko789/singctl/internal/config"
	"github.com/eremenko789/singctl/internal/output"
	"github.com/spf13/cobra"
)

var sectionColumns = []output.Column{
	{Key: "id", Title: "ID"},
	{Key: "title", Title: "Title"},
	{Key: "parent", Title: "Parent"},
	{Key: "parentOrder", Title: "Order"},
	{Key: "removed", Title: "Removed?"},
}

func sectionsToRecordSet(sections []api.Section) output.RecordSet {
	rows := make([]map[string]any, 0, len(sections))
	for _, s := range sections {
		rows = append(rows, sectionToRow(s))
	}
	return output.RecordSet{Columns: sectionColumns, Rows: rows}
}

func sectionToRecordSet(s api.Section) output.RecordSet {
	return output.RecordSet{
		Columns: sectionColumns,
		Rows:    []map[string]any{sectionToRow(s)},
	}
}

func sectionToRow(s api.Section) map[string]any {
	return map[string]any{
		"id":          nullIfEmpty(s.ID),
		"title":       nullIfEmpty(s.Title),
		"parent":      nullIfEmpty(s.Parent),
		"parentOrder": s.ParentOrder,
		"removed":     s.Removed,
	}
}

func renderSectionRecordSet(cmd *cobra.Command, settings cfgpkg.EffectiveSettings, set output.RecordSet, singleObject bool) error {
	return output.Render(cmd.OutOrStdout(), set, buildProjectRenderOptions(cmd, settings, singleObject))
}
