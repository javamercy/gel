package cli

import (
	"github.com/spf13/cobra"
)

func newLsTreeCommand() *cobra.Command {
	var recursive bool
	var showTrees bool
	var nameOnly bool

	lsTreeCommand := &cobra.Command{
		Use:   "ls-tree",
		Short: "List the contents of a tree",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	lsTreeCommand.Flags().BoolVarP(
		&recursive,
		"recursive",
		"r",
		false,
		"Recursively list subtrees",
	)
	lsTreeCommand.Flags().BoolVarP(
		&showTrees,
		"show-trees",
		"t",
		false,
		"Show tree objects in the listing",
	)
	lsTreeCommand.Flags().BoolVarP(
		&nameOnly,
		"name-only",
		"n",
		false,
		"Show only names of the entries",
	)
	return lsTreeCommand
}
