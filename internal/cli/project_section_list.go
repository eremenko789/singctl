package cli

import (
	"context"
	"fmt"

	"github.com/eremenko789/singctl/internal/api"
	"github.com/spf13/cobra"
)

func newProjectSectionListCmd() *cobra.Command {
	var (
		removed bool
		limit   int
		offset  int
	)

	cmd := &cobra.Command{
		Use:   "list <PROJECT_ID>",
		Short: "Список секций проекта",
		Long: `Показать секции (task groups) указанного проекта.

Обязателен PROJECT_ID. Флаги: --removed, --limit (1…1000), --offset (≥0).`,
		Example: `  singctl project section list P-uuid
  singctl project section list P-uuid --removed --limit 20 -o json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectID, err := requireProjectID(args[0])
			if err != nil {
				return err
			}
			query, err := buildSectionListQuery(cmd, projectID, removed, limit, offset)
			if err != nil {
				return err
			}
			session, settings, err := openAPISession()
			if err != nil {
				return err
			}
			sections, err := session.ListSections(context.Background(), query)
			if err != nil {
				return err
			}
			return renderSectionRecordSet(cmd, settings, sectionsToRecordSet(sections), false)
		},
	}

	f := cmd.Flags()
	f.BoolVar(&removed, "removed", false, "включать удалённые секции")
	f.IntVar(&limit, "limit", 0, "максимум записей (1…1000)")
	f.IntVar(&offset, "offset", 0, "смещение (≥0)")
	return cmd
}

func buildSectionListQuery(cmd *cobra.Command, parent string, removed bool, limit, offset int) (api.SectionListQuery, error) {
	q := api.SectionListQuery{Parent: parent}
	if cmd.Flags().Changed("removed") {
		r := removed
		q.IncludeRemoved = &r
	}
	if cmd.Flags().Changed("limit") {
		if limit < 1 || limit > 1000 {
			return q, fmt.Errorf("--limit должен быть от 1 до 1000")
		}
		l := limit
		q.MaxCount = &l
	}
	if cmd.Flags().Changed("offset") {
		if offset < 0 {
			return q, fmt.Errorf("--offset должен быть ≥ 0")
		}
		o := offset
		q.Offset = &o
	}
	return q, nil
}
