package cli

import (
	"Gel/internal/app"
	"Gel/internal/domain"
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
)

func newHashObjectCommand(provider *repositoryProvider) *cobra.Command {
	var write bool

	hashObjectCommand := &cobra.Command{
		Use:   "hash-object <file>...",
		Short: "Compute the hash of a file",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			paths := make([]domain.AbsolutePath, 0, len(args))
			for _, arg := range args {
				abs, err := filepath.Abs(arg)
				if err != nil {
					return err
				}

				path, err := domain.NewAbsolutePath(abs)
				if err != nil {
					return err
				}
				paths = append(paths, path)
			}

			repository, err := provider.Load()
			if err != nil {
				return err
			}

			hashObject := app.NewHashObject(repository.objectStore)
			result, err := hashObject.Run(
				app.HashObjectInput{
					Paths: paths,
					Write: write,
				},
			)

			for _, hash := range result.Hashes {
				if _, err := fmt.Fprintln(cmd.OutOrStdout(), hash.Hex()); err != nil {
					return fmt.Errorf(
						"write output: %w",
						err,
					)
				}
			}
			return err
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
