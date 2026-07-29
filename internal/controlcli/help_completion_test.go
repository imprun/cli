package controlcli

import (
	"bytes"
	"strings"
	"testing"
)

func TestImprunCommandHelpDoesNotRequireConfigurationOrAuthentication(t *testing.T) {
	t.Setenv("IMPRUN_CONFIG", t.TempDir())
	var stdout, stderr bytes.Buffer
	exit := RunImprun(
		[]string{"app", "publish", "--help"},
		strings.NewReader(""),
		&stdout,
		&stderr,
	)
	if exit != ExitOK {
		t.Fatalf("exit=%d stderr=%s", exit, stderr.String())
	}
	if !strings.Contains(stdout.String(), "imprun app publish [path]") ||
		!strings.Contains(stdout.String(), "--allow-dirty") ||
		!strings.Contains(stdout.String(), "exact-commit") {
		t.Fatalf("help = %s", stdout.String())
	}
}

func TestImprunAuthLogoutHelpExplainsRemoteRevocationBoundary(t *testing.T) {
	t.Setenv("IMPRUN_CONFIG", t.TempDir())
	var stdout, stderr bytes.Buffer
	exit := RunImprun(
		[]string{"auth", "logout", "--help"},
		strings.NewReader(""),
		&stdout,
		&stderr,
	)
	if exit != ExitOK ||
		!strings.Contains(stdout.String(), "--local-only") ||
		!strings.Contains(stdout.String(), "central browser session") {
		t.Fatalf("exit=%d stdout=%q stderr=%q", exit, stdout.String(), stderr.String())
	}
}

func TestImprunContextDeleteHelpExplainsCredentialAndCurrentContextSafety(t *testing.T) {
	t.Setenv("IMPRUN_CONFIG", t.TempDir())
	var stdout, stderr bytes.Buffer
	exit := RunImprun(
		[]string{"context", "delete", "--help"},
		strings.NewReader(""),
		&stdout,
		&stderr,
	)
	if exit != ExitOK ||
		!strings.Contains(stdout.String(), "imprun context delete <name> --yes") ||
		!strings.Contains(stdout.String(), "logged out first") ||
		!strings.Contains(stdout.String(), "Select another context") ||
		stderr.Len() != 0 {
		t.Fatalf("exit=%d stdout=%q stderr=%q", exit, stdout.String(), stderr.String())
	}
}

func TestImprunVersionDoesNotRequireConfigurationOrAuthentication(t *testing.T) {
	for _, args := range [][]string{{"version"}, {"--version"}} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			t.Setenv("IMPRUN_CONFIG", t.TempDir())
			var stdout, stderr bytes.Buffer
			exit := RunImprun(args, strings.NewReader(""), &stdout, &stderr)
			if exit != ExitOK || strings.TrimSpace(stdout.String()) != Version {
				t.Fatalf("exit=%d stdout=%q stderr=%q", exit, stdout.String(), stderr.String())
			}
		})
	}
}

func TestImprunCompletionDoesNotRequireConfigurationOrAuthentication(t *testing.T) {
	t.Setenv("IMPRUN_CONFIG", t.TempDir())
	for _, test := range []struct {
		shell string
		want  string
	}{
		{shell: "bash", want: "complete -F _imprun_complete imprun"},
		{shell: "zsh", want: "#compdef imprun"},
		{shell: "fish", want: "complete -c imprun"},
		{shell: "powershell", want: "Register-ArgumentCompleter"},
	} {
		t.Run(test.shell, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			exit := RunImprun(
				[]string{"completion", test.shell},
				strings.NewReader(""),
				&stdout,
				&stderr,
			)
			if exit != ExitOK || !strings.Contains(stdout.String(), test.want) ||
				(test.shell != "zsh" && !strings.Contains(stdout.String(), "delete")) {
				t.Fatalf("exit=%d stdout=%q stderr=%q", exit, stdout.String(), stderr.String())
			}
		})
	}
}

func TestImprunUnknownHelpTopicReturnsUsageFailureWithoutConfiguration(t *testing.T) {
	t.Setenv("IMPRUN_CONFIG", t.TempDir())
	var stdout, stderr bytes.Buffer
	exit := RunImprun(
		[]string{"help", "does-not-exist"},
		strings.NewReader(""),
		&stdout,
		&stderr,
	)
	if exit != ExitUsage || !strings.Contains(stderr.String(), "unknown help topic") {
		t.Fatalf("exit=%d stdout=%q stderr=%q", exit, stdout.String(), stderr.String())
	}
}
