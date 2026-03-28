package views

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/aidev/cli/internal/api"
)

// LoginState represents the current state of the device flow
type LoginState int

const (
	LoginStateWelcome    LoginState = iota
	LoginStateInitiating
	LoginStateWaiting
	LoginStateError
)

type LoginModel struct {
	client         *api.Client
	baseURL        string
	version        string
	width          int
	height         int
	state          LoginState
	deviceCode     string
	userCode       string
	verificationURI string
	pollInterval   int
	errorMsg       string
	lastError      error
}

// Message types
type deviceCodeMsg struct {
	DeviceCode      string
	UserCode        string
	VerificationURI string
	Interval        int
}

type devicePollResultMsg struct {
	Token string
	Err   error
}

type loginErrorMsg string

// LoginSuccessMsg is sent when login succeeds
type LoginSuccessMsg struct {
	Token string
}

// NewLoginModel creates a new login view
func NewLoginModel(client *api.Client, baseURL string, version string) *LoginModel {
	return &LoginModel{
		client:  client,
		baseURL: baseURL,
		version: version,
		state:   LoginStateWelcome,
	}
}

func (m *LoginModel) Init() tea.Cmd {
	return nil
}

func (m *LoginModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		case "enter", " ":
			// Welcome state: start login
			if m.state == LoginStateWelcome {
				m.state = LoginStateInitiating
				return m, m.initiateDeviceLogin()
			}
			// Error state: allow retry
			if m.state == LoginStateError {
				m.state = LoginStateInitiating
				m.errorMsg = ""
				m.lastError = nil
				return m, m.initiateDeviceLogin()
			}
		}

	case deviceCodeMsg:
		m.deviceCode = msg.DeviceCode
		m.userCode = msg.UserCode
		m.verificationURI = msg.VerificationURI
		m.pollInterval = msg.Interval
		m.state = LoginStateWaiting
		m.errorMsg = ""

		// Open browser
		openBrowser(m.verificationURI)

		// Start polling
		return m, m.pollDeviceAuth()

	case devicePollResultMsg:
		if msg.Err != nil {
			if api.IsAuthorizationPending(msg.Err) {
				// Still waiting, schedule next poll
				return m, m.pollDeviceAuth()
			}
			// Error
			m.state = LoginStateError
			m.lastError = msg.Err
			m.errorMsg = formatDeviceError(msg.Err)
			return m, nil
		}
		// Success
		return m, func() tea.Msg {
			return LoginSuccessMsg{Token: msg.Token}
		}

	case LoginSuccessMsg:
		// Re-emit upward
		return m, func() tea.Msg {
			return msg
		}

	case loginErrorMsg:
		m.state = LoginStateError
		m.errorMsg = string(msg)
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

func (m *LoginModel) View() string {
	var sb strings.Builder

	switch m.state {
	case LoginStateWelcome:
		return m.renderWelcome()

	case LoginStateInitiating:
		sb.WriteString("AIDev Login")
		sb.WriteString("\n\n")
		sb.WriteString(StyleHint.Render("Initiating browser authentication..."))
		sb.WriteString("\n\n")

	case LoginStateWaiting:
		sb.WriteString("AIDev Login")
		sb.WriteString("\n\n")
		sb.WriteString(StyleHint.Render("Opening browser for authentication..."))
		sb.WriteString("\n\n")

		sb.WriteString(StyleLabel.Render("Code:"))
		sb.WriteString("\n")
		sb.WriteString(StyleSuccess.Render("  " + m.userCode))
		sb.WriteString("\n\n")

		sb.WriteString(StyleLabel.Render("Or visit:"))
		sb.WriteString("\n")
		sb.WriteString(StyleLink.Render("  " + m.verificationURI))
		sb.WriteString("\n\n")

		sb.WriteString(StyleWarning.Render("Waiting for authorization..."))
		sb.WriteString("\n")

	case LoginStateError:
		sb.WriteString("AIDev Login")
		sb.WriteString("\n\n")
		sb.WriteString(StyleError.Render("Error: " + m.errorMsg))
		sb.WriteString("\n\n")
		sb.WriteString(StyleHint.Render("[Enter] Retry  [Ctrl+C] Exit"))
		sb.WriteString("\n")
	}

	sb.WriteString("\n")
	sb.WriteString(StyleHint.Render("[Ctrl+C] Cancel"))

	return StyleBorderBox.Render(sb.String())
}

// Private methods

func (m *LoginModel) renderWelcome() string {
	var sb strings.Builder

	// App name in large spaced letters with sky blue color
	appName := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#87CEEB")).
		Render("A I D E V")

	sb.WriteString(appName)
	sb.WriteString("\n")

	// Subtitle
	subtitle := lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorFg).
		Render("AI Development Sandbox")

	sb.WriteString(subtitle)
	sb.WriteString("\n")

	// Version
	versionStr := lipgloss.NewStyle().
		Foreground(ColorDimmed).
		Render("v" + m.version)

	sb.WriteString(versionStr)
	sb.WriteString("\n\n")

	// Separator
	separator := strings.Repeat("─", 45)
	sb.WriteString(StyleHint.Render(separator))
	sb.WriteString("\n\n")

	// Marketing copy
	marketing := `Cloud workstations for AI-assisted development.

Pre-loaded with Claude Code, Codex, and
developer tools. SSH in from anywhere.
Spin up new instances in minutes.`

	sb.WriteString(StyleHint.Render(marketing))
	sb.WriteString("\n\n")

	// Separator
	sb.WriteString(StyleHint.Render(separator))
	sb.WriteString("\n\n")

	// Button
	button := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#87CEEB")).
		Padding(0, 4).
		Bold(true).
		Foreground(lipgloss.Color("#87CEEB")).
		Render("Sign in with browser")

	// Center the button
	buttonPadded := lipgloss.NewStyle().
		Align(lipgloss.Center).
		Render(button)

	sb.WriteString(buttonPadded)
	sb.WriteString("\n\n")

	// Footer hints
	sb.WriteString(StyleHint.Render("[Enter] Sign in  [Ctrl+C] Exit"))

	return StyleBorderBox.Render(sb.String())
}

func (m *LoginModel) initiateDeviceLogin() tea.Cmd {
	return func() tea.Msg {
		resp, err := m.client.DeviceAuthorize()
		if err != nil {
			return loginErrorMsg(fmt.Sprintf("Failed to initiate login: %v", err))
		}

		return deviceCodeMsg{
			DeviceCode:      resp.DeviceCode,
			UserCode:        resp.UserCode,
			VerificationURI: resp.VerificationURI,
			Interval:        resp.Interval,
		}
	}
}

func (m *LoginModel) pollDeviceAuth() tea.Cmd {
	deviceCode := m.deviceCode
	interval := m.pollInterval

	return tea.Tick(time.Duration(interval)*time.Second, func(t time.Time) tea.Msg {
		resp, err := m.client.DevicePoll(deviceCode)
		if err != nil {
			return devicePollResultMsg{Err: err}
		}

		return devicePollResultMsg{Token: resp.Token}
	})
}

func openBrowser(url string) {
	switch runtime.GOOS {
	case "darwin":
		exec.Command("open", url).Start()
	case "linux":
		exec.Command("xdg-open", url).Start()
	case "windows":
		exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	}
}

func formatDeviceError(err error) string {
	if httpErr, ok := err.(*api.HTTPError); ok {
		if strings.Contains(httpErr.Body, "expired_token") {
			return "Device code expired. Please try again."
		}
		if strings.Contains(httpErr.Body, "access_denied") {
			return "Authorization denied."
		}
	}
	return fmt.Sprintf("%v", err)
}
