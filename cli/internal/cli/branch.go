package cli

import (
	"github.com/spf13/cobra"
)

func newBranchCommand() *cobra.Command {
	var delete_ bool

	branchCmd := &cobra.Command{
		Use:   "branch",
		Short: "List, create, or delete branches",
		Args:  cobra.RangeArgs(0, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	branchCmd.Flags().BoolVarP(
		&delete_,
		"delete",
		"d",
		false,
		"Delete all branches",
	)
	return branchCmd
}
