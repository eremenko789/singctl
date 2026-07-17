package cli

import (
	"context"

	"github.com/spf13/cobra"
)

func newProjectSectionGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <SECTION_ID>",
		Short: "Показать секцию по ID",
		Long:  "Загрузить одну секцию и вывести её в выбранном формате (-o).",
		Example: `  singctl project section get Q-uuid
  singctl project section get Q-uuid -o json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := requireSectionID(args[0])
			if err != nil {
				return err
			}
			session, settings, err := openAPISession()
			if err != nil {
				return err
			}
			section, err := session.GetSection(context.Background(), id)
			if err != nil {
				return err
			}
			return renderSectionRecordSet(cmd, settings, sectionToRecordSet(section), true)
		},
	}
}
