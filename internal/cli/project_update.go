package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

func newProjectUpdateCmd() *cobra.Command {
	var f projectWriteFlags
	cmd := &cobra.Command{
		Use:   "update <ID>",
		Short: "Обновить проект",
		Long: `Частично обновить проект. Нужен хотя бы один write-флаг.

--note передаётся as-is; API может ожидать delta-формат заметки.
--emoji: unicode (💞) или hex (1f49e).`,
		Example: `  singctl project update P-uuid --title "Новое имя"
  singctl project update P-uuid --emoji 1f49e -o json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := requireProjectID(args[0])
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
			project, err := session.UpdateProject(context.Background(), id, in)
			if err != nil {
				return err
			}
			return renderProjectRecordSet(cmd, settings, projectToRecordSet(project), true)
		},
	}
	bindProjectWriteFlags(cmd, &f)
	return cmd
}
