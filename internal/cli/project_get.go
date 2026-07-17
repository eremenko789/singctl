package cli

import (
	"context"

	"github.com/spf13/cobra"
)

func newProjectGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <ID>",
		Short: "Показать проект по ID",
		Long:  "Загрузить один проект и вывести его в выбранном формате (-o).",
		Example: `  singctl project get P-uuid
  singctl project get P-uuid -o json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := requireProjectID(args[0])
			if err != nil {
				return err
			}
			session, settings, err := openAPISession()
			if err != nil {
				return err
			}
			project, err := session.GetProject(context.Background(), id)
			if err != nil {
				return err
			}
			return renderProjectRecordSet(cmd, settings, projectToRecordSet(project), true)
		},
	}
}
