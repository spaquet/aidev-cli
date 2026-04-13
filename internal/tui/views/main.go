package views

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/aidev/cli/internal/api"
	"github.com/aidev/cli/internal/models"
)

// Modal represents an overlay dialog
type Modal int

const (
	ModalNone Modal = iota
	ModalConfirmDelete
	ModalResize
	ModalForward
)

// PortForward represents an active port forward
type PortForward struct {
	LocalPort  int
	RemotePort int
	StopFunc   func() error
}

// MainModel is the main view (instance list + detail)
type MainModel struct {
	client        *api.Client
	user          *models.User
	list          *InstanceListModel
	detail        *InstanceDetailModel
	modal         Modal
	confirmDialog *ConfirmDialogModel
	resizeModal   *ResizeModalModel
	forwardModal  *ForwardModalModel
	notifications *NotificationManager
	width         int
	height        int
	showHelp      bool
	operationMsg  string                  // Status message
	portForwards  map[string]*PortForward // instance ID -> port forward
}

// NewMainModel creates the main view
func NewMainModel(client *api.Client, user *models.User, width, height int) *MainModel {
	return &MainModel{
		client:        client,
		user:          user,
		list:          NewInstanceListModel(client, width*40/100, height),
		detail:        NewInstanceDetailModel(width*60/100, height),
		notifications: NewNotificationManager(),
		width:         width,
		height:        height,
		showHelp:      true,
		modal:         ModalNone,
		portForwards:  make(map[string]*PortForward),
	}
}

func (m *MainModel) Init() tea.Cmd {
	return m.list.Init()
}

