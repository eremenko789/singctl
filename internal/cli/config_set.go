package cli

import (
	"fmt"

	cfgpkg "github.com/eremenko789/singctl/internal/config"
	"github.com/spf13/cobra"
)

func newConfigSetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Изменить значение в локальной конфигурации",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
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

			if err := cfgpkg.SetConfigValue(&cfg, args[0], args[1]); err != nil {
				return err
			}
			if err := cfgpkg.SaveConfig(resolved.Path, cfg); err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Параметр %s сохранён.\n", args[0])
			return nil
		},
	}
}
