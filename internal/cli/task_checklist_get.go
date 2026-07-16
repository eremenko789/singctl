package cli

import (
	"context"

	"github.com/spf13/cobra"
)

func newTaskChecklistGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <CHECKLIST_ITEM_ID>",
		Short: "Показать пункт чек-листа по ID",
		Long:  "Загрузить один пункт чек-листа и вывести его в выбранном формате (-o).",
		Example: `  singctl task checklist get C-uuid
  singctl task checklist get C-uuid -o json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := requireChecklistItemID(args[0])
			if err != nil {
				return err
			}
			session, settings, err := openAPISession()
			if err != nil {
				return err
			}
			item, err := session.GetChecklistItem(context.Background(), id)
			if err != nil {
				return err
			}
			return renderTaskRecordSet(cmd, settings, checklistItemToRecordSet(item), true)
		},
	}
}
