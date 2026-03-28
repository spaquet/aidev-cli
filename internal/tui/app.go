package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/aidev/cli/internal/api"
	"github.com/aidev/cli/internal/auth"
	"github.com/aidev/cli/internal/models"
	"github.com/aidev/cli/internal/tui/views"
)

type Screen int

const (
	ScreenLogin Screen = iota
	ScreenMain
)

// AppModel is the root Bubble Tea model
type AppModel struct {
	screen      Screen
	baseURL     string
	token       string
	config      *models.Config
	apiClient   *api.Client
	authStore   *auth.Store
	width       int
	height      int
	SSHInstance *models.Instance // Set when SSH is requested

	// Views
	loginView *views.LoginModel
	mainView  *views.MainModel
}

// NewAppModel creates the root app model
func NewAppModel(baseURL string, apiClient *api.Client, authStore *auth.Store) *AppModel {
	return &AppModel{
		baseURL:   baseURL,
		apiClient: apiClient,
		authStore: authStore,
		screen:    ScreenLogin,
	}
}

// Init initializes the app
func (m *AppModel) Init() tea.Cmd {
	// Try to load existing config
	config, err := m.authStore.Load()
	if err != nil {
		if !auth.IsNoConfigError(err) {
			// Log error but continue
		}
		// No config, show login
		m.loginView = views.NewLoginModel(m.apiClient, m.baseURL)
		m.screen = ScreenLogin
		return m.loginView.Init()
	}

	// Check if token is expired
	if m.authStore.IsTokenExpired(config) {
		// Try to refresh
		return func() tea.Msg {
			return refreshTokenCmd{config: config}
		}
	}

	// Token is valid
	m.config = config
	m.token = config.Token
	m.apiClient.SetToken(m.token)
	m.screen = ScreenMain

	// Initialize main view with a placeholder user (we'll fetch real data in Phase 3)
	user := &models.User{
		ID:    "unknown",
		Email: config.UserEmail,
		Name:  config.UserEmail,
	}
	m.mainView = views.NewMainModel(m.apiClient, user, m.width, m.height)
	return m.mainView.Init()
}

// Update handles messages
func (m *AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case refreshTokenCmd:
		return m, m.handleRefreshToken(msg)

	case refreshTokenResponseMsg:
		if msg.err != nil {
			// Refresh failed, go to login
			m.screen = ScreenLogin
			m.loginView = views.NewLoginModel(m.apiClient, m.baseURL)
			return m, m.loginView.Init()
		}
		// Refresh succeeded
		m.config = msg.config
		m.token = msg.config.Token
		m.apiClient.SetToken(m.token)
		m.screen = ScreenMain

		user := &models.User{
			ID:    "unknown",
			Email: msg.config.UserEmail,
			Name:  msg.config.UserEmail,
		}
		m.mainView = views.NewMainModel(m.apiClient, user, m.width, m.height)
		return m, m.mainView.Init()

	case views.LoginSuccessMsg:
		// Login succeeded
		return m, m.handleLoginSuccess(msg)

	case views.SSHConnectMsg:
		// SSH requested - store instance and quit TUI
		instance := msg.Instance
		m.SSHInstance = &instance
		return m, tea.Quit
	}

	// Route to current screen
	switch m.screen {
	case ScreenLogin:
		if m.loginView != nil {
			var cmd tea.Cmd
			model, cmd := m.loginView.Update(msg)
			if loginModel, ok := model.(*views.LoginModel); ok {
				m.loginView = loginModel
			}
			return m, cmd
		}

	case ScreenMain:
		if m.mainView != nil {
			var cmd tea.Cmd
			model, cmd := m.mainView.Update(msg)
			if mainModel, ok := model.(*views.MainModel); ok {
				m.mainView = mainModel
			}
			return m, cmd
		}
	}

	return m, nil
}

// View renders the current screen
func (m *AppModel) View() string {
	switch m.screen {
	case ScreenLogin:
		if m.loginView != nil {
			return m.loginView.View()
		}
		return "Loading..."
	case ScreenMain:
		if m.mainView != nil {
			return m.mainView.View()
		}
		return "Loading..."
	}
	return "Unknown screen"
}

// Private handlers

func (m *AppModel) handleLoginSuccess(msg views.LoginSuccessMsg) tea.Cmd {
	return func() tea.Msg {
		// Save config (note: expires_at will be set from JWT decode in Phase 3)
		config := &models.Config{
			BaseURL:        m.baseURL,
			Token:          msg.Token,
			TokenExpiresAt: "", // Will be set after JWT decode
			UserEmail:      "", // Will be set after JWT decode
		}
		m.authStore.Save(config)
		m.config = config
		m.token = msg.Token
		m.apiClient.SetToken(m.token)
		m.screen = ScreenMain

		user := &models.User{
			ID:    "unknown",
			Email: "unknown",
			Name:  "User",
		}
		m.mainView = views.NewMainModel(m.apiClient, user, m.width, m.height)
		return m.mainView.Init()
	}
}

func (m *AppModel) handleRefreshToken(msg refreshTokenCmd) tea.Cmd {
	return func() tea.Msg {
		resp, err := m.apiClient.Refresh(msg.config.Token)
		if err != nil {
			return refreshTokenResponseMsg{err: err}
		}

		// Update config
		config := &models.Config{
			BaseURL:        m.baseURL,
			Token:          resp.Token,
			TokenExpiresAt: resp.ExpiresAt,
			UserEmail:      msg.config.UserEmail,
		}
		m.authStore.Save(config)

		return refreshTokenResponseMsg{config: config}
	}
}

// Message types

type refreshTokenCmd struct {
	config *models.Config
}

type refreshTokenResponseMsg struct {
	config *models.Config
	err    error
}
