package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/aidev/cli/internal/api"
	"github.com/aidev/cli/internal/auth"
	"github.com/aidev/cli/internal/commands"
	"github.com/aidev/cli/internal/ssh"
)

var (
	version = "0.1.0"
	baseURL = "https://api.sandbox.example.com"
)

func main() {
	rootCmd := &cobra.Command{
		Use:     "aidev",
		Short:   "AIDev CLI - Manage your AI Dev Sandbox instances",
		Version: version,
		Run: func(cmd *cobra.Command, args []string) {
			// Default: launch TUI
			launchTUI(baseURL)
		},
	}

	// Global flags
	rootCmd.PersistentFlags().StringVar(&baseURL, "api", baseURL, "API base URL")

	// Subcommands
	rootCmd.AddCommand(commands.NewTUICmd())
	rootCmd.AddCommand(commands.NewLoginCmd())
	rootCmd.AddCommand(commands.NewSSHCmd())
	rootCmd.AddCommand(commands.NewForwardCmd())
	rootCmd.AddCommand(commands.NewInstancesCmd())
	rootCmd.AddCommand(commands.NewConfigCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func launchTUI(baseURL string) {
	authStore, err := auth.NewStore()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize config store: %v\n", err)
		os.Exit(1)
	}

	apiClient := api.NewClient(baseURL)
	sshInstance := commands.RunTUI(apiClient, authStore, baseURL)

	// If user initiated SSH connection, handle it
	if sshInstance != nil {
		err := ssh.Connect(ssh.ConnectOptions{
			Host: sshInstance.SSHHost,
			Port: sshInstance.SSHPort,
			User: sshInstance.SSHUser,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "SSH connection error: %v\n", err)
			os.Exit(1)
		}
	}
}
