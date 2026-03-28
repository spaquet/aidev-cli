package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewForwardCmd creates the forward subcommand
func NewForwardCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "forward <instance-name> <local-port> [remote-port]",
		Short: "Forward a local port to a remote instance",
		Args:  cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("Forward to %s (coming in Phase 4)\n", args[0])
			return nil
		},
	}
}
