package cli

import (
	"context"
	"fmt"

	"github.com/eremenko789/singctl/internal/api"
	"github.com/spf13/cobra"
)

func newProjectListCmd() *cobra.Command {
	var (
		archived bool
		removed  bool
		limit    int
		offset   int
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "Список проектов",
		Long: `Показать список проектов с фильтрами.

Флаги: --archived, --removed, --limit (1…1000), --offset (≥0).
Shared/collaborative projects API не возвращает.`,
		Example: `  singctl project list
  singctl project list --archived --limit 20 -o json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			query, err := buildProjectListQuery(cmd, archived, removed, limit, offset)
			if err != nil {
				return err
			}
			session, settings, err := openAPISession()
			if err != nil {
				return err
			}
			projects, err := session.ListProjects(context.Background(), query)
			if err != nil {
				return err
			}
			return renderProjectRecordSet(cmd, settings, projectsToRecordSet(projects), false)
		},
	}

	f := cmd.Flags()
	f.BoolVar(&archived, "archived", false, "включать архивные проекты")
	f.BoolVar(&removed, "removed", false, "включать удалённые в корзину")
	f.IntVar(&limit, "limit", 0, "максимум записей (1…1000)")
	f.IntVar(&offset, "offset", 0, "смещение (≥0)")
	return cmd
}

func buildProjectListQuery(cmd *cobra.Command, archived, removed bool, limit, offset int) (api.ProjectListQuery, error) {
	var q api.ProjectListQuery
	if cmd.Flags().Changed("archived") {
		a := archived
		q.IncludeArchived = &a
	}
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
