package cli

import (
	"github.com/eremenko789/singctl/internal/api"
	"github.com/eremenko789/singctl/internal/output"
)

var kanbanColumns = []output.Column{
	{Key: "id", Title: "ID"},
	{Key: "taskId", Title: "Task"},
	{Key: "statusId", Title: "Column"},
	{Key: "kanbanOrder", Title: "Order"},
}

func kanbanLinksToRecordSet(links []api.KanbanLink) output.RecordSet {
	rows := make([]map[string]any, 0, len(links))
	for _, l := range links {
		rows = append(rows, kanbanLinkToRow(l))
	}
	return output.RecordSet{Columns: kanbanColumns, Rows: rows}
}

func kanbanLinkToRecordSet(l api.KanbanLink) output.RecordSet {
	return output.RecordSet{
		Columns: kanbanColumns,
		Rows:    []map[string]any{kanbanLinkToRow(l)},
	}
}

func kanbanLinkToRow(l api.KanbanLink) map[string]any {
	return map[string]any{
		"id":          nullIfEmpty(l.ID),
		"taskId":      nullIfEmpty(l.TaskID),
		"statusId":    nullIfEmpty(l.StatusID),
		"kanbanOrder": l.KanbanOrder,
	}
}
