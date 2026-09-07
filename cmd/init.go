package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/cloudsprints/sprintctl/internal/mtcapi"
	"github.com/cloudsprints/sprintctl/internal/types"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:     "init [lesson-token]",
	Short:   "Initialize a lab environment",
	Args:    cobra.MaximumNArgs(1),
	Example: "sprintctl init                        # auto-detect active lab\nsprintctl init cm4ppz694200blze51ts1234  # explicit lesson token\nsprintctl init --admin cmj3wql250016ru5xw0vifo2b\nsprintctl init --project cmj3abc123      # admin only, pulls all labs",
	Run: func(cmd *cobra.Command, args []string) {
		publicOnly, _ := cmd.Flags().GetBool("public-only")
		adminMode, _ := cmd.Flags().GetBool("admin")
		projectID, _ := cmd.Flags().GetString("project")

		apiClient := mtcapi.New(apiBaseURL())

		// Project mode: download all labs from a project
		if projectID != "" {
			runProjectInit(apiClient, projectID)
			return
		}

		// Single lab mode - auto-detect or use provided token
		if len(args) == 0 && !adminMode {
			token, source, activeLesson, err := resolveLessonToken(args, apiClient)
			if err != nil {
				fmt.Println("Error fetching active lab:", err)
				fmt.Println("\nPlease open a lab in the UI first, or provide a lesson token explicitly:")
				fmt.Println("  sprintctl init <lesson-token>")
				os.Exit(1)
			}
			args = []string{token}
			printDetectionBanner(source, activeLesson)
		}

		if len(args) == 0 {
			fmt.Println("Error: lesson-token is required (or use --project)")
			fmt.Println("Usage: sprintctl init <lesson-token>")
			fmt.Println("       sprintctl init --project <project-id>")
			os.Exit(1)
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
				os.Exit(1)
			}
			labTitle = lessonInfo.Title
			fmt.Printf("Initializing lab: %s\n", labTitle)

			// Get all files including solutions via admin endpoint
			var filesErr error
			files, filesErr = apiClient.GetAdminLabFiles(lessonToken)
			if filesErr != nil {
				fmt.Printf("Error getting lab files: %s\n", filesErr)
				os.Exit(1)
			}
		} else {
			// Normal mode: use user_lesson_id
			fmt.Println("Fetching lab information...")
			labInfo, err := apiClient.GetLabInfo(lessonToken)
			if err != nil {
				fmt.Printf("Error getting lab information: %s\n", err)
				os.Exit(1)
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
				os.Exit(1)
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
				os.Exit(1)
			}
		}

		// Download files
		fmt.Printf("Downloading %d files...\n", len(files))

		failed := downloadLabFiles(files, labDir)
		if failed > 0 {
			fmt.Printf("\n%d of %d files failed to download.\n", failed, len(files))
			fmt.Println("Re-run 'sprintctl init' to retry.")
			os.Exit(1)
		}

		fmt.Printf("\nLab initialized successfully in %s\n", labDir)
		fmt.Println("You can now cd into the directory and start working on the lab.")
	},
}

// downloadLabFiles downloads files concurrently into labDir and returns the
// number of failures. Files under "public/" are extracted to the lab root.
func downloadLabFiles(files []types.LabFile, labDir string) int {
	const maxConcurrent = 6

	var wg sync.WaitGroup
	sem := make(chan struct{}, maxConcurrent)
	var mu sync.Mutex
	failed := 0

	for _, file := range files {
		targetPath := strings.TrimPrefix(file.Path, "public/")
		filePath := filepath.Join(labDir, targetPath)

		wg.Add(1)
		go func(file types.LabFile, filePath string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			err := os.MkdirAll(filepath.Dir(filePath), 0755)
			if err == nil {
				err = downloadFile(file.URL, filePath)
			}

			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				failed++
				fmt.Printf("✗ %s: %s\n", file.Path, err)
			} else {
				fmt.Printf("✓ %s\n", file.Path)
			}
		}(file, filePath)
	}
	wg.Wait()

	return failed
}

func downloadFile(url, filePath string) error {
	// Get the data before touching the destination so a failed request
	// doesn't leave an empty file behind
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Create the file
	out, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
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
