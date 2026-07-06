package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/cloudsprints/sprintctl/internal/types"
)

// lessonJSON is the machine-readable envelope for `status --json` and
// `lesson --json`. Consumed by the CloudSprints IDE extension; fields are
// additive-only so external consumers never break.
type lessonJSON struct {
	LessonToken string `json:"lesson_token"`
	types.Lesson
}

func printLessonJSON(lessonToken string, lesson types.Lesson) {
	out, err := json.MarshalIndent(lessonJSON{LessonToken: lessonToken, Lesson: lesson}, "", "  ")
	if err != nil {
		exitJSONError(err)
	}
	fmt.Println(string(out))
}

// exitJSONError reports errors on stderr as JSON so machine consumers can
// distinguish them from payload output, then exits non-zero.
func exitJSONError(err error) {
	msg, _ := json.Marshal(map[string]string{"error": err.Error()})
	fmt.Fprintln(os.Stderr, string(msg))
	os.Exit(1)
}
