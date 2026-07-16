package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func newTaskKanbanGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <LINK_ID>",
		Short: "Получить канбан-связь по ID",
		Long:  "Показать одну связь задача↔колонка по идентификатору связи.",
		Example: `  singctl task kanban get KTS-uuid
  singctl task kanban get KTS-uuid -o json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := requireKanbanLinkID(args[0])
			if err != nil {
				return err
			}
			session, settings, err := openAPISession()
			if err != nil {
				return err
			}
			link, err := session.GetKanbanLink(context.Background(), id)
			if err != nil {
				return err
			}
			return renderTaskRecordSet(cmd, settings, kanbanLinkToRecordSet(link), true)
		},
	}
}

func requireKanbanLinkID(id string) (string, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return "", fmt.Errorf("id канбан-связи не задан")
	}
	return id, nil
}
