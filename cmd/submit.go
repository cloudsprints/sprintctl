package cmd

import (
	"fmt"

	"github.com/cloudsprints/sprintctl/internal/mtcapi"
	"github.com/cloudsprints/sprintctl/internal/styles"
	"github.com/cloudsprints/sprintctl/internal/tui"
	"github.com/cloudsprints/sprintctl/internal/types"
	"github.com/erikgeiser/promptkit/confirmation"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var submitCmd = &cobra.Command{
	Use:     "submit [lesson-token]",
	Short:   "Submit a lesson for grading",
	Args:    cobra.MaximumNArgs(1),
	Example: "sprintctl submit\nsprintctl submit cm4ppz694200blze51ts1234",
	Run: func(cmd *cobra.Command, args []string) {
		// Determine lesson token
		var lessonToken string
		apiClient := mtcapi.New(viper.GetString("api_base_url"))

		if len(args) == 0 {
			// No token provided - auto-detect active lab
			fmt.Println("No lesson token provided. Fetching most recently accessed lab...")
			activeLesson, err := apiClient.GetActiveLesson()
			if err != nil {
				fmt.Println("Error fetching active lab:", err)
				fmt.Println("\nPlease open a lab in the UI first, or provide a lesson token explicitly:")
				fmt.Println("  sprintctl submit <lesson-token>")
				return
			}
			lessonToken = activeLesson.LessonToken
			fmt.Printf("\n📚 Auto-detected lab: %s\n", activeLesson.Title)
			if activeLesson.CourseTitle != "" {
				fmt.Printf("   Course: %s\n", activeLesson.CourseTitle)
			}
			fmt.Println()
		} else {
			lessonToken = args[0]
		}

		reset, _ := cmd.Flags().GetBool("reset")
		lesson, err := apiClient.GetLesson(lessonToken)
		if err != nil {
			fmt.Println("Error getting lesson:", err)
			return
		}

		if reset {
			lesson, err = apiClient.ResetLesson(lessonToken)
			if err != nil {
				fmt.Println(styles.ErrorStyle.Render(" ERROR "), "resetting lesson:", err)
				return
			}
			fmt.Println("\n" + styles.SuccessStyle.Render(" LESSON RESET "))
			printTasksTable(lesson.Tasks)
			return
		}

		printTasksTable(lesson.Tasks)
		fmt.Println()

		// Run the submission flow
		runSubmissionFlow(lessonToken, lesson, apiClient)
	},
}

func init() {
	rootCmd.AddCommand(submitCmd)
	submitCmd.Flags().BoolP("reset", "r", false, "Reset the lesson tasks")
}

func runSubmissionFlow(lessonToken string, lesson types.Lesson, apiClient *mtcapi.MtcApiClient) {
	// Display submission info
	fmt.Println(styles.SectionHeaderStyle.Render("🚀 SUBMIT FOR GRADING"))
	fmt.Printf("Your lesson will be validated with %d command(s) and submitted for AI grading.\n\n", len(lesson.CliCommands))

	input := confirmation.New("Submit lesson for grading?", confirmation.Yes)
	ready, err := input.RunPrompt()
	if err != nil {
		fmt.Println("Error getting confirmation:", err)
		return
	}
	if !ready {
		fmt.Println(styles.WarningStyle.Render(" ABORTED "))
		return
	}

	cliCommandResults, aborted := runCommandsWithProgress(lesson.CliCommands, true)
	if aborted {
		fmt.Println(styles.WarningStyle.Render(" ABORTED "))
		return
	}

	for _, cliCommandResult := range cliCommandResults {
		if commandFailed(cliCommandResult) {
			fmt.Println("\n" + styles.ErrorStyle.Render(" VALIDATION FAILED "))
			fmt.Printf("Command exited with code %d\n\n", cliCommandResult.ExitCode)
			if cliCommandResult.Stderr != "" {
				fmt.Println(styles.ErrorStyle.Render(" ERROR OUTPUT "))
				fmt.Println(cliCommandResult.Stderr)
				fmt.Println()
			}
			if cliCommandResult.Stdout != "" {
				fmt.Println(styles.InfoStyle.Render(" OUTPUT "))
				fmt.Println(cliCommandResult.Stdout)
				fmt.Println()
			}
			fmt.Println(styles.WarningStyle.Render(" SUBMISSION ABORTED "))
			fmt.Println("Fix the errors above and try again.")
			fmt.Println()
			return
		}
	}

	fmt.Println("\nSubmitting for AI grading...")

	lesson, err = apiClient.SubmitLesson(lessonToken, cliCommandResults)
	if err != nil {
		fmt.Println("\n" + styles.ErrorStyle.Render(" SUBMISSION ERROR "))
		fmt.Println(err)
		return
	}

	fmt.Println("\n" + styles.SuccessStyle.Render(" GRADING COMPLETE! "))

	// Show grading results summary
	completed := 0
	failed := 0
	for _, task := range lesson.Tasks {
		switch task.Status {
		case "COMPLETED":
			completed++
		case "FAILED":
			failed++
		}
	}

	summaryText := fmt.Sprintf("Tasks Completed: %d/%d", completed, len(lesson.Tasks))
	if failed > 0 {
		summaryText += fmt.Sprintf(" | Failed: %d", failed)
	}

	fmt.Println(summaryText)
	fmt.Println(styles.ProgressBar(completed, len(lesson.Tasks), 40))
	fmt.Println()

	fmt.Println()
	// Launch TUI for interactive grading report
	shouldResubmit, err := tui.RunGradingReport(lesson.Tasks, lessonToken)

	// Clear screen after TUI exits (whether quitting or resubmitting)
	fmt.Print("\033[H\033[2J")

	if err != nil {
		// Fallback to table view if TUI fails
		fmt.Println(styles.WarningStyle.Render(" TUI ERROR "))
		fmt.Printf("Error: %v\n\n", err)
		printTasksTable(lesson.Tasks)
		fmt.Println()
	} else if shouldResubmit {
		// User wants to resubmit - refetch lesson state and run the flow again
		fmt.Println(styles.InfoStyle.Render(" RESUBMITTING LESSON "))
		fmt.Println()
		fresh, err := apiClient.GetLesson(lessonToken)
		if err != nil {
			fmt.Println("Error getting lesson:", err)
			return
		}
		runSubmissionFlow(lessonToken, fresh, apiClient)
	}
}

func printTasksTable(tasks []types.Task) {
	fmt.Println("\n" + styles.SectionHeaderStyle.Render("📋 TASK STATUS"))

	for _, task := range tasks {
		statusIcon := styles.StatusIcon(task.Status)
		statusStyle := styles.StatusText(task.Status)

		taskLine := fmt.Sprintf("%s %s %s",
			statusIcon,
			task.Title,
			statusStyle.Render(task.Status))

		fmt.Println(styles.ListItemStyle.Render(taskLine))
	}
	fmt.Println()
}
