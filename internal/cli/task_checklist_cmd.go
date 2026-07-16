package cli

import (
	"errors"

	"github.com/spf13/cobra"
)

func newTaskChecklistCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "checklist",
		Short: "Чек-лист задачи",
		Long: `Управление пунктами чек-листа задачи.

Подкоманды: list, get, add, update, delete.
Родитель пункта — всегда задача (TASK_ID). Без --order и пагинации list.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return errors.New("укажите подкоманду: list, get, add, update или delete")
		},
	}
	cmd.AddCommand(newTaskChecklistListCmd())
	cmd.AddCommand(newTaskChecklistGetCmd())
	cmd.AddCommand(newTaskChecklistAddCmd())
	cmd.AddCommand(newTaskChecklistUpdateCmd())
	cmd.AddCommand(newTaskChecklistDeleteCmd())
	return cmd
}
