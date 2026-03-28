package views

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
	"github.com/aidev/cli/internal/api"
)

type LoginModel struct {
	client       *api.Client
	baseURL      string
	width        int
	height       int
	mode         LoginMode // email or apikey
	email        textinput.Model
	password     textinput.Model
	apiKey       textinput.Model
	errorMsg     string
	loading      bool
	focusIndex   int
	fields       []interface{} // Current mode's fields
}

type LoginMode int

const (
	LoginModeEmail LoginMode = iota
	LoginModeAPIKey
)

// LoginSuccessMsg is sent when login succeeds
type LoginSuccessMsg struct {
	Token string
}

type loginErrorMsg string

// NewLoginModel creates a new login view
func NewLoginModel(client *api.Client, baseURL string) *LoginModel {
	email := textinput.New()
	email.Placeholder = "alice@example.com"
	email.CharLimit = 254

	password := textinput.New()
	password.Placeholder = "••••••••"
	password.EchoCharacter = '•'
	password.EchoMode = textinput.EchoPassword

	apiKey := textinput.New()
	apiKey.Placeholder = "aidev_sk_..."
	apiKey.CharLimit = 100

	m := &LoginModel{
		client:    client,
		baseURL:   baseURL,
		mode:      LoginModeEmail,
		email:     email,
		password:  password,
		apiKey:    apiKey,
		focusIndex: 0,
	}

	m.updateFields()
	m.email.Focus()

	return m
}

func (m *LoginModel) updateFields() {
	if m.mode == LoginModeEmail {
		m.fields = []interface{}{&m.email, &m.password}
	} else {
		m.fields = []interface{}{&m.apiKey}
	}
}

func (m *LoginModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m *LoginModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit

		case "tab":
			m.focusIndex = (m.focusIndex + 1) % len(m.fields)
			m.updateFocus()
			return m, nil

		case "shift+tab":
			m.focusIndex = (m.focusIndex - 1 + len(m.fields)) % len(m.fields)
			m.updateFocus()
			return m, nil

		case "enter":
			if m.loading {
				return m, nil
			}
			m.loading = true
			m.errorMsg = ""

			if m.mode == LoginModeEmail {
				return m, m.attemptEmailLogin()
			} else {
				return m, m.attemptAPIKeyLogin()
			}
		}

	case LoginSuccessMsg:
		// Login succeeded; signal parent to transition
		m.loading = false
		return m, func() tea.Msg {
			// Signal parent to handle login success
			return msg
		}

	case loginErrorMsg:
		m.errorMsg = string(msg)
		m.loading = false
		m.password.SetValue("") // Clear password on error
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	// Update focused field
	if m.focusIndex < len(m.fields) {
		if input, ok := m.fields[m.focusIndex].(*textinput.Model); ok {
			var cmd tea.Cmd
			*input, cmd = input.Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

func (m *LoginModel) View() string {
	return m.render()
}

func (m *LoginModel) render() string {
	var sb strings.Builder

	// Title
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorFg).
		MarginBottom(2).
		Render("🔐 AIDev Login")

	sb.WriteString(title)
	sb.WriteString("\n")

	// Mode indicator
	modeEmail := "Email"
	modeAPIKey := "API Key"
	if m.mode == LoginModeEmail {
		modeEmail = StyleSuccess.Render(modeEmail)
	} else {
		modeAPIKey = StyleSuccess.Render(modeAPIKey)
	}

	sb.WriteString(modeEmail)
	sb.WriteString(" • ")
	sb.WriteString(modeAPIKey)
	sb.WriteString("\n\n")

	// Form fields
	if m.mode == LoginModeEmail {
		sb.WriteString(StyleLabel.Render("Email:"))
		sb.WriteString("\n")
		sb.WriteString(m.email.View())
		sb.WriteString("\n\n")

		sb.WriteString(StyleLabel.Render("Password:"))
		sb.WriteString("\n")
		sb.WriteString(m.password.View())
		sb.WriteString("\n\n")
	} else {
		sb.WriteString(StyleLabel.Render("API Key:"))
		sb.WriteString("\n")
		sb.WriteString(m.apiKey.View())
		sb.WriteString("\n\n")
	}

	// Error message
	if m.errorMsg != "" {
		sb.WriteString(StyleError.Render("❌ " + m.errorMsg))
		sb.WriteString("\n\n")
	}

	// Loading
	if m.loading {
		sb.WriteString(StyleWarning.Render("⏳ Logging in..."))
		sb.WriteString("\n\n")
	}

	// Instructions
	sb.WriteString(StyleHint.Render("[Tab] Next field • [Enter] Sign in • [Ctrl+C] Exit"))

	return StyleBorderBox.Render(sb.String())
}

func (m *LoginModel) updateFocus() {
	m.email.Blur()
	m.password.Blur()
	m.apiKey.Blur()

	if m.focusIndex < len(m.fields) {
		if input, ok := m.fields[m.focusIndex].(*textinput.Model); ok {
			input.Focus()
		}
	}
}

func (m *LoginModel) attemptEmailLogin() tea.Cmd {
	email := strings.TrimSpace(m.email.Value())
	password := m.password.Value()

	if email == "" || password == "" {
		return func() tea.Msg {
			return loginErrorMsg("Email and password are required")
		}
	}

	return func() tea.Msg {
		resp, err := m.client.Login(email, password, "")
		if err != nil {
			if api.IsUnauthorized(err) {
				return loginErrorMsg("Invalid email or password")
			}
			return loginErrorMsg(fmt.Sprintf("Login failed: %v", err))
		}

		return LoginSuccessMsg{Token: resp.Token}
	}
}

func (m *LoginModel) attemptAPIKeyLogin() tea.Cmd {
	apiKey := strings.TrimSpace(m.apiKey.Value())

	if apiKey == "" {
		return func() tea.Msg {
			return loginErrorMsg("API key is required")
		}
	}

	if !strings.HasPrefix(apiKey, "aidev_sk_") {
		return func() tea.Msg {
			return loginErrorMsg("API key must start with 'aidev_sk_'")
		}
	}

	return func() tea.Msg {
		resp, err := m.client.Login("", "", apiKey)
		if err != nil {
			if api.IsUnauthorized(err) {
				return loginErrorMsg("Invalid API key")
			}
			return loginErrorMsg(fmt.Sprintf("Login failed: %v", err))
		}

		return LoginSuccessMsg{Token: resp.Token}
	}
}
