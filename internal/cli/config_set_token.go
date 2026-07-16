package cli

import (
	"fmt"

	cfgpkg "github.com/eremenko789/singctl/internal/config"
	"github.com/spf13/cobra"
)

func newConfigSetTokenCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set-token <TOKEN>",
		Short: "Сохранить API-токен в локальный конфиг",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			token, err := cfgpkg.NormalizeStoredToken(args[0])
			if err != nil {
				return err
			}

			opts, err := cfgpkg.ResolveOptionsFromEnv(Opts.ConfigPath)
			if err != nil {
				return fmt.Errorf("определить путь конфигурации: %w", err)
			}
			resolved, err := cfgpkg.ResolveWritePath(opts)
			if err != nil {
				return fmt.Errorf("определить путь конфигурации: %w", err)
			}

			cfg := cfgpkg.DefaultConfig()
			if resolved.Exists {
				cfg, err = cfgpkg.LoadConfig(resolved.Path)
				if err != nil {
					return err
				}
			}
			cfg.API.Token = token

			if err := cfgpkg.SaveConfig(resolved.Path, cfg); err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Токен сохранён.")
			return nil
		},
	}
}
