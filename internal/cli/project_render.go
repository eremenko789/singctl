package cli

import (
	"github.com/eremenko789/singctl/internal/api"
	"github.com/eremenko789/singctl/internal/output"
)

var projectColumns = []output.Column{
	{Key: "id", Title: "ID"},
	{Key: "title", Title: "Title"},
	{Key: "emoji", Title: "Emoji"},
	{Key: "color", Title: "Color"},
	{Key: "isNotebook", Title: "Notebook?"},
	{Key: "parent", Title: "Parent"},
	{Key: "journalDate", Title: "Archived"},
	{Key: "deleteDate", Title: "Trash"},
}

func projectsToRecordSet(projects []api.Project) output.RecordSet {
	rows := make([]map[string]any, 0, len(projects))
	for _, p := range projects {
		rows = append(rows, projectToRow(p))
	}
	return output.RecordSet{Columns: projectColumns, Rows: rows}
}

func projectToRecordSet(p api.Project) output.RecordSet {
	return output.RecordSet{
		Columns: projectColumns,
		Rows:    []map[string]any{projectToRow(p)},
	}
}

func projectToRow(p api.Project) map[string]any {
	return map[string]any{
		"id":          nullIfEmpty(p.ID),
		"title":       nullIfEmpty(p.Title),
		"emoji":       nullIfEmpty(p.Emoji),
		"color":       nullIfEmpty(p.Color),
		"isNotebook":  p.IsNotebook,
		"parent":      nullIfEmpty(p.Parent),
		"journalDate": nullIfEmpty(p.JournalDate),
		"deleteDate":  nullIfEmpty(p.DeleteDate),
	}
}
