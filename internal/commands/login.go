package commands

import (
	"fmt"
	"os/exec"
	"runtime"
	"time"

	"github.com/spf13/cobra"
	"github.com/aidev/cli/internal/api"
	"github.com/aidev/cli/internal/auth"
	"github.com/aidev/cli/internal/models"
)

// NewLoginCmd creates the login subcommand
func NewLoginCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate using your browser",
		RunE: func(cmd *cobra.Command, args []string) error {
			baseURL, _ := cmd.Flags().GetString("api")
			authStore, err := auth.NewStore()
			if err != nil {
				return fmt.Errorf("failed to initialize config: %w", err)
			}

			apiClient := api.NewClient(baseURL)

			// Initiate device authorization
			fmt.Println("Initiating browser authentication...")
			deviceResp, err := apiClient.DeviceAuthorize()
			if err != nil {
				return fmt.Errorf("failed to initiate authentication: %w", err)
			}

			// Display code and URL
			fmt.Println()
			fmt.Println("Your authorization code:")
			fmt.Printf("  %s\n", deviceResp.UserCode)
			fmt.Println()
			fmt.Printf("Visit: %s\n", deviceResp.VerificationURI)
			fmt.Println()

			// Open browser automatically
			openBrowser(deviceResp.VerificationURI)

			// Poll for result
			fmt.Println("Waiting for authorization...")
			timeout := time.Now().Add(time.Duration(deviceResp.ExpiresIn) * time.Second)

			for {
				// Check timeout
				if time.Now().After(timeout) {
					return fmt.Errorf("authentication timeout")
				}

				// Poll
				loginResp, err := apiClient.DevicePoll(deviceResp.DeviceCode)
				if err != nil {
					if api.IsAuthorizationPending(err) {
						// Still waiting, sleep and retry
						time.Sleep(time.Duration(deviceResp.Interval) * time.Second)
						continue
					}
					// Other error
					if apiErr, ok := err.(*api.HTTPError); ok {
						return fmt.Errorf("authentication failed: %s", apiErr.Body)
					}
					return fmt.Errorf("authentication failed: %w", err)
				}

				// Success - save config
				config := &models.Config{
					BaseURL:        baseURL,
					Token:          loginResp.Token,
					TokenExpiresAt: loginResp.ExpiresAt,
					UserEmail:      loginResp.User.Email,
				}
				if err := authStore.Save(config); err != nil {
					return fmt.Errorf("failed to save config: %w", err)
				}

				fmt.Printf("Logged in as %s\n", loginResp.User.Email)
				return nil
			}
		},
	}

	return cmd
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
