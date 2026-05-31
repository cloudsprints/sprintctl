package cmd

import (
	"fmt"
	"strings"

	"github.com/erikgeiser/promptkit/textinput"
	"github.com/morethancertified/sprintctl/internal/auth"
	"github.com/morethancertified/sprintctl/internal/styles"
	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login [email]",
	Short: "Authenticate with CloudSprints using email OTP",
	Long: `Authenticate with CloudSprints using a one-time password sent to your email.
You can provide your email as an argument or enter it when prompted.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var email string
		
		// Get email from args or prompt
		if len(args) > 0 {
			email = args[0]
		} else {
			input := textinput.New("Enter your email:")
			input.Placeholder = "user@example.com"
			
			var err error
			email, err = input.RunPrompt()
			if err != nil {
				fmt.Printf("Error getting email: %v\n", err)
				return
			}
		}
		
		// Validate email format (basic check)
		if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
			fmt.Println(styles.ErrorStyle.Render(" INVALID EMAIL "))
			fmt.Println(styles.BoxStyle.Render("Please enter a valid email address"))
			return
		}
		
		// Check if already authenticated
		if auth.IsAuthenticated() {
			fmt.Println(styles.WarningStyle.Render(" ALREADY AUTHENTICATED "))
			fmt.Println(styles.BoxStyle.Render("You are already authenticated!\nRun 'sprintctl logout' to sign out first."))
			return
		}
		
		// Request OTP
		fmt.Println(styles.InfoStyle.Render(" SENDING OTP "))
		fmt.Println(styles.BoxStyle.Render(fmt.Sprintf("Sending one-time password to: %s", email)))
		err := auth.LoginWithOTP(email)
		if err != nil {
			fmt.Println(styles.ErrorStyle.Render(" OTP ERROR "), err)
			return
		}
		
		fmt.Println(styles.SuccessStyle.Render(" OTP SENT! "))
		fmt.Println(styles.BoxStyle.Render("Check your email for the 6-digit verification code"))
		
		// Prompt for OTP code
		codeInput := textinput.New("Enter the 6-digit code from your email:")
		codeInput.Placeholder = "123456"
		
		code, err := codeInput.RunPrompt()
		if err != nil {
			fmt.Printf("Error getting code: %v\n", err)
			return
		}
		
		// Verify OTP
		fmt.Println(styles.InfoStyle.Render(" VERIFYING CODE "))
		err = auth.VerifyOTP(email, code)
		if err != nil {
			fmt.Println(styles.ErrorStyle.Render(" AUTHENTICATION FAILED "), err)
			return
		}
		
		fmt.Println(styles.SuccessStyle.Render(" AUTHENTICATION SUCCESS! "))
		fmt.Println(styles.BoxStyle.Render("You can now use sprintctl to submit lessons and access your data.\n\nNext steps:\n• Run 'sprintctl submit <lesson-token>' to grade a lesson\n• Run 'sprintctl status' to view cached results"))
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
}
