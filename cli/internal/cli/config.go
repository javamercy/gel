package cli

import (
	"github.com/spf13/cobra"
)

func newConfigCommand() *cobra.Command {
	var list bool

	configCommand := &cobra.Command{
		Use:   "config [key] [value]",
		Short: "Get or set repository or global options",
		Args:  cobra.MaximumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	configCommand.Flags().BoolVarP(
		&list,
		"list",
		"l",
		false,
		"List all config values",
	)
	return configCommand
}
