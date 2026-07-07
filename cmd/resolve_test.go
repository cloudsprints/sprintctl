package cmd

import (
	"testing"
)

// The API-fallback path is deliberately not covered here: it needs a stored
// session (real keychain access on dev machines) and is the long-standing
// legacy path. These tests pin the resolution order in front of it: explicit
// argument first, then the machine's injected lesson.

func TestResolveLessonTokenArgWins(t *testing.T) {
	t.Setenv(machineLessonEnv, "cmenvenvenvenvenv")

	token, source, activeLesson, err := resolveLessonToken([]string{"cmargargargargarg"}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "cmargargargargarg" {
		t.Errorf("token = %q, want the explicit argument", token)
	}
	if source != tokenFromArg {
		t.Errorf("source = %v, want tokenFromArg", source)
	}
	if activeLesson != nil {
		t.Errorf("activeLesson should be nil for the argument path")
	}
}

func TestResolveLessonTokenMachineEnvWins(t *testing.T) {
	t.Setenv(machineLessonEnv, "cmenvenvenvenvenv")

	token, source, activeLesson, err := resolveLessonToken(nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "cmenvenvenvenvenv" {
		t.Errorf("token = %q, want the machine env lesson", token)
	}
	if source != tokenFromMachine {
		t.Errorf("source = %v, want tokenFromMachine", source)
	}
	if activeLesson != nil {
		t.Errorf("activeLesson should be nil for the machine env path")
	}
}

func TestResolveLessonTokenIgnoresInvalidMachineEnv(t *testing.T) {
	// Not a CUID (empty or malformed) must not be trusted as a lesson token.
	for _, invalid := range []string{"", "short", "not-a-cuid-x"} {
		t.Setenv(machineLessonEnv, invalid)

		// nil apiClient: reaching the API fallback would panic, which is
		// exactly what we assert does NOT happen for the argument path...
		token, source, _, err := resolveLessonToken([]string{"cmargargargargarg"}, nil)
		if err != nil || token != "cmargargargargarg" || source != tokenFromArg {
			t.Errorf("env %q: argument path broken: token=%q source=%v err=%v", invalid, token, source, err)
		}
	}
}
