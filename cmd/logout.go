package cmd

import (
	"fmt"

	"github.com/morethancertified/sprintctl/internal/auth"
	"github.com/morethancertified/sprintctl/internal/styles"
	"github.com/spf13/cobra"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Sign out of CloudSprints",
	Long:  "Remove stored authentication credentials and sign out of CloudSprints.",
	Run: func(cmd *cobra.Command, args []string) {
		if !auth.IsAuthenticated() {
			fmt.Println(styles.WarningStyle.Render(" NOT AUTHENTICATED "))
			fmt.Println(styles.BoxStyle.Render("You are not currently authenticated.\nRun 'sprintctl login' to authenticate."))
			return
		}
		
		err := auth.Logout()
		if err != nil {
			fmt.Println(styles.ErrorStyle.Render(" LOGOUT ERROR "), err)
			return
		}
		
		fmt.Println(styles.SuccessStyle.Render(" SIGNED OUT! "))
		fmt.Println(styles.BoxStyle.Render("Successfully signed out of CloudSprints.\nRun 'sprintctl login' to authenticate again."))
	},
}

func init() {
	rootCmd.AddCommand(logoutCmd)
}
