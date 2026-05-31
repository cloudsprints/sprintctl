package tui

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/morethancertified/sprintctl/internal/styles"
	"github.com/morethancertified/sprintctl/internal/types"
)

// Note: dynamic sizing is handled via WindowSizeMsg; no fixed widths

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(styles.PrimaryBlue).
			Bold(true).
			PaddingLeft(1)

	itemStyle = lipgloss.NewStyle().
			PaddingLeft(2)

	selectedItemStyle = lipgloss.NewStyle().
				Foreground(styles.PrimaryBlue).
				Bold(true)

	paginationStyle = list.DefaultStyles().PaginationStyle.
			PaddingLeft(2).
			Foreground(styles.Gray)

	quitTextStyle = lipgloss.NewStyle().
			Margin(1, 0, 1, 2).
			Foreground(styles.PrimaryGreen).
			Bold(true)
)

type item struct {
	task types.Task
}

func (i item) FilterValue() string { return i.task.Title }

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	statusIcon := styles.StatusIcon(i.task.Status)
	str := fmt.Sprintf("%s %s", statusIcon, i.task.Title)

	if index == m.Index() {
		str = "▶ " + str
		fmt.Fprint(w, selectedItemStyle.Render(str))
	} else {
		fmt.Fprint(w, itemStyle.Render(str))
	}
}

type model struct {
	list               list.Model
	viewport           viewport.Model
	quitting           bool
	tasks              []types.Task
	showingExplanation bool
	lessonToken        string
	resubmitting       bool
	width              int
	height             int
	topHeight          int
	bottomHeight       int
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Ensure minimum dimensions
		if m.width < 10 {
			m.width = 10
		}
		if m.height < 10 {
			m.height = 10
		}

		// Reserve space for title, separator and footer
		reservedHeight := 6 // title + separator + footer + padding lines
		availableHeight := m.height - reservedHeight
		if availableHeight < 6 {
			availableHeight = 6
		}

		// Dynamically size task list based on number of tasks
		// Each task takes 1 line, plus 3 lines for title/header/padding
		// The list component reserves ~2 lines for title internally
		taskCount := len(m.tasks)
		neededTaskHeight := taskCount + 3 // tasks + title + header padding

		// Cap task list height to leave at least 5 lines for viewport
		minViewportHeight := 5
		maxTaskHeight := availableHeight - minViewportHeight
		if maxTaskHeight < 3 {
			maxTaskHeight = 3
		}

		// Use the smaller of needed height or max allowed
		m.topHeight = neededTaskHeight
		if m.topHeight > maxTaskHeight {
			m.topHeight = maxTaskHeight
		}
		if m.topHeight < 3 {
			m.topHeight = 3
		}

		// Give all remaining space to viewport for AI explanation
		m.bottomHeight = availableHeight - m.topHeight
		if m.bottomHeight < 3 {
			m.bottomHeight = 3
		}

		// Set list dimensions
		listWidth := m.width
		if listWidth < 20 {
			listWidth = 20
		}
		m.list.SetWidth(listWidth)
		m.list.SetHeight(m.topHeight)

		// Set viewport dimensions
		viewportWidth := m.width - 2
		if viewportWidth < 20 {
			viewportWidth = 20
		}
		m.viewport.Width = viewportWidth
		m.viewport.Height = m.bottomHeight
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			// Toggle AI explanation display
			m.showingExplanation = !m.showingExplanation
			m.updateDetail()
			return m, nil

		case "r":
			// Resubmit lesson
			m.resubmitting = true
			return m, tea.Quit

		case "j", "down":
			// Always navigate task list - AI feedback updates automatically
			m.list.CursorDown()
			m.updateDetail()
			return m, nil

		case "k", "up":
			// Always navigate task list - AI feedback updates automatically
			m.list.CursorUp()
			m.updateDetail()
			return m, nil

		case "pgdown", "ctrl+d":
			if m.showingExplanation {
				m.viewport.HalfViewDown()
			}
			return m, nil

		case "pgup", "ctrl+u":
			if m.showingExplanation {
				m.viewport.HalfViewUp()
			}
			return m, nil

		case "g":
			if m.showingExplanation {
				m.viewport.GotoTop()
			}
			return m, nil

		case "G":
			if m.showingExplanation {
				m.viewport.GotoBottom()
			}
			return m, nil
		}
	}

	return m, nil
}

