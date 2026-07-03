package cmd

import (
	"testing"
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

func TestRunValidationCommandNotFound(t *testing.T) {
	result := runValidationCommand("definitely-not-a-real-command-xyz", 0, 1)
	if result.ExitCode == 0 {
		t.Fatal("expected non-zero exit code for missing command")
	}
}
