package cli

import (
	"Gel/internal/domain"
	"Gel/internal/staging"
	"fmt"

	"github.com/spf13/cobra"
)

var (
	updateIndexAddFlag    bool
	updateIndexRemoveFlag bool
)

// updateIndexCmd updates index entries directly for the provided path arguments.
var updateIndexCmd = &cobra.Command{
	Use:   "update-index <file>...",
	Short: "Update the index with the current state of the working directory",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		normPaths := make([]domain.NormalizedPath, len(args))
		for i, path := range args {
			resolved, err := repositoryPathResolver.Resolve(path)
			if err != nil {
				return err
			}
			normPaths[i] = resolved.Normalized
		}

		paths, err := updateIndexService.UpdateIndex(
			normPaths,
			staging.UpdateIndexOptions{
				Add:    updateIndexAddFlag,
				Remove: updateIndexRemoveFlag,
				Write:  true,
			},
		)
		if err != nil {
			return err
		}
		for _, path := range paths {
			fmt.Println(path)
		}
		return nil
	},
}

func init() {
	updateIndexCmd.Flags().BoolVarP(&updateIndexAddFlag, "add", "a", false, "Add specified files to the index")
	updateIndexCmd.Flags().BoolVarP(
		&updateIndexRemoveFlag, "remove", "r", false, "Remove specified files from the index",
	)
	rootCmd.AddCommand(updateIndexCmd)
}
