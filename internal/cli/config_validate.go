package cli

import (
	"errors"
	"fmt"

	cfgpkg "github.com/eremenko789/singctl/internal/config"
	"github.com/spf13/cobra"
)

func newConfigValidateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "validate",
		Short: "Локально проверить готовность конфигурации",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			settings, err := cfgpkg.LoadEffectiveSettings(Opts.ConfigPath, Opts.Token)
			if err != nil {
				return err
			}
			if settings.Config.API.Token == "" {
				return errors.New("токен не задан; используйте 'singctl config set-token'")
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Локальная проверка конфигурации пройдена. Удалённая проверка API пока работает как заглушка и не выполнялась.")
			return nil
		},
	}
}
