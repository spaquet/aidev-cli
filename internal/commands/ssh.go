package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewSSHCmd creates the ssh subcommand
func NewSSHCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "ssh <instance-name>",
		Short: "SSH into an instance",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("SSH to %s (coming in Phase 4)\n", args[0])
			return nil
		},
	}
}
