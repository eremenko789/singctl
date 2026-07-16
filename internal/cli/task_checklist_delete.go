package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func newTaskChecklistDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "delete <CHECKLIST_ITEM_ID>",
		Short:   "Удалить пункт чек-листа",
		Long:    "Безвозвратно удалить пункт чек-листа. При успехе stdout пуст.",
		Example: `  singctl task checklist delete C-uuid`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := requireChecklistItemID(args[0])
			if err != nil {
				return err
			}
			session, _, err := openAPISession()
			if err != nil {
				return err
			}
			return session.DeleteChecklistItem(context.Background(), id)
		},
	}
}

func requireChecklistItemID(id string) (string, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return "", fmt.Errorf("id пункта чек-листа не задан")
	}
	return id, nil
}
