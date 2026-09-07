package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

// chdir switches the working directory for a test and restores it after
// (testing.Chdir needs go1.24; go.mod still targets 1.23).
func chdir(t *testing.T, dir string) {
	t.Helper()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })
}

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

func TestResolveLessonTokenMarkerBeatsAPI(t *testing.T) {
	t.Setenv(machineLessonEnv, "")
	dir := t.TempDir()
	if err := writeLabMarker(dir, labMarker{UserLessonID: "cmdirdirdirdirdir", Title: "Marker lab"}); err != nil {
		t.Fatal(err)
	}
	// Run from a nested folder: the marker must be found in a parent.
	nested := filepath.Join(dir, "public", "src")
	if err := os.MkdirAll(nested, 0755); err != nil {
		t.Fatal(err)
	}
	chdir(t, nested)

	// nil apiClient: reaching the API fallback would panic.
	token, source, activeLesson, err := resolveLessonToken(nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "cmdirdirdirdirdir" || source != tokenFromDir || activeLesson != nil {
		t.Errorf("got token=%q source=%v activeLesson=%v, want the directory marker", token, source, activeLesson)
	}
}

func TestResolveLessonTokenMachineEnvBeatsMarker(t *testing.T) {
	dir := t.TempDir()
	if err := writeLabMarker(dir, labMarker{UserLessonID: "cmdirdirdirdirdir"}); err != nil {
		t.Fatal(err)
	}
	chdir(t, dir)
	t.Setenv(machineLessonEnv, "cmenvenvenvenvenv")

	token, source, _, err := resolveLessonToken(nil, nil)
	if err != nil || token != "cmenvenvenvenvenv" || source != tokenFromMachine {
		t.Errorf("got token=%q source=%v err=%v, want the machine env lesson", token, source, err)
	}
}

func TestFindLabMarkerIgnoresGarbage(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, labMarkerFile), []byte("{not json"), 0644); err != nil {
		t.Fatal(err)
	}
	chdir(t, dir)
	if m := findLabMarker(); m != nil {
		t.Errorf("garbage marker should be ignored, got %+v", m)
	}
}
