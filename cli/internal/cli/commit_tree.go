package cli

import (
	"Gel/internal/app"
	"Gel/internal/domain"
	"fmt"

	"github.com/spf13/cobra"
)

func newCommitTreeCommand(provider *repositoryProvider) *cobra.Command {
	var message string
	var parents []string

	commitTreeCommand := &cobra.Command{
		Use:   "commit-tree <tree-hash>",
		Short: "Create a new commit object from a tree object",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			treeHash, err := domain.ParseHash(args[0])
			if err != nil {
				return fmt.Errorf("parse tree hash: %w", err)
			}

			parentHashes := make([]domain.Hash, 0, len(parents))
			for i, value := range parents {
				hash, err := domain.ParseHash(value)
				if err != nil {
					return fmt.Errorf(
						"parse parent %d hash %q: %w",
						i,
						value,
						err,
					)
				}
				parentHashes = append(parentHashes, hash)
			}

			repository, err := provider.Load()
			if err != nil {
				return err
			}

			commitTree := app.NewCommitTree(
				repository.configStore,
				repository.objectStore,
			)
			result, err := commitTree.Run(
				app.CommitTreeInput{
					TreeHash:     treeHash,
					ParentHashes: parentHashes,
					Message:      message,
				},
			)
			if err != nil {
				return err
			}

			if _, err := fmt.Fprintln(
				cmd.OutOrStdout(),
				result.CommitHash.Hex(),
			); err != nil {
				return fmt.Errorf("write commit hash: %w", err)
			}
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
	commitTreeCommand.Flags().StringArrayVarP(
		&parents,
		"parent",
		"p",
		nil,
		"Parent commit; may be repeated",
	)
	return commitTreeCommand
}
