package cli

import (
	"github.com/spf13/cobra"
)

func newUpdateRefCommand() *cobra.Command {
	var delete_ bool

	updateRefCommand := &cobra.Command{
		Use:   "update-ref",
		Short: "Update a reference",
		Args:  cobra.RangeArgs(1, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	updateRefCommand.Flags().BoolVarP(
		&delete_,
		"delete",
		"d",
		false,
		"Delete the reference instead of updating it",
	)
	return updateRefCommand
}
