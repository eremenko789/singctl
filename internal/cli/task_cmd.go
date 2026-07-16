package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/eremenko789/singctl/internal/api"
	cfgpkg "github.com/eremenko789/singctl/internal/config"
	"github.com/eremenko789/singctl/internal/output"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
)

func newTaskCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "task",
		Short: "Управление задачами",
		Long: `Команды для списка, просмотра и изменения задач SingularityApp.

Подкоманды: list, get, create, update, delete, archive, trash, checklist, kanban, move.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return errors.New("укажите подкоманду: list, get, create, update, delete, archive, trash, checklist, kanban или move")
		},
	}
	cmd.AddCommand(newTaskListCmd())
	cmd.AddCommand(newTaskGetCmd())
	cmd.AddCommand(newTaskCreateCmd())
	cmd.AddCommand(newTaskUpdateCmd())
	cmd.AddCommand(newTaskDeleteCmd())
	cmd.AddCommand(newTaskArchiveCmd())
	cmd.AddCommand(newTaskTrashCmd())
	cmd.AddCommand(newTaskChecklistCmd())
	cmd.AddCommand(newTaskKanbanCmd())
	cmd.AddCommand(newTaskMoveCmd())
	return cmd
}

func openAPISession() (*api.Session, cfgpkg.EffectiveSettings, error) {
	settings, err := cfgpkg.LoadEffectiveSettings(Opts.ConfigPath, Opts.Token)
	if err != nil {
		return nil, settings, api.Classify(err)
	}
	if settings.Config.API.Token == "" {
		return nil, settings, api.Classify(errors.New("токен не задан; используйте 'singctl config set-token'"))
	}
	session, err := api.NewFromSettings(settings)
	if err != nil {
		return nil, settings, api.Classify(err)
	}
	return session, settings, nil
}

func requireTaskID(id string) (string, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return "", fmt.Errorf("id задачи не задан")
	}
	return id, nil
}

func writerIsTTY(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	fd := f.Fd()
	return isatty.IsTerminal(fd) || isatty.IsCygwinTerminal(fd)
}

func buildTaskRenderOptions(cmd *cobra.Command, settings cfgpkg.EffectiveSettings, singleObject bool) output.RenderOptions {
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

func renderTaskRecordSet(cmd *cobra.Command, settings cfgpkg.EffectiveSettings, set output.RecordSet, singleObject bool) error {
	return output.Render(cmd.OutOrStdout(), set, buildTaskRenderOptions(cmd, settings, singleObject))
}
