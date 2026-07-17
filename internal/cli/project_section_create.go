package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/eremenko789/singctl/internal/api"
	"github.com/spf13/cobra"
)

type sectionWriteFlags struct {
	title  string
	parent string
}

func (f *sectionWriteFlags) toInput(cmd *cobra.Command, requireTitle bool, parentFromArg string) (api.SectionWriteInput, error) {
	var in api.SectionWriteInput
	flags := cmd.Flags()

	if requireTitle {
		title := strings.TrimSpace(f.title)
		if title == "" {
			return in, fmt.Errorf("--title обязателен")
		}
		in.Title = &title
		pid, err := requireProjectID(parentFromArg)
		if err != nil {
			return in, err
		}
		in.Parent = &pid
	} else {
		if flags.Changed("title") {
			title := strings.TrimSpace(f.title)
			if title == "" {
				return in, fmt.Errorf("--title не может быть пустым")
			}
			in.Title = &title
		}
		if flags.Changed("parent") {
			parent := strings.TrimSpace(f.parent)
			if parent == "" {
				return in, fmt.Errorf("--parent не может быть пустым")
			}
			in.Parent = &parent
		}
	}
	return in, nil
}

func (f *sectionWriteFlags) anyWriteFlagSet(cmd *cobra.Command) bool {
	for _, name := range []string{"title", "parent"} {
		if cmd.Flags().Changed(name) {
			return true
		}
	}
	return false
}

func newProjectSectionCreateCmd() *cobra.Command {
	var f sectionWriteFlags
	cmd := &cobra.Command{
		Use:   "create <PROJECT_ID>",
		Short: "Создать секцию в проекте",
		Long: `Создать секцию в указанном проекте. Обязательны PROJECT_ID и --title.

Родительский проект задаётся только позиционным аргументом, без флага --parent.`,
		Example: `  singctl project section create P-uuid --title "Inbox"
  singctl project section create P-uuid --title "Done" -o json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			in, err := f.toInput(cmd, true, args[0])
			if err != nil {
				return err
			}
			session, settings, err := openAPISession()
			if err != nil {
				return err
			}
			section, err := session.CreateSection(context.Background(), in)
			if err != nil {
				return err
			}
			return renderSectionRecordSet(cmd, settings, sectionToRecordSet(section), true)
		},
	}
	flags := cmd.Flags()
	flags.StringVar(&f.title, "title", "", "заголовок секции")
	_ = cmd.MarkFlagRequired("title")
	return cmd
}

func newProjectSectionUpdateCmd() *cobra.Command {
	var f sectionWriteFlags
	cmd := &cobra.Command{
		Use:   "update <SECTION_ID>",
		Short: "Обновить секцию",
		Long: `Частично обновить секцию. Нужен хотя бы один write-флаг.

--parent переносит секцию в другой проект.`,
		Example: `  singctl project section update Q-uuid --title "Новое имя"
  singctl project section update Q-uuid --parent P-other -o json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := requireSectionID(args[0])
			if err != nil {
				return err
			}
			if !f.anyWriteFlagSet(cmd) {
				return fmt.Errorf("укажите хотя бы один флаг для обновления")
			}
			in, err := f.toInput(cmd, false, "")
			if err != nil {
				return err
			}
			session, settings, err := openAPISession()
			if err != nil {
				return err
			}
			section, err := session.UpdateSection(context.Background(), id, in)
			if err != nil {
				return err
			}
			return renderSectionRecordSet(cmd, settings, sectionToRecordSet(section), true)
		},
	}
	flags := cmd.Flags()
	flags.StringVar(&f.title, "title", "", "новый заголовок секции")
	flags.StringVar(&f.parent, "parent", "", "перенести секцию в другой проект (PROJECT_ID)")
	return cmd
}
