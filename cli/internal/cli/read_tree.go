package cli

import (
	"Gel/internal/app"
	"Gel/internal/domain"
	"fmt"

	"github.com/spf13/cobra"
)

func newReadTreeCommand(provider *repositoryProvider) *cobra.Command {
	var empty bool

	readTreeCommand := &cobra.Command{
		Use:   "read-tree [--empty | <tree-hash>]",
		Short: "Read tree objects into the index",
		Args: func(cmd *cobra.Command, args []string) error {
			if empty {
				return cobra.NoArgs(cmd, args)
			}
			return cobra.ExactArgs(1)(cmd, args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			var hash domain.Hash
			if !empty {
				var err error
				hash, err = domain.ParseHash(args[0])
				if err != nil {
					return fmt.Errorf(
						"parse tree hash: %w",
						err,
					)
				}
				if hash.IsZero() {
					return fmt.Errorf("tree hash cannot be zero")
				}
			}

			repository, err := provider.Load()
			if err != nil {
				return err
			}

			readTree := app.NewReadTree(
				repository.indexStore,
				repository.objectStore,
			)
			return readTree.Run(
				app.ReadTreeInput{
					Hash:  hash,
					Empty: empty,
				},
			)
		},
	}

	readTreeCommand.Flags().BoolVar(
		&empty,
		"empty",
		false,
		"Empty the index instead of reading a tree",
	)
	return readTreeCommand
}
