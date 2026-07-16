package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/eremenko789/singctl/internal/api"
	"github.com/spf13/cobra"
)

func newTaskKanbanListCmd() *cobra.Command {
	var (
		taskID   string
		statusID string
	)
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Список канбан-связей",
		Long: `Показать связи задача↔колонка.

Опциональные фильтры --task и --status. Без pre-check задачи/колонки.
Пагинация и includeRemoved не поддерживаются.`,
		Example: `  singctl task kanban list
  singctl task kanban list --task T-uuid -o json
  singctl task kanban list --status KS-uuid`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			query := api.KanbanLinkListQuery{}
			if cmd.Flags().Changed("task") {
				tid := strings.TrimSpace(taskID)
				if tid == "" {
					return fmt.Errorf("--task не может быть пустым")
				}
				query.TaskID = tid
			}
			if cmd.Flags().Changed("status") {
				sid := strings.TrimSpace(statusID)
				if sid == "" {
					return fmt.Errorf("--status не может быть пустым")
				}
				query.StatusID = sid
			}
			session, settings, err := openAPISession()
			if err != nil {
				return err
			}
			links, err := session.ListKanbanLinks(context.Background(), query)
			if err != nil {
				return err
			}
			return renderTaskRecordSet(cmd, settings, kanbanLinksToRecordSet(links), false)
		},
	}
	cmd.Flags().StringVar(&taskID, "task", "", "фильтр по ID задачи")
	cmd.Flags().StringVar(&statusID, "status", "", "фильтр по ID колонки (statusId)")
	return cmd
}
