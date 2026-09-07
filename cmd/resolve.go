package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/cloudsprints/sprintctl/internal/mtcapi"
	"github.com/cloudsprints/sprintctl/internal/styles"
	"github.com/cloudsprints/sprintctl/internal/types"
)

// machineLessonEnv is set on CloudSprints cloud IDE machines by the lab
// workspace launcher: it names the user_lesson whose credentials are
// injected into this machine. When present it is ground truth for "the
// active lab" — the API's click-history detection can lag behind what this
// machine is actually provisioned for.
const machineLessonEnv = "CLOUDSPRINTS_USER_LESSON_ID"

type tokenSource int

const (
	tokenFromArg tokenSource = iota
	tokenFromMachine
	tokenFromDir
	tokenFromAPI
)

// resolveLessonToken picks the lesson to operate on: an explicit argument
// wins, then the machine's injected lesson (cloud IDE), then the lab marker
// `sprintctl init` left in the current directory (or a parent), then the
// API's active-lab detection. activeLesson is non-nil only on the API path.
func resolveLessonToken(args []string, apiClient *mtcapi.MtcApiClient) (string, tokenSource, *types.ActiveLesson, error) {
	if len(args) > 0 {
		return args[0], tokenFromArg, nil, nil
	}

	if token := os.Getenv(machineLessonEnv); mtcapi.ValidCUID(token) {
		return token, tokenFromMachine, nil, nil
	}

	if marker := findLabMarker(); marker != nil {
		return marker.UserLessonID, tokenFromDir, nil, nil
	}

	activeLesson, err := apiClient.GetActiveLesson()
	if err != nil {
		return "", tokenFromAPI, nil, fmt.Errorf("fetching active lab: %w", err)
	}
	return activeLesson.LessonToken, tokenFromAPI, &activeLesson, nil
}

// printDetectionBanner explains where the auto-detected lab came from so a
// surprising target is visible before anything runs against it.
func printDetectionBanner(source tokenSource, activeLesson *types.ActiveLesson) {
	if source == tokenFromMachine {
		fmt.Println("\n📚 Using the lab assigned to this workspace")
		fmt.Println()
		return
	}
	if source == tokenFromDir {
		if marker := findLabMarker(); marker != nil && marker.Title != "" {
			fmt.Printf("\n📚 Using the lab initialized in this directory: %s\n", marker.Title)
		} else {
			fmt.Println("\n📚 Using the lab initialized in this directory")
		}
		fmt.Println()
		return
	}
	if activeLesson != nil {
		fmt.Printf("\n📚 Auto-detected lab: %s\n", activeLesson.Title)
		if activeLesson.CourseTitle != "" {
			fmt.Printf("   Course: %s\n", activeLesson.CourseTitle)
		}
		fmt.Println()
	}
}

// printAuthError reports a missing or stale login and returns true when err
// is an auth error — a signed-out user must not be told "no active lab" and
// sent hunting for a lab they already opened.
func printAuthError(err error) bool {
	if err == nil || !strings.Contains(err.Error(), "sprintctl login") {
		return false
	}
	fmt.Println(styles.ErrorStyle.Render(" NOT AUTHENTICATED "))
	fmt.Println(styles.BoxStyle.Render(err.Error()))
	return true
}
