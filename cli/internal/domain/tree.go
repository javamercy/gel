package domain

import (
	"bytes"
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
	if err := validateTreeEntries(entriesCopy); err != nil {
		return nil, fmt.Errorf(
			"%w: construct tree: %w",
			ErrInvalidTree,
			err,
		)
	}

	slices.SortFunc(entriesCopy, compareTreeEntries)

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

// EncodeTree encodes tree using the canonical Gel tree-body format.
func EncodeTree(tree *Tree) ([]byte, error) {
	if tree == nil {
		return nil, fmt.Errorf(
			"%w: tree is nil",
			ErrInvalidTree,
		)
	}
	if err := validateTreeEntries(tree.entries); err != nil {
		return nil, fmt.Errorf(
			"%w: validate entries: %w",
			ErrInvalidTree,
			err,
		)
	}
	return encodeTreeEntries(tree.entries), nil
}

// DecodeTree decodes and validates a canonical Gel tree body.
//
// Encoded entries must already be in canonical order.
func DecodeTree(data []byte) (*Tree, error) {
	entries, err := decodeTreeEntries(data)
	if err != nil {
		return nil, fmt.Errorf(
			"%w: decode entries: %w",
			ErrInvalidTree,
			err,
		)
	}
	return &Tree{
		entries: entries,
	}, nil
}

func encodeTreeEntries(entries []TreeEntry) []byte {
	encoded := make([]byte, 0)
	for _, entry := range entries {
		encoded = append(
			encoded,
			[]byte(fmt.Sprintf(
				"%s %s\x00%s",
				entry.mode.String(),
				entry.name,
				entry.hash.String(),
			))...,
		)
	}
	return encoded
}

func decodeTreeEntries(body []byte) ([]TreeEntry, error) {
	var entries []TreeEntry
	seenNames := make(map[string]struct{})
	previousSortKey := ""
	hasPrevious := false

	for offset := 0; offset < len(body); {
		entryOffset := offset
		spaceIndex := bytes.IndexByte(body[offset:], ' ')
		if spaceIndex == -1 {
			return nil, fmt.Errorf(
				"entry at offset %d: missing mode separator",
				entryOffset,
			)
		}

		modeEnd := offset + spaceIndex
		modeText := string(body[offset:modeEnd])
		mode, err := ParseFileMode(modeText)
		if err != nil {
			return nil, fmt.Errorf(
				"entry at offset %d: %w",
				entryOffset,
				err,
			)
		}

		offset = modeEnd + 1
		nameStart := offset
		nulIndex := bytes.IndexByte(body[offset:], 0)
		if nulIndex == -1 {
			return nil, fmt.Errorf(
				"entry at offset %d: missing name terminator",
				entryOffset,
			)
		}

		nameEnd := offset + nulIndex
		name := string(body[nameStart:nameEnd])
		offset = nameEnd + 1

		if len(body)-offset < HashByteLength {
			return nil, fmt.Errorf(
				"entry %q at offset %d: truncated hash: got %d bytes, want %d",
				name,
				entryOffset,
				len(body)-offset,
				HashByteLength,
			)
		}

		hash, err := NewHashFromBytes(body[offset : offset+HashByteLength])
		if err != nil {
			return nil, fmt.Errorf(
				"entry %q at offset %d: parse hash: %w",
				name,
				entryOffset,
				err,
			)
		}

		offset += HashByteLength

		entry := TreeEntry{
			mode: mode,
			hash: hash,
			name: name,
		}
		if err := validateTreeEntry(entry); err != nil {
			return nil, fmt.Errorf(
				"entry at offset %d: %w",
				entryOffset,
				err,
			)
		}
		if _, exists := seenNames[name]; exists {
			return nil, fmt.Errorf(
				"duplicate entry name %q",
				name,
			)
		}

		seenNames[name] = struct{}{}

		sortKey := treeEntrySortKey(entry.name, entry.mode.IsDirectory())
		if hasPrevious && strings.Compare(previousSortKey, sortKey) > 0 {
			return nil, fmt.Errorf(
				"entries are not in canonical order: %q before %q",
				previousSortKey,
				sortKey,
			)
		}

		previousSortKey = sortKey
		hasPrevious = true
		entries = append(entries, entry)
	}
	return entries, nil
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

		prev := entries[i-1]
		curr := entries[i]
		if compareTreeEntries(prev, curr) > 0 {
			return fmt.Errorf(
				"entry %q appears before earlier entry %q",
				prev.name,
				curr.name,
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
