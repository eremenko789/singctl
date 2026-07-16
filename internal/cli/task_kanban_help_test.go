package cli

import (
	"strings"
	"testing"
)

func TestTaskKanbanHelpDiscoverability(t *testing.T) {
	stdout, stderr, err := executeForTest([]string{"task", "kanban", "--help"})
	if err != nil {
		t.Fatalf("err=%v stderr=%q", err, stderr)
	}
	lower := strings.ToLower(stdout)
	for _, name := range []string{"list", "get", "create", "update", "delete"} {
		if !hasCommandLine(lower, name) {
			t.Errorf("kanban --help missing %q:\n%s", name, stdout)
		}
	}

	cases := map[string][]string{
		"list":   {"--task", "--status"},
		"get":    {"LINK_ID", "ID"},
		"create": {"--task", "--column", "--order"},
		"update": {"--task", "--column", "--order"},
		"delete": {"LINK_ID", "ID"},
	}
	for sub, needles := range cases {
		out, _, err := executeForTest([]string{"task", "kanban", sub, "--help"})
		if err != nil {
			t.Fatalf("%s --help: %v", sub, err)
		}
		l := strings.ToLower(out)
		found := false
		for _, n := range needles {
			if strings.Contains(l, strings.ToLower(n)) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("%s --help missing one of %v:\n%s", sub, needles, out)
		}
		for _, bad := range []string{"--limit", "--offset", "--removed", "tui", "project column"} {
			if strings.Contains(l, bad) {
				t.Errorf("%s --help must not claim %q:\n%s", sub, bad, out)
			}
		}
	}

	moveHelp, _, err := executeForTest([]string{"task", "move", "--help"})
	if err != nil {
		t.Fatalf("move --help: %v", err)
	}
	ml := strings.ToLower(moveHelp)
	if !strings.Contains(ml, "--column") {
		t.Fatalf("move --help missing --column:\n%s", moveHelp)
	}
	if strings.Contains(ml, "--order") {
		// Allow mention only outside Flags; Flags must not define --order.
		flagIdx := strings.Index(ml, "flags:")
		if flagIdx >= 0 && strings.Contains(ml[flagIdx:], "--order") {
			t.Fatalf("move --help must not document --order flag:\n%s", moveHelp)
		}
	}

	stdout, _, err = executeForTest([]string{"task", "kanban", "nosuch"})
	if err == nil || ExitCode(err) != 1 {
		t.Fatalf("unknown subcommand: %v", err)
	}
	if strings.TrimSpace(stdout) != "" {
		t.Fatalf("stdout=%q", stdout)
	}
}
