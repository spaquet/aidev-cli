package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewInstancesCmd creates the instances subcommand
func NewInstancesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "instances",
		Short: "Manage instances",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List all instances",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("List instances (coming in Phase 2)")
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "create",
		Short: "Create a new instance",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Create instance (coming later)")
			return nil
		},
	})

	return cmd
}
