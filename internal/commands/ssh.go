package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/aidev/cli/internal/api"
	"github.com/aidev/cli/internal/auth"
	"github.com/aidev/cli/internal/models"
	"github.com/aidev/cli/internal/ssh"
)

// NewSSHCmd creates the ssh subcommand
func NewSSHCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "ssh <instance-name>",
		Short: "SSH into an instance directly (no TUI)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			instanceName := args[0]
			baseURL, _ := cmd.Flags().GetString("api")

			// Load auth config
			authStore, err := auth.NewStore()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			config, err := authStore.Load()
			if err != nil {
				return fmt.Errorf("not logged in. Run 'aidev login' first")
			}

			// Create API client with token
			apiClient := api.NewClient(baseURL)
			apiClient.SetToken(config.Token)

			// Fetch instances
			resp, err := apiClient.GetInstances()
			if err != nil {
				return fmt.Errorf("failed to fetch instances: %w", err)
			}

			// Find instance by name
			var instance *models.Instance
			for i := range resp.Instances {
				if resp.Instances[i].Name == instanceName {
					instance = &resp.Instances[i]
					break
				}
			}

			if instance == nil {
				return fmt.Errorf("instance %q not found", instanceName)
			}

			if instance.Status != "running" {
				return fmt.Errorf("instance %q is not running (status: %s)", instanceName, instance.Status)
			}

			// Connect via SSH
			return ssh.Connect(ssh.ConnectOptions{
				Host: instance.SSHHost,
				Port: instance.SSHPort,
				User: instance.SSHUser,
			})
		},
	}
}
