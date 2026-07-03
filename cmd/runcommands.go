package cmd

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/cloudsprints/sprintctl/internal/types"
)

// commandTimeout bounds each validation command so a hung command (e.g. a CLI
// waiting on credentials or network) can't stall the submission forever.
const commandTimeout = 120 * time.Second

// runValidationCommand executes a single validation command with a live
// spinner and a timeout, returning the captured result.
func runValidationCommand(command string, idx, total int) types.CLICommandResult {
	result := types.CLICommandResult{Command: command}

	label := fmt.Sprintf("[%d/%d] Running validation…", idx+1, total)
	stop := startSpinner(label)

	ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sh", "-c", "LANG=en_US.UTF-8 "+command)
	b, err := cmd.Output()

	if ctx.Err() == context.DeadlineExceeded {
		result.ExitCode = 124
		result.Stderr = fmt.Sprintf("command timed out after %s", commandTimeout)
	} else if ee, ok := err.(*exec.ExitError); ok {
		result.ExitCode = ee.ExitCode()
		result.Stderr = strings.TrimRight(string(ee.Stderr), "\n\t\r")
	} else if err != nil {
		result.ExitCode = -69
	} else {
		result.Stdout = strings.TrimRight(string(b), "\n\t\r")
	}

	stop(result.ExitCode == 0)
	return result
}

// startSpinner renders an animated spinner next to label until the returned
// stop function is called, which replaces it with a ✓ or ✗ marker.
func startSpinner(label string) func(ok bool) {
	frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	done := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		i := 0
		fmt.Printf("\r\033[2K%s %s", frames[0], label)
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				i++
				fmt.Printf("\r\033[2K%s %s", frames[i%len(frames)], label)
			}
		}
	}()

	return func(ok bool) {
		close(done)
		wg.Wait()
		icon := "✓"
		if !ok {
			icon = "✗"
		}
		fmt.Printf("\r\033[2K%s %s\n", icon, label)
	}
}
