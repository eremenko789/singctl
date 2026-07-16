package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/eremenko789/singctl/internal/api"
	"github.com/spf13/cobra"
)

type taskWriteFlags struct {
	title       string
	project     string
	parent      string
	start       string
	note        string
	priority    int
	isNote      bool
	archiveDate string
	deleteDate  string
}

func bindTaskWriteFlags(cmd *cobra.Command, f *taskWriteFlags) {
	flags := cmd.Flags()
	flags.StringVar(&f.title, "title", "", "заголовок задачи")
	flags.StringVar(&f.project, "project", "", "ID проекта")
	flags.StringVar(&f.parent, "parent", "", "ID родительской задачи")
	flags.StringVar(&f.start, "start", "", "дата старта (YYYY-MM-DD)")
	flags.StringVar(&f.note, "note", "", "заметка as-is (API может ожидать delta-формат)")
	flags.IntVar(&f.priority, "priority", 0, "приоритет: 0 (high), 1 (normal), 2 (low)")
	flags.BoolVar(&f.isNote, "is-note", false, "задача-заметка")
	flags.StringVar(&f.archiveDate, "archive-date", "", "дата архива journalDate (YYYY-MM-DD)")
	flags.StringVar(&f.deleteDate, "delete-date", "", "дата корзины deleteDate (YYYY-MM-DD)")
}

func (f *taskWriteFlags) toInput(cmd *cobra.Command, requireTitle bool) (api.TaskWriteInput, error) {
	var in api.TaskWriteInput
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

	if flags.Changed("project") {
		p := f.project
		in.ProjectID = &p
	}
	if flags.Changed("parent") {
		p := f.parent
		in.Parent = &p
	}
	if flags.Changed("start") {
		if _, err := api.ParseDate(f.start); err != nil {
			return in, err
		}
		s := f.start
		in.Start = &s
	}
	if flags.Changed("note") {
		n := f.note
		in.Note = &n
	}
	if flags.Changed("priority") {
		if f.priority < 0 || f.priority > 2 {
			return in, fmt.Errorf("--priority должен быть 0, 1 или 2")
		}
		p := f.priority
		in.Priority = &p
	}
	if flags.Changed("is-note") {
		b := f.isNote
		in.IsNote = &b
	}
	if flags.Changed("archive-date") {
		if _, err := api.ParseDate(f.archiveDate); err != nil {
			return in, err
		}
		d := f.archiveDate
		in.JournalDate = &d
	}
	if flags.Changed("delete-date") {
		if _, err := api.ParseDate(f.deleteDate); err != nil {
			return in, err
		}
		d := f.deleteDate
		in.DeleteDate = &d
	}
	return in, nil
}

func (f *taskWriteFlags) anyWriteFlagSet(cmd *cobra.Command) bool {
	for _, name := range []string{
		"title", "project", "parent", "start", "note",
		"priority", "is-note", "archive-date", "delete-date",
	} {
		if cmd.Flags().Changed(name) {
			return true
		}
	}
	return false
}

func newTaskCreateCmd() *cobra.Command {
	var f taskWriteFlags
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Создать задачу",
		Long: `Создать задачу. Обязателен --title.

--note передаётся as-is; API может ожидать delta-формат заметки.
--delete-date на create выполняется как create, затем update (границы OpenAPI).`,
		Example: `  singctl task create --title "Купить молоко"
  singctl task create --title "Note" --is-note --project P-uuid -o json`,
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
			task, err := session.CreateTask(context.Background(), in)
			if err != nil {
				return err
			}
			return renderTaskRecordSet(cmd, settings, taskToRecordSet(task), true)
		},
	}
	bindTaskWriteFlags(cmd, &f)
	_ = cmd.MarkFlagRequired("title")
	return cmd
}
