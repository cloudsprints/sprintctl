package cmd

import (
	"fmt"

	"github.com/morethancertified/sprintctl/internal/auth"
	"github.com/morethancertified/sprintctl/internal/styles"
	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authentication commands",
	Long:  "Commands for managing authentication with CloudSprints.",
}

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show authentication status",
	Run: func(cmd *cobra.Command, args []string) {
		if auth.IsAuthenticated() {
			fmt.Println(styles.SuccessStyle.Render(" AUTHENTICATED "))
			fmt.Println(styles.BoxStyle.Render("You are signed in to CloudSprints.\n\nAvailable commands:\n• sprintctl submit <lesson-token>\n• sprintctl status [lesson-token]\n• sprintctl logout"))
		} else {
			fmt.Println(styles.ErrorStyle.Render(" NOT AUTHENTICATED "))
			fmt.Println(styles.BoxStyle.Render("You are not signed in to CloudSprints.\n\nTo authenticate:\n• Run 'sprintctl login' to sign in with OTP\n• Run 'sprintctl auth token <token>' for manual token"))
		}
	},
}

var authTokenCmd = &cobra.Command{
	Use:   "token <access_token>",
	Short: "Manually set authentication token (beta)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		token := args[0]
		
		if len(token) < 10 {
			fmt.Println(styles.ErrorStyle.Render(" INVALID TOKEN "))
			fmt.Println(styles.BoxStyle.Render("Token must be at least 10 characters long"))
			return
		}
		
		err := auth.StoreToken(token)
		if err != nil {
			fmt.Println(styles.ErrorStyle.Render(" STORAGE ERROR "), err)
			return
		}
		
		fmt.Println(styles.SuccessStyle.Render(" TOKEN STORED! "))
		fmt.Println(styles.BoxStyle.Render("Authentication token stored successfully!\nYou can now use sprintctl commands."))
	},
}

func init() {
	authCmd.AddCommand(authStatusCmd)
	authCmd.AddCommand(authTokenCmd)
	rootCmd.AddCommand(authCmd)
}
