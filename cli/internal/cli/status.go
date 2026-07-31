package cli

import (
	"github.com/spf13/cobra"
)

func newStatusCommand() *cobra.Command {
	statusCommand := &cobra.Command{
		Use:   "status",
		Short: "Show the working tree status",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	return statusCommand
}
