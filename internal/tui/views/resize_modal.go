package views

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TierOption represents a machine tier
type TierOption struct {
	Tier    string
	Price   string
	VCPU    int
	RAM     string
	Disk    string
	Details string
}

var tierOptions = []TierOption{
	{
		Tier:    "starter",
		Price:   "$9/mo",
		VCPU:    2,
		RAM:     "4 GB",
		Disk:    "40 GB",
		Details: "Perfect for learning and light development",
	},
	{
		Tier:    "builder",
		Price:   "$25/mo",
		VCPU:    4,
		RAM:     "8 GB",
		Disk:    "80 GB",
		Details: "Best for active development and testing",
	},
	{
		Tier:    "pro",
		Price:   "$59/mo",
		VCPU:    8,
		RAM:     "16 GB",
		Disk:    "160 GB",
		Details: "Maximum performance for production work",
	},
}

// ResizeModalModel allows selecting a new tier
type ResizeModalModel struct {
	currentTier string
	selectedIdx int
	width       int
	height      int
}

// ResizeResponse is sent when user selects a tier
type ResizeResponse struct {
	NewTier string
	Confirm bool
}

// NewResizeModal creates a resize dialog
func NewResizeModal(currentTier string) *ResizeModalModel {
	// Find current tier index
	selectedIdx := 0
	for i, opt := range tierOptions {
		if opt.Tier == currentTier {
			selectedIdx = i
			break
		}
	}

	return &ResizeModalModel{
		currentTier: currentTier,
		selectedIdx: selectedIdx,
	}
}

func (m *ResizeModalModel) Init() tea.Cmd {
	return nil
}

func (m *ResizeModalModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.selectedIdx > 0 {
				m.selectedIdx--
			}
		case "down", "j":
			if m.selectedIdx < len(tierOptions)-1 {
				m.selectedIdx++
			}
		case "enter":
			return m, func() tea.Msg {
				return ResizeResponse{
					NewTier: tierOptions[m.selectedIdx].Tier,
					Confirm: true,
				}
			}
		case "esc":
			return m, func() tea.Msg {
				return ResizeResponse{Confirm: false}
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

func (m *ResizeModalModel) View() string {
	var sb strings.Builder

	sb.WriteString(lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorFg).
		Render("📦 Resize Instance"))
	sb.WriteString("\n")
	sb.WriteString(StyleHint.Render("Current tier: " + m.currentTier))
	sb.WriteString("\n\n")

	for i, opt := range tierOptions {
		isSelected := i == m.selectedIdx
		isCurrentTier := opt.Tier == m.currentTier

		var style lipgloss.Style
		if isSelected {
			style = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#000000")).
				Background(lipgloss.Color("#87CEEB")).
				Bold(true).
				Padding(1, 2)
		} else if isCurrentTier {
			style = lipgloss.NewStyle().
				Foreground(ColorWarning).
				Padding(1, 2)
		} else {
			style = lipgloss.NewStyle().
				Foreground(ColorFg).
				Padding(1, 2)
		}

		prefix := "  "
		if isSelected {
			prefix = "▶ "
		}

		line := prefix + opt.Tier + " (" + opt.Price + ")"
		if isCurrentTier {
			line += " [current]"
		}

		sb.WriteString(style.Render(line))
		sb.WriteString("\n")

		if isSelected {
			vcpuLabel := "vCPU"
			if opt.VCPU > 1 {
				vcpuLabel = "vCPUs"
			}
			specLine := fmt.Sprintf("%d %s • %s RAM • %s Disk", opt.VCPU, vcpuLabel, opt.RAM, opt.Disk)
			sb.WriteString(StyleHint.Render(opt.Details))
			sb.WriteString("\n")
			sb.WriteString(StyleHint.Render(specLine))
			sb.WriteString("\n")
		}
	}

	sb.WriteString("\n")
	sb.WriteString(StyleWarning.Render("⚠️  Resize requires a reboot (~2 minutes)"))
	sb.WriteString("\n\n")
	sb.WriteString("[↑↓] Select  [Enter] Confirm  [Esc] Cancel")

	content := StyleBorderBox.Render(sb.String())

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		content,
	)
}
