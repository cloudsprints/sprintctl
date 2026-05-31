package styles

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
)

// Color palette for consistent theming
var (
	// Primary colors
	PrimaryGreen    = lipgloss.Color("#25A065")
	PrimaryBlue     = lipgloss.Color("#1E3A8A")
	PrimaryPurple   = lipgloss.Color("#8B5CF6")
	
	// Status colors
	SuccessGreen    = lipgloss.Color("#10B981")
	ErrorRed        = lipgloss.Color("#EF4444")
	WarningYellow   = lipgloss.Color("#F59E0B")
	InfoBlue        = lipgloss.Color("#3B82F6")
	
	// Neutral colors
	White           = lipgloss.Color("#FFFFFF")
	LightGray       = lipgloss.Color("#F3F4F6")
	Gray            = lipgloss.Color("#6B7280")
	DarkGray        = lipgloss.Color("#374151")
	Black           = lipgloss.Color("#111827")
	
	// Accent colors
	Cyan            = lipgloss.Color("#06B6D4")
	Pink            = lipgloss.Color("#EC4899")
	Orange          = lipgloss.Color("#F97316")
)

// Header styles
var (
	// Main title style with solid background
	TitleStyle = lipgloss.NewStyle().
		Foreground(White).
		Background(PrimaryGreen).
		Padding(1, 3).
		Bold(true).
		Align(lipgloss.Center)

	// Compact logo style for CLI help
	LogoStyle = lipgloss.NewStyle().
		Foreground(Cyan).
		Bold(true)

	// Subtitle style
	SubtitleStyle = lipgloss.NewStyle().
		Foreground(Gray).
		Italic(true).
		Margin(0, 0, 0, 0)

	// Hint style for actionable tips
	HintStyle = lipgloss.NewStyle().
		Foreground(PrimaryGreen).
		Italic(true)

	// Section header style
	SectionHeaderStyle = lipgloss.NewStyle().
		Foreground(PrimaryBlue).
		Bold(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		BorderForeground(PrimaryBlue).
		Padding(0, 0, 0, 1).
		Margin(1, 0, 1, 0)
)

// Status styles
var (
	SuccessStyle = lipgloss.NewStyle().
		Foreground(White).
		Background(SuccessGreen).
		Padding(0, 1).
		Bold(true)

	ErrorStyle = lipgloss.NewStyle().
		Foreground(White).
		Background(ErrorRed).
		Padding(0, 1).
		Bold(true)

	WarningStyle = lipgloss.NewStyle().
		Foreground(Black).
		Background(WarningYellow).
		Padding(0, 1).
		Bold(true)

	InfoStyle = lipgloss.NewStyle().
		Foreground(White).
		Background(InfoBlue).
		Padding(0, 1).
		Bold(true)
)

// Box styles
var (
	// Main content box
	BoxStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(PrimaryBlue).
		Padding(1, 2).
		Margin(1, 0)

	// Highlighted box for important content
	HighlightBoxStyle = lipgloss.NewStyle().
		Border(lipgloss.ThickBorder()).
		BorderForeground(PrimaryGreen).
		Background(DarkGray).
		Foreground(White).
		Padding(1, 2).
		Margin(1, 0)

	// Warning box
	WarningBoxStyle = lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(WarningYellow).
		Background(DarkGray).
		Foreground(White).
		Padding(1, 2).
		Margin(1, 0)

	// Error box
	ErrorBoxStyle = lipgloss.NewStyle().
		Border(lipgloss.ThickBorder()).
		BorderForeground(ErrorRed).
		Background(DarkGray).
		Foreground(White).
		Padding(1, 2).
		Margin(1, 0)
)

// List styles
var (
	ListItemStyle = lipgloss.NewStyle().
		PaddingLeft(2).
		Foreground(DarkGray)

	SelectedListItemStyle = lipgloss.NewStyle().
		PaddingLeft(1).
		Foreground(PrimaryBlue).
		Bold(true).
		Background(DarkGray)

	BulletStyle = lipgloss.NewStyle().
		Foreground(PrimaryGreen).
		Bold(true)
)

// Command styles
var (
	CommandStyle = lipgloss.NewStyle().
		Foreground(White).
		Background(DarkGray).
		Padding(0, 1).
		Italic(true)

	CodeBlockStyle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(Gray).
		Background(Black).
		Foreground(White).
		Padding(1, 2).
		Margin(1, 0)
)

// Progress styles
var (
	ProgressBarStyle = lipgloss.NewStyle().
		Background(Black).
		Height(1)

	ProgressFillStyle = lipgloss.NewStyle().
		Background(PrimaryGreen).
		Height(1)
)

// Interactive styles
var (
	ButtonStyle = lipgloss.NewStyle().
		Foreground(White).
		Background(PrimaryBlue).
		Padding(0, 2).
		Margin(0, 1).
		Bold(true)

	SelectedButtonStyle = lipgloss.NewStyle().
		Foreground(White).
		Background(PrimaryGreen).
		Padding(0, 2).
		Margin(0, 1).
		Bold(true).
		Border(lipgloss.NormalBorder()).
		BorderForeground(White)

	InputStyle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(PrimaryBlue).
		Padding(0, 1)
)

// Helper functions for common patterns
func StatusIcon(status string) string {
	switch status {
	case "COMPLETED":
		return "✅"
	case "FAILED":
		return "❌"
	case "IN_PROGRESS":
		return "🔄"
	case "PENDING":
		return "⏳"
	default:
		return "❓"
	}
}

func StatusText(status string) lipgloss.Style {
	switch status {
	case "COMPLETED":
		return SuccessStyle
	case "FAILED":
		return ErrorStyle
	case "IN_PROGRESS":
		return InfoStyle
	case "PENDING":
		return WarningStyle
	default:
		return lipgloss.NewStyle().Foreground(Gray)
	}
}

// Create a banner with ASCII art
func CreateBanner(text string) string {
	banner := lipgloss.NewStyle().
		Foreground(White).
		Background(PrimaryBlue).
		Padding(2, 4).
		Bold(true).
		Align(lipgloss.Center).
		Width(60)

	return banner.Render(text)
}

// Create a separator line
func Separator(width int) string {
	return lipgloss.NewStyle().
		Foreground(PrimaryBlue).
		Render(lipgloss.PlaceHorizontal(width, lipgloss.Center, "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
}

// Create a progress bar
func ProgressBar(current, total int, width int) string {
	if total == 0 {
		return ""
	}
	
	percentage := float64(current) / float64(total)
	filled := int(percentage * float64(width))
	
	bar := ""
	for i := 0; i < width; i++ {
		if i < filled {
			bar += "█"
		} else {
			bar += "░"
		}
	}
	
	progressStyle := lipgloss.NewStyle().
		Foreground(PrimaryGreen)
	
	emptyStyle := lipgloss.NewStyle().
		Foreground(Gray)
	
	filledPart := progressStyle.Render(bar[:filled])
	emptyPart := emptyStyle.Render(bar[filled:])
	
	return filledPart + emptyPart + fmt.Sprintf(" %d/%d (%.1f%%)", current, total, percentage*100)
}
