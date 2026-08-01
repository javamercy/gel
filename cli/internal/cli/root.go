package cli

import (
	"Gel/internal/core"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var commandsWithoutRepository = map[string]bool{
	"init": true,
	"help": true,
}

func newRootCommand() *cobra.Command {
	rootCommand := &cobra.Command{
		Use:   "Gel",
		Short: "An Agentic Version Control System",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if commandsWithoutRepository[cmd.Name()] {
				return nil
			}
			return initialize()
		},
	}

	rootCommand.AddCommand(
		newInitCommand(),
	)
	return rootCommand
}

func Execute() int {
	cmd := newRootCommand()
	cmd.SilenceErrors = true

	if err := cmd.Execute(); err != nil {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "%v\n", err)
		return 1
	}
	return 0
}

func initialize() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	workspace, err := core.DiscoverWorkspace(cwd)
	if err != nil {
		return err
	}

	_ = workspace
	return nil
}
