package cli

import (
	"context"

	"github.com/spf13/cobra"
)

func newTaskGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <ID>",
		Short: "Показать задачу по ID",
		Long:  "Загрузить одну задачу и вывести её в выбранном формате (-o).",
		Example: `  singctl task get T-uuid
  singctl task get T-uuid -o json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := requireTaskID(args[0])
			if err != nil {
				return err
			}
			session, settings, err := openAPISession()
			if err != nil {
				return err
			}
			task, err := session.GetTask(context.Background(), id)
			if err != nil {
				return err
			}
			return renderTaskRecordSet(cmd, settings, taskToRecordSet(task), true)
		},
	}
}
