package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Brand Colors
	ColorPrimary   = lipgloss.Color("#00E5FF") // Electric Cyan
	ColorSecondary = lipgloss.Color("#7C4DFF") // Deep Violet
	ColorSuccess   = lipgloss.Color("#00E676") // Neon Green
	ColorWarning   = lipgloss.Color("#FFD600") // Warm Amber
	ColorSubtle    = lipgloss.Color("#757575") // Muted Slate
	ColorWhite     = lipgloss.Color("#FFFFFF")

	// Styles
	StyleTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary).
			MarginBottom(1)

	StyleSubtitle = lipgloss.NewStyle().
			Foreground(ColorSubtle).
			MarginBottom(1)

	StyleSelected = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true)

	StyleNormal = lipgloss.NewStyle().
			Foreground(ColorWhite)

	StyleBadge = lipgloss.NewStyle().
			Background(ColorSecondary).
			Foreground(ColorWhite).
			Padding(0, 1).
			Bold(true)

	StyleSuccess = lipgloss.NewStyle().
			Foreground(ColorSuccess).
			Bold(true)

	StyleBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorSecondary).
			Padding(1, 2)
)
