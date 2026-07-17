package cli

import (
	"context"

	"github.com/eremenko789/singctl/internal/api"
	"github.com/spf13/cobra"
)

func newProjectArchiveCmd() *cobra.Command {
	var date string
	cmd := &cobra.Command{
		Use:   "archive <ID>",
		Short: "Архивировать проект",
		Long:  "Установить journalDate. Без --date используется сегодняшняя локальная дата (YYYY-MM-DD).",
		Example: `  singctl project archive P-uuid
  singctl project archive P-uuid --date 2026-07-16 -o json`,
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
			project, err := session.ArchiveProject(context.Background(), id, d)
			if err != nil {
				return err
			}
			return renderProjectRecordSet(cmd, settings, projectToRecordSet(project), true)
		},
	}
	cmd.Flags().StringVar(&date, "date", "", "дата архива YYYY-MM-DD (по умолчанию сегодня)")
	return cmd
}

func resolveProjectDateFlag(cmd *cobra.Command, date string) (string, error) {
	if !cmd.Flags().Changed("date") {
		return api.TodayCalendarDate(), nil
	}
	if _, err := api.ParseDate(date); err != nil {
		return "", err
	}
	return date, nil
}
