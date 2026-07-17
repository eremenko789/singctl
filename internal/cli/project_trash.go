package cli

import (
	"context"

	"github.com/spf13/cobra"
)

func newProjectTrashCmd() *cobra.Command {
	var date string
	cmd := &cobra.Command{
		Use:   "trash <ID>",
		Short: "Переместить проект в корзину",
		Long:  "Установить deleteDate. Без --date используется сегодняшняя локальная дата (YYYY-MM-DD).",
		Example: `  singctl project trash P-uuid
  singctl project trash P-uuid --date 2026-07-16 -o json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := requireProjectID(args[0])
			if err != nil {
				return err
			}
			d, err := resolveProjectDateFlag(cmd, date)
			if err != nil {
				return err
			}
			session, settings, err := openAPISession()
			if err != nil {
				return err
			}
			project, err := session.TrashProject(context.Background(), id, d)
			if err != nil {
				return err
			}
			return renderProjectRecordSet(cmd, settings, projectToRecordSet(project), true)
		},
	}
	cmd.Flags().StringVar(&date, "date", "", "дата корзины YYYY-MM-DD (по умолчанию сегодня)")
	return cmd
}
