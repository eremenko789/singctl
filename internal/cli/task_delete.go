package cli

import (
	"context"

	"github.com/spf13/cobra"
)

func newTaskDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "delete <ID>",
		Short:   "Удалить задачу навсегда",
		Long:    "Безвозвратно удалить задачу. При успехе stdout пуст.",
		Example: `  singctl task delete T-uuid`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := requireTaskID(args[0])
			if err != nil {
				return err
			}
			session, _, err := openAPISession()
			if err != nil {
				return err
			}
			return session.DeleteTask(context.Background(), id)
		},
	}
}
