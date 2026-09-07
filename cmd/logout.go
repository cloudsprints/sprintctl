package cmd

import (
	"fmt"
	"strings"

	"github.com/cloudsprints/sprintctl/internal/auth"
	"github.com/cloudsprints/sprintctl/internal/styles"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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

		appRoot := strings.TrimSuffix(viper.GetString("api_base_url"), "/api/v1")
		err := auth.Logout(appRoot)
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
