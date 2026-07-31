package cli

import (
	"github.com/spf13/cobra"
)

func newSymbolicRefCommand() *cobra.Command {
	var short bool

	symbolicRefCommand := &cobra.Command{
		Use:   "symbolic-ref <name> [ref]",
		Short: "Read or update symbolic references",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	symbolicRefCommand.Flags().BoolVarP(
		&short,
		"short",
		"s",
		false,
		"Shorten refs/heads/<name> to <name> when reading",
	)
	return symbolicRefCommand
}
