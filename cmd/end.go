package cmd

import (
	"fmt"

	"github.com/cloudsprints/sprintctl/internal/mtcapi"
	"github.com/cloudsprints/sprintctl/internal/styles"
	"github.com/erikgeiser/promptkit/confirmation"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var endCmd = &cobra.Command{
	Use:     "end [lesson-token]",
	Short:   "End the current lab and destroy its workspace",
	Args:    cobra.MaximumNArgs(1),
	Example: "sprintctl end",
	Run: func(cmd *cobra.Command, args []string) {
		apiClient := mtcapi.New(viper.GetString("api_base_url"))

		token, source, activeLesson, err := resolveLessonToken(args, apiClient)
		if err != nil {
			fmt.Println("Error fetching active lab:", err)
			fmt.Println("\nOpen a lab first, or pass a lesson token: sprintctl end <lesson-token>")
			return
		}
		if source != tokenFromArg {
			printDetectionBanner(source, activeLesson)
		}

		if yes, _ := cmd.Flags().GetBool("yes"); !yes {
			input := confirmation.New(
				"End this lab? Your workspace will be destroyed (submit first if you want a grade)",
				confirmation.No,
			)
			ok, err := input.RunPrompt()
			if err != nil {
				fmt.Println("Error getting confirmation:", err)
				return
			}
			if !ok {
				fmt.Println(styles.WarningStyle.Render(" ABORTED "))
				return
			}
		}

		fmt.Println("\nEnding lab and tearing down workspace...")
		if err := apiClient.EndLab(token); err != nil {
			fmt.Println("\n" + styles.ErrorStyle.Render(" END LAB ERROR "))
			fmt.Println(err)
			return
		}

		fmt.Println("\n" + styles.SuccessStyle.Render(" LAB ENDED "))
		fmt.Println("Your workspace is being destroyed. You can relaunch this lab any time.")
	},
}

func init() {
	rootCmd.AddCommand(endCmd)
	endCmd.Flags().BoolP("yes", "y", false, "Skip the confirmation prompt")
}
