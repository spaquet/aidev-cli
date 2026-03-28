package components

import "github.com/charmbracelet/lipgloss"

// StatusColor returns the color for an instance status
func StatusColor(status string) lipgloss.Color {
	switch status {
	case "running":
		return lipgloss.Color("#90EE90") // Green
	case "stopped":
		return lipgloss.Color("#A9A9A9") // Gray
	case "starting", "stopping", "rebooting", "updating":
		return lipgloss.Color("#FFD700") // Yellow
	case "deleted":
		return lipgloss.Color("#FF6B6B") // Red
	default:
		return lipgloss.Color("#808080") // Gray
	}
}

// StatusBadge returns a formatted status badge
func StatusBadge(status string) string {
	symbol := "●"
	color := StatusColor(status)

	return lipgloss.NewStyle().
		Foreground(color).
		Bold(true).
		Render(symbol + " " + status)
}

// TierColor returns the background color for a tier
func TierColor(tier string) lipgloss.Color {
	switch tier {
	case "builder":
		return lipgloss.Color("#9370DB") // Purple
	case "pro":
		return lipgloss.Color("#FFB347") // Gold
	default:
		return lipgloss.Color("#6495ED") // Blue (starter)
	}
}

// TierBadge returns a formatted tier badge
func TierBadge(tier string) string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#000000")).
		Background(TierColor(tier)).
		Bold(true).
		Padding(0, 1).
		Render(tier)
}

// DiskBar returns a simple ASCII bar showing disk usage
func DiskBar(used, total int) string {
	if total == 0 {
		return "─────"
	}

	barWidth := 5
	filled := (used * barWidth) / total
	if filled > barWidth {
		filled = barWidth
	}

	bar := ""
	for i := 0; i < filled; i++ {
		bar += "█"
	}
	for i := filled; i < barWidth; i++ {
		bar += "░"
	}

	// Color based on usage percentage
	pct := (used * 100) / total
	var color lipgloss.Color
	if pct > 80 {
		color = lipgloss.Color("#FF6B6B") // Red
	} else if pct > 60 {
		color = lipgloss.Color("#FFD700") // Yellow
	} else {
		color = lipgloss.Color("#51CF66") // Green
	}

	return lipgloss.NewStyle().
		Foreground(color).
		Render(bar)
}
