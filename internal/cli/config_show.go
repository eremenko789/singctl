package cli

import (
	"errors"
	"fmt"

	cfgpkg "github.com/eremenko789/singctl/internal/config"
	"github.com/spf13/cobra"
)

func newConfigShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Показать effective-конфигурацию",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			settings, err := cfgpkg.LoadEffectiveSettings(Opts.ConfigPath, Opts.Token)
			if err != nil {
				return err
			}
			if !settings.FromFile {
				return errors.New("конфиг не найден")
			}

			format := resolveShowOutputFormat(cmd)
			return renderConfigShow(cmd, settings.Config, format)
		},
	}
}

func resolveShowOutputFormat(cmd *cobra.Command) OutputFormat {
	flag := cmd.Flags().Lookup("output")
	if flag == nil || !flag.Changed {
		return OutputYAML
	}
	return Opts.Output
}

func renderConfigShow(cmd *cobra.Command, cfg cfgpkg.Document, format OutputFormat) error {
	switch format {
	case OutputYAML:
		return renderConfigYAML(cmd.OutOrStdout(), cfg)
	case OutputJSON:
		return renderConfigJSON(cmd.OutOrStdout(), cfg)
	case OutputCSV:
		return renderConfigCSV(cmd.OutOrStdout(), cfg)
	case OutputTable:
		return renderConfigTable(cmd.OutOrStdout(), cfg)
	default:
		return fmt.Errorf("неподдерживаемый формат вывода %q", format)
	}
}
