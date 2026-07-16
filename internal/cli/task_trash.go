package cli

import (
	"context"

	"github.com/spf13/cobra"
)

func newTaskTrashCmd() *cobra.Command {
	var date string
	cmd := &cobra.Command{
		Use:   "trash <ID>",
		Short: "Переместить задачу в корзину",
		Long:  "Установить deleteDate. Без --date используется сегодняшняя локальная дата (YYYY-MM-DD).",
		Example: `  singctl task trash T-uuid
  singctl task trash T-uuid --date 2026-07-16 -o json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := requireTaskID(args[0])
			if err != nil {
				return err
			}
			d, err := resolveTaskDateFlag(cmd, date)
			if err != nil {
				return err
			}
			session, settings, err := openAPISession()
			if err != nil {
				return err
			}
			task, err := session.TrashTask(context.Background(), id, d)
			if err != nil {
				return err
			}
			return renderTaskRecordSet(cmd, settings, taskToRecordSet(task), true)
		},
	}
	cmd.Flags().StringVar(&date, "date", "", "дата корзины YYYY-MM-DD (по умолчанию сегодня)")
	return cmd
}
