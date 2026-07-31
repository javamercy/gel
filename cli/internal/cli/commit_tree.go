package cli

import (
	"github.com/spf13/cobra"
)

func newCommitTreeCommand() *cobra.Command {
	var message string
	var parents []string

	commitTreeCommand := &cobra.Command{
		Use:   "commit-tree <tree-hash>",
		Short: "Create a new commit object from a tree object",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	commitTreeCommand.Flags().StringVarP(
		&message,
		"message",
		"m",
		"",
		"Commit message",
	)
	commitTreeCommand.Flags().StringSliceVarP(
		&parents,
		"parent",
		"p",
		nil,
		"Parent commit(s)",
	)
	return commitTreeCommand
}
