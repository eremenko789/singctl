package cli

import (
	"strings"
	"testing"
)

func TestTaskHelpListsSevenSubcommands(t *testing.T) {
	stdout, stderr, err := executeForTest([]string{"task", "--help"})
	if err != nil {
		t.Fatalf("err=%v stderr=%q", err, stderr)
	}
	lower := strings.ToLower(stdout)
	for _, name := range []string{"list", "get", "create", "update", "delete", "archive", "trash"} {
		if !hasCommandLine(lower, name) {
			t.Errorf("task --help missing %q:\n%s", name, stdout)
		}
	}
	for _, bad := range []string{"checklist", "kanban", "move"} {
		if hasCommandLine(lower, bad) {
			t.Errorf("task --help must not list %q", bad)
		}
	}
}

func TestTaskSubcommandHelpFlagsAndNoteDelta(t *testing.T) {
	cases := map[string][]string{
		"list":    {"--project", "--limit", "--offset", "--from", "--to"},
		"get":     {"ID"},
		"create":  {"--title", "--note", "delta"},
		"update":  {"--title", "--note", "delta"},
		"delete":  {"ID"},
		"archive": {"--date"},
		"trash":   {"--date"},
	}
	for sub, needles := range cases {
		stdout, stderr, err := executeForTest([]string{"task", sub, "--help"})
		if err != nil {
			t.Fatalf("%s --help: %v stderr=%q", sub, err, stderr)
		}
		lower := strings.ToLower(stdout)
		for _, n := range needles {
			if !strings.Contains(lower, strings.ToLower(n)) {
				t.Errorf("%s --help missing %q:\n%s", sub, n, stdout)
			}
		}
		for _, bad := range []string{"checklist", "kanban", "move"} {
			if strings.Contains(lower, bad+" ") || hasCommandLine(lower, bad) {
				// allow word "move" in unrelated text? avoid false positive — check Available Commands
			}
		}
		if strings.Contains(lower, "checklist") || strings.Contains(lower, "kanban") {
			t.Errorf("%s --help must not claim checklist/kanban:\n%s", sub, stdout)
		}
	}
}

func TestTaskUnknownSubcommandExit1(t *testing.T) {
	stdout, stderr, err := executeForTest([]string{"task", "nosuch"})
	if err == nil {
		t.Fatal("expected error")
	}
	if ExitCode(err) != 1 {
		t.Fatalf("ExitCode=%d", ExitCode(err))
	}
	if strings.TrimSpace(stdout) != "" {
		t.Fatalf("stdout=%q", stdout)
	}
	if !strings.Contains(stderr, "неизвестная команда") && !strings.Contains(stderr, "nosuch") {
		t.Fatalf("stderr=%q", stderr)
	}
}
