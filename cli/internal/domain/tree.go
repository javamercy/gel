package domain

import (
	"bytes"
	"errors"
	"fmt"
	"slices"
	"strings"
)

var (
	// ErrInvalidTreeEntryName is returned when a tree entry name is invalid
	// (empty, contains slash, null bytes, or path traversal components).
	ErrInvalidTreeEntryName = errors.New("invalid tree entry name")

	// ErrDuplicateTreeEntryName is returned when a tree contains duplicate entry names.
	ErrDuplicateTreeEntryName = errors.New("invalid tree format: duplicate entry name")

	// ErrTreeMissingModeSeparator is returned when a tree entry mode is not followed by a space.
	ErrTreeMissingModeSeparator = errors.New("invalid tree format: missing space after mode")

	// ErrTreeMissingNullByte is returned when a tree entry name is not null-terminated.
	ErrTreeMissingNullByte = errors.New("invalid tree format: missing null byte after name")

	// ErrTreeEntriesNotSorted is returned when tree entries are not in canonical order.
	ErrTreeEntriesNotSorted = errors.New("invalid tree format: entries not in canonical order")

	// ErrTreeTruncatedHash is returned when a tree entry hash is incomplete.
	ErrTreeTruncatedHash = errors.New("invalid tree format: truncated hash")
)

// TreeEntry represents a single entry within a tree object.
// Each entry corresponds to a file or subdirectory.
type TreeEntry struct {
	// Mode is the file mode.
	Mode FileMode

	// Hash is the SHA-256 content hash of the referenced object.
	Hash Hash

	// Name is the entry's filename, not a path.
	Name string
}

// NewTreeEntry returns a TreeEntry without validating name or mode.
// Tree entry validation is performed by NewTree and NewTreeFromEntries.
func NewTreeEntry(mode FileMode, hash Hash, name string) TreeEntry {
	return TreeEntry{
		Mode: mode,
		Hash: hash,
		Name: name,
	}
}

// Tree represents a directory structure in the object database.
type Tree struct {
	body    []byte
	entries []TreeEntry
}

// NewTree parses raw tree body bytes and returns a validated Tree.
// The input is copied to prevent external mutation.
func NewTree(body []byte) (*Tree, error) {
	bodyCopy := append([]byte(nil), body...)
	entries, err := parseTreeEntries(bodyCopy)
	if err != nil {
		return nil, err
	}

	return &Tree{
		body: bodyCopy, entries: entries,
	}, nil
}

// NewTreeFromEntries validates entries and returns a Tree built from them.
// Entries are serialized in canonical tree order.
func NewTreeFromEntries(entries []TreeEntry) (*Tree, error) {
	entriesCopy := append([]TreeEntry(nil), entries...)
	if err := validateTreeEntries(entriesCopy); err != nil {
		return nil, err
	}

	SortTreeEntries(entriesCopy)

	return &Tree{
		body:    serializeTreeEntries(entriesCopy),
		entries: append([]TreeEntry(nil), entriesCopy...),
	}, nil
}

// Body returns a defensive copy of the raw tree body bytes.
func (t *Tree) Body() []byte {
	return append([]byte(nil), t.body...)
}

// Type returns the domain object type for Tree.
func (t *Tree) Type() ObjectType {
	return ObjectTypeTree
}

// Size returns the byte length of the raw tree body.
func (t *Tree) Size() int {
	return len(t.body)
}

func (t *Tree) Entries() []TreeEntry {
	return append([]TreeEntry(nil), t.entries...)
}

// Serialize returns the full object serialization in the form "<type> <size>\x00<body>".
func (t *Tree) Serialize() []byte {
	return serializeObject(ObjectTypeTree, t.body)
}

func serializeTreeEntries(entries []TreeEntry) []byte {
	var buffer bytes.Buffer
	for _, entry := range entries {
		buffer.WriteString(entry.Mode.String())
		buffer.WriteByte(' ')
		buffer.WriteString(entry.Name)
		buffer.WriteByte(0)
		buffer.Write(entry.Hash[:])
	}
	return buffer.Bytes()
}

