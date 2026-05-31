package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cloudsprints/sprintctl/internal/styles"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

var Version = "v0.0.0"

var rootCmd = &cobra.Command{
	Use:     "sprintctl",
	Short:   "CloudSprints CLI for lab grading",
	Long: fmt.Sprintf(`%s
%s

%s`,
		styles.LogoStyle.Render("⚡ sprintctl"),
		styles.SubtitleStyle.Render("CloudSprints CLI"),
		styles.HintStyle.Render("Run 'sprintctl grade' to submit your lab for grading")),
	Version: Version,
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.CompletionOptions.HiddenDefaultCmd = true
	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true})
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/.config/sprintctl/config.json)")
	rootCmd.PersistentFlags().StringP("api-base-url", "l", viper.GetString("api_base_url"), "API base URL")
	viper.BindPFlag("api_base_url", rootCmd.PersistentFlags().Lookup("api-base-url"))
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
		err := viper.ReadInConfig()
		cobra.CheckErr(err)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		configDir := filepath.Join(home, ".config", "sprintctl")
		err = os.MkdirAll(configDir, os.ModePerm)
		cobra.CheckErr(err)

		viper.SetConfigName("config")
		viper.SetConfigType("json")
		viper.AddConfigPath(configDir)
		viper.AutomaticEnv()
		viper.SetEnvPrefix("SPRINTCTL")

		if err := viper.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); ok {
				viper.Set("api_base_url", "https://cloudsprints.com/api/v1")
				err := viper.SafeWriteConfigAs(filepath.Join(configDir, "config.json"))
				cobra.CheckErr(err)
			} else {
				cobra.CheckErr(err)
			}
		}
	}
}

func Execute() {
	err := rootCmd.Execute()
	cobra.CheckErr(err)
}
