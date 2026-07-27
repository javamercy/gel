package cli

import (
	"Gel/internal/domain"

	"github.com/spf13/cobra"
)

var (
	commitTreeMessageFlag string
	commitTreeParentsFlag []string
)

// commitTreeCmd creates a commit object directly from a tree hash.
var commitTreeCmd = &cobra.Command{
	Use:   "commit-tree <tree-hash>",
	Short: "Create a new commit object from a tree object",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		hash, err := domain.ParseHash(args[0])
		if err != nil {
			return err
		}

		var parentHashes []domain.Hash
		for _, parent := range commitTreeParentsFlag {
			parentHash, err := domain.ParseHash(parent)
			if err != nil {
				return err
			}
			parentHashes = append(parentHashes, parentHash)
		}

		commitHash, err := commitTreeService.CommitTree(hash, commitTreeMessageFlag, parentHashes)
		if err != nil {
			return err
		}

		cmd.Println(commitHash)
		return nil
	},
}

func init() {
	commitTreeCmd.Flags().StringVarP(&commitTreeMessageFlag, "message", "m", "", "Commit message")
	commitTreeCmd.Flags().StringSliceVarP(&commitTreeParentsFlag, "parent", "p", nil, "Parent commit(s)")
	rootCmd.AddCommand(commitTreeCmd)
}
