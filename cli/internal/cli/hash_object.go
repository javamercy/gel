package cli

import (
	"github.com/spf13/cobra"
)

func newHashObjectCommand() *cobra.Command {
	var write bool

	hashObjectCommand := &cobra.Command{
		Use:   "hash-object <file>...",
		Short: "Compute the hash of a file",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	hashObjectCommand.Flags().BoolVarP(
		&write,
		"write",
		"w",
		false,
		"Write the object to the object database",
	)
	return hashObjectCommand
}
