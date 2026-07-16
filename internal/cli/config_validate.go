package cli

import (
	"context"
	"errors"
	"fmt"

	"github.com/eremenko789/singctl/internal/api"
	cfgpkg "github.com/eremenko789/singctl/internal/config"
	"github.com/spf13/cobra"
)

func newConfigValidateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "validate",
		Short: "Проверить конфигурацию и доступность API",
		Long:  "Локально проверяет наличие токена и выполняет удалённую проверку доступности API (GET /v2/project).",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			settings, err := cfgpkg.LoadEffectiveSettings(Opts.ConfigPath, Opts.Token)
			if err != nil {
				return err
			}
			if settings.Config.API.Token == "" {
				return api.Classify(errors.New("токен не задан; используйте 'singctl config set-token'"))
			}

			session, err := api.NewFromSettings(settings)
			if err != nil {
				return api.Classify(err)
			}
			if err := session.ValidateConnectivity(context.Background()); err != nil {
				// ValidateConnectivity already returns ClassifiedError; keep As-friendly.
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Удалённая проверка API успешно пройдена.")
			return nil
		},
	}
}
