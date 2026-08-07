package cli

import (
	"Gel/internal/app"
	"Gel/internal/domain"
	"errors"
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

func newCatFileCommand(provider *repositoryProvider) *cobra.Command {
	var type_ bool
	var pretty bool
	var size bool
	var exists bool

	catFileCmd := &cobra.Command{
		Use:   "cat-file <hash>",
		Short: "Display the content of a Git object",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if exists && (type_ || pretty || size) {
				return errors.New(
					"cannot use --exists with --type, --pretty, or --size",
				)
			}
			if !exists && !type_ && !size && !pretty {
				return errors.New(
					"specify at least one of --type, --size, --pretty, or --exists",
				)
			}

			hash, err := domain.ParseHash(args[0])
			if err != nil {
				return fmt.Errorf("parse hash: %w", err)
			}

			repository, err := provider.Load()
			if err != nil {
				return fmt.Errorf("load repository: %w", err)
			}

			catFile := app.NewCatFile(repository.objectStore)
			result, err := catFile.Run(app.CatFileInput{Hash: hash})
			if err != nil {
				return fmt.Errorf("cat-file: %w", err)
			}

			if exists {
				return nil
			}

			out := cmd.OutOrStdout()

			if type_ {
				if _, err := fmt.Fprintln(out, result.Object.Type()); err != nil {
					return fmt.Errorf("write object type: %w", err)
				}
			}
			if size {
				if _, err := fmt.Fprintln(out, len(result.Body)); err != nil {
					return fmt.Errorf("write object size: %w", err)
				}
			}
			if pretty {
				if err := writePrettyObject(out, result.Object, result.Body); err != nil {
					return err
				}
			}
			return nil
		},
	}

	catFileCmd.Flags().BoolVarP(
		&type_,
		"type",
		"t",
		false,
		"Show the object type",
	)
	catFileCmd.Flags().BoolVarP(
		&pretty,
		"pretty",
		"p",
		false,
		"Pretty-print the object content",
	)
	catFileCmd.Flags().BoolVarP(
		&size,
		"size",
		"s",
		false,
		"Show the object size",
	)
	catFileCmd.Flags().BoolVarP(
		&exists,
		"exists",
		"e",
		false,
		"Check if the object exists",
	)

	return catFileCmd
}

func writePrettyObject(
	out io.Writer,
	object domain.Object,
	body []byte,
) error {
	switch obj := object.(type) {
	case *domain.Blob, *domain.Commit:
		return writeAll(out, body)

	case *domain.Tree:
		for _, entry := range obj.Entries() {
			objectType, err := entry.Mode().ObjectType()
			if err != nil {
				return fmt.Errorf(
					"determine object type for tree entry %q: %w",
					entry.Name(),
					err,
				)
			}
			if _, err := fmt.Fprintf(
				out,
				"%s %s %s\t%s\n",
				entry.Mode(),
				objectType,
				entry.Hash(),
				entry.Name(),
			); err != nil {
				return fmt.Errorf(
					"write tree entry %q: %w",
					entry.Name(),
					err,
				)
			}
		}
	default:
		return fmt.Errorf(
			"pretty-print unsupported object type %T",
			object,
		)
	}
	return nil
}

func writeAll(out io.Writer, content []byte) error {
	n, err := out.Write(content)
	if err != nil {
		return fmt.Errorf("write object content: %w", err)
	}
	if n != len(content) {
		return io.ErrShortWrite
	}
	return nil
}
