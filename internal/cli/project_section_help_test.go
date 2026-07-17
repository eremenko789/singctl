package cli

import (
	"strings"
	"testing"
)

func TestProjectSectionHelp(t *testing.T) {
	stdout, stderr, err := executeForTest([]string{"project", "--help"})
	if err != nil {
		t.Fatalf("err=%v stderr=%q", err, stderr)
	}
	if !hasCommandLine(strings.ToLower(stdout), "section") {
		t.Fatalf("project --help missing section:\n%s", stdout)
	}

	stdout, stderr, err = executeForTest([]string{"project", "section", "--help"})
	if err != nil {
		t.Fatalf("err=%v stderr=%q", err, stderr)
	}
	lower := strings.ToLower(stdout)
	for _, name := range []string{"list", "get", "create", "update", "delete"} {
		if !hasCommandLine(lower, name) {
			t.Errorf("section --help missing %q:\n%s", name, stdout)
		}
	}
	if hasCommandLine(lower, "column") {
		t.Errorf("section --help must not list column:\n%s", stdout)
	}

	cases := map[string][]string{
		"list":   {"PROJECT_ID", "--removed", "--limit", "--offset"},
		"get":    {"SECTION_ID"},
		"create": {"PROJECT_ID", "--title"},
		"update": {"--title", "--parent", "перенос"},
		"delete": {"SECTION_ID"},
	}
	for sub, needles := range cases {
		stdout, stderr, err = executeForTest([]string{"project", "section", sub, "--help"})
		if err != nil {
			t.Fatalf("%s --help: %v stderr=%q", sub, err, stderr)
		}
		lower = strings.ToLower(stdout)
		for _, n := range needles {
			if !strings.Contains(lower, strings.ToLower(n)) {
				t.Errorf("%s --help missing %q:\n%s", sub, n, stdout)
			}
		}
		if strings.Contains(lower, "flags:") {
			flagsPart := lower[strings.Index(lower, "flags:"):]
			if idx := strings.Index(flagsPart, "global flags:"); idx >= 0 {
				flagsPart = flagsPart[:idx]
			}
			if strings.Contains(flagsPart, "column") {
				t.Errorf("%s --help must not advertise column in flags:\n%s", sub, stdout)
			}
		}
	}
}

func TestProjectSectionUnknownSubcommand(t *testing.T) {
	stdout, stderr, err := executeForTest([]string{"project", "section", "nosuch"})
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
