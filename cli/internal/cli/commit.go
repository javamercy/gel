package cli

import (
	"Gel/internal/app"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

func newCommitCommand(provider *repositoryProvider) *cobra.Command {
	var message string
	commitCommand := &cobra.Command{
		Use:   "commit -m <message>",
		Short: "Record changes to the repository",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !cmd.Flags().Changed("message") {
				return errors.New("commit message is required")
			}

			repository, err := provider.Load()
			if err != nil {
				return err
			}

			commit := app.NewCommit(
				repository.refStore,
				repository.indexStore,
				repository.objectStore,
				repository.configStore,
			)

			result, err := commit.Run(
				app.CommitInput{Message: message},
			)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(
				cmd.OutOrStdout(),
				result.CommitHash.Hex(),
			)
			return err
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
