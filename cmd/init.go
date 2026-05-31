package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/morethancertified/sprintctl/internal/mtcapi"
	"github.com/morethancertified/sprintctl/internal/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var initCmd = &cobra.Command{
	Use:     "init <lesson-token>",
	Short:   "Initialize a lab environment",
	Args:    cobra.MaximumNArgs(1),
	Example: "sprintctl init cm4ppz694200blze51ts1234\nsprintctl init --admin cmj3wql250016ru5xw0vifo2b\nsprintctl init --project cmj3abc123  # admin only, pulls all labs",
	Run: func(cmd *cobra.Command, args []string) {
		publicOnly, _ := cmd.Flags().GetBool("public-only")
		adminMode, _ := cmd.Flags().GetBool("admin")
		projectID, _ := cmd.Flags().GetString("project")

		apiClient := mtcapi.New(viper.GetString("api_base_url"))

		// Project mode: download all labs from a project
		if projectID != "" {
			runProjectInit(apiClient, projectID)
			return
		}

		// Single lab mode requires a lesson token
		if len(args) == 0 {
			fmt.Println("Error: lesson-token is required (or use --project)")
			fmt.Println("Usage: sprintctl init <lesson-token>")
			fmt.Println("       sprintctl init --project <project-id>")
			return
		}

		lessonToken := args[0]

		var labTitle string
		var files []types.LabFile

		if adminMode {
			// Admin mode: use lesson_id directly with admin endpoints
			fmt.Println("🔐 Admin mode: fetching lab files directly...")

			// Get lesson info via admin endpoint
			lessonInfo, err := apiClient.GetAdminLessonInfo(lessonToken)
			if err != nil {
				fmt.Printf("Error getting lesson information: %s\n", err)
				return
			}
			labTitle = lessonInfo.Title
			fmt.Printf("Initializing lab: %s\n", labTitle)

			// Get all files including solutions via admin endpoint
			var filesErr error
			files, filesErr = apiClient.GetAdminLabFiles(lessonToken)
			if filesErr != nil {
				fmt.Printf("Error getting lab files: %s\n", filesErr)
				return
			}
		} else {
			// Normal mode: use user_lesson_id
			fmt.Println("Fetching lab information...")
			labInfo, err := apiClient.GetLabInfo(lessonToken)
			if err != nil {
				fmt.Printf("Error getting lab information: %s\n", err)
				return
			}
			labTitle = labInfo.Title
			fmt.Printf("Initializing lab: %s\n", labTitle)

			// Get file listing
			fmt.Println("Fetching lab files...")
			var filesErr error
			if publicOnly {
				files, filesErr = apiClient.GetLabPublicFiles(lessonToken)
			} else {
				files, filesErr = apiClient.GetLabFiles(lessonToken)
			}
			if filesErr != nil {
				fmt.Printf("Error getting lab files: %s\n", filesErr)
				return
			}
		}

		// Create lab directory
		labDir := sanitizeDirectoryName(labTitle)
		if dirFlag, _ := cmd.Flags().GetString("dir"); dirFlag != "" {
			labDir = dirFlag
		}

		// Clean up the directory name
		labDir = filepath.Clean(labDir)

		// Create the directory if it doesn't exist
		if _, err := os.Stat(labDir); os.IsNotExist(err) {
			if err := os.MkdirAll(labDir, 0755); err != nil {
				fmt.Printf("Error creating directory %s: %s\n", labDir, err)
				return
			}
		}

		// Download files
		fmt.Printf("Downloading %d files...\n", len(files))

		for i, file := range files {
			// Determine the target path
			var filePath string

			// If it's a public file, extract it to the root of the lab directory
			if strings.HasPrefix(file.Path, "public/") {
				// Remove the "public/" prefix
				targetPath := strings.TrimPrefix(file.Path, "public/")
				filePath = filepath.Join(labDir, targetPath)
				fmt.Printf("[%d/%d] Downloading %s to %s...\n", i+1, len(files), file.Path, targetPath)
			} else {
				// Keep the original path for other files
				filePath = filepath.Join(labDir, file.Path)
				fmt.Printf("[%d/%d] Downloading %s...\n", i+1, len(files), file.Path)
			}

			// Create subdirectories if needed
			fileDir := filepath.Dir(filePath)
			if err := os.MkdirAll(fileDir, 0755); err != nil {
				fmt.Printf("Error creating directory %s: %s\n", fileDir, err)
				continue
			}

			// Download file
			if err := downloadFile(file.URL, filePath); err != nil {
				fmt.Printf("Error downloading %s: %s\n", file.Path, err)
				continue
			}
		}

		// No need to save lab metadata

		fmt.Printf("\nLab initialized successfully in %s\n", labDir)
		fmt.Println("You can now cd into the directory and start working on the lab.")
	},
}

