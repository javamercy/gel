package cli

import (
	"github.com/spf13/cobra"
)

func newReadTreeCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "read-tree <tree-hash>",
		Short: "Read tree objects into the index",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
}
