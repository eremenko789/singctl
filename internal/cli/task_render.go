package cli

import (
	"github.com/eremenko789/singctl/internal/api"
	"github.com/eremenko789/singctl/internal/output"
)

var taskColumns = []output.Column{
	{Key: "id", Title: "ID"},
	{Key: "title", Title: "Title"},
	{Key: "projectId", Title: "Project"},
	{Key: "parent", Title: "Parent"},
	{Key: "priority", Title: "Priority"},
	{Key: "start", Title: "Start"},
	{Key: "journalDate", Title: "Archived"},
	{Key: "deleteDate", Title: "Trash"},
	{Key: "isNote", Title: "Note?"},
}

func tasksToRecordSet(tasks []api.Task) output.RecordSet {
	rows := make([]map[string]any, 0, len(tasks))
	for _, t := range tasks {
		rows = append(rows, taskToRow(t))
	}
	return output.RecordSet{Columns: taskColumns, Rows: rows}
}

func taskToRecordSet(t api.Task) output.RecordSet {
	return output.RecordSet{
		Columns: taskColumns,
		Rows:    []map[string]any{taskToRow(t)},
	}
}

func taskToRow(t api.Task) map[string]any {
	row := map[string]any{
		"id":          nullIfEmpty(t.ID),
		"title":       nullIfEmpty(t.Title),
		"projectId":   nullIfEmpty(t.ProjectID),
		"parent":      nullIfEmpty(t.Parent),
		"start":       nullIfEmpty(t.Start),
		"journalDate": nullIfEmpty(t.JournalDate),
		"deleteDate":  nullIfEmpty(t.DeleteDate),
		"isNote":      t.IsNote,
	}
	if t.Priority != nil {
		row["priority"] = *t.Priority
	} else {
		row["priority"] = nil
	}
	return row
}

func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}
