package views

import (
	"fmt"
	"strconv"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// ForwardModalModel allows user to set up port forwarding
type ForwardModalModel struct {
	localPort  textinput.Model
	remotePort textinput.Model
	focusIdx   int
	width      int
	height     int
}

// ForwardResponse is sent when user confirms port forwarding
type ForwardResponse struct {
	LocalPort  int
	RemotePort int
	Confirm    bool
}

// NewForwardModal creates the port forwarding modal
func NewForwardModal(defaultPort int) *ForwardModalModel {
	localInput := textinput.New()
	localInput.Placeholder = "3000"
	localInput.SetValue(fmt.Sprintf("%d", defaultPort))
	localInput.CharLimit = 5
	localInput.Focus()

	remoteInput := textinput.New()
	remoteInput.Placeholder = "3000"
	remoteInput.SetValue(fmt.Sprintf("%d", defaultPort))
	remoteInput.CharLimit = 5

	return &ForwardModalModel{
		localPort:  localInput,
		remotePort: remoteInput,
		focusIdx:   0,
	}
}

func (m *ForwardModalModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m *ForwardModalModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "tab":
			m.focusIdx = (m.focusIdx + 1) % 2
			m.updateFocus()
		case "shift+tab":
			m.focusIdx = (m.focusIdx - 1 + 2) % 2
			m.updateFocus()
		case "enter":
			// Validate and confirm
			localPort, err1 := strconv.Atoi(m.localPort.Value())
			remotePort, err2 := strconv.Atoi(m.remotePort.Value())

			if err1 != nil || err2 != nil || localPort < 1024 || localPort > 65535 || remotePort < 1 || remotePort > 65535 {
				return m, nil // Invalid input, stay on modal
			}

			return m, func() tea.Msg {
				return ForwardResponse{
					LocalPort:  localPort,
					RemotePort: remotePort,
					Confirm:    true,
				}
			}

		case "esc":
			return m, func() tea.Msg {
				return ForwardResponse{Confirm: false}
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	// Update focused field
	if m.focusIdx == 0 {
		var cmd tea.Cmd
		m.localPort, cmd = m.localPort.Update(msg)
		return m, cmd
	} else {
		var cmd tea.Cmd
		m.remotePort, cmd = m.remotePort.Update(msg)
		return m, cmd
	}
}

func (m *ForwardModalModel) View() tea.View {
	var sb strings.Builder

	sb.WriteString(lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorFg).
		Render("Local Port Forwarding"))
	sb.WriteString("\n\n")

	sb.WriteString(StyleLabel.Render("Local port (on your machine):"))
	sb.WriteString("\n")
	sb.WriteString(m.localPort.View())
	sb.WriteString("\n\n")

	sb.WriteString(StyleLabel.Render("Remote port (on VM):"))
	sb.WriteString("\n")
	sb.WriteString(m.remotePort.View())
	sb.WriteString("\n\n")

	sb.WriteString(StyleHint.Render("Forward localhost:LOCAL → VM:REMOTE"))
	sb.WriteString("\n\n")

	sb.WriteString(StyleWarning.Render("Port forwarding runs in background"))
	sb.WriteString("\n")
	sb.WriteString(StyleHint.Render("Press [F] to stop forwarding"))
	sb.WriteString("\n\n")

	sb.WriteString("[Tab] Next field  [Enter] Forward  [Esc] Cancel")

	content := StyleBorderBox.Render(sb.String())

	return tea.NewView(lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		content,
	))
}

func (m *ForwardModalModel) updateFocus() {
	m.localPort.Blur()
	m.remotePort.Blur()

	if m.focusIdx == 0 {
		m.localPort.Focus()
	} else {
		m.remotePort.Focus()
	}
}