func downloadFile(url, filePath string) error {
	// Create the file
	out, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Writer the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

// sanitizeDirectoryName cleans up a string to be used as a directory name
func sanitizeDirectoryName(name string) string {
	// Remove quotes
	name = strings.Trim(name, "\"'`")

	// Replace problematic characters with underscores
	replacer := strings.NewReplacer(
		"/", "_",
		"\\", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
		" ", "_", // Replace spaces with underscores
	)

	// Ensure the name doesn't have any remaining problematic characters
	sanitized := replacer.Replace(name)

	// Convert to lowercase
	sanitized = strings.ToLower(sanitized)

	// If the name is empty after sanitization, use a default name
	if sanitized == "" {
		return "lab"
	}

	return sanitized
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().BoolP("public-only", "p", false, "Download only public files")
	initCmd.Flags().StringP("dir", "d", "", "Directory to initialize the lab in (defaults to lab title)")
	initCmd.Flags().BoolP("admin", "a", false, "Admin mode: use lesson_id directly (includes solution files)")
	initCmd.Flags().StringP("project", "P", "", "Download all labs from a project (admin only, includes solutions)")
}

// runProjectInit downloads all labs from a project
func runProjectInit(apiClient *mtcapi.MtcApiClient, projectID string) {
	fmt.Println("🔐 Admin mode: fetching all labs from project...")

	projectLabs, err := apiClient.GetAdminProjectLabs(projectID)
	if err != nil {
		fmt.Printf("Error getting project labs: %s\n", err)
		return
	}

	fmt.Printf("\n📦 Project: %s\n", projectLabs.ProjectTitle)
	fmt.Printf("   Labs: %d\n\n", len(projectLabs.Lessons))

	if len(projectLabs.Lessons) == 0 {
		fmt.Println("No labs found in this project.")
		return
	}

	// Create project directory
	projectDir := sanitizeDirectoryName(projectLabs.ProjectTitle)
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		fmt.Printf("Error creating project directory: %s\n", err)
		return
	}

	// Download each lab
	for _, lesson := range projectLabs.Lessons {
		fmt.Printf("📚 %s\n", lesson.Title)

		if len(lesson.Files) == 0 {
			fmt.Println("   (no files)")
			continue
		}

		// Create lab subdirectory
		labDir := filepath.Join(projectDir, sanitizeDirectoryName(lesson.Title))
		if err := os.MkdirAll(labDir, 0755); err != nil {
			fmt.Printf("   Error creating directory: %s\n", err)
			continue
		}

		// Download files for this lab
		for _, file := range lesson.Files {
			var filePath string
			if strings.HasPrefix(file.Path, "public/") {
				filePath = filepath.Join(labDir, strings.TrimPrefix(file.Path, "public/"))
			} else {
				filePath = filepath.Join(labDir, file.Path)
			}

			// Create parent directories
			if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
				fmt.Printf("   Error creating directory for %s: %s\n", file.Path, err)
				continue
			}

			// Download file
			resp, err := http.Get(file.URL)
			if err != nil {
				fmt.Printf("   Error downloading %s: %s\n", file.Path, err)
				continue
			}

			out, err := os.Create(filePath)
			if err != nil {
				resp.Body.Close()
				fmt.Printf("   Error creating %s: %s\n", file.Path, err)
				continue
			}

			_, err = io.Copy(out, resp.Body)
			resp.Body.Close()
			out.Close()

			if err != nil {
				fmt.Printf("   Error writing %s: %s\n", file.Path, err)
				continue
			}

			fmt.Printf("   ✓ %s\n", file.Path)
		}
	}

	fmt.Printf("\n✅ Project initialized in %s/\n", projectDir)
}
