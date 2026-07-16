package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/eremenko789/singctl/internal/api"
	"github.com/spf13/cobra"
)

func newTaskChecklistUpdateCmd() *cobra.Command {
	var (
		title  string
		done   bool
		undone bool
	)
	cmd := &cobra.Command{
		Use:   "update <CHECKLIST_ITEM_ID>",
		Short: "Обновить пункт чек-листа",
		Long: `Частично обновить пункт: --title, --done и/или --undone.

Нужен хотя бы один изменяющий флаг. --done и --undone взаимоисключающие.
Смена parent и parentOrder с CLI не поддерживаются.`,
		Example: `  singctl task checklist update C-uuid --done
  singctl task checklist update C-uuid --title "Новый текст" -o json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := requireChecklistItemID(args[0])
			if err != nil {
				return err
			}
			flags := cmd.Flags()
			if !flags.Changed("title") && !flags.Changed("done") && !flags.Changed("undone") {
				return fmt.Errorf("укажите хотя бы один флаг: --title, --done или --undone")
			}
			if flags.Changed("done") && flags.Changed("undone") {
				return fmt.Errorf("--done и --undone взаимоисключающие")
			}
			var in api.ChecklistWriteInput
			if flags.Changed("title") {
				t := strings.TrimSpace(title)
				if t == "" {
					return fmt.Errorf("--title не может быть пустым")
				}
				in.Title = &t
			}
			if flags.Changed("done") {
				d := true
				in.Done = &d
			}
			if flags.Changed("undone") {
				d := false
				in.Done = &d
			}
			session, settings, err := openAPISession()
			if err != nil {
				return err
			}
			item, err := session.UpdateChecklistItem(context.Background(), id, in)
			if err != nil {
				return err
			}
			return renderTaskRecordSet(cmd, settings, checklistItemToRecordSet(item), true)
		},
	}
	cmd.Flags().StringVar(&title, "title", "", "новый заголовок")
	cmd.Flags().BoolVar(&done, "done", false, "отметить выполненным")
	cmd.Flags().BoolVar(&undone, "undone", false, "снять отметку выполнения")
	// Bind vars so cobra Changed() tracks flags; values forced true/false above.
	_ = done
	_ = undone
	return cmd
}
