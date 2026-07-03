package cmd

import (
	"strings"
	"testing"
	"time"

	"github.com/cloudsprints/sprintctl/internal/types"
)

func TestExecuteCommandSuccess(t *testing.T) {
	result := executeCommand("echo hello")
	if result.ExitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", result.ExitCode)
	}
	if result.Stdout != "hello" {
		t.Fatalf("expected stdout %q, got %q", "hello", result.Stdout)
	}
}

func TestExecuteCommandFailure(t *testing.T) {
	result := executeCommand("echo oops >&2; exit 3")
	if result.ExitCode != 3 {
		t.Fatalf("expected exit code 3, got %d", result.ExitCode)
	}
	if result.Stderr != "oops" {
		t.Fatalf("expected stderr %q, got %q", "oops", result.Stderr)
	}
}

func TestExecuteCommandTimeout(t *testing.T) {
	original := commandTimeout
	commandTimeout = 1 * time.Second
	defer func() { commandTimeout = original }()

	start := time.Now()
	result := executeCommand("sleep 30")
	elapsed := time.Since(start)

	if result.ExitCode != 124 {
		t.Fatalf("expected exit code 124 for timeout, got %d", result.ExitCode)
	}
	if !strings.Contains(result.Stderr, "timed out") {
		t.Fatalf("expected timeout message in stderr, got %q", result.Stderr)
	}
	if elapsed > 5*time.Second {
		t.Fatalf("command was not killed promptly, took %s", elapsed)
	}
}

func TestExecuteCommandNotFound(t *testing.T) {
	result := executeCommand("definitely-not-a-real-command-xyz")
	if result.ExitCode == 0 {
		t.Fatal("expected non-zero exit code for missing command")
	}
}

func TestCommandFailed(t *testing.T) {
	if commandFailed(types.CLICommandResult{ExitCode: 0, Stdout: "all good"}) {
		t.Fatal("expected success result to not be failed")
	}
	if !commandFailed(types.CLICommandResult{ExitCode: 1}) {
		t.Fatal("expected non-zero exit code to be failed")
	}
	if !commandFailed(types.CLICommandResult{ExitCode: 0, Stdout: "validation failed"}) {
		t.Fatal("expected 'validation failed' output to be failed")
	}
}
