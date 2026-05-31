package cmd

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/erikgeiser/promptkit/confirmation"
	"github.com/morethancertified/sprintctl/internal/mtcapi"
	"github.com/morethancertified/sprintctl/internal/styles"
	"github.com/morethancertified/sprintctl/internal/tui"
	"github.com/morethancertified/sprintctl/internal/types"
	"github.com/morethancertified/sprintctl/internal/widgets"
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
				fmt.Println("  mtc submit <lesson-token>")
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
		runSubmissionFlow(lessonToken, apiClient)
	},
}

func init() {
	rootCmd.AddCommand(submitCmd)
	submitCmd.Flags().BoolP("reset", "r", false, "Reset the lesson tasks")
}

func runSubmissionFlow(lessonToken string, apiClient *mtcapi.MtcApiClient) {
	// Get fresh lesson data
	lesson, err := apiClient.GetLesson(lessonToken)
	if err != nil {
		fmt.Println("Error getting lesson:", err)
		return
	}

	// Display submission info
	fmt.Println(styles.SectionHeaderStyle.Render("🚀 SUBMIT FOR GRADING"))
	fmt.Printf("Your lesson will be validated with %d command(s) and submitted for AI grading.\n\n", len(lesson.CliCommands))
	
	// Main submission confirmation with option to view commands
	input := confirmation.New("Submit lesson for grading? (type 'n' to view commands first)", confirmation.Yes)
	ready, err := input.RunPrompt()
	if err != nil {
		fmt.Println("Error getting confirmation:", err)
		return
	}
	
	// Check if user typed 'c' to view commands
	// Note: promptkit doesn't expose the raw input, so we'll use a different approach
	// If user said no, ask if they want to see commands
	if !ready {
		showCmd := confirmation.New("View validation commands?", confirmation.Yes)
		showCommands, err := showCmd.RunPrompt()
		if err != nil {
			fmt.Println("Error getting input:", err)
			return
		}
		
		if showCommands {
			fmt.Println("\n" + styles.InfoStyle.Render(" VALIDATION COMMANDS "))
			for i, command := range lesson.CliCommands {
				fmt.Printf("  %d. %s\n", i+1, styles.CommandStyle.Render(command))
			}
			fmt.Println()
			
			// Ask again after showing commands
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
		} else {
			fmt.Println(styles.WarningStyle.Render(" ABORTED "))
			return
		}
	}

	widgets.RunProgressBar()

	cliCommandResults := []types.CLICommandResult{}
	for _, command := range lesson.CliCommands {
		cliCommandResult := types.CLICommandResult{
			Command: command,
		}

		cmd := exec.Command("sh", "-c", "LANG=en_US.UTF-8 "+command)

		b, err := cmd.Output()
		if ee, ok := err.(*exec.ExitError); ok {
			cliCommandResult.ExitCode = ee.ExitCode()
			cliCommandResult.Stderr = strings.TrimRight(string(ee.Stderr), "\n\t\r")
		} else if err != nil {
			cliCommandResult.ExitCode = -69
		} else {
			cliCommandResult.Stdout = strings.TrimRight(string(b), "\n\t\r")
		}

		cliCommandResults = append(cliCommandResults, cliCommandResult)

		// Check if command failed - abort submission if:
		// 1. Non-zero exit code, OR
		// 2. Output contains "validation failed" (from || echo "validation failed")
		validationFailed := cliCommandResult.ExitCode != 0 || 
			strings.Contains(cliCommandResult.Stdout, "validation failed")

		if validationFailed {
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
		// User wants to resubmit - run submission flow again
		fmt.Println(styles.InfoStyle.Render(" RESUBMITTING LESSON "))
		fmt.Println()
		runSubmissionFlow(lessonToken, apiClient)
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
