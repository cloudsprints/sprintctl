package cmd

import (
	"fmt"

	"github.com/cloudsprints/sprintctl/internal/mtcapi"
	"github.com/cloudsprints/sprintctl/internal/styles"
	"github.com/cloudsprints/sprintctl/internal/tui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var statusCmd = &cobra.Command{
	Use:   "status [lesson-token]",
	Short: "Show the status of a lesson (auto-detects the active lab if no token is provided)",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		apiClient := mtcapi.New(viper.GetString("api_base_url"))

		var lessonToken string
		if len(args) > 0 {
			lessonToken = args[0]
		} else {
			activeLesson, err := apiClient.GetActiveLesson()
			if err != nil {
				fmt.Println(styles.ErrorStyle.Render(" NO ACTIVE LAB "))
				fmt.Println(styles.BoxStyle.Render("Could not detect an active lab.\n\nOpen a lab in the UI, or pass a token explicitly:\n  sprintctl status <lesson-token>"))
				return
			}
			lessonToken = activeLesson.LessonToken
			fmt.Printf("\n📚 Auto-detected lab: %s\n", activeLesson.Title)
			if activeLesson.CourseTitle != "" {
				fmt.Printf("   Course: %s\n", activeLesson.CourseTitle)
			}
			fmt.Println()
		}

		lesson, err := apiClient.GetLesson(lessonToken)
		if err != nil {
			fmt.Println(styles.ErrorStyle.Render(" API ERROR "), err)
			return
		}

		fmt.Println(styles.InfoStyle.Render(" STATUS "))
		fmt.Println(styles.BoxStyle.Render(fmt.Sprintf("Lesson Token: %s", lessonToken)))

		// Launch TUI for interactive grading report.
		// status doesn't support resubmit, so pass empty token and ignore resubmit result.
		_, err = tui.RunGradingReport(lesson.Tasks, "")
		if err != nil {
			fmt.Println(styles.WarningStyle.Render(" TUI ERROR "), err)
			fmt.Println(styles.SectionHeaderStyle.Render("📋 FALLBACK OUTPUT"))
			printTasksTable(lesson.Tasks)
		}
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
