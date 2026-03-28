package commands

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/aidev/cli/internal/api"
	"github.com/aidev/cli/internal/auth"
	"github.com/aidev/cli/internal/models"
	"github.com/aidev/cli/internal/tui"
)

// NewTUICmd creates the TUI subcommand
func NewTUICmd(version string) *cobra.Command {
	return &cobra.Command{
		Use:   "tui",
		Short: "Launch the interactive TUI (default)",
		Run: func(cmd *cobra.Command, args []string) {
			baseURL, _ := cmd.Flags().GetString("api")
			authStore, err := auth.NewStore()
			if err != nil {
				cobra.CompError("Failed to initialize config: " + err.Error())
				return
			}

			apiClient := api.NewClient(baseURL)
			RunTUI(apiClient, authStore, baseURL, version)
		},
	}
}

// RunTUI launches the Bubble Tea application
// Returns sshInstance if user initiated SSH connection
func RunTUI(apiClient *api.Client, authStore *auth.Store, baseURL string, version string) *models.Instance {
	appModel := tui.NewAppModel(baseURL, apiClient, authStore, version)

	p := tea.NewProgram(appModel, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		panic(err)
	}

	return appModel.SSHInstance
}
