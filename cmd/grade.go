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

// gradeCmd handles grading - either as alias for submit or admin mode
var gradeCmd = &cobra.Command{
	Use:     "grade [lesson-token]",
	Short:   "Grade a lesson (or use --admin for direct lesson grading)",
	Args:    cobra.MaximumNArgs(1),
	Example: "mtc grade\nmtc grade cm4ppz694200blze51ts1234\nmtc grade --admin cmj3wql250016ru5xw0vifo2b",
	Run: func(cmd *cobra.Command, args []string) {
		adminMode, _ := cmd.Flags().GetBool("admin")

		if adminMode {
			// Admin mode: grade directly using lesson_id
			if len(args) == 0 {
				fmt.Println("Error: lesson_id is required in admin mode")
				fmt.Println("Usage: sprintctl grade --admin <lesson_id>")
				return
			}
			runAdminGrading(args[0])
		} else {
			// Normal mode: use submitCmd
			submitCmd.Run(cmd, args)
		}
	},
}

func init() {
	rootCmd.AddCommand(gradeCmd)
	gradeCmd.Flags().BoolP("reset", "r", false, "Reset the lesson tasks")
	gradeCmd.Flags().BoolP("admin", "a", false, "Admin mode: grade using lesson_id directly (no enrollment required)")
}

func runAdminGrading(lessonID string) {
	apiClient := mtcapi.New(viper.GetString("api_base_url"))

	fmt.Println("🔐 Admin mode: grading lesson directly...")

	// Get lesson info to get CLI commands
	lessonInfo, err := apiClient.GetAdminLessonInfo(lessonID)
	if err != nil {
		fmt.Printf("Error getting lesson information: %s\n", err)
		return
	}

	fmt.Printf("\n📚 Lesson: %s\n", lessonInfo.Title)
	fmt.Printf("   Tasks: %d\n", len(lessonInfo.Tasks))
	fmt.Printf("   CLI Commands: %d\n\n", len(lessonInfo.CliCommands))

	// Show tasks
	fmt.Println(styles.SectionHeaderStyle.Render("📋 TASKS TO GRADE"))
	for _, task := range lessonInfo.Tasks {
		fmt.Printf("  • %s\n", task.Title)
	}
	fmt.Println()

	// Confirm submission
	input := confirmation.New("Run CLI commands and grade?", confirmation.Yes)
	ready, err := input.RunPrompt()
	if err != nil || !ready {
		fmt.Println(styles.WarningStyle.Render(" ABORTED "))
		return
	}

	widgets.RunProgressBar()

	// Run CLI commands
	cliCommandResults := []types.CLICommandResult{}
	for _, command := range lessonInfo.CliCommands {
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
	}

	// Submit for grading via admin endpoint
	fmt.Println("\nSubmitting for AI grading...")
	result, err := apiClient.AdminGradeLesson(lessonID, cliCommandResults)
	if err != nil {
		fmt.Println(styles.ErrorStyle.Render(" GRADING ERROR "))
		fmt.Println(err)
		return
	}

	// Convert admin results to types.Task for TUI compatibility
	tasks := make([]types.Task, len(result.Results))
	for i, r := range result.Results {
		status := "FAILED"
		if r.Passed {
			status = "COMPLETED"
		}
		tasks[i] = types.Task{
			ID:            fmt.Sprintf("task-%d", r.TaskNumber),
			Title:         r.Title,
			Status:        status,
			AiExplanation: r.Explanation,
		}
	}

	fmt.Println("\n" + styles.SuccessStyle.Render(" GRADING COMPLETE! "))

	// Show summary
	summaryText := fmt.Sprintf("Tasks Completed: %d/%d", result.Summary.Passed, result.Summary.Total)
	if result.Summary.Failed > 0 {
		summaryText += fmt.Sprintf(" | Failed: %d", result.Summary.Failed)
	}
	fmt.Println(summaryText)
	fmt.Println(styles.ProgressBar(result.Summary.Passed, result.Summary.Total, 40))
	fmt.Println()

	// Launch TUI for interactive grading report (same as regular grade)
	_, err = tui.RunGradingReport(tasks, lessonID)

	// Clear screen after TUI exits
	fmt.Print("\033[H\033[2J")

	if err != nil {
		// Fallback to table view if TUI fails
		fmt.Println(styles.WarningStyle.Render(" TUI ERROR "))
		fmt.Printf("Error: %v\n\n", err)
		printAdminTasksTable(tasks)
	}
}

func printAdminTasksTable(tasks []types.Task) {
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
