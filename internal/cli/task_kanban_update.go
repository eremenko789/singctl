package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/eremenko789/singctl/internal/api"
	"github.com/spf13/cobra"
)

func newTaskKanbanUpdateCmd() *cobra.Command {
	var (
		taskID   string
		columnID string
		order    float32
	)
	cmd := &cobra.Command{
		Use:   "update <LINK_ID>",
		Short: "Обновить канбан-связь",
		Long: `Частично обновить связь: --task, --column и/или --order.

Нужен хотя бы один изменяющий флаг. Pre-check задачи не выполняется.`,
		Example: `  singctl task kanban update KTS-uuid --column KS-uuid
  singctl task kanban update KTS-uuid --order 2 -o json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := requireKanbanLinkID(args[0])
			if err != nil {
				return err
			}
			changedTask := cmd.Flags().Changed("task")
			changedCol := cmd.Flags().Changed("column")
			changedOrder := cmd.Flags().Changed("order")
			if !changedTask && !changedCol && !changedOrder {
				return fmt.Errorf("укажите хотя бы один флаг: --task, --column или --order")
			}
			in := api.KanbanLinkWriteInput{}
			if changedTask {
				taskID = strings.TrimSpace(taskID)
				if taskID == "" {
					return fmt.Errorf("--task не может быть пустым")
				}
				in.TaskID = &taskID
			}
			if changedCol {
				columnID = strings.TrimSpace(columnID)
				if columnID == "" {
					return fmt.Errorf("--column не может быть пустым")
				}
				in.StatusID = &columnID
			}
			if changedOrder {
				if order < 0 {
					return fmt.Errorf("--order должен быть ≥ 0")
				}
				o := order
				in.KanbanOrder = &o
			}
			session, settings, err := openAPISession()
			if err != nil {
				return err
			}
			link, err := session.UpdateKanbanLink(context.Background(), id, in)
			if err != nil {
				return err
			}
			return renderTaskRecordSet(cmd, settings, kanbanLinkToRecordSet(link), true)
		},
	}
	cmd.Flags().StringVar(&taskID, "task", "", "новый ID задачи")
	cmd.Flags().StringVar(&columnID, "column", "", "новый ID колонки (statusId)")
	cmd.Flags().Float32Var(&order, "order", 0, "порядок в колонке (kanbanOrder)")
	return cmd
}
