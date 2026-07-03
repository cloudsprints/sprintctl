package cmd

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/cloudsprints/sprintctl/internal/types"
)

// commandTimeout bounds each validation command so a hung command (e.g. a CLI
// waiting on credentials or network) can't stall the submission forever.
// Variable rather than const so tests can shorten it.
var commandTimeout = 120 * time.Second

// executeCommand runs a single validation command with a timeout and returns
// the captured result.
func executeCommand(command string) types.CLICommandResult {
	result := types.CLICommandResult{Command: command}

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

	return result
}

// commandFailed reports whether a validation command result should be treated
// as a failure: non-zero exit code, or output containing "validation failed"
// (from lab commands of the form `<check> || echo "validation failed"`).
func commandFailed(r types.CLICommandResult) bool {
	return r.ExitCode != 0 || strings.Contains(r.Stdout, "validation failed")
}

// runCommandsWithProgress executes commands sequentially while animating the
// gradient progress bar, advancing it as each command actually completes.
// It returns the collected results and whether the user aborted with ctrl+c.
// When stopOnFailure is true, execution stops at the first failing command.
func runCommandsWithProgress(commands []string, stopOnFailure bool) ([]types.CLICommandResult, bool) {
	if len(commands) == 0 {
		return nil, false
	}

	m := progressRunModel{
		progress:      progress.New(progress.WithGradient("#00f1ff", "#ff00ed")),
		commands:      commands,
		stopOnFailure: stopOnFailure,
	}

	final, err := tea.NewProgram(m).Run()
	if err != nil {
		// No usable terminal for the progress bar - run the commands plainly
		results := []types.CLICommandResult{}
		for _, command := range commands {
			result := executeCommand(command)
			results = append(results, result)
			if stopOnFailure && commandFailed(result) {
				break
			}
		}
		return results, false
	}

	fm := final.(progressRunModel)
	return fm.results, fm.aborted
}

const (
	progressPadding  = 2
	progressMaxWidth = 80
)

var progressLabelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#626262"))

type cmdDoneMsg struct{ result types.CLICommandResult }
type progressFinishedMsg struct{}

type progressRunModel struct {
	progress      progress.Model
	commands      []string
	results       []types.CLICommandResult
	stopOnFailure bool
	aborted       bool
	done          bool
	failed        bool
}

func execCommandCmd(command string) tea.Cmd {
	return func() tea.Msg {
		return cmdDoneMsg{result: executeCommand(command)}
	}
}

func (m progressRunModel) Init() tea.Cmd {
	return execCommandCmd(m.commands[0])
}

func (m progressRunModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			m.aborted = true
			return m, tea.Quit
		}
		return m, nil

	case tea.WindowSizeMsg:
		m.progress.Width = msg.Width - progressPadding*2 - 4
		if m.progress.Width > progressMaxWidth {
			m.progress.Width = progressMaxWidth
		}
		return m, nil

	case cmdDoneMsg:
		m.results = append(m.results, msg.result)
		percent := float64(len(m.results)) / float64(len(m.commands))
		cmds := []tea.Cmd{m.progress.SetPercent(percent)}

		if commandFailed(msg.result) {
			m.failed = true
		}

		if (m.failed && m.stopOnFailure) || len(m.results) == len(m.commands) {
			m.done = true
			// Let the bar animate to its final position before quitting
			cmds = append(cmds, tea.Tick(600*time.Millisecond, func(time.Time) tea.Msg {
				return progressFinishedMsg{}
			}))
		} else {
			cmds = append(cmds, execCommandCmd(m.commands[len(m.results)]))
		}
		return m, tea.Batch(cmds...)

	case progressFinishedMsg:
		return m, tea.Quit

	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		return m, cmd

	default:
		return m, nil
	}
}

func (m progressRunModel) View() string {
	pad := strings.Repeat(" ", progressPadding)

	label := fmt.Sprintf("Validating… (%d/%d)", len(m.results), len(m.commands))
	if m.done {
		if m.failed {
			label = fmt.Sprintf("Validation stopped (%d/%d)", len(m.results), len(m.commands))
		} else {
			label = fmt.Sprintf("Validation complete (%d/%d)", len(m.results), len(m.commands))
		}
	}

	return "\n" +
		pad + m.progress.View() + "\n\n" +
		pad + progressLabelStyle.Render(label) + "\n"
}
