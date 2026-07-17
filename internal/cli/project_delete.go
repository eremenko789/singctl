package cli

import (
	"context"

	"github.com/spf13/cobra"
)

func newProjectDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "delete <ID>",
		Short:   "Удалить проект навсегда",
		Long:    "Безвозвратно удалить проект. При успехе stdout пуст.",
		Example: `  singctl project delete P-uuid`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := requireProjectID(args[0])
			if err != nil {
				return err
			}
			session, _, err := openAPISession()
			if err != nil {
				return err
			}
			return session.DeleteProject(context.Background(), id)
		},
	}
}
