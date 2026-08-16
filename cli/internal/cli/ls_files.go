package cli

import (
	"Gel/internal/app"
	"fmt"

	"github.com/spf13/cobra"
)

func newLsFilesCommand(provider *repositoryProvider) *cobra.Command {
	var stage bool
	var unmerged bool
	var zeroTerminated bool

	lsFilesCommand := &cobra.Command{
		Use:   "ls-files [-s|--stage] [-u|--unmerged] [-z] [<path>...]",
		Short: "List files in the index",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			repository, err := provider.Load()
			if err != nil {
				return err
			}

			lsFiles, err := app.NewLsFiles(
				repository.indexStore,
				repository.workspace.RepoRoot(),
				repository.workingDir,
			)
			if err != nil {
				return err
			}

			result, err := lsFiles.Run(
				app.LsFilesInput{
					Pathspecs: args,
					Unmerged:  unmerged,
				},
			)
			if err != nil {
				return err
			}

			terminator := "\n"
			if zeroTerminated {
				terminator = "\x00"
			}

			for _, entry := range result.Entries {
				if stage || unmerged {
					_, err = fmt.Fprintf(
						cmd.OutOrStdout(),
						"%s %s %d\t%s%s",
						entry.Mode().String(),
						entry.Hash().Hex(),
						entry.Stage(),
						entry.Path().String(),
						terminator,
					)
				} else {
					_, err = fmt.Fprint(
						cmd.OutOrStdout(),
						entry.Path().String(),
						terminator,
					)
				}

				if err != nil {
					return fmt.Errorf(
						"write ls-files output: %w",
						err,
					)
				}
			}
			return nil
		},
	}

	lsFilesCommand.Flags().BoolVarP(
		&stage,
		"stage",
		"s",
		false,
		"Show mode, object hash, and stage",
	)
	lsFilesCommand.Flags().BoolVarP(
		&unmerged,
		"unmerged",
		"u",
		false,
		"Show unmerged entries with stage details",
	)
	lsFilesCommand.Flags().BoolVarP(
		&zeroTerminated,
		"zero",
		"z",
		false,
		"Terminate entries with NUL instead of newline",
	)
	return lsFilesCommand
}
