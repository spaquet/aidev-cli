package views

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/aidev/cli/internal/api"
	"github.com/aidev/cli/internal/models"
)

// MainModel is the main view (instance list + detail)
type MainModel struct {
	client       *api.Client
	user         *models.User
	list         *InstanceListModel
	width        int
	height       int
	showHelp     bool
}

// NewMainModel creates the main view
func NewMainModel(client *api.Client, user *models.User, width, height int) *MainModel {
	return &MainModel{
		client:   client,
		user:     user,
		list:     NewInstanceListModel(client, width, height),
		width:    width,
		height:   height,
		showHelp: true,
	}
}

func (m *MainModel) Init() tea.Cmd {
	return m.list.Init()
}

func (m *MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "?":
			m.showHelp = !m.showHelp
			return m, nil

		default:
			// Pass to list
			model, cmd := m.list.Update(msg)
			if listModel, ok := model.(*InstanceListModel); ok {
				m.list = listModel
			}
			return m, cmd
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		model, cmd := m.list.Update(msg)
		if listModel, ok := model.(*InstanceListModel); ok {
			m.list = listModel
		}
		return m, cmd

	default:
		// Pass to list
		model, cmd := m.list.Update(msg)
		if listModel, ok := model.(*InstanceListModel); ok {
			m.list = listModel
		}
		return m, cmd
	}
}

func (m *MainModel) View() string {
	var sb strings.Builder

	// Top bar
	topBar := lipgloss.NewStyle().
		Foreground(ColorFg).
		Background(lipgloss.Color("#1e1e1e")).
		Padding(0, 1).
		Render("AIDev • " + m.user.Email)

	sb.WriteString(topBar)
	sb.WriteString("\n")

	// Content
	sb.WriteString(m.list.View())
	sb.WriteString("\n\n")

	// Footer
	footer := ""
	if m.showHelp {
		footer = "[↑↓] Navigate  [Enter] Connect  [c]onnect [f]orward [u]pdate [d]elete  [Ctrl+R] Refresh  [?] Help [q]uit"
	} else {
		footer = "[?] Show help  [q]uit"
	}

	footerStyle := lipgloss.NewStyle().
		Foreground(ColorDimmed).
		Padding(0, 1).
		Render(footer)

	sb.WriteString(footerStyle)

	return sb.String()
}
