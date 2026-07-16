package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/eremenko789/singctl/internal/api"
	"github.com/spf13/cobra"
)

func newTaskKanbanCreateCmd() *cobra.Command {
	var (
		taskID   string
		columnID string
		order    float32
	)
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Создать канбан-связь",
		Long: `Создать связь задачи с канбан-колонкой.

Обязательны --task и --column. Опционально --order (kanbanOrder ≥ 0).
Сначала проверяется существование задачи (как task get).
Уникальность связи на клиенте не проверяется.`,
		Example: `  singctl task kanban create --task T-uuid --column KS-uuid
  singctl task kanban create --task T-uuid --column KS-uuid --order 1 -o json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			taskID = strings.TrimSpace(taskID)
			columnID = strings.TrimSpace(columnID)
			if taskID == "" {
				return fmt.Errorf("--task обязателен и не может быть пустым")
			}
			if columnID == "" {
				return fmt.Errorf("--column обязателен и не может быть пустым")
			}
			if cmd.Flags().Changed("order") && order < 0 {
				return fmt.Errorf("--order должен быть ≥ 0")
			}
			session, settings, err := openAPISession()
			if err != nil {
				return err
			}
			if _, err := session.GetTask(context.Background(), taskID); err != nil {
				return err
			}
			in := api.KanbanLinkWriteInput{
				TaskID:   &taskID,
				StatusID: &columnID,
			}
			if cmd.Flags().Changed("order") {
				o := order
				in.KanbanOrder = &o
			}
			link, err := session.CreateKanbanLink(context.Background(), in)
			if err != nil {
				return err
			}
			return renderTaskRecordSet(cmd, settings, kanbanLinkToRecordSet(link), true)
		},
	}
	cmd.Flags().StringVar(&taskID, "task", "", "ID задачи")
	cmd.Flags().StringVar(&columnID, "column", "", "ID канбан-колонки (statusId)")
	cmd.Flags().Float32Var(&order, "order", 0, "порядок задачи в колонке (kanbanOrder)")
	_ = cmd.MarkFlagRequired("task")
	_ = cmd.MarkFlagRequired("column")
	return cmd
}
