package cli

import (
	"Gel/internal/app"
	"fmt"

	"github.com/spf13/cobra"
)

func newWriteTreeCommand(provider *repositoryProvider) *cobra.Command {
	return &cobra.Command{
		Use:   "write-tree",
		Short: "Write the current index as a tree object",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			repository, err := provider.Load()
			if err != nil {
				return err
			}

			writeTree := app.NewWriteTree(
				repository.indexStore,
				repository.objectStore,
			)

			result, err := writeTree.Run()
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(
				cmd.OutOrStdout(),
				result.Hash.Hex(),
			)
			return err
		},
	}
}
