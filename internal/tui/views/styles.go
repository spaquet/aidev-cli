package views

import "github.com/charmbracelet/lipgloss"

// Colors used in views
var (
	ColorFg      = lipgloss.Color("#f0f0f0")
	ColorDimmed  = lipgloss.Color("#808080")
	ColorError   = lipgloss.Color("#FF6B6B")
	ColorSuccess = lipgloss.Color("#51CF66")
	ColorWarning = lipgloss.Color("#FFD700")
	ColorBorder  = lipgloss.Color("#404040")
)

// Styles for login view
var (
	StyleBorderBox = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorBorder).
		Padding(1, 2)

	StyleLabel = lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorFg)

	StyleHint = lipgloss.NewStyle().
		Foreground(ColorDimmed).
		Italic(true)

	StyleError = lipgloss.NewStyle().
		Foreground(ColorError).
		Bold(true)

	StyleSuccess = lipgloss.NewStyle().
		Foreground(ColorSuccess).
		Bold(true)

	StyleWarning = lipgloss.NewStyle().
		Foreground(ColorWarning).
		Bold(true)
)
