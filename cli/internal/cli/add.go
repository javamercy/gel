package cli

import (
	"Gel/internal/app"
	"fmt"

	"github.com/spf13/cobra"
)

func newAddCommand(provider *repositoryProvider) *cobra.Command {
	var dryRun bool
	var verbose bool

	addCommand := &cobra.Command{
		Use:   "add <pathspec>...",
		Short: "Add file contents to the index",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repository, err := provider.Load()
			if err != nil {
				return err
			}

			resolver, err := app.NewPathspecResolver(
				repository.workspace.RepoRoot(),
				repository.workingDir,
			)
			add := app.NewAdd(
				repository.indexStore,
				repository.objectStore,
				resolver,
			)
			result, err := add.Run(
				app.AddInput{
					Pathspecs: args,
					DryRun:    dryRun,
				},
			)
			if err != nil {
				return err
			}

			if dryRun || verbose {
				for _, path := range result.Staged {
					if _, err := fmt.Fprintf(
						cmd.OutOrStdout(),
						"ADD %s\n",
						path,
					); err != nil {
						return fmt.Errorf("write add output: %w", err)
					}
				}

				for _, path := range result.Removed {
					if _, err := fmt.Fprintf(
						cmd.OutOrStdout(),
						"REMOVE %s\n",
						path,
					); err != nil {
						return fmt.Errorf("write add output: %w", err)
					}
				}
			}
			return nil
		},
	}

	addCommand.Flags().BoolVarP(
		&dryRun,
		"dry-run",
		"n",
		false,
		"Show changes without updating the index",
	)
	addCommand.Flags().BoolVarP(
		&verbose,
		"verbose",
		"v",
		false,
		"Show staged paths",
	)
	return addCommand
}
