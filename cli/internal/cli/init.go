package cli

import (
	"Gel/internal/app"
	"Gel/internal/domain"
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
)

func newInitCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "init [path]",
		Short: "Initialize a new gel repository",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := "."
			if len(args) == 1 {
				path = args[0]
			}

			absPath, err := filepath.Abs(path)
			if err != nil {
				return err
			}

			root, err := domain.NewAbsolutePath(absPath)
			if err != nil {
				return err
			}

			workspace, err := domain.NewWorkspace(root)
			if err != nil {
				return err
			}

			result, err := app.Init(workspace)
			if err != nil {
				return err
			}

			if result.Reinitialized {
				_, err := fmt.Fprintf(
					cmd.OutOrStdout(), "Reinitialized existing gel repository in %s\n",
					result.GelDir,
				)
				return err
			}

			_, err = fmt.Fprintf(
				cmd.OutOrStdout(), "Initialized empty gel repository in %s\n",
				result.GelDir,
			)
			return err
		},
	}
}
