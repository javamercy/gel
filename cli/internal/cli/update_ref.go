package cli

import (
	"Gel/internal/domain"
	"fmt"

	"github.com/spf13/cobra"
)

var (
	updateRefDeleteFlag bool
)

// updateRefCmd updates or deletes direct refs under refs/.
// Modes:
//   - update-ref <ref> <new-hash>
//   - update-ref <ref> <new-hash> <old-hash>
//   - update-ref -d <ref>
//   - update-ref -d <ref> <old-hash>
var updateRefCmd = &cobra.Command{
	Use:   "update-ref",
	Short: "Update a reference",
	Args:  cobra.RangeArgs(1, 3),
	RunE: func(cmd *cobra.Command, args []string) error {
		ref := args[0]
		switch len(args) {
		case 1:
			if updateRefDeleteFlag {
				return updateRefService.Delete(ref)
			}
			return fmt.Errorf("update-ref: missing new hash argument")
		case 2:
			hashStr := args[1]
			hash, err := domain.ParseHash(hashStr)
			if err != nil {
				return fmt.Errorf("update-ref: %w", err)
			}
			if updateRefDeleteFlag {
				return updateRefService.DeleteSafe(ref, hash)
			}
			return updateRefService.Update(ref, hash)
		case 3:
			if updateRefDeleteFlag {
				return fmt.Errorf("update-ref: --delete accepts at most one hash argument")
			}

			newHashStr := args[1]
			oldHashStr := args[2]
			newHash, err := domain.ParseHash(newHashStr)
			if err != nil {
				return fmt.Errorf("update-ref: %w", err)
			}
			oldHash, err := domain.ParseHash(oldHashStr)
			if err != nil {
				return fmt.Errorf("update-ref: %w", err)
			}
			return updateRefService.UpdateSafe(ref, newHash, oldHash)
		}
		return nil
	},
}

func init() {
	updateRefCmd.Flags().BoolVarP(
		&updateRefDeleteFlag, "delete", "d", false, "Delete the reference instead of updating it",
	)
	rootCmd.AddCommand(updateRefCmd)
}
