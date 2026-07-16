package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

func newTaskUpdateCmd() *cobra.Command {
	var f taskWriteFlags
	cmd := &cobra.Command{
		Use:   "update <ID>",
		Short: "Обновить задачу",
		Long: `Частично обновить задачу. Нужен хотя бы один write-флаг.

--note передаётся as-is; API может ожидать delta-формат заметки.`,
		Example: `  singctl task update T-uuid --title "Новое имя"
  singctl task update T-uuid --priority 0 -o json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := requireTaskID(args[0])
			if err != nil {
				return err
			}
			if !f.anyWriteFlagSet(cmd) {
				return fmt.Errorf("укажите хотя бы один флаг для обновления")
			}
			in, err := f.toInput(cmd, false)
			if err != nil {
				return err
			}
			session, settings, err := openAPISession()
			if err != nil {
				return err
			}
			task, err := session.UpdateTask(context.Background(), id, in)
			if err != nil {
				return err
			}
			return renderTaskRecordSet(cmd, settings, taskToRecordSet(task), true)
		},
	}
	bindTaskWriteFlags(cmd, &f)
	return cmd
}
