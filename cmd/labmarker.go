package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// labMarkerFile is written into a lab directory by `sprintctl init` so that
// later commands run from inside that directory (grade, submit, status,
// lesson) target the initialized lab instead of guessing from click history.
// Course labs graded from a student's own terminal have no machine env and
// the API's "most recently clicked" detection is easily clobbered by opening
// any other ticket in the UI.
const labMarkerFile = ".cloudsprints-lab.json"

// Walk at most this many parent directories looking for a marker, so a grade
// run from a nested lab folder still resolves without scanning to /.
const labMarkerMaxDepth = 8

type labMarker struct {
	UserLessonID string `json:"user_lesson_id"`
	Title        string `json:"title,omitempty"`
	APIBaseURL   string `json:"api_base_url,omitempty"`
}

func writeLabMarker(dir string, m labMarker) error {
	payload, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, labMarkerFile), append(payload, '\n'), 0644)
}

// findLabMarker looks for a marker in the working directory or one of its
// parents. Returns nil when there is none (or the marker is unusable).
func findLabMarker() *labMarker {
	dir, err := os.Getwd()
	if err != nil {
		return nil
	}
	for i := 0; i <= labMarkerMaxDepth; i++ {
		data, err := os.ReadFile(filepath.Join(dir, labMarkerFile))
		if err == nil {
			var m labMarker
			if json.Unmarshal(data, &m) == nil && m.UserLessonID != "" {
				return &m
			}
			return nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return nil
		}
		dir = parent
	}
	return nil
}
