package cli

import (
	"Gel/internal/domain"

	"github.com/spf13/cobra"
)

// readTreeCmd loads a tree object into the index.
var readTreeCmd = &cobra.Command{
	Use:   "read-tree <tree-hash>",
	Short: "Read tree objects into the index",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		hash, err := domain.ParseHash(args[0])
		if err != nil {
			return err
		}
		return readTreeService.ReadTree(hash)
	},
}

func init() {
	rootCmd.AddCommand(readTreeCmd)
}
