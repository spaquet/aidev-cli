package tui

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

// Colors
var (
	ColorBg      = lipgloss.Color("#1e1e1e")
	ColorFg      = lipgloss.Color("#f0f0f0")
	ColorPrimary = lipgloss.Color("#87CEEB") // Sky blue
	ColorSuccess = lipgloss.Color("#51CF66") // Bright green
	ColorError   = lipgloss.Color("#FF6B6B") // Red
	ColorWarning = lipgloss.Color("#FFD700") // Yellow
	ColorDimmed  = lipgloss.Color("#808080") // Gray
	ColorBorder  = lipgloss.Color("#404040") // Dark gray

	ColorRunning  = lipgloss.Color("#90EE90") // Green
	ColorStopped  = lipgloss.Color("#A9A9A9") // Gray
	ColorStarting = lipgloss.Color("#FFD700") // Yellow

	ColorTierStarterBg = lipgloss.Color("#6495ED") // Blue
	ColorTierBuilderBg = lipgloss.Color("#9370DB") // Purple
	ColorTierProBg     = lipgloss.Color("#FFB347") // Gold
)

var _ color.Color = lipgloss.Color("")

// Styles
var (
	StyleTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorFg).
			Background(ColorBg).
			MarginBottom(1)

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

	StyleBorderBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorder).
			Padding(1, 2)

	StyleLink = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Underline(true)

	StyleFieldInput = lipgloss.NewStyle().
			Foreground(ColorFg).
			Background(ColorBg).
			Padding(0)
)

// Badge styles
func StyleStatusBadge(status string) lipgloss.Style {
	var c color.Color = ColorStopped
	switch status {
	case "running":
		c = ColorRunning
	case "starting", "stopping", "rebooting", "updating":
		c = ColorWarning
	}

	return lipgloss.NewStyle().
		Foreground(c).
		Bold(true)
}

func StyleTierBadge(tier string) lipgloss.Style {
	var bgColor color.Color = ColorTierStarterBg
	switch tier {
	case "builder":
		bgColor = ColorTierBuilderBg
	case "pro":
		bgColor = ColorTierProBg
	}

	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#000000")).
		Background(bgColor).
		Bold(true).
		Padding(0, 1)
}

// Responsive dimensions
func GetMainViewDimensions(width, height int) (listWidth, detailWidth, detailHeight int) {
	// Left pane takes 40%, right pane takes 60%
	// Reserve 2 for border
	listWidth = (width * 40 / 100) - 1
	detailWidth = (width * 60 / 100) - 1
	detailHeight = height - 3 // Account for header and footer

	return
}
