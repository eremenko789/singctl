package cli

import (
	"errors"
	"fmt"
	"os"
	"strings"

	cfgpkg "github.com/eremenko789/singctl/internal/config"
	"github.com/eremenko789/singctl/internal/output"
	"github.com/spf13/cobra"
)

func newProjectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "project",
		Short: "Управление проектами",
		Long: `Команды для списка, просмотра и изменения проектов SingularityApp.

Подкоманды: list, get, create, update, delete, archive, trash.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return errors.New("укажите подкоманду: list, get, create, update, delete, archive или trash")
		},
	}
	cmd.AddCommand(newProjectListCmd())
	cmd.AddCommand(newProjectGetCmd())
	cmd.AddCommand(newProjectCreateCmd())
	cmd.AddCommand(newProjectUpdateCmd())
	cmd.AddCommand(newProjectDeleteCmd())
	cmd.AddCommand(newProjectArchiveCmd())
	cmd.AddCommand(newProjectTrashCmd())
	return cmd
}

func requireProjectID(id string) (string, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return "", fmt.Errorf("id проекта не задан")
	}
	return id, nil
}

func buildProjectRenderOptions(cmd *cobra.Command, settings cfgpkg.EffectiveSettings, singleObject bool) output.RenderOptions {
	flag := cmd.Root().PersistentFlags().Lookup("output")
	flagSet := flag != nil && flag.Changed
	format := output.ResolveFormat(flagSet, Opts.Output.String(), settings.Config.Output.Format)
	color := output.ColorEnabled(
		writerIsTTY(cmd.OutOrStdout()),
		Opts.NoColor,
		os.Getenv("NO_COLOR"),
		settings.Config.Output.Color,
	)
	return output.RenderOptions{
		Format:       format,
		Color:        color,
		DateLayout:   settings.Config.Output.DateFormat,
		SingleObject: singleObject,
	}
}

func renderProjectRecordSet(cmd *cobra.Command, settings cfgpkg.EffectiveSettings, set output.RecordSet, singleObject bool) error {
	return output.Render(cmd.OutOrStdout(), set, buildProjectRenderOptions(cmd, settings, singleObject))
}
