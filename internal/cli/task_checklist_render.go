package cli

import (
	"github.com/eremenko789/singctl/internal/api"
	"github.com/eremenko789/singctl/internal/output"
)

var checklistColumns = []output.Column{
	{Key: "id", Title: "ID"},
	{Key: "title", Title: "Title"},
	{Key: "done", Title: "Done"},
	{Key: "parent", Title: "Parent"},
	{Key: "parentOrder", Title: "Order"},
}

func checklistItemsToRecordSet(items []api.ChecklistItem) output.RecordSet {
	rows := make([]map[string]any, 0, len(items))
	for _, it := range items {
		rows = append(rows, checklistItemToRow(it))
	}
	return output.RecordSet{Columns: checklistColumns, Rows: rows}
}

func checklistItemToRecordSet(it api.ChecklistItem) output.RecordSet {
	return output.RecordSet{
		Columns: checklistColumns,
		Rows:    []map[string]any{checklistItemToRow(it)},
	}
}

func checklistItemToRow(it api.ChecklistItem) map[string]any {
	return map[string]any{
		"id":          nullIfEmpty(it.ID),
		"title":       nullIfEmpty(it.Title),
		"done":        it.Done,
		"parent":      nullIfEmpty(it.Parent),
		"parentOrder": it.ParentOrder,
	}
}
