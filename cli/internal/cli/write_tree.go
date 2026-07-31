package cli

import (
	"github.com/spf13/cobra"
)

func newWriteTreeCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "write-tree",
		Short: "Write the current index as a tree object",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
}
