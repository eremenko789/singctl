package cli

import (
	"context"

	"github.com/spf13/cobra"
)

func newTaskKanbanDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "delete <LINK_ID>",
		Short:   "Удалить канбан-связь",
		Long:    "Безвозвратно удалить связь задача↔колонка. При успехе stdout пуст.",
		Example: `  singctl task kanban delete KTS-uuid`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := requireKanbanLinkID(args[0])
			if err != nil {
				return err
			}
			session, _, err := openAPISession()
			if err != nil {
				return err
			}
			return session.DeleteKanbanLink(context.Background(), id)
		},
	}
}
