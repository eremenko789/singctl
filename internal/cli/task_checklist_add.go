package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/eremenko789/singctl/internal/api"
	"github.com/spf13/cobra"
)

func newTaskChecklistAddCmd() *cobra.Command {
	var (
		title string
		done  bool
	)
	cmd := &cobra.Command{
		Use:   "add <TASK_ID>",
		Short: "Добавить пункт чек-листа",
		Long: `Создать пункт чек-листа у задачи. Обязателен --title (непустой).

Сначала проверяется существование задачи (как task get).
Опционально --done. Порядок пункта (parentOrder) не задаётся с CLI.`,
		Example: `  singctl task checklist add T-uuid --title "Купить молоко"
  singctl task checklist add T-uuid --title "Done already" --done -o json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			taskID, err := requireTaskID(args[0])
			if err != nil {
				return err
			}
			title = strings.TrimSpace(title)
			if title == "" {
				return fmt.Errorf("--title обязателен и не может быть пустым")
			}
			session, settings, err := openAPISession()
			if err != nil {
				return err
			}
			if _, err := session.GetTask(context.Background(), taskID); err != nil {
				return err
			}
			in := api.ChecklistWriteInput{
				Parent: &taskID,
				Title:  &title,
			}
			if cmd.Flags().Changed("done") {
				d := done
				in.Done = &d
			}
			item, err := session.CreateChecklistItem(context.Background(), in)
			if err != nil {
				return err
			}
			return renderTaskRecordSet(cmd, settings, checklistItemToRecordSet(item), true)
		},
	}
	cmd.Flags().StringVar(&title, "title", "", "заголовок пункта")
	cmd.Flags().BoolVar(&done, "done", false, "отметить выполненным при создании")
	_ = cmd.MarkFlagRequired("title")
	return cmd
}