func (m *model) updateDetail() {
	if len(m.list.Items()) == 0 {
		m.viewport.SetContent("")
		m.viewport.GotoTop()
		return
	}

	selectedItem := m.list.SelectedItem()
	if selectedItem == nil {
		m.viewport.SetContent("")
		m.viewport.GotoTop()
		return
	}

	task := selectedItem.(item).task

	// Always clear viewport first and reset scroll position to prevent content remnants
	m.viewport.SetContent("")
	m.viewport.GotoTop()

	var content string

	if m.showingExplanation {
		if task.AiExplanation != "" {
			// Render markdown with glamour for proper code highlighting
			renderWidth := m.viewport.Width
			if renderWidth <= 0 {
				renderWidth = 80
			}
			renderer, err := glamour.NewTermRenderer(
				glamour.WithAutoStyle(),
				glamour.WithWordWrap(renderWidth-4),
			)
			if err == nil {
				renderedContent, err := renderer.Render(task.AiExplanation)
				if err == nil {
					content = lipgloss.NewStyle().Bold(true).Render("🤖 AI Feedback:") + "\n" + renderedContent
				} else {
					// Fallback to plain text if rendering fails
					content = lipgloss.NewStyle().Bold(true).Render("🤖 AI Feedback:") + "\n\n" + task.AiExplanation
				}
			} else {
				// Fallback to plain text if renderer creation fails
				content = lipgloss.NewStyle().Bold(true).Render("🤖 AI Feedback:") + "\n\n" + task.AiExplanation
			}
		} else {
			content = lipgloss.NewStyle().Foreground(styles.Gray).Render("No AI feedback available for this task.")
		}
	} else {
		if task.AiExplanation != "" {
			content = lipgloss.NewStyle().Foreground(styles.Gray).Italic(true).Render("Press Enter to view AI feedback...")
		} else {
			content = lipgloss.NewStyle().Foreground(styles.Gray).Render("No AI feedback available.")
		}
	}

	m.viewport.SetContent(content)
}

func (m model) View() string {
	if m.quitting {
		return quitTextStyle.Render("Thanks for using sprintctl!")
	}

	// Top: tasks list
	listView := m.list.View()

	// Bottom: AI feedback
	feedbackContent := m.viewport.View()

	// Simple separator line
	separatorWidth := m.width - 2
	if separatorWidth < 1 {
		separatorWidth = 1
	}
	separator := lipgloss.NewStyle().
		Foreground(styles.Gray).
		Render(strings.Repeat("─", separatorWidth))

	// Footer with instructions
	footerText := "[j/k] Navigate  [Enter] Toggle feedback  [r] Resubmit  [q] Quit"
	if m.showingExplanation {
		footerText = "[j/k] Navigate  [PgUp/PgDn] Scroll feedback  [Enter] Hide  [r] Resubmit  [q] Quit"
	}
	footer := lipgloss.NewStyle().
		Foreground(styles.Gray).
		Italic(true).
		Render(footerText)

	// Combine everything
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		listView,
		"",
		separator,
		"",
		feedbackContent,
		"",
		footer,
	)

	return content
}

func RunGradingReport(tasks []types.Task, lessonToken string) (bool, error) {
	items := make([]list.Item, len(tasks))
	for i, task := range tasks {
		items[i] = item{task: task}
	}

	const defaultWidth = 80
	// Initial height should accommodate all tasks plus title overhead
	// The list reserves ~2 lines for title, so we need tasks + 2
	initialHeight := len(items) + 4

	l := list.New(items, itemDelegate{}, defaultWidth, initialHeight)
	l.Title = "🚀 Grading Report"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	l.SetShowPagination(false) // Disable pagination dots since we want to show all tasks
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle

	m := model{
		list:        l,
		tasks:       tasks,
		lessonToken: lessonToken,
	}
	// Initial viewport; will be resized on WindowSizeMsg
	m.viewport = viewport.New(80, 10)
	m.updateDetail()

	p := tea.NewProgram(m, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return false, err
	}

	// Check if user wants to resubmit
	if finalModel, ok := finalModel.(model); ok {
		return finalModel.resubmitting, nil
	}

	return false, nil
}
