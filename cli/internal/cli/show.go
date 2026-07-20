package cli

import (
	"Gel/internal/inspect"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// init registers the show command.
func init() {
	rootCmd.AddCommand(showCmd)
}

// showCmd displays commit, tree, or blob objects for a reference or object hash.
var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Show various types of objects",
	Args:  cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		objectRef := ""
		if len(args) == 1 {
			objectRef = args[0]
		}

		result, err := showService.Show(objectRef, inspect.ShowOptions{})
		if err != nil {
			return err
		}
		return printShowResult(cmd, result)
	},
}

// printShowResult dispatches object-specific show output rendering.
func printShowResult(cmd *cobra.Command, result *inspect.ShowResult) error {
	switch result.Mode {
	case inspect.ShowModeCommit:
		return printShowCommit(cmd, result.Commit)
	case inspect.ShowModeTree:
		printShowTree(cmd, result.Tree)
		return nil
	case inspect.ShowModeBlob:
		printShowBlob(cmd, result.Blob)
		return nil
	default:
		return fmt.Errorf("show: unsupported show result mode")
	}
}

// printShowCommit renders commit header, message, and patch output.
func printShowCommit(cmd *cobra.Command, r *inspect.ShowCommitResult) error {
	date := ""

	if r.Branch != "" {
		cmd.Printf("commit %s (%s)\n", r.Hash, r.Branch)
	} else {
		cmd.Printf("commit %s\n", r.Hash)
	}
	cmd.Printf("Author: %s <%s>\n", r.Commit.Author().Name(), r.Commit.Author().Email())
	cmd.Printf("Date:   %s\n\n", date)

	for _, line := range strings.Split(r.Commit.Message(), "\n") {
		cmd.Printf("    %s\n", line)
	}
	cmd.Printf("\n")

	printDiffResults(cmd, r.Diff)
	return nil
}

// printShowTree renders tree header and direct entry names.
func printShowTree(cmd *cobra.Command, r *inspect.ShowTreeResult) {
	cmd.Printf("tree %s\n\n", r.Hash)
	for _, entry := range r.TreeEntries {
		name := entry.Name()
		if entry.Mode().IsDirectory() {
			name += "/"
		}
		cmd.Printf("%s\n", name)
	}
}

// printShowBlob renders raw blob content.
func printShowBlob(cmd *cobra.Command, r *inspect.ShowBlobResult) {
	cmd.Print(string(r.Body))
}
