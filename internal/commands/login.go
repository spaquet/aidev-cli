package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/aidev/cli/internal/api"
	"github.com/aidev/cli/internal/auth"
	"github.com/aidev/cli/internal/models"
)

// NewLoginCmd creates the login subcommand
func NewLoginCmd() *cobra.Command {
	var email, password, apiKey string

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Login and store credentials",
		RunE: func(cmd *cobra.Command, args []string) error {
			baseURL, _ := cmd.Flags().GetString("api")
			authStore, err := auth.NewStore()
			if err != nil {
				return fmt.Errorf("failed to initialize config: %w", err)
			}

			apiClient := api.NewClient(baseURL)
			resp, err := apiClient.Login(email, password, apiKey)
			if err != nil {
				return fmt.Errorf("login failed: %w", err)
			}

			// Save config
			config := &models.Config{
				BaseURL:        baseURL,
				Token:          resp.Token,
				TokenExpiresAt: resp.ExpiresAt,
				UserEmail:      resp.User.Email,
			}
			if err := authStore.Save(config); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			fmt.Printf("✅ Logged in as %s\n", resp.User.Email)
			return nil
		},
	}

	cmd.Flags().StringVar(&email, "email", "", "Email address")
	cmd.Flags().StringVar(&password, "password", "", "Password")
	cmd.Flags().StringVar(&apiKey, "api-key", "", "API key (instead of email/password)")

	return cmd
}
