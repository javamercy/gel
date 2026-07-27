package cli

import (
	"Gel/internal/domain"
	"Gel/internal/inspect"

	"github.com/spf13/cobra"
)

var (
	catFileTypeFlag   bool
	catFilePrettyFlag bool
	catFileSizeFlag   bool
	catFileExistsFlag bool
)

// catFileCmd inspects objects in the repository object database by hash.
var catFileCmd = &cobra.Command{
	Use:   "cat-file <hash>",
	Short: "Display the content of a Git object",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		hash, err := domain.ParseHash(args[0])
		if err != nil {
			return err
		}
		return catFileService.CatFile(
			cmd.OutOrStdout(),
			hash,
			inspect.CatFileOptions{
				ObjectType: catFileTypeFlag,
				Pretty:     catFilePrettyFlag,
				Size:       catFileSizeFlag,
				Exists:     catFileExistsFlag,
			},
		)
	},
}

func init() {
	catFileCmd.Flags().BoolVarP(&catFileTypeFlag, "type", "t", false, "Show the object type")
	catFileCmd.Flags().BoolVarP(&catFilePrettyFlag, "pretty", "p", false, "Pretty-print the object content")
	catFileCmd.Flags().BoolVarP(&catFileSizeFlag, "size", "s", false, "Show the object size")
	catFileCmd.Flags().BoolVarP(&catFileExistsFlag, "exists", "e", false, "Check if the object exists")
	rootCmd.AddCommand(catFileCmd)
}
