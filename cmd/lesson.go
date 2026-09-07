package cmd

import (
	"fmt"

	"github.com/cloudsprints/sprintctl/internal/mtcapi"
	"github.com/cloudsprints/sprintctl/internal/styles"
	"github.com/spf13/cobra"
)

var lessonCmd = &cobra.Command{
	Use:     "lesson [lesson-token]",
	Short:   "Show the lab's ticket/scenario content (auto-detects the active lab if no token is provided)",
	Args:    cobra.MaximumNArgs(1),
	Example: "sprintctl lesson\nsprintctl lesson --json\nsprintctl lesson cm4ppz694200blze51ts1234",
	Run: func(cmd *cobra.Command, args []string) {
		apiClient := mtcapi.New(apiBaseURL())
		jsonOut, _ := cmd.Flags().GetBool("json")

		lessonToken, _, _, err := resolveLessonToken(args, apiClient)
		if err != nil {
			if jsonOut {
				exitJSONError(fmt.Errorf("could not detect an active lab: %w", err))
			}
			if printAuthError(err) {
				return
			}
			fmt.Println(styles.ErrorStyle.Render(" NO ACTIVE LAB "))
			fmt.Println(styles.BoxStyle.Render("Could not detect an active lab.\n\nOpen a lab in the UI, or pass a token explicitly:\n  sprintctl lesson <lesson-token>"))
			return
		}

		lesson, err := apiClient.GetLesson(lessonToken)
		if err != nil {
			if jsonOut {
				exitJSONError(err)
			}
			fmt.Println(styles.ErrorStyle.Render(" API ERROR "), err)
			return
		}

		if jsonOut {
			printLessonJSON(lessonToken, lesson)
			return
		}

		if lesson.Title != "" {
			fmt.Println(styles.SectionHeaderStyle.Render("🎫 " + lesson.Title))
		}
		if lesson.Course != nil && lesson.Course.Title != "" {
			fmt.Printf("Course: %s\n", lesson.Course.Title)
		}
		if lesson.HelpdeskPriority > 0 {
			fmt.Printf("Priority: P%d\n", lesson.HelpdeskPriority)
		}
		if lesson.ReporterPersona != nil {
			fmt.Printf("Reported by: %s\n", lesson.ReporterPersona.Name)
		}
		if lesson.LabContent != "" {
			fmt.Println()
			fmt.Println(lesson.LabContent)
		} else if lesson.Title == "" {
			fmt.Println(styles.WarningStyle.Render(" NO CONTENT "), "This server does not provide lesson content; use the lab page in the UI.")
		}
	},
}

func init() {
	rootCmd.AddCommand(lessonCmd)
	lessonCmd.Flags().Bool("json", false, "Output machine-readable JSON instead of formatted text")
}
