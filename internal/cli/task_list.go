package cli

import (
	"context"
	"fmt"

	"github.com/eremenko789/singctl/internal/api"
	"github.com/spf13/cobra"
)

func newTaskListCmd() *cobra.Command {
	var (
		project       string
		parent        string
		from          string
		to            string
		archived      bool
		removed       bool
		limit         int
		offset        int
		allRecurrence bool
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "Список задач",
		Long: `Показать список задач с фильтрами.

Флаги: --project, --parent, --from, --to, --archived, --removed,
--limit (1…1000), --offset (≥0), --all-recurrence.`,
		Example: `  singctl task list
  singctl task list --project P-uuid --limit 20 -o json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			query, err := buildTaskListQuery(cmd, project, parent, from, to, archived, removed, limit, offset, allRecurrence)
			if err != nil {
				return err
			}
			session, settings, err := openAPISession()
			if err != nil {
				return err
			}
			tasks, err := session.ListTasks(context.Background(), query)
			if err != nil {
				return err
			}
			return renderTaskRecordSet(cmd, settings, tasksToRecordSet(tasks), false)
		},
	}

	f := cmd.Flags()
	f.StringVar(&project, "project", "", "фильтр по ID проекта")
	f.StringVar(&parent, "parent", "", "фильтр по ID родительской задачи")
	f.StringVar(&from, "from", "", "startDateFrom (YYYY-MM-DD)")
	f.StringVar(&to, "to", "", "startDateTo (YYYY-MM-DD)")
	f.BoolVar(&archived, "archived", false, "включать архивные задачи")
	f.BoolVar(&removed, "removed", false, "включать удалённые в корзину")
	f.IntVar(&limit, "limit", 0, "максимум записей (1…1000)")
	f.IntVar(&offset, "offset", 0, "смещение (≥0)")
	f.BoolVar(&allRecurrence, "all-recurrence", false, "все экземпляры повторяющихся задач")
	return cmd
}

func buildTaskListQuery(
	cmd *cobra.Command,
	project, parent, from, to string,
	archived, removed bool,
	limit, offset int,
	allRecurrence bool,
) (api.TaskListQuery, error) {
	var q api.TaskListQuery
	if cmd.Flags().Changed("project") {
		p := project
		q.ProjectID = &p
	}
	if cmd.Flags().Changed("parent") {
		p := parent
		q.Parent = &p
	}
	if cmd.Flags().Changed("from") {
		if _, err := api.ParseDate(from); err != nil {
			return q, err
		}
		f := from
		q.StartFrom = &f
	}
	if cmd.Flags().Changed("to") {
		if _, err := api.ParseDate(to); err != nil {
			return q, err
		}
		t := to
		q.StartTo = &t
	}
	if cmd.Flags().Changed("archived") {
		a := archived
		q.IncludeArchived = &a
	}
	if cmd.Flags().Changed("removed") {
		r := removed
		q.IncludeRemoved = &r
	}
	if cmd.Flags().Changed("all-recurrence") {
		a := allRecurrence
		q.IncludeAllRecurrence = &a
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
