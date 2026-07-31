package cli

import (
	"github.com/spf13/cobra"
)

func newDiffCommand() *cobra.Command {
	var staged bool

	diffCommand := &cobra.Command{
		Use:   "diff",
		Short: "Show changes between commits, commit and working tree, etc",
		Args:  cobra.RangeArgs(0, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	diffCommand.Flags().BoolVarP(
		&staged,
		"staged",
		"s",
		false,
		"Show diff between HEAD and Index",
	)
	return diffCommand
}
