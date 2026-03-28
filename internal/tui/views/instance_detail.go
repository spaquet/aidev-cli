package views

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
	"github.com/aidev/cli/internal/models"
	"github.com/aidev/cli/internal/tui/components"
)

// InstanceDetailModel shows detailed info about the selected instance
type InstanceDetailModel struct {
	instance *models.Instance
	viewport viewport.Model
	width    int
	height   int
}

// NewInstanceDetailModel creates the detail pane
func NewInstanceDetailModel(width, height int) *InstanceDetailModel {
	vp := viewport.New(width, height)
	vp.HighPerformanceRendering = false

	return &InstanceDetailModel{
		viewport: vp,
		width:    width,
		height:   height,
	}
}

func (m *InstanceDetailModel) Init() tea.Cmd {
	return nil
}

func (m *InstanceDetailModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k", "pgup":
			m.viewport.LineUp(3)
		case "down", "j", "pgdn":
			m.viewport.LineDown(3)
		}
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m *InstanceDetailModel) View() string {
	if m.instance == nil {
		return StyleHint.Render("Select an instance to view details")
	}

	content := m.renderDetail()
	m.viewport.SetContent(content)
	return m.viewport.View()
}

func (m *InstanceDetailModel) SetInstance(inst *models.Instance) {
	m.instance = inst
}

// Private methods

func (m *InstanceDetailModel) renderDetail() string {
	if m.instance == nil {
		return ""
	}

	var sb strings.Builder
	inst := m.instance

	// Title
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorFg).
		MarginBottom(1).
		Render(inst.Name)
	sb.WriteString(title)
	sb.WriteString("\n")

	// Status and tier in one row
	statusBadge := components.StatusBadge(inst.Status)
	tierBadge := components.TierBadge(inst.Tier)
	sb.WriteString(statusBadge)
	sb.WriteString("  ")
	sb.WriteString(tierBadge)
	sb.WriteString("\n\n")

	// Metadata section
	sb.WriteString(renderSection("Instance Metadata", []string{
		fmt.Sprintf("ID: %s", inst.ID),
		fmt.Sprintf("Region: %s", inst.Region),
		fmt.Sprintf("Created: %s", inst.CreatedAt),
	}))

	// SSH Connection section
	sshCmd := fmt.Sprintf("ssh -i ~/.ssh/id_ed25519 %s@%s", inst.SSHUser, inst.SSHHost)
	if inst.SSHPort != 22 {
		sshCmd = fmt.Sprintf("ssh -i ~/.ssh/id_ed25519 -p %d %s@%s", inst.SSHPort, inst.SSHUser, inst.SSHHost)
	}

	sb.WriteString("\n")
	sb.WriteString(renderSection("SSH Connection", []string{
		fmt.Sprintf("Host: %s", inst.SSHHost),
		fmt.Sprintf("Port: %d", inst.SSHPort),
		fmt.Sprintf("User: %s", inst.SSHUser),
		fmt.Sprintf("Key: ~/.ssh/id_ed25519 (or your default key)"),
		"",
		"Command:",
		StyleLink.Render(sshCmd),
	}))

	// Disk section
	diskPct := 0
	if inst.DiskGB > 0 {
		diskPct = (inst.DiskUsedGB * 100) / inst.DiskGB
	}
	diskBar := components.DiskBar(inst.DiskUsedGB, inst.DiskGB)

	sb.WriteString("\n")
	sb.WriteString(renderSection("Storage", []string{
		fmt.Sprintf("Disk: %s %d/%d GB (%d%%)", diskBar, inst.DiskUsedGB, inst.DiskGB, diskPct),
	}))

	// Image section
	imageStr := inst.ImageVersion
	if inst.ImageUpdateAvailable {
		imageStr += " (update available)"
	}

	sb.WriteString("\n")
	sb.WriteString(renderSection("Image", []string{
		fmt.Sprintf("Version: %s", imageStr),
	}))

	// Tools section
	if len(inst.InstalledTools) > 0 {
		toolList := make([]string, len(inst.InstalledTools))
		for i, tool := range inst.InstalledTools {
			toolList[i] = "• " + tool
		}
		sb.WriteString("\n")
		sb.WriteString(renderSection("Installed Tools", toolList))
	}

	// Public URLs section
	if len(inst.PublicURLs) > 0 {
		urlList := make([]string, len(inst.PublicURLs))
		for i, url := range inst.PublicURLs {
			urlList[i] = StyleLink.Render(url)
		}
		sb.WriteString("\n")
		sb.WriteString(renderSection("Public URLs", urlList))
	}

	// Actions section
	sb.WriteString("\n")
	actions := []string{
		"[c] Connect via SSH",
		"[f] Forward local port",
		"[u] Update image",
		"[d] Delete instance",
		"[s] Start",
		"[S] Stop",
		"[r] Restart",
		"[R] Resize tier",
		"[e] Expose port",
	}

	// Disable actions based on status
	if inst.Status != "running" {
		actions = append([]string{"⚠️  Instance is not running"}, actions...)
	}

	sb.WriteString(renderSection("Actions", actions))

	return sb.String()
}

func renderSection(title string, items []string) string {
	var sb strings.Builder

	// Section title
	sectionTitle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#87CEEB")).
		Render("▸ " + title)

	sb.WriteString(sectionTitle)
	sb.WriteString("\n")

	// Items with indentation
	for _, item := range items {
		if item == "" {
			sb.WriteString("\n")
		} else {
			sb.WriteString("  " + item + "\n")
		}
	}

	return sb.String()
}
