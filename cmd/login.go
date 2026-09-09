package cmd

import (
	"errors"
	"fmt"
	"os/exec"
	"runtime"

	"github.com/cloudsprints/sprintctl/internal/auth"
	"github.com/cloudsprints/sprintctl/internal/styles"
	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with CloudSprints",
	Long: `Authenticate with CloudSprints.

Opens your browser to approve the login using your existing CloudSprints
session. Sign in on the web first if you aren't already.`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		// Check if already authenticated before asking for anything
		if auth.IsAuthenticated() {
			fmt.Println(styles.WarningStyle.Render(" ALREADY AUTHENTICATED "))
			fmt.Println(styles.BoxStyle.Render("You are already authenticated!\nRun 'sprintctl logout' to sign out first."))
			return
		}

		appRoot := appRoot()

		if err := runDeviceLogin(appRoot); err != nil {
			fmt.Println(styles.ErrorStyle.Render(" LOGIN FAILED "), err)
			if errors.Is(err, auth.ErrDeviceFlowUnsupported) {
				fmt.Println(styles.BoxStyle.Render("This server doesn't support browser login. Check that sprintctl points at CloudSprints (see 'sprintctl env')."))
			}
			return
		}
		printLoginSuccess()
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

func printLoginSuccess() {
	fmt.Println(styles.SuccessStyle.Render(" AUTHENTICATION SUCCESS! "))
	fmt.Println(styles.BoxStyle.Render(envLine() + "\n\nYou can now use sprintctl to submit lessons and access your data.\n\nNext steps:\n• Run 'sprintctl submit <lesson-token>' to grade a lesson\n• Run 'sprintctl status' to view cached results"))
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
	rootCmd.AddCommand(loginCmd)
}
