package cli

import (
	"github.com/spf13/cobra"
)

func newSwitchCommand() *cobra.Command {
	var create bool
	var force bool

	switchCommand := &cobra.Command{
		Use:   "switch",
		Short: "Switch branches or restore working tree files",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	switchCommand.Flags().BoolVarP(
		&create,
		"create",
		"c",
		false,
		"Create the new branch",
	)
	switchCommand.Flags().BoolVarP(
		&force,
		"force",
		"f",
		false,
		"Switch even if the index or the working tree differs from HEAD",
	)
	return switchCommand
}
