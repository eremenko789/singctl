package cli

import (
	"fmt"

	"github.com/eremenko789/singctl/internal/buildinfo"
	"github.com/spf13/cobra"
)

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Показать версию singctl",
		Long:  "Печатает имя CLI, версию и метаданные сборки (commit, date).",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			_, err := fmt.Fprint(cmd.OutOrStdout(), buildinfo.Format())
			return err
		},
	}
}
