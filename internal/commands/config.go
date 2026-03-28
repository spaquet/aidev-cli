package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/aidev/cli/internal/auth"
)

// NewConfigCmd creates the config subcommand
func NewConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "show",
		Short: "Show current configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			authStore, err := auth.NewStore()
			if err != nil {
				return fmt.Errorf("failed to initialize config: %w", err)
			}

			config, err := authStore.Load()
			if err != nil {
				if auth.IsNoConfigError(err) {
					fmt.Println("No configuration found. Run 'aidev login' first.")
					return nil
				}
				return fmt.Errorf("failed to load config: %w", err)
			}

			fmt.Printf("Base URL: %s\n", config.BaseURL)
			fmt.Printf("User Email: %s\n", config.UserEmail)
			fmt.Printf("Token Expires At: %s\n", config.TokenExpiresAt)
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "logout",
		Short: "Log out and delete credentials",
		RunE: func(cmd *cobra.Command, args []string) error {
			authStore, err := auth.NewStore()
			if err != nil {
				return fmt.Errorf("failed to initialize config: %w", err)
			}

			if err := authStore.Delete(); err != nil {
				return fmt.Errorf("failed to delete config: %w", err)
			}

			fmt.Println("✅ Logged out")
			return nil
		},
	})

	return cmd
}
