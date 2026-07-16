package cli

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/eremenko789/singctl/internal/buildinfo"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Opts holds parsed global flags for the current process.
var Opts = GlobalOptions{
	Output: OutputTable,
}

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "singctl",
		Short:         "CLI для SingularityApp",
		Long:          "singctl — инструмент командной строки для работы с SingularityApp.\nВ этой версии доступны справка, версия и глобальные флаги; сущности и TUI появятся позже.",
		SilenceErrors: true,
		SilenceUsage:  true,
		Version:       buildinfo.Version,
		RunE: func(cmd *cobra.Command, args []string) error {
			return errors.New("команда не указана: TUI ещё не реализован. Укажите команду или используйте --help")
		},
	}

	cmd.SetVersionTemplate(buildinfo.Format())
	cmd.SetFlagErrorFunc(func(c *cobra.Command, err error) error {
		return fmt.Errorf("%s\nСм. '%s --help'", localizeFlagError(err), c.CommandPath())
	})

	opts := &Opts
	*opts = GlobalOptions{
		Output: OutputTable,
	}

	pf := cmd.PersistentFlags()
	pf.StringVar(&opts.ConfigPath, "config", "", "путь к файлу конфигурации")
	pf.StringVar(&opts.Token, "token", "", "API-токен (не сохраняется в F01)")
	pf.VarP(&opts.Output, "output", "o", "формат вывода: table, json, yaml, csv")
	pf.BoolVar(&opts.NoColor, "no-color", false, "отключить цветной вывод")
	pf.BoolVar(&opts.Debug, "debug", false, "включить отладочный режим")

	_ = viper.BindPFlag("config", pf.Lookup("config"))
	_ = viper.BindPFlag("token", pf.Lookup("token"))
	_ = viper.BindPFlag("output", pf.Lookup("output"))
	_ = viper.BindPFlag("no-color", pf.Lookup("no-color"))
	_ = viper.BindPFlag("debug", pf.Lookup("debug"))

	cmd.AddCommand(newVersionCmd())
	cmd.AddCommand(newConfigCmd())

	return cmd
}

// Execute runs the root command with process args and writes errors to stderr.
func Execute() error {
	cmd := newRootCmd()
	err := cmd.Execute()
	if err != nil {
		msg := localizeCobraError(err)
		fmt.Fprintln(os.Stderr, "Ошибка:", msg)
		return err
	}
	return nil
}

// executeForTest runs with explicit args and captured streams (for unit tests).
func executeForTest(args []string) (stdout, stderr string, err error) {
	var outBuf, errBuf strings.Builder
	cmd := newRootCmd()
	cmd.SetOut(&outBuf)
	cmd.SetErr(&errBuf)
	cmd.SetArgs(args)

	err = cmd.Execute()
	if err != nil {
		msg := localizeCobraError(err)
		fmt.Fprintln(&errBuf, "Ошибка:", msg)
	}
	return outBuf.String(), errBuf.String(), err
}

func localizeCobraError(err error) string {
	if err == nil {
		return ""
	}
	s := err.Error()
	lower := strings.ToLower(s)

	if strings.Contains(lower, "unknown command") {
		// cobra: `unknown command "foo" for "singctl"`
		parts := strings.SplitN(s, "\"", 3)
		name := ""
		if len(parts) >= 2 {
			name = parts[1]
		}
		if name != "" {
			return fmt.Sprintf("неизвестная команда %q. См. 'singctl --help'", name)
		}
		return "неизвестная команда. См. 'singctl --help'"
	}

	return s
}

func localizeFlagError(err error) string {
	if err == nil {
		return ""
	}
	s := err.Error()
	// Prefer our RU validation text when pflag wraps it.
	if idx := strings.Index(s, "недопустимый формат"); idx >= 0 {
		return s[idx:]
	}
	if strings.Contains(strings.ToLower(s), "unknown flag") {
		if flag := extractQuoted(s); flag != "" {
			return fmt.Sprintf("неизвестный флаг %s", flag)
		}
		// cobra often formats: `unknown flag: --foo`
		if i := strings.Index(strings.ToLower(s), "unknown flag:"); i >= 0 {
			rest := strings.TrimSpace(s[i+len("unknown flag:"):])
			if rest != "" {
				return fmt.Sprintf("неизвестный флаг %s", strings.Fields(rest)[0])
			}
		}
		return "неизвестный флаг"
	}
	return s
}

func extractQuoted(s string) string {
	i := strings.Index(s, "\"")
	if i < 0 {
		return ""
	}
	j := strings.Index(s[i+1:], "\"")
	if j < 0 {
		return ""
	}
	return s[i : i+1+j+1]
}
