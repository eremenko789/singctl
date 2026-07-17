package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/eremenko789/singctl/internal/api"
	"github.com/spf13/cobra"
)

type projectWriteFlags struct {
	title    string
	note     string
	notebook bool
	emoji    string
	color    string
	parent   string
}

func bindProjectWriteFlags(cmd *cobra.Command, f *projectWriteFlags) {
	flags := cmd.Flags()
	flags.StringVar(&f.title, "title", "", "заголовок проекта")
	flags.StringVar(&f.note, "note", "", "заметка as-is (API может ожидать delta-формат)")
	flags.BoolVar(&f.notebook, "notebook", false, "проект-блокнот (isNotebook)")
	flags.StringVar(&f.emoji, "emoji", "", "emoji: unicode (например 💞) или hex (1f49e)")
	flags.StringVar(&f.color, "color", "", "цвет проекта (as-is)")
	flags.StringVar(&f.parent, "parent", "", "ID родительского проекта")
}

func (f *projectWriteFlags) toInput(cmd *cobra.Command, requireTitle bool) (api.ProjectWriteInput, error) {
	var in api.ProjectWriteInput
	flags := cmd.Flags()

	if requireTitle {
		title := strings.TrimSpace(f.title)
		if title == "" {
			return in, fmt.Errorf("--title обязателен")
		}
		in.Title = &title
	} else if flags.Changed("title") {
		title := strings.TrimSpace(f.title)
		if title == "" {
			return in, fmt.Errorf("--title не может быть пустым")
		}
		in.Title = &title
	}

	if flags.Changed("note") {
		n := f.note
		in.Note = &n
	}
	if flags.Changed("notebook") {
		b := f.notebook
		in.IsNotebook = &b
	}
	if flags.Changed("emoji") {
		hex, err := api.NormalizeProjectEmoji(f.emoji)
		if err != nil {
			return in, err
		}
		in.Emoji = &hex
	}
	if flags.Changed("color") {
		c := f.color
		in.Color = &c
	}
	if flags.Changed("parent") {
		p := f.parent
		in.Parent = &p
	}
	return in, nil
}

func (f *projectWriteFlags) anyWriteFlagSet(cmd *cobra.Command) bool {
	for _, name := range []string{"title", "note", "notebook", "emoji", "color", "parent"} {
		if cmd.Flags().Changed(name) {
			return true
		}
	}
	return false
}

func newProjectCreateCmd() *cobra.Command {
	var f projectWriteFlags
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Создать проект",
		Long: `Создать проект. Обязателен --title.

--note передаётся as-is; API может ожидать delta-формат заметки.
--emoji: unicode (💞) или hex (1f49e); иначе ошибка до сети.`,
		Example: `  singctl project create --title "Inbox"
  singctl project create --title "Notes" --notebook --emoji 💞 -o json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			in, err := f.toInput(cmd, true)
			if err != nil {
				return err
			}
			session, settings, err := openAPISession()
			if err != nil {
				return err
			}
			project, err := session.CreateProject(context.Background(), in)
			if err != nil {
				return err
			}
			return renderProjectRecordSet(cmd, settings, projectToRecordSet(project), true)
		},
	}
	bindProjectWriteFlags(cmd, &f)
	_ = cmd.MarkFlagRequired("title")
	return cmd
}
