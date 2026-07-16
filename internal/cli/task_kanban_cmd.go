package cli

import (
	"errors"

	"github.com/spf13/cobra"
)

func newTaskKanbanCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kanban",
		Short: "Канбан-связь задачи с колонкой",
		Long: `Управление связями задача↔канбан-колонка (/v2/kanban-task-status).

Подкоманды: list, get, create, update, delete.
Перемещение задачи: singctl task move (create/update связи).
Управление колонками проекта — отдельно (project column).`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return errors.New("укажите подкоманду: list, get, create, update или delete")
		},
	}
	cmd.AddCommand(newTaskKanbanListCmd())
	cmd.AddCommand(newTaskKanbanGetCmd())
	cmd.AddCommand(newTaskKanbanCreateCmd())
	cmd.AddCommand(newTaskKanbanUpdateCmd())
	cmd.AddCommand(newTaskKanbanDeleteCmd())
	return cmd
}