func parseTreeEntries(body []byte) ([]TreeEntry, error) {
	var entries []TreeEntry
	seenNames := make(map[string]struct{})
	previousSortName := ""
	hasPrevious := false
	offset := 0
	for offset < len(body) {
		spaceOffset := bytes.IndexByte(body[offset:], ' ')
		if spaceOffset == -1 {
			return nil, fmt.Errorf("%w at offset %d", ErrTreeMissingModeSeparator, offset)
		}

		modeText := string(body[offset : offset+spaceOffset])
		fileMode, err := ParseFileMode(modeText)
		if err != nil {
			return nil, err
		}

		offset += spaceOffset + 1
		nameOffset := offset
		nullOffset := bytes.IndexByte(body[offset:], 0)
		if nullOffset == -1 {
			return nil, fmt.Errorf("%w at offset %d", ErrTreeMissingNullByte, nameOffset)
		}

		name := string(body[offset : offset+nullOffset])
		if err := validateTreeEntryName(name); err != nil {
			return nil, fmt.Errorf("%w: %q", err, name)
		}
		if err := validateUniqueTreeEntryName(seenNames, name); err != nil {
			return nil, err
		}

		sortName := treeEntrySortKey(name, fileMode.IsDirectory())
		if hasPrevious && previousSortName > sortName {
			return nil, fmt.Errorf("%w: %q before %q", ErrTreeEntriesNotSorted, previousSortName, sortName)
		}

		previousSortName = sortName
		hasPrevious = true
		offset += nullOffset + 1
		if offset+SHA256ByteLength > len(body) {
			return nil, fmt.Errorf(
				"%w: expected %d bytes at offset %d, got %d",
				ErrTreeTruncatedHash,
				SHA256ByteLength,
				offset,
				len(body)-offset,
			)
		}

		hash, err := NewHashFromBytes(body[offset : offset+SHA256ByteLength])
		if err != nil {
			return nil, err
		}
		offset += SHA256ByteLength
		entries = append(entries, NewTreeEntry(fileMode, hash, name))
	}
	return entries, nil
}

func validateTreeEntries(entries []TreeEntry) error {
	seenNames := make(map[string]struct{}, len(entries))
	for _, entry := range entries {
		if !entry.Mode.IsValid() {
			return fmt.Errorf("%w: %s", ErrInvalidFileMode, entry.Mode)
		}
		if err := validateTreeEntryName(entry.Name); err != nil {
			return fmt.Errorf("%w: %q", err, entry.Name)
		}
		if err := validateUniqueTreeEntryName(seenNames, entry.Name); err != nil {
			return err
		}
	}
	return nil
}

func validateTreeEntryName(name string) error {
	if name == "" || name == "." || name == ".." {
		return ErrInvalidTreeEntryName
	}
	if strings.Contains(name, "/") {
		return ErrInvalidTreeEntryName
	}
	if strings.Contains(name, "\x00") {
		return ErrInvalidTreeEntryName
	}
	return nil
}

func validateUniqueTreeEntryName(seenNames map[string]struct{}, name string) error {
	if _, exists := seenNames[name]; exists {
		return fmt.Errorf("%w: %q", ErrDuplicateTreeEntryName, name)
	}
	seenNames[name] = struct{}{}
	return nil
}

func SortTreeEntries(entries []TreeEntry) {
	slices.SortFunc(
		entries, func(a, b TreeEntry) int {
			return strings.Compare(
				treeEntrySortKey(a.Name, a.Mode.IsDirectory()),
				treeEntrySortKey(b.Name, b.Mode.IsDirectory()),
			)
		},
	)
}
func treeEntrySortKey(name string, isDirectory bool) string {
	if isDirectory {
		return name + "/"
	}
	return name
}
