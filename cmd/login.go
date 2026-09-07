package cmd

import (
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/cloudsprints/sprintctl/internal/auth"
	"github.com/cloudsprints/sprintctl/internal/styles"
	"github.com/erikgeiser/promptkit/textinput"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var loginCmd = &cobra.Command{
	Use:   "login [email]",
	Short: "Authenticate with CloudSprints",
	Long: `Authenticate with CloudSprints.

By default this opens your browser to approve the login using your existing
CloudSprints session. Pass --otp (or an email argument) to sign in with a
one-time password sent to your email instead.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Check if already authenticated before asking for anything
		if auth.IsAuthenticated() {
			fmt.Println(styles.WarningStyle.Render(" ALREADY AUTHENTICATED "))
			fmt.Println(styles.BoxStyle.Render("You are already authenticated!\nRun 'sprintctl logout' to sign out first."))
			return
		}

		appRoot := strings.TrimSuffix(viper.GetString("api_base_url"), "/api/v1")

		// An explicit email argument implies the OTP flow, since the browser
		// flow authenticates whoever is signed in to the browser session
		useOTP, _ := cmd.Flags().GetBool("otp")
		if !useOTP && len(args) == 0 {
			err := runDeviceLogin(appRoot)
			if err == nil {
				printLoginSuccess()
				return
			}
			if !errors.Is(err, auth.ErrDeviceFlowUnsupported) {
				fmt.Println(styles.ErrorStyle.Render(" LOGIN FAILED "), err)
				return
			}
			fmt.Println(styles.InfoStyle.Render(" FALLING BACK TO EMAIL CODE LOGIN "))
		}

		if runOTPLogin(appRoot, args) {
			printLoginSuccess()
		}
	},
}

// runDeviceLogin drives the browser-based device flow
func runDeviceLogin(appRoot string) error {
	deviceAuth, err := auth.StartDeviceAuth(appRoot)
	if err != nil {
		return err
	}

	fmt.Println(styles.InfoStyle.Render(" BROWSER LOGIN "))
	fmt.Println(styles.BoxStyle.Render(fmt.Sprintf(
		"Confirmation code: %s\n\nOpening your browser to approve this login:\n%s\n\nIf the browser doesn't open, visit the link manually.",
		deviceAuth.UserCode, deviceAuth.VerificationURIComplete)))

	if err := openBrowser(deviceAuth.VerificationURIComplete); err != nil {
		fmt.Println(styles.WarningStyle.Render(" COULD NOT OPEN BROWSER "))
	}

	fmt.Println(styles.InfoStyle.Render(" WAITING FOR APPROVAL "))
	return auth.PollDeviceAuth(appRoot, deviceAuth)
}

// runOTPLogin drives the email one-time-password flow; returns true on success
func runOTPLogin(appRoot string, args []string) bool {
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
			return false
		}
	}

	// Validate email format (basic check)
	if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
		fmt.Println(styles.ErrorStyle.Render(" INVALID EMAIL "))
		fmt.Println(styles.BoxStyle.Render("Please enter a valid email address"))
		return false
	}

	proxyURL := appRoot + "/api/auth/cli-otp"

	// Request OTP
	fmt.Println(styles.InfoStyle.Render(" SENDING OTP "))
	fmt.Println(styles.BoxStyle.Render(fmt.Sprintf("Sending one-time password to: %s", email)))
	err := auth.LoginWithOTP(email, proxyURL)
	if err != nil {
		fmt.Println(styles.ErrorStyle.Render(" OTP ERROR "), err)
		return false
	}

	fmt.Println(styles.SuccessStyle.Render(" OTP SENT! "))
	fmt.Println(styles.BoxStyle.Render("Check your email for the 6-digit verification code"))

	// Prompt for OTP code
	codeInput := textinput.New("Enter the 6-digit code from your email:")
	codeInput.Placeholder = "123456"

	code, err := codeInput.RunPrompt()
	if err != nil {
		fmt.Printf("Error getting code: %v\n", err)
		return false
	}

	// Verify OTP
	fmt.Println(styles.InfoStyle.Render(" VERIFYING CODE "))
	err = auth.VerifyOTP(email, code, proxyURL)
	if err != nil {
		fmt.Println(styles.ErrorStyle.Render(" AUTHENTICATION FAILED "), err)
		return false
	}

	return true
}

func printLoginSuccess() {
	fmt.Println(styles.SuccessStyle.Render(" AUTHENTICATION SUCCESS! "))
	fmt.Println(styles.BoxStyle.Render("You can now use sprintctl to submit lessons and access your data.\n\nNext steps:\n• Run 'sprintctl submit <lesson-token>' to grade a lesson\n• Run 'sprintctl status' to view cached results"))
}

// openBrowser opens the default browser at the given URL
func openBrowser(url string) error {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", url).Start()
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	default:
		return exec.Command("xdg-open", url).Start()
	}
}

func init() {
	loginCmd.Flags().Bool("otp", false, "Sign in with an emailed one-time password instead of the browser")
	rootCmd.AddCommand(loginCmd)
}
