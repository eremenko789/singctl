package cli

import "github.com/spf13/cobra"

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Управление локальной конфигурацией и токеном",
		Long:  "Команды для просмотра, изменения и локальной проверки конфигурации singctl.",
	}

	cmd.AddCommand(newConfigSetTokenCmd())
	cmd.AddCommand(newConfigShowCmd())
	cmd.AddCommand(newConfigSetCmd())
	cmd.AddCommand(newConfigValidateCmd())

	return cmd
}
