package cli

import (
	"github.com/spf13/cobra"
)

func newRestoreCommand() *cobra.Command {
	var staged bool
	var source string

	restoreCommand := &cobra.Command{
		Use:   "restore",
		Short: "Restore working tree files",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	restoreCommand.Flags().BoolVarP(
		&staged, "staged",
		"s",
		false,
		"Restore staged files",
	)
	restoreCommand.Flags().StringVarP(
		&source, "source",
		"S",
		"",
		"Restore from specified source",
	)
	return restoreCommand
}
