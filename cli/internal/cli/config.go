package cli

import (
	"Gel/internal/app"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func newConfigCommand(provider *repositoryProvider) *cobra.Command {
	var list bool

	configCommand := &cobra.Command{
		Use:   "config [section.key] [value]",
		Short: "Get or set repository configuration",
		Args: func(cmd *cobra.Command, args []string) error {
			if list {
				if len(args) != 0 {
					return errors.New("--list does not accept arguments")
				}
				return nil
			}
			if len(args) < 1 || len(args) > 2 {
				return errors.New(
					"config expects <section.key> or <section.key> <value>",
				)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			repository, err := provider.Load()
			if err != nil {
				return err
			}

			config, err := app.NewConfig(repository.configStore)
			if err != nil {
				return err
			}

			if list {
				entries, err := config.List()
				if err != nil {
					return err
				}
				for _, entry := range entries {
					if _, err := fmt.Fprintf(
						cmd.OutOrStdout(),
						"%s.%s=%s\n",
						entry.Section,
						entry.Key,
						entry.Value,
					); err != nil {
						return fmt.Errorf(
							"write config: %w",
							err,
						)
					}
				}
				return nil
			}

			section, key, err := parseConfigKey(args[0])
			if err != nil {
				return err
			}

			if len(args) == 2 {
				return config.Set(section, key, args[1])
			}

			value, found, err := config.Get(section, key)
			if err != nil {
				return err
			}
			if !found {
				return fmt.Errorf("config key %q is not set", args[0])
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), value)
			return err
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

func parseConfigKey(value string) (string, string, error) {
	section, key, found := strings.Cut(value, ".")
	if !found || section == "" || key == "" || strings.Contains(key, ".") {
		return "", "", fmt.Errorf(
			"invalid config key %q: expected section.key",
			value,
		)
	}
	return section, key, nil
}
