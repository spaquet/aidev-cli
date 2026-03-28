package views

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ConfirmDialogModel is a yes/no confirmation dialog
type ConfirmDialogModel struct {
	title   string
	message string
	width   int
	height  int
}

// ConfirmResponse is sent when user confirms or cancels
type ConfirmResponse struct {
	Confirmed bool
	Action    string // The action that was confirmed (e.g., "delete", "restart")
}

// NewConfirmDialog creates a confirmation dialog
func NewConfirmDialog(title, message string) *ConfirmDialogModel {
	return &ConfirmDialogModel{
		title:   title,
		message: message,
	}
}

func (m *ConfirmDialogModel) Init() tea.Cmd {
	return nil
}

func (m *ConfirmDialogModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "y", "enter":
			return m, func() tea.Msg {
				return ConfirmResponse{Confirmed: true}
			}
		case "n", "esc":
			return m, func() tea.Msg {
				return ConfirmResponse{Confirmed: false}
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

func (m *ConfirmDialogModel) View() string {
	var sb strings.Builder

	sb.WriteString(m.title)
	sb.WriteString("\n\n")
	sb.WriteString(m.message)
	sb.WriteString("\n\n")
	sb.WriteString("[y] Yes  [n] No  [Esc] Cancel")

	content := StyleBorderBox.Render(sb.String())

	// Center on screen
	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		content,
	)
}
