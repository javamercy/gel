package cli

import (
	"Gel/internal/domain"
	"Gel/internal/tree"

	"github.com/spf13/cobra"
)

var (
	lsTreeRecursiveFlag bool
	lsTreeShowTreesFlag bool
	lsTreeNameOnlyFlag  bool
)

// lsTreeCmd lists entries from a tree object by hash.
var lsTreeCmd = &cobra.Command{
	Use:   "ls-tree",
	Short: "List the contents of a tree",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		hash, err := domain.ParseHash(args[0])
		if err != nil {
			return err
		}
		contents, err := lsTreeService.LsTree(
			hash, tree.LsTreeOptions{
				Recursive: lsTreeRecursiveFlag,
				ShowTrees: lsTreeShowTreesFlag,
				NameOnly:  lsTreeNameOnlyFlag,
			},
		)
		if err != nil {
			return err
		}
		for _, result := range contents {
			cmd.Println(result)
		}
		return nil
	},
}

func init() {
	lsTreeCmd.Flags().BoolVarP(&lsTreeRecursiveFlag, "recursive", "r", false, "Recursively list subtrees")
	lsTreeCmd.Flags().BoolVarP(&lsTreeShowTreesFlag, "show-trees", "t", false, "Show tree objects in the listing")
	lsTreeCmd.Flags().BoolVarP(&lsTreeNameOnlyFlag, "name-only", "n", false, "Show only names of the entries")
	rootCmd.AddCommand(lsTreeCmd)
}
