package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/creativeprojects/go-selfupdate"
	"github.com/spf13/cobra"
)

func update(version string) error {
	latest, found, err := selfupdate.DetectLatest(context.Background(), selfupdate.ParseSlug("cloudsprints/sprintctl"))
	if err != nil {
		return fmt.Errorf("error occurred while detecting version: %w", err)
	}
	if !found {
		return fmt.Errorf("latest version for %s/%s could not be found from github repository", runtime.GOOS, runtime.GOARCH)
	}

	if latest.LessOrEqual(version) {
		fmt.Printf("✅ sprintctl is already up to date (version %s)\n", version)
		return nil
	}

	exe, err := selfupdate.ExecutablePath()
	if err != nil {
		return errors.New("could not locate executable path")
	}
	if err := selfupdate.UpdateTo(context.Background(), latest.AssetURL, latest.AssetName, exe); err != nil {
		if errors.Is(err, os.ErrPermission) {
			return fmt.Errorf("permission denied writing to %s.\n"+
				"sprintctl is installed in a directory your user can't write to. Re-run the update with elevated privileges:\n\n"+
				"    sudo sprintctl update\n",
				filepath.Dir(exe))
		}
		return fmt.Errorf("error occurred while updating binary: %w", err)
	}
	fmt.Printf("🎉 Successfully updated sprintctl to version %s\n", latest.Version())
	return nil
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update sprintctl to the latest version",
	Run: func(cmd *cobra.Command, args []string) {
		if err := update(Version); err != nil {
			fmt.Println(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
