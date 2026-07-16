package cli

import (
	"strings"
	"testing"
)

func TestTaskChecklistHelp(t *testing.T) {
	stdout, stderr, err := executeForTest([]string{"task", "--help"})
	if err != nil {
		t.Fatalf("err=%v stderr=%q", err, stderr)
	}
	if !hasCommandLine(strings.ToLower(stdout), "checklist") {
		t.Fatalf("task --help missing checklist:\n%s", stdout)
	}

	stdout, stderr, err = executeForTest([]string{"task", "checklist", "--help"})
	if err != nil {
		t.Fatalf("err=%v stderr=%q", err, stderr)
	}
	lower := strings.ToLower(stdout)
	for _, name := range []string{"list", "get", "add", "update", "delete"} {
		if !hasCommandLine(lower, name) {
			t.Errorf("checklist --help missing %q:\n%s", name, stdout)
		}
	}

	cases := map[string][]string{
		"list":   {"TASK_ID"},
		"get":    {"CHECKLIST_ITEM_ID", "ID"},
		"add":    {"--title", "--done"},
		"update": {"--title", "--done", "--undone"},
		"delete": {"CHECKLIST_ITEM_ID", "ID"},
	}
	for sub, needles := range cases {
		stdout, stderr, err = executeForTest([]string{"task", "checklist", sub, "--help"})
		if err != nil {
			t.Fatalf("%s --help: %v stderr=%q", sub, err, stderr)
		}
		lower = strings.ToLower(stdout)
		found := false
		for _, n := range needles {
			if strings.Contains(lower, strings.ToLower(n)) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("%s --help missing one of %v:\n%s", sub, needles, stdout)
		}
		for _, bad := range []string{"--order", "kanban", "tui"} {
			if strings.Contains(lower, bad) {
				t.Errorf("%s --help must not mention %q:\n%s", sub, bad, stdout)
			}
		}
		// Flags section must not advertise pagination/removed (Long may mention unsupported).
		if strings.Contains(lower, "flags:") {
			flagsPart := lower[strings.Index(lower, "flags:"):]
			if idx := strings.Index(flagsPart, "global flags:"); idx >= 0 {
				flagsPart = flagsPart[:idx]
			}
			for _, bad := range []string{"--limit", "--offset", "--removed", "--order"} {
				if strings.Contains(flagsPart, bad) {
					t.Errorf("%s flags must not include %q:\n%s", sub, bad, stdout)
				}
			}
		}
	}
}

func TestTaskChecklistUnknownSubcommand(t *testing.T) {
	stdout, stderr, err := executeForTest([]string{"task", "checklist", "nosuch"})
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
