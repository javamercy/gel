package cli

import (
	"github.com/spf13/cobra"
)

func newCatFileCommand() *cobra.Command {
	var type_ bool
	var pretty bool
	var size bool
	var exists bool

	catFileCmd := &cobra.Command{
		Use:   "cat-file <hash>",
		Short: "Display the content of a Git object",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	catFileCmd.Flags().BoolVarP(
		&type_,
		"type",
		"t",
		false,
		"Show the object type",
	)
	catFileCmd.Flags().BoolVarP(
		&pretty,
		"pretty",
		"p",
		false,
		"Pretty-print the object content",
	)
	catFileCmd.Flags().BoolVarP(
		&size,
		"size",
		"s",
		false,
		"Show the object size",
	)
	catFileCmd.Flags().BoolVarP(
		&exists,
		"exists",
		"e",
		false,
		"Check if the object exists",
	)
	return catFileCmd
}