func (m *MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Route to modal if one is open
	if m.modal != ModalNone {
		return m.handleModalUpdate(msg)
	}

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "?":
			m.showHelp = !m.showHelp
			return m, nil

		case "enter":
			// Select instance
			selected := m.list.SelectedInstance()
			if selected != nil {
				m.detail.SetInstance(selected)
			}
			return m, nil

		// Operations
		case "c":
			// Connect via SSH
			selected := m.list.SelectedInstance()
			if selected != nil && selected.Status == "running" {
				// Signal parent to handle SSH connection and exit TUI
				return m, func() tea.Msg {
					return SSHConnectMsg{Instance: *selected}
				}
			}
			return m, nil

		case "d":
			// Delete confirmation
			selected := m.list.SelectedInstance()
			if selected != nil {
				m.modal = ModalConfirmDelete
				m.confirmDialog = NewConfirmDialog(
					"Delete Instance?",
					fmt.Sprintf("Are you sure you want to permanently delete %q?\nAll data will be lost.", selected.Name),
				)
				return m, m.confirmDialog.Init()
			}
			return m, nil

		case "R":
			// Resize tier
			selected := m.list.SelectedInstance()
			if selected != nil {
				m.modal = ModalResize
				m.resizeModal = NewResizeModal(selected.Tier)
				return m, m.resizeModal.Init()
			}
			return m, nil

		case "s":
			// Start instance
			selected := m.list.SelectedInstance()
			if selected != nil && selected.Status != "running" {
				m.operationMsg = fmt.Sprintf("Starting %s...", selected.Name)
				return m, m.startInstance(selected.ID)
			}
			return m, nil

		case "S":
			// Stop instance
			selected := m.list.SelectedInstance()
			if selected != nil && selected.Status == "running" {
				m.operationMsg = fmt.Sprintf("Stopping %s...", selected.Name)
				return m, m.stopInstance(selected.ID)
			}
			return m, nil

		case "r":
			// Restart instance
			selected := m.list.SelectedInstance()
			if selected != nil && selected.Status == "running" {
				m.operationMsg = fmt.Sprintf("Restarting %s...", selected.Name)
				return m, m.restartInstance(selected.ID)
			}
			return m, nil

		case "u":
			// Update image
			selected := m.list.SelectedInstance()
			if selected != nil && selected.ImageUpdateAvailable {
				m.operationMsg = fmt.Sprintf("Triggering image update on %s...", selected.Name)
				return m, m.updateImage(selected.ID)
			}
			return m, nil

		case "f":
			// Port forwarding
			selected := m.list.SelectedInstance()
			if selected != nil && selected.Status == "running" {
				m.modal = ModalForward
				m.forwardModal = NewForwardModal(3000)
				return m, m.forwardModal.Init()
			}
			return m, nil

		default:
			// Pass to list
			model, cmd := m.list.Update(msg)
			if listModel, ok := model.(*InstanceListModel); ok {
				m.list = listModel
				// Update detail when list selection changes
				if selected := m.list.SelectedInstance(); selected != nil {
					m.detail.SetInstance(selected)
				}
			}
			return m, cmd
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.width = msg.Width * 40 / 100
		m.list.height = msg.Height - 5
		m.detail.width = msg.Width * 60 / 100
		m.detail.height = msg.Height - 5

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

func (m *MainModel) View() tea.View {
	if m.modal != ModalNone {
		return tea.NewView(m.renderModalOverlay())
	}

	var sb strings.Builder

	topBar := lipgloss.NewStyle().
		Foreground(ColorFg).
		Background(lipgloss.Color("#1e1e1e")).
		Padding(0, 1).
		Render("AIDev • " + m.user.Email)

	sb.WriteString(topBar)
	sb.WriteString("\n")

	listWidth := m.width * 40 / 100
	detailWidth := m.width * 60 / 100

	listContent := m.list.View().Content
	detailContent := m.detail.View().Content

	// Pad content to widths
	listPadded := lipgloss.NewStyle().Width(listWidth - 1).Render(listContent)
	detailPadded := lipgloss.NewStyle().Width(detailWidth - 1).Render(detailContent)

	combined := lipgloss.JoinHorizontal(lipgloss.Top,
		listPadded,
		detailPadded,
	)

	sb.WriteString(combined)
	sb.WriteString("\n")

	m.notifications.Cleanup()
	if m.notifications.Count() > 0 {
		sb.WriteString(m.notifications.Render(m.width))
	}

	if m.operationMsg != "" {
		sb.WriteString(lipgloss.NewStyle().
			Foreground(ColorWarning).
			Render(m.operationMsg))
		sb.WriteString("\n")
	}

	footer := ""
	if m.showHelp {
		footer = "[↑↓] Navigate  [Enter] Select  [c]onnect [s]tart [S]top [r]estart"
		footer += "\n[u]pdate [d]elete [R]esize [f]orward [e]xpose  [Ctrl+R] Refresh  [?] Help [q]uit"
	} else {
		footer = "[?] Show help  [q]uit"
	}

	footerStyle := lipgloss.NewStyle().
		Foreground(ColorDimmed).
		Padding(0, 1).
		Render(footer)

	sb.WriteString(footerStyle)

	return tea.NewView(sb.String())
}

// Private methods

func (m *MainModel) handleModalUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.modal {
	case ModalConfirmDelete:
		if resp, ok := msg.(ConfirmResponse); ok {
			if resp.Confirmed {
				selected := m.list.SelectedInstance()
				if selected != nil {
					m.operationMsg = fmt.Sprintf("Deleting %s...", selected.Name)
					m.modal = ModalNone
					return m, m.deleteInstance(selected.ID)
				}
			}
			m.modal = ModalNone
			m.confirmDialog = nil
			return m, nil
		}
		model, cmd := m.confirmDialog.Update(msg)
		if confirm, ok := model.(*ConfirmDialogModel); ok {
			m.confirmDialog = confirm
		}
		return m, cmd

	case ModalResize:
		if resp, ok := msg.(ResizeResponse); ok {
			if resp.Confirm {
				selected := m.list.SelectedInstance()
				if selected != nil && selected.Tier != resp.NewTier {
					m.operationMsg = fmt.Sprintf("Resizing to %s...", resp.NewTier)
					m.modal = ModalNone
					req := &models.UpdateInstanceRequest{Tier: &resp.NewTier}
					return m, m.resizeInstance(selected.ID, req)
				}
			}
			m.modal = ModalNone
			m.resizeModal = nil
			return m, nil
		}
		model, cmd := m.resizeModal.Update(msg)
		if resize, ok := model.(*ResizeModalModel); ok {
			m.resizeModal = resize
		}
		return m, cmd

	case ModalForward:
		if resp, ok := msg.(ForwardResponse); ok {
			if resp.Confirm {
				selected := m.list.SelectedInstance()
				if selected != nil {
					m.modal = ModalNone
					return m, m.startPortForward(selected, resp.LocalPort, resp.RemotePort)
				}
			}
			m.modal = ModalNone
			m.forwardModal = nil
			return m, nil
		}
		model, cmd := m.forwardModal.Update(msg)
		if forward, ok := model.(*ForwardModalModel); ok {
			m.forwardModal = forward
		}
		return m, cmd
	}

	return m, nil
}

