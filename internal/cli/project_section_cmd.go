package cli

import (
	"errors"

	"github.com/spf13/cobra"
)

func newProjectSectionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "section",
		Short: "Секции проекта",
		Long: `Управление секциями (task groups) проекта SingularityApp.

Подкоманды: list, get, create, update, delete.
Секция принадлежит одному проекту; list и create требуют PROJECT_ID.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return errors.New("укажите подкоманду: list, get, create, update или delete")
		},
	}
	cmd.AddCommand(newProjectSectionListCmd())
	cmd.AddCommand(newProjectSectionGetCmd())
	cmd.AddCommand(newProjectSectionCreateCmd())
	cmd.AddCommand(newProjectSectionUpdateCmd())
	cmd.AddCommand(newProjectSectionDeleteCmd())
	return cmd
}

func requireSectionID(id string) (string, error) {
	return requireProjectID(id)
}
