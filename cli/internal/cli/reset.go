package cli

import (
	"github.com/spf13/cobra"
)

func newResetCommand() *cobra.Command {
	var hard bool
	var soft bool
	var mixed bool

	resetCommand := &cobra.Command{
		Use:   "reset [target]",
		Short: "Reset the current HEAD to a specified state",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	resetCommand.Flags().BoolVarP(
		&soft, "soft",
		"S",
		false,
		"Move HEAD only; keep index and working tree",
	)
	resetCommand.Flags().BoolVarP(
		&mixed,
		"mixed",
		"M",
		false,
		"Move HEAD and reset index; keep working tree (default)",
	)
	resetCommand.Flags().BoolVarP(
		&hard,
		"hard",
		"H",
		false,
		"Move HEAD, reset index, and discard working tree changes",
	)
	return resetCommand
}