func (m *MainModel) renderModalOverlay() string {
	switch m.modal {
	case ModalConfirmDelete:
		if m.confirmDialog != nil {
			return m.confirmDialog.View().Content
		}
	case ModalResize:
		if m.resizeModal != nil {
			return m.resizeModal.View().Content
		}
	}
	return ""
}

// API operation commands

func (m *MainModel) deleteInstance(id string) tea.Cmd {
	return func() tea.Msg {
		err := m.client.DeleteInstance(id)
		if err != nil {
			return operationErrorMsg{err: err, operation: "delete"}
		}
		// Reload list
		return m.list.loadInstances()
	}
}

func (m *MainModel) startInstance(id string) tea.Cmd {
	return func() tea.Msg {
		_, err := m.client.StartInstance(id)
		if err != nil {
			return operationErrorMsg{err: err, operation: "start"}
		}
		return m.list.loadInstances()
	}
}

func (m *MainModel) stopInstance(id string) tea.Cmd {
	return func() tea.Msg {
		_, err := m.client.StopInstance(id)
		if err != nil {
			return operationErrorMsg{err: err, operation: "stop"}
		}
		return m.list.loadInstances()
	}
}

func (m *MainModel) restartInstance(id string) tea.Cmd {
	return func() tea.Msg {
		_, err := m.client.RestartInstance(id)
		if err != nil {
			return operationErrorMsg{err: err, operation: "restart"}
		}
		return m.list.loadInstances()
	}
}

func (m *MainModel) updateImage(id string) tea.Cmd {
	return func() tea.Msg {
		_, err := m.client.TriggerImageUpdate(id)
		if err != nil {
			return operationErrorMsg{err: err, operation: "update"}
		}
		return m.list.loadInstances()
	}
}

func (m *MainModel) resizeInstance(id string, req *models.UpdateInstanceRequest) tea.Cmd {
	return func() tea.Msg {
		_, err := m.client.UpdateInstance(id, req)
		if err != nil {
			return operationErrorMsg{err: err, operation: "resize"}
		}
		return m.list.loadInstances()
	}
}

func (m *MainModel) startPortForward(inst *models.Instance, localPort, remotePort int) tea.Cmd {
	return func() tea.Msg {
		// Import ssh package to use port forwarding
		return portForwardStartedMsg{
			instanceID: inst.ID,
			localPort:  localPort,
			remotePort: remotePort,
		}
	}
}

// operationErrorMsg is sent when an operation fails
type operationErrorMsg struct {
	err       error
	operation string
}

// SSHConnectMsg is sent when user requests SSH connection
type SSHConnectMsg struct {
	Instance models.Instance
}

// portForwardStartedMsg is sent when port forwarding starts
type portForwardStartedMsg struct {
	instanceID string
	localPort  int
	remotePort int
}
