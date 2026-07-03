package cmd

import (
	"strings"
	"testing"
	"time"
)

func TestRunValidationCommandSuccess(t *testing.T) {
	result := runValidationCommand("echo hello", 0, 1)
	if result.ExitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", result.ExitCode)
	}
	if result.Stdout != "hello" {
		t.Fatalf("expected stdout %q, got %q", "hello", result.Stdout)
	}
}

func TestRunValidationCommandFailure(t *testing.T) {
	result := runValidationCommand("echo oops >&2; exit 3", 0, 1)
	if result.ExitCode != 3 {
		t.Fatalf("expected exit code 3, got %d", result.ExitCode)
	}
	if result.Stderr != "oops" {
		t.Fatalf("expected stderr %q, got %q", "oops", result.Stderr)
	}
}

func TestRunValidationCommandTimeout(t *testing.T) {
	original := commandTimeout
	commandTimeout = 1 * time.Second
	defer func() { commandTimeout = original }()

	start := time.Now()
	result := runValidationCommand("sleep 30", 0, 1)
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

func TestRunValidationCommandNotFound(t *testing.T) {
	result := runValidationCommand("definitely-not-a-real-command-xyz", 0, 1)
	if result.ExitCode == 0 {
		t.Fatal("expected non-zero exit code for missing command")
	}
}
