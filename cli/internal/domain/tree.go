package domain

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"unicode/utf8"
)

// ErrInvalidTree indicates malformed, inconsistent, or non-canonical tree data.
var ErrInvalidTree = errors.New("invalid tree")

// TreeEntry represents a single entry within a tree object.
type TreeEntry struct {
	mode FileMode
	hash Hash
	name string
}

// NewTreeEntry validates and constructs a tree entry.
func NewTreeEntry(
	mode FileMode,
	hash Hash,
	name string,
) (TreeEntry, error) {
	entry := TreeEntry{
		mode: mode,
		hash: hash,
		name: name,
	}

	if err := validateTreeEntry(entry); err != nil {
		return TreeEntry{}, fmt.Errorf(
			"create tree entry: %w",
			err,
		)
	}
	return entry, nil
}

// Mode returns the entry's gel file mode.
func (e TreeEntry) Mode() FileMode {
	return e.mode
}

// Hash returns the entry's object hash.
func (e TreeEntry) Hash() Hash {
	return e.hash
}

// Name returns the entry's file or directory name.
func (e TreeEntry) Name() string {
	return e.name
}

// Tree represents an immutable directory structure.
//
// The zero value is a valid empty tree.
type Tree struct {
	entries []TreeEntry
}

// NewTreeFromEntries validates entries and constructs a canonical tree.
//
// The provided slice is copied. Entries may be provided in any order.
func NewTreeFromEntries(entries []TreeEntry) (*Tree, error) {
	entriesCopy := slices.Clone(entries)
	slices.SortFunc(entriesCopy, compareTreeEntries)

	if err := validateTreeEntries(entriesCopy); err != nil {
		return nil, fmt.Errorf(
			"%w: construct tree: %w",
			ErrInvalidTree,
			err,
		)
	}
	return &Tree{
		entries: entriesCopy,
	}, nil
}

// Type returns ObjectTypeTree.
func (t *Tree) Type() ObjectType {
	return ObjectTypeTree
}

func (t *Tree) isObject() {}

// Entries returns a copy of the entries in canonical order.
func (t *Tree) Entries() []TreeEntry {
	return slices.Clone(t.entries)
}

func validateTreeEntries(entries []TreeEntry) error {
	seenNames := make(map[string]struct{}, len(entries))
	for i, entry := range entries {
		if err := validateTreeEntry(entry); err != nil {
			return fmt.Errorf(
				"entry %d: %w",
				i,
				err,
			)
		}
		if _, exists := seenNames[entry.name]; exists {
			return fmt.Errorf(
				"duplicate entry name %q",
				entry.name,
			)
		}

		seenNames[entry.name] = struct{}{}

		if i > 0 && compareTreeEntries(entries[i-1], entry) > 0 {
			return fmt.Errorf(
				"entry %q appears before earlier entry %q",
				entries[i-1].name,
				entry.name,
			)
		}
	}
	return nil
}

func validateTreeEntry(entry TreeEntry) error {
	if !entry.mode.IsValid() {
		return fmt.Errorf(
			"invalid mode %#o: %w",
			uint32(entry.mode),
			ErrInvalidFileMode,
		)
	}
	if entry.hash.IsZero() {
		return fmt.Errorf(
			"entry hash cannot be zero",
		)
	}
	if err := validateTreeEntryName(entry.name); err != nil {
		return err
	}
	return nil
}

func validateTreeEntryName(name string) error {
	switch {
	case name == "":
		return fmt.Errorf("name is empty")
	case name == "." || name == "..":
		return fmt.Errorf(
			"traversal component %q is not allowed",
			name,
		)
	case !utf8.ValidString(name):
		return fmt.Errorf(
			"name %q is not valid UTF-8",
			name,
		)
	case strings.ContainsRune(name, '/'):
		return fmt.Errorf(
			"%q contains slash",
			name,
		)
	case strings.ContainsRune(name, '\\'):
		return fmt.Errorf(
			"%q contains backslash",
			name,
		)
	case strings.ContainsRune(name, 0):
		return fmt.Errorf(
			"%q contains NUL",
			name,
		)
	default:
		return nil
	}
}

func compareTreeEntries(a, b TreeEntry) int {
	return strings.Compare(
		treeEntrySortKey(
			a.name,
			a.mode.IsDirectory(),
		),
		treeEntrySortKey(
			b.name,
			b.mode.IsDirectory(),
		),
	)
}

func treeEntrySortKey(name string, isDirectory bool) string {
	if isDirectory {
		return name + "/"
	}
	return name
}
