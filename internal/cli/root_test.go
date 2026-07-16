package cli

import (
	"strings"
	"testing"

	"github.com/eremenko789/singctl/internal/buildinfo"
)

func TestHelpRussianNoEntityCommands(t *testing.T) {
	stdout, _, err := executeForTest([]string{"--help"})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	out := stdout
	if !strings.Contains(out, "singctl") {
		t.Errorf("help missing product name: %q", out)
	}
	for _, needle := range []string{
		"путь к файлу конфигурации",
		"API-токен",
		"формат вывода",
		"отключить цветной вывод",
		"отладочный режим",
		"--config",
		"--token",
		"--output",
		"--no-color",
		"--debug",
	} {
		if !strings.Contains(out, needle) {
			t.Errorf("help missing %q:\n%s", needle, out)
		}
	}

	// F07: exit codes briefly documented in root help
	for _, code := range []string{"0", "1", "2", "3"} {
		if !strings.Contains(out, code) {
			t.Errorf("help must mention exit code %s:\n%s", code, out)
		}
	}
	if !strings.Contains(out, "scriptability") && !strings.Contains(out, "Коды выхода") {
		t.Errorf("help must mention exit codes / scriptability docs:\n%s", out)
	}

	forbidden := []string{"project", "habit", "tag", "time", "tui"}
	lower := strings.ToLower(out)
	for _, name := range forbidden {
		if hasCommandLine(lower, name) {
			t.Errorf("help must not list entity/TUI command %q", name)
		}
	}
	if !hasCommandLine(lower, "task") {
		t.Errorf("help must list task command:\n%s", out)
	}
}

func hasCommandLine(help, name string) bool {
	for _, line := range strings.Split(help, "\n") {
		fields := strings.Fields(line)
		if len(fields) > 0 && fields[0] == name {
			return true
		}
	}
	return false
}

func containsFold(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

func TestBareInvokeNonzero(t *testing.T) {
	stdout, stderr, err := executeForTest(nil)
	if err == nil {
		t.Fatal("expected error on bare invoke")
	}
	if strings.TrimSpace(stdout) != "" {
		t.Errorf("stdout should be empty, got %q", stdout)
	}
	if !strings.Contains(stderr, "команда не указана") && !strings.Contains(stderr, "TUI") {
		t.Errorf("stderr missing RU bare-invoke message: %q", stderr)
	}
}

func TestUnknownCommand(t *testing.T) {
	stdout, stderr, err := executeForTest([]string{"nosuch"})
	if err == nil {
		t.Fatal("expected error for unknown command")
	}
	if ExitCode(err) != 1 {
		t.Fatalf("ExitCode = %d, want 1", ExitCode(err))
	}
	if strings.TrimSpace(stdout) != "" {
		t.Errorf("stdout should be empty, got %q", stdout)
	}
	if !strings.Contains(stderr, "неизвестная команда") {
		t.Errorf("stderr missing RU unknown-command message: %q", stderr)
	}
	if !strings.Contains(stderr, "--help") {
		t.Errorf("stderr should hint --help: %q", stderr)
	}
}

func TestVersionParity(t *testing.T) {
	outSub, errSub, err := executeForTest([]string{"version"})
	if err != nil {
		t.Fatalf("version subcommand: %v stderr=%q", err, errSub)
	}
	outFlag, errFlag, err := executeForTest([]string{"--version"})
	if err != nil {
		t.Fatalf("--version: %v stderr=%q", err, errFlag)
	}
	if outSub != outFlag {
		t.Errorf("version parity failed:\nsubcommand:\n%q\n--version:\n%q", outSub, outFlag)
	}
	expected := buildinfo.Format()
	if outSub != expected {
		t.Errorf("stdout != buildinfo.Format():\ngot:\n%q\nwant:\n%q", outSub, expected)
	}
	for _, field := range []string{"singctl", buildinfo.Version, "commit", "date"} {
		if !containsFold(outSub, field) {
			t.Errorf("version output missing %q: %q", field, outSub)
		}
	}
}

func TestValidGlobalFlagsAccepted(t *testing.T) {
	cases := [][]string{
		{"--config", "/tmp/example.yaml", "--token", "TOKEN", "--output", "json", "--no-color", "--debug", "version"},
		{"-o", "yaml", "--help"},
		{"--output", "table", "version"},
		{"--output", "csv", "--version"},
	}
	for _, args := range cases {
		stdout, stderr, err := executeForTest(args)
		if err != nil {
			t.Errorf("args %v: unexpected error %v; stderr=%q stdout=%q", args, err, stderr, stdout)
		}
	}
}

func TestInvalidOutputBlocksHelpAndVersion(t *testing.T) {
	cases := [][]string{
		{"--output", "xml", "--help"},
		{"-o", "xml", "version"},
		{"--output", "xml", "--version"},
	}
	for _, args := range cases {
		stdout, stderr, err := executeForTest(args)
		if err == nil {
			t.Errorf("args %v: expected validation error", args)
			continue
		}
		if ExitCode(err) != 1 {
			t.Errorf("args %v: ExitCode = %d, want 1", args, ExitCode(err))
		}
		if strings.TrimSpace(stdout) != "" {
			t.Errorf("args %v: stdout should be empty, got %q", args, stdout)
		}
		if !strings.Contains(stderr, "недопустимый формат") && !strings.Contains(stderr, "формат вывода") {
			t.Errorf("args %v: missing RU validation message in stderr=%q", args, stderr)
		}
		// must not print full help or version body
		if strings.Contains(stdout, "Available Commands") || strings.Contains(stdout, "Доступн") {
			t.Errorf("args %v: help body leaked to stdout=%q", args, stdout)
		}
		if strings.Contains(stdout, "commit:") || strings.Contains(stdout, "version "+buildinfo.Version) {
			t.Errorf("args %v: version body leaked to stdout=%q", args, stdout)
		}
	}
}

func TestUnknownFlag(t *testing.T) {
	stdout, stderr, err := executeForTest([]string{"--unknown-flag"})
	if err == nil {
		t.Fatal("expected error for unknown flag")
	}
	if ExitCode(err) != 1 {
		t.Fatalf("ExitCode = %d, want 1", ExitCode(err))
	}
	if strings.TrimSpace(stdout) != "" {
		t.Errorf("stdout should be empty, got %q", stdout)
	}
	if strings.TrimSpace(stderr) == "" {
		t.Error("expected stderr message for unknown flag")
	}
}
