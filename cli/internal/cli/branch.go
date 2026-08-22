package cli

import (
	"Gel/internal/app"
	"fmt"

	"github.com/spf13/cobra"
)

func newBranchCommand(provider *repositoryProvider) *cobra.Command {
	var deleteBranch bool
	var forceDelete bool

	branchCmd := &cobra.Command{
		Use:   "branch [-d | -D] [<branch-name> [<start-point>]]",
		Short: "List, create, or delete branches",
		Args: func(cmd *cobra.Command, args []string) error {
			if deleteBranch || forceDelete {
				return cobra.ExactArgs(1)(cmd, args)
			}
			if len(args) == 0 {
				return nil
			}
			return cobra.RangeArgs(1, 2)(cmd, args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			repository, err := provider.Load()
			if err != nil {
				return err
			}

			branch := app.NewBranch(
				repository.refStore,
				repository.objectStore,
			)

			switch {
			case len(args) == 0:
				result, err := branch.List()
				if err != nil {
					return err
				}

				for _, branch := range result.Branches {
					prefix := "  "
					if branch.IsCurrent {
						prefix = "* "
					}
					if _, err = fmt.Fprintf(
						cmd.OutOrStdout(),
						"%s%s\n",
						prefix,
						branch.Name,
					); err != nil {
						return fmt.Errorf(
							"write branch list: %w",
							err,
						)
					}
				}

			case len(args) == 1:
				if deleteBranch || forceDelete {
					return branch.Delete(args[0], forceDelete)
				}
				return branch.Create(args[0], "")

			default:
				return branch.Create(args[0], args[1])
			}
			return nil
		},
	}

	branchCmd.Flags().BoolVarP(
		&deleteBranch,
		"delete",
		"d",
		false,
		"Delete the branch",
	)
	branchCmd.Flags().BoolVarP(
		&forceDelete,
		"force-delete",
		"D",
		false,
		"Delete even if the branch is not fully merged",
	)
	return branchCmd
}
