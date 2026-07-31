package cli

import (
	"github.com/spf13/cobra"
)

func newUpdateIndexCommand() *cobra.Command {
	var add bool
	var remove bool

	updateIndexCommand := &cobra.Command{
		Use:   "update-index <file>...",
		Short: "Update the index with the current state of the working directory",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	updateIndexCommand.Flags().BoolVarP(
		&add,
		"add",
		"a",
		false,
		"Add specified files to the index",
	)
	updateIndexCommand.Flags().BoolVarP(
		&remove,
		"remove",
		"r",
		false,
		"Remove specified files from the index",
	)
	return updateIndexCommand
}
