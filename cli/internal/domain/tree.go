package domain

import (
	"bytes"
	"errors"
	"fmt"
	"slices"
	"strings"
)

var (
	// ErrInvalidTree indicates malformed, inconsistent, or non-canonical tree data.
	ErrInvalidTree = errors.New("invalid tree")
)

// TreeEntry represents a single entry within a tree object.
// Each entry corresponds to a file or subdirectory.
type TreeEntry struct {
	mode FileMode
	hash Hash
	name string
}

// NewTreeEntry validates and constructs a tree entry.
func NewTreeEntry(mode FileMode, hash Hash, name string) (TreeEntry, error) {
	if !mode.IsValid() {
		return TreeEntry{}, fmt.Errorf(
			"%w: %#o",
			ErrInvalidFileMode,
			uint32(mode),
		)
	}
	if err := validateTreeEntryName(name); err != nil {
		return TreeEntry{}, err
	}
	return TreeEntry{
		mode: mode,
		hash: hash,
		name: name,
	}, nil
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

// Tree represents a directory structure stored in the object database.
type Tree struct {
	body    []byte
	entries []TreeEntry
}

// ParseTree parses and validates a serialized tree body.
//
// The input is defensively copied. Serialized entries must already be in canonical order.
func ParseTree(body []byte) (*Tree, error) {
	bodyCopy := bytes.Clone(body)
	entries, err := parseTreeEntries(bodyCopy)
	if err != nil {
		return nil, fmt.Errorf(
			"%w: %v",
			ErrInvalidTree,
			err,
		)
	}
	return &Tree{
		body:    bodyCopy,
		entries: entries,
	}, nil
}

// NewTreeFromEntries validates entries and constructs a tree.
//
// Entries may be provided in any order. The resulting tree body uses canonical tree-entry ordering.
func NewTreeFromEntries(entries []TreeEntry) (*Tree, error) {
	entriesCopy := slices.Clone(entries)
	if err := validateTreeEntries(entriesCopy); err != nil {
		return nil, err
	}

	SortTreeEntries(entriesCopy)

	return &Tree{
		body:    serializeTreeEntries(entriesCopy),
		entries: entriesCopy,
	}, nil
}

// Body returns a defensive copy of the raw tree body.
func (t *Tree) Body() []byte {
	return bytes.Clone(t.body)
}

// Type returns ObjectTypeTree.
func (t *Tree) Type() ObjectType {
	return ObjectTypeTree
}

// Size returns the byte length of the serialized tree body.
func (t *Tree) Size() int {
	return len(t.body)
}

// Entries returns a copy of the entries in canonical order.
func (t *Tree) Entries() []TreeEntry {
	return slices.Clone(t.entries)
}

// Serialize returns the tree in "<type> <size>\x00<body>" format.
func (t *Tree) Serialize() []byte {
	return serializeObject(ObjectTypeTree, t.body)
}

func serializeTreeEntries(entries []TreeEntry) []byte {
	var buffer bytes.Buffer
	for _, entry := range entries {
		buffer.WriteString(entry.mode.String())
		buffer.WriteByte(' ')
		buffer.WriteString(entry.name)
		buffer.WriteByte(0)
		buffer.Write(entry.hash[:])
	}
	return buffer.Bytes()
}

func parseTreeEntries(body []byte) ([]TreeEntry, error) {
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

		offset += modeEnd + 1
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

		entry, err := NewTreeEntry(mode, hash, name)
		if err != nil {
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
		if hasPrevious && previousSortKey > sortKey {
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
		if !entry.mode.IsValid() {
			return fmt.Errorf(
				"%w: entry %d has invalid mode %#o",
				ErrInvalidTree,
				i,
				uint32(entry.mode),
			)
		}
		if err := validateTreeEntryName(entry.name); err != nil {
			return fmt.Errorf(
				"%w: entry %d: %v",
				ErrInvalidTree,
				i,
				err,
			)
		}
		if _, exists := seenNames[entry.name]; exists {
			return fmt.Errorf(
				"%w: duplicate entry name %q",
				ErrInvalidTree,
				entry.name,
			)
		}
		seenNames[entry.name] = struct{}{}
	}
	return nil
}

func validateTreeEntryName(name string) error {
	switch {
	case name == "":
		return fmt.Errorf(
			"%w: name is empty",
			ErrInvalidTree,
		)
	case name == "." || name == "..":
		return fmt.Errorf(
			"%w: traversal component %q is not allowed",
			ErrInvalidTree,
			name,
		)
	case strings.ContainsRune(name, '/'):
		return fmt.Errorf(
			"%w: %q contains slash",
			ErrInvalidTree,
			name,
		)
	case strings.ContainsRune(name, '\\'):
		return fmt.Errorf(
			"%w: %q contains backslash",
			ErrInvalidTree,
			name,
		)
	case strings.ContainsRune(name, 0):
		return fmt.Errorf(
			"%w: %q contains NUL",
			ErrInvalidTree,
			name,
		)
	default:
		return nil
	}
}

func SortTreeEntries(entries []TreeEntry) {
	slices.SortFunc(
		entries, func(a, b TreeEntry) int {
			return strings.Compare(
				treeEntrySortKey(a.name, a.mode.IsDirectory()),
				treeEntrySortKey(b.name, b.mode.IsDirectory()),
			)
		},
	)

}

func SortTreeEntriesByName(entries []TreeEntry) {
	slices.SortFunc(
		entries, func(a, b TreeEntry) int {
			return strings.Compare(a.name, b.name)
		},
	)
}

func treeEntrySortKey(name string, isDirectory bool) string {
	if isDirectory {
		return name + "/"
	}
	return name
}
