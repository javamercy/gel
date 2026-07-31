package cli

import (
	"github.com/spf13/cobra"
)

func newLogCommand() *cobra.Command {
	var limit int
	var oneline bool
	var since string
	var until string

	logCommand := &cobra.Command{
		Use:   "log",
		Short: "Show commit logs",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	logCommand.Flags().IntVarP(
		&limit,
		"limit",
		"n",
		0,
		"Maximum number of commits to list",
	)
	logCommand.Flags().BoolVarP(
		&oneline,
		"oneline",
		"1",
		false,
		"Show oneline commit summary",
	)
	logCommand.Flags().StringVarP(
		&since,
		"since",
		"S",
		"",
		"Only commits after (inclusive) this date",
	)
	logCommand.Flags().StringVarP(
		&until,
		"until",
		"U",
		"",
		"Only commits before (inclusive) this date",
	)
	return logCommand
}
