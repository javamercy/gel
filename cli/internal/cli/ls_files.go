package cli

import (
	"github.com/spf13/cobra"
)

func newLsFilesCommand() *cobra.Command {
	var stage bool
	var cached bool
	var deleted bool
	var modified bool

	lsFilesCommand := &cobra.Command{
		Use:   "ls-files",
		Short: "List all files tracked by Gel in the current repository",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	lsFilesCommand.Flags().BoolVarP(
		&cached,
		"cached",
		"c",
		false,
		"Show cached files in the index",
	)
	lsFilesCommand.Flags().BoolVarP(
		&stage,
		"stage",
		"s",
		false,
		"Show staged files",
	)
	lsFilesCommand.Flags().BoolVarP(
		&modified,
		"modified",
		"m",
		false,
		"Show modified files",
	)
	lsFilesCommand.Flags().BoolVarP(
		&deleted,
		"deleted",
		"d",
		false,
		"Show deleted files",
	)
	return lsFilesCommand
}
