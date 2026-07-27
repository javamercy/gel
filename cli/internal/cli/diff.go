package cli

import (
	"Gel/internal/core"
	"Gel/internal/diff"
	"Gel/internal/domain"

	"github.com/spf13/cobra"
)

var (
	diffStagedFlag bool
)

// diffCmd shows textual diffs between snapshots such as index, working tree, and commits.
var diffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Show changes between commits, commit and working tree, etc",
	Args:  cobra.RangeArgs(0, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		var results []*diff.DiffResult
		var err error

		switch len(args) {
		case 0:
			mode := diff.DiffModeIndexVsWorkingTree
			if diffStagedFlag {
				mode = diff.DiffModeHEADVsIndex
			}
			results, err = diffService.Diff(diff.DiffOptions{Mode: mode})
		case 1:
			arg := args[0]
			if arg == domain.HeadFileName {
				results, err = diffService.Diff(diff.DiffOptions{Mode: diff.DiffModeHeadVsWorkingTree})
			} else {
				baseCommitHash, hashErr := domain.ParseHash(arg)
				if hashErr != nil {
					return hashErr
				}
				results, err = diffService.Diff(
					diff.DiffOptions{
						Mode: diff.DiffModeCommitVsWorkingTree, BaseCommitHash: baseCommitHash,
					},
				)
			}
		case 2:
			baseCommitHash, hashErr := domain.ParseHash(args[0])
			if hashErr != nil {
				return err
			}
			targetCommitHash, hashErr := domain.ParseHash(args[1])
			if hashErr != nil {
				return err
			}
			results, err = diffService.Diff(
				diff.DiffOptions{
					Mode: diff.DiffModeCommitVsCommit, BaseCommitHash: baseCommitHash,
					TargetCommitHash: targetCommitHash,
				},
			)
		}
		if err != nil {
			return err
		}
		printDiffResults(cmd, results)
		return nil
	},
}

// printDiffResults prints each file-level diff result with file header and hunks.
func printDiffResults(cmd *cobra.Command, results []*diff.DiffResult) {
	for _, result := range results {
		switch result.Status {
		case diff.DiffStatusAdded:
			printAddedFileHeader(cmd, result.OldPath, result.NewPath, result.NewHash)
		case diff.DiffStatusModified:
			printModifiedFileHeader(cmd, result.OldPath, result.NewPath, result.OldHash, result.NewHash)
		case diff.DiffStatusDeleted:
			printDeletedFileHeader(cmd, result.OldPath, result.OldHash)
		}
		printHunks(cmd, result.Hunks)
	}
}

// printAddedFileHeader prints patch header lines for a newly added file.
func printAddedFileHeader(cmd *cobra.Command, oldPath, newPath domain.NormalizedPath, hash domain.Hash) {
	cmd.Printf("%sdiff --gel a/%s b/%s%s\n", core.ColorBold, oldPath, newPath, core.ColorReset)
	cmd.Printf("%snew file mode %s%s\n", core.ColorBold, domain.FileModeRegular, core.ColorReset)
	cmd.Printf("%sindex 00000000..%s%s\n", core.ColorBold, hash, core.ColorReset)
	cmd.Printf("%s--- /dev/null%s\n", core.ColorBold, core.ColorReset)
	cmd.Printf("%s+++ b/%s%s\n", core.ColorBold, newPath, core.ColorReset)
}

// printDeletedFileHeader prints patch header lines for a deleted file.
func printDeletedFileHeader(cmd *cobra.Command, oldPath domain.NormalizedPath, oldHash domain.Hash) {
	cmd.Printf("%sdiff --gel a/%s b/%s%s\n", core.ColorBold, oldPath, oldPath, core.ColorReset)
	cmd.Printf("%sdeleted file mode %s%s\n", core.ColorBold, domain.FileModeRegular, core.ColorReset)
	cmd.Printf("%sindex %s..00000000%s\n", core.ColorBold, oldHash, core.ColorReset)
	cmd.Printf("%s--- a/%s%s\n", core.ColorBold, oldPath, core.ColorReset)
	cmd.Printf("%s+++ /dev/null%s\n", core.ColorBold, core.ColorReset)
}

// printModifiedFileHeader prints patch header lines for a modified file.
func printModifiedFileHeader(
	cmd *cobra.Command,
	oldPath, newPath domain.NormalizedPath,
	oldHash, newHash domain.Hash,
) {
	cmd.Printf("%sdiff --gel a/%s b/%s%s\n", core.ColorBold, oldPath, newPath, core.ColorReset)
	cmd.Printf("%sindex %s..%s %s%s\n", core.ColorBold, oldHash, newHash, domain.FileModeRegular, core.ColorReset)
	cmd.Printf("%s--- a/%s%s\n", core.ColorBold, oldPath, core.ColorReset)
	cmd.Printf("%s+++ b/%s%s\n", core.ColorBold, newPath, core.ColorReset)
}

// printHunks prints unified-style hunk ranges and per-line operations.
func printHunks(cmd *cobra.Command, hunks []*diff.Hunk) {
	for _, hunk := range hunks {
		cmd.Printf("@@ -%d,%d +%d,%d @@\n", hunk.OldStart, hunk.OldLength, hunk.NewStart, hunk.NewLength)
		for _, line := range hunk.Lines {
			var prefix string
			var color string
			switch line.OperationType {
			case diff.OpTypeMatch:
				prefix = " "
				color = ""
			case diff.OpTypeInsertion:
				prefix = "+ "
				color = core.ColorGreen
			case diff.OpTypeDeletion:
				prefix = "- "
				color = core.ColorRed
			}
			cmd.Printf("%s%s%s%s\n", color, prefix, line.Content, core.ColorReset)
		}
	}
}

// init registers the diff command and staged flag.
func init() {
	diffCmd.Flags().BoolVarP(
		&diffStagedFlag, "staged", "s", false,
		"Show diff between HEAD and Index",
	)
	rootCmd.AddCommand(diffCmd)
}
