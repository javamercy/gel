package cli

import (
	"github.com/spf13/cobra"
)

func newRemoveCommand() *cobra.Command {
	var cached bool
	var dryRun bool
	var recursive bool
	var force bool

	removeCommand := &cobra.Command{
		Use:   "rm <pathspec>...",
		Short: "Remove a file or directory",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	removeCommand.Flags().BoolVar(
		&cached,
		"cached",
		false,
		"Remove from cache only",
	)
	removeCommand.Flags().BoolVarP(
		&dryRun,
		"dry-run",
		"n",
		false,
		"Show what would be removed without actually removing",
	)
	removeCommand.Flags().BoolVarP(
		&recursive,
		"recursive",
		"r",
		false,
		"Remove directories and their contents recursively",
	)
	removeCommand.Flags().BoolVarP(
		&force,
		"force",
		"f",
		false,
		"Ignore nonexistent files and arguments, never prompt",
	)
	return removeCommand
}
