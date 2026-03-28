package views

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	"github.com/aidev/cli/internal/api"
	"github.com/aidev/cli/internal/models"
	"github.com/aidev/cli/internal/tui/components"
)

// InstanceListModel shows a table of instances
type InstanceListModel struct {
	client    *api.Client
	table     table.Model
	instances []models.Instance
	loading   bool
	errorMsg  string
	width     int
	height    int
	lastSync  time.Time
}

type instancesLoadedMsg struct {
	instances []models.Instance
	err       error
}

// NewInstanceListModel creates a new instance list view
func NewInstanceListModel(client *api.Client, width, height int) *InstanceListModel {
	columns := []table.Column{
		{Title: "NAME", Width: 20},
		{Title: "STATUS", Width: 12},
		{Title: "TIER", Width: 10},
		{Title: "REGION", Width: 12},
		{Title: "DISK", Width: 12},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows([]table.Row{}),
		table.WithFocused(true),
		table.WithHeight(height - 5),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		Foreground(lipgloss.Color("#f0f0f0")).
		Bold(true).
		Padding(0, 1)

	s.Selected = s.Selected.
		Foreground(lipgloss.Color("#000000")).
		Background(lipgloss.Color("#87CEEB")).
		Bold(true).
		Padding(0, 1)

	t.SetStyles(s)

	m := &InstanceListModel{
		client:   client,
		table:    t,
		width:    width,
		height:   height,
		loading:  true,
		lastSync: time.Now(),
	}

	return m
}

func (m *InstanceListModel) Init() tea.Cmd {
	return m.loadInstances()
}

func (m *InstanceListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			m.table.MoveUp(1)
		case "down", "j":
			m.table.MoveDown(1)
		case "ctrl+r":
			// Force refresh
			m.loading = true
			return m, m.loadInstances()
		}

	case instancesLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.errorMsg = fmt.Sprintf("Failed to load instances: %v", msg.err)
			return m, nil
		}

		m.instances = msg.instances
		m.errorMsg = ""
		m.lastSync = time.Now()
		m.updateTableRows()
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.table.SetHeight(msg.Height - 5)
	}

	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m *InstanceListModel) View() string {
	var sb strings.Builder

	// Header
	header := lipgloss.NewStyle().
		Foreground(ColorFg).
		Bold(true).
		Render(fmt.Sprintf("📦 Instances (%d)", len(m.instances)))

	if m.loading {
		header += " ⟳ Loading..."
	} else if m.lastSync.Year() > 1 {
		syncTime := time.Since(m.lastSync)
		header += fmt.Sprintf(" (synced %v ago)", syncTime.Round(time.Second))
	}

	sb.WriteString(header)
	sb.WriteString("\n\n")

	// Error message
	if m.errorMsg != "" {
		sb.WriteString(StyleError.Render("❌ " + m.errorMsg))
		sb.WriteString("\n\n")
	}

	// Table
	if len(m.instances) == 0 {
		sb.WriteString(StyleHint.Render("No instances found. Create one with 'aidev instances create'"))
	} else {
		sb.WriteString(m.table.View())
	}

	return sb.String()
}

func (m *InstanceListModel) SelectedInstance() *models.Instance {
	if len(m.instances) == 0 {
		return nil
	}

	cursorRow := m.table.Cursor()
	if cursorRow >= len(m.instances) {
		return nil
	}

	return &m.instances[cursorRow]
}

// Private methods

func (m *InstanceListModel) updateTableRows() {
	rows := make([]table.Row, len(m.instances))

	for i, inst := range m.instances {
		statusBadge := components.StatusBadge(inst.Status)
		tierBadge := components.TierBadge(inst.Tier)
		diskBar := components.DiskBar(inst.DiskUsedGB, inst.DiskGB)
		diskStr := fmt.Sprintf("%s %d/%d GB", diskBar, inst.DiskUsedGB, inst.DiskGB)

		rows[i] = table.Row{
			inst.Name,
			statusBadge,
			tierBadge,
			inst.Region,
			diskStr,
		}
	}

	m.table.SetRows(rows)
}

func (m *InstanceListModel) loadInstances() tea.Cmd {
	return func() tea.Msg {
		resp, err := m.client.GetInstances()
		if err != nil {
			return instancesLoadedMsg{err: err}
		}

		return instancesLoadedMsg{instances: resp.Instances}
	}
}
