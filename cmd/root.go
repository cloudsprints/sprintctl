package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cloudsprints/sprintctl/internal/auth"
	"github.com/cloudsprints/sprintctl/internal/styles"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

var Version = "v0.0.0"

var rootCmd = &cobra.Command{
	Use:   "sprintctl",
	Short: "CloudSprints CLI for lab grading",
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
	rootCmd.PersistentFlags().StringP("api-base-url", "l", "", "API base URL (overrides --env)")
	viper.BindPFlag("api_base_url", rootCmd.PersistentFlags().Lookup("api-base-url"))
	rootCmd.PersistentFlags().StringP("env", "e", "", "environment: prod, staging or local (see 'sprintctl env')")
	viper.BindPFlag("env", rootCmd.PersistentFlags().Lookup("env"))
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
				// SetDefault, NOT Set: Set outranks env vars, which made every FIRST
				// sprintctl run in a fresh box ignore SPRINTCTL_API_BASE_URL and
				// phone prod (dev/staging labs only ever pulled files by the accident
				// of cloned userLesson ids existing in prod). Defaults rank below env.
				viper.SetDefault("env", defaultEnv)
				err := viper.SafeWriteConfigAs(filepath.Join(configDir, "config.json"))
				cobra.CheckErr(err)
			} else {
				cobra.CheckErr(err)
			}
		}
	}

	// An explicit URL (flag or SPRINTCTL_API_BASE_URL) outranks the named
	// environment; a URL that merely sits in the config file is the legacy
	// pre-profile setting and ranks below --env / SPRINTCTL_ENV / config env.
	explicitURL := ""
	if rootCmd.PersistentFlags().Changed("api-base-url") || os.Getenv("SPRINTCTL_API_BASE_URL") != "" {
		explicitURL = viper.GetString("api_base_url")
	}
	legacyURL := ""
	if viper.InConfig("api_base_url") {
		legacyURL = viper.GetString("api_base_url")
	}
	currentEnv = resolveEnvironment(explicitURL, viper.GetString("env"), legacyURL)
	auth.TokenScope = tokenScopeFor(currentEnv)
}

func Execute() {
	err := rootCmd.Execute()
	cobra.CheckErr(err)
}
