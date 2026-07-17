package cli

import (
	"context"

	"github.com/spf13/cobra"
)

func newProjectSectionDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "delete <SECTION_ID>",
		Short:   "Удалить секцию навсегда",
		Long:    "Безвозвратно удалить секцию. При успехе stdout пуст.",
		Example: `  singctl project section delete Q-uuid`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := requireSectionID(args[0])
			if err != nil {
				return err
			}
			session, _, err := openAPISession()
			if err != nil {
				return err
			}
			return session.DeleteSection(context.Background(), id)
		},
	}
}
