package cli

import (
	"github.com/spf13/cobra"
)

func newAddCommand() *cobra.Command {
	var dryRun bool
	var verbose bool

	addCommand := &cobra.Command{
		Use:   "add <pathspec>...",
		Short: "Add file contents to the index",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	addCommand.Flags().BoolVarP(
		&dryRun,
		"dry-run",
		"n",
		false,
		"Dry run the add operation without making any changes",
	)
	addCommand.Flags().BoolVarP(
		&verbose,
		"verbose",
		"v",
		false,
		"Show verbose output of the add operation",
	)
	return addCommand
}
