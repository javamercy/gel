package cli

import "github.com/spf13/cobra"

func newCommitCommand() *cobra.Command {
	var message string
	commitCommand := &cobra.Command{
		Use:   "commit",
		Short: "Record changes to the repository",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	commitCommand.Flags().StringVarP(
		&message,
		"message",
		"m",
		"",
		"Commit message",
	)
	return commitCommand
}
