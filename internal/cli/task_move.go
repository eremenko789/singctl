package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func newTaskMoveCmd() *cobra.Command {
	var columnID string
	cmd := &cobra.Command{
		Use:   "move <TASK_ID>",
		Short: "Переместить задачу в канбан-колонку",
		Long: `Переместить задачу в колонку: создать связь или обновить единственную.

Если активных связей нет — create; если одна — update statusId (в т.ч. та же колонка);
если несколько — ошибка (используйте task kanban list / update).
Порядок в колонке с CLI не задаётся. Сначала проверяется существование задачи.`,
		Example: `  singctl task move T-uuid --column KS-uuid
  singctl task move T-uuid --column KS-uuid -o json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			taskID, err := requireTaskID(args[0])
			if err != nil {
				return err
			}
			columnID = strings.TrimSpace(columnID)
			if columnID == "" {
				return fmt.Errorf("--column обязателен и не может быть пустым")
			}
			session, settings, err := openAPISession()
			if err != nil {
				return err
			}
			if _, err := session.GetTask(context.Background(), taskID); err != nil {
				return err
			}
			link, err := session.MoveTaskToKanban(context.Background(), taskID, columnID)
			if err != nil {
				return err
			}
			return renderTaskRecordSet(cmd, settings, kanbanLinkToRecordSet(link), true)
		},
	}
	cmd.Flags().StringVar(&columnID, "column", "", "ID канбан-колонки (statusId)")
	_ = cmd.MarkFlagRequired("column")
	return cmd
}
