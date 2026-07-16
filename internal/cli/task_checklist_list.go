package cli

import (
	"context"

	"github.com/eremenko789/singctl/internal/api"
	"github.com/spf13/cobra"
)

func newTaskChecklistListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list <TASK_ID>",
		Short: "Список пунктов чек-листа задачи",
		Long: `Показать пункты чек-листа задачи.

Сначала проверяется существование задачи (как task get).
Фильтр — только parent = TASK_ID; пагинация и includeRemoved не поддерживаются.`,
		Example: `  singctl task checklist list T-uuid
  singctl task checklist list T-uuid -o json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			taskID, err := requireTaskID(args[0])
			if err != nil {
				return err
			}
			session, settings, err := openAPISession()
			if err != nil {
				return err
			}
			if _, err := session.GetTask(context.Background(), taskID); err != nil {
				return err
			}
			items, err := session.ListChecklistItems(context.Background(), api.ChecklistListQuery{Parent: taskID})
			if err != nil {
				return err
			}
			return renderTaskRecordSet(cmd, settings, checklistItemsToRecordSet(items), false)
		},
	}
}
