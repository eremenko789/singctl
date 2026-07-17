package cli

import (
	"strings"
	"testing"
)

func TestProjectHelpListsSubcommands(t *testing.T) {
	stdout, stderr, err := executeForTest([]string{"project", "--help"})
	if err != nil {
		t.Fatalf("err=%v stderr=%q", err, stderr)
	}
	lower := strings.ToLower(stdout)
	for _, name := range []string{"list", "get", "create", "update", "delete", "archive", "trash", "section"} {
		if !hasCommandLine(lower, name) {
			t.Errorf("project --help missing %q:\n%s", name, stdout)
		}
	}
	for _, forbidden := range []string{"column"} {
		if hasCommandLine(lower, forbidden) {
			t.Errorf("project --help must not list %q as command:\n%s", forbidden, stdout)
		}
	}
	if strings.Contains(lower, "shared") && strings.Contains(lower, "collaborative") {
		// soft: must not promise shared projects as available results
		t.Log("help mentions shared; ensure it does not promise availability")
	}
}

func TestProjectSubcommandHelpFlags(t *testing.T) {
	cases := map[string][]string{
		"list":    {"--archived", "--removed", "--limit", "--offset"},
		"get":     {"ID"},
		"create":  {"--title", "--note", "delta", "--emoji", "💞", "1f49e"},
		"update":  {"--title", "--note", "delta", "--emoji"},
		"delete":  {"ID"},
		"archive": {"--date"},
		"trash":   {"--date"},
	}
	for sub, needles := range cases {
		stdout, stderr, err := executeForTest([]string{"project", sub, "--help"})
		if err != nil {
			t.Fatalf("%s --help: %v stderr=%q", sub, err, stderr)
		}
		lower := strings.ToLower(stdout)
		for _, n := range needles {
			if !strings.Contains(lower, strings.ToLower(n)) {
				t.Errorf("%s --help missing %q:\n%s", sub, n, stdout)
			}
		}
		if strings.Contains(lower, "section") && strings.Contains(lower, "available") {
			t.Errorf("%s --help must not advertise section as available", sub)
		}
	}
}

func TestProjectUnknownSubcommandExit1(t *testing.T) {
	stdout, stderr, err := executeForTest([]string{"project", "nosuch"})
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
